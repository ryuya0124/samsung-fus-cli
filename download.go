package main

import (
	"context"
	"crypto/aes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (c *fusClient) download(ctx context.Context, o options, m metadata) (downloadReport, error) {
	name := m.Fields["BINARY_NAME"]
	modelPath := m.Fields["MODEL_PATH"]
	deviceType := m.Fields["DEVICE_MODEL_TYPE"]
	if name == "" || modelPath == "" || (!m.Legacy && deviceType == "") {
		return downloadReport{}, errors.New("FUS response is missing download fields")
	}
	if filepath.Base(name) != name || strings.ContainsAny(name, `/\\`) {
		return downloadReport{}, errors.New("FUS returned an unsafe binary name")
	}
	expectedSize, err := strconv.ParseInt(m.Fields["BINARY_BYTE_SIZE"], 10, 64)
	if err != nil || expectedSize <= 0 {
		return downloadReport{}, errors.New("FUS returned an invalid file size")
	}
	expectedCRC, err := strconv.ParseUint(m.Fields["BINARY_CRC"], 10, 32)
	if err != nil {
		return downloadReport{}, errors.New("FUS returned an invalid CRC32")
	}
	if err := c.initializeDownload(ctx, o, m); err != nil {
		return downloadReport{}, fmt.Errorf("initialize download: %w", err)
	}

	ap := strings.Split(o.version, "/")[0]
	destination := filepath.Join(o.outDir, fmt.Sprintf("%s-%s-%s", o.model, o.csc, ap))
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return downloadReport{}, err
	}
	encryptedPath := filepath.Join(destination, name)
	if err := c.downloadFile(ctx, modelPath+name, encryptedPath, expectedSize, m.Legacy); err != nil {
		return downloadReport{}, err
	}

	report, err := verifyAndDecrypt(encryptedPath, destination, o, m, uint32(expectedCRC))
	if err != nil {
		return report, err
	}
	reportData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return report, err
	}
	reportPath := filepath.Join(destination, "download-report.json")
	if err := os.WriteFile(reportPath, append(reportData, '\n'), 0o644); err != nil {
		return report, err
	}
	return report, nil
}

func (c *fusClient) downloadFile(ctx context.Context, remotePath, localPath string, expectedSize int64, legacy bool) error {
	var offset int64
	if stat, err := os.Stat(localPath); err == nil {
		offset = stat.Size()
		if offset > expectedSize {
			return fmt.Errorf("existing file is larger than expected: %s", localPath)
		}
		if offset == expectedSize {
			fmt.Println("Encrypted package already downloaded; verifying.")
			return nil
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	baseURL := downloadBaseURL
	agent := userAgent
	if legacy {
		baseURL = legacyDownloadBaseURL
		agent = "Kiss2.0_FUS"
	}
	downloadURL, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	query := downloadURL.Query()
	query.Set("file", remotePath)
	downloadURL.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.authorization())
	req.Header.Set("User-Agent", agent)
	req.Header.Set("Cache-Control", "no-cache")
	if offset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
		fmt.Printf("Resuming at %.2f GB\n", float64(offset)/1e9)
	}

	downloadClient := *c.http
	downloadClient.Timeout = 0
	resp, err := downloadClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if offset > 0 && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("server did not accept resume (HTTP %d); existing file was preserved", resp.StatusCode)
	}
	if offset == 0 && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	file, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, 4*1024*1024)
	done := offset
	started := time.Now()
	lastUpdate := started
	for {
		count, readErr := resp.Body.Read(buffer)
		if count > 0 {
			if _, err := file.Write(buffer[:count]); err != nil {
				return err
			}
			done += int64(count)
			if time.Since(lastUpdate) >= 2*time.Second {
				speed := float64(done-offset) / time.Since(started).Seconds() / 1e6
				fmt.Printf("\r%6.2f%%  %.2f / %.2f GB  %.1f MB/s", float64(done)*100/float64(expectedSize), float64(done)/1e9, float64(expectedSize)/1e9, speed)
				lastUpdate = time.Now()
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return readErr
		}
	}
	fmt.Println()
	if err := file.Sync(); err != nil {
		return err
	}
	if done != expectedSize {
		return fmt.Errorf("download length mismatch: got %d, expected %d", done, expectedSize)
	}
	return nil
}

func verifyAndDecrypt(encryptedPath, destination string, o options, m metadata, expectedCRC uint32) (downloadReport, error) {
	stat, err := os.Stat(encryptedPath)
	if err != nil {
		return downloadReport{}, err
	}
	report := downloadReport{
		Model: o.model, CSC: o.csc, Version: o.version,
		EncryptedFile: abs(encryptedPath), Bytes: stat.Size(), ExpectedCRC32: expectedCRC,
	}

	source, err := os.Open(encryptedPath)
	if err != nil {
		return report, err
	}
	defer source.Close()

	sha := sha256.New()
	crc := crc32.NewIEEE()
	var zipPath, temporaryPath string
	var destinationFile *os.File
	if !o.noDecrypt {
		zipPath = filepath.Join(destination, strings.TrimSuffix(filepath.Base(encryptedPath), ".enc4"))
		temporaryPath = zipPath + ".part"
		if _, err := os.Stat(zipPath); err == nil && !o.force {
			return report, fmt.Errorf("ZIP already exists: %s (use --force to replace)", zipPath)
		}
		_ = os.Remove(temporaryPath)
		destinationFile, err = os.OpenFile(temporaryPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return report, err
		}
		defer destinationFile.Close()
	}

	var block cipherBlock
	if !o.noDecrypt {
		logic := m.Fields["LOGIC_VALUE_FACTORY"]
		if logic == "" {
			logic = m.Fields["LOGIC_VALUE_HOME"]
		}
		keyVersion := m.Results["LATEST_FW_VERSION"]
		if keyVersion == "" {
			keyVersion = m.Results["BINARY_SW_VERSION"]
		}
		if keyVersion == "" {
			keyVersion = o.version
		}
		key, err := decryptionKey(keyVersion, logic)
		if err != nil {
			return report, err
		}
		block, err = aes.NewCipher(key)
		if err != nil {
			return report, err
		}
	}

	buffer := make([]byte, 8*1024*1024)
	var processed int64
	for {
		count, readErr := io.ReadFull(source, buffer)
		if readErr != nil && !errors.Is(readErr, io.EOF) && !errors.Is(readErr, io.ErrUnexpectedEOF) {
			return report, readErr
		}
		if count == 0 {
			break
		}
		chunk := buffer[:count]
		_, _ = sha.Write(chunk)
		_, _ = crc.Write(chunk)
		processed += int64(count)

		if !o.noDecrypt {
			if count%aes.BlockSize != 0 {
				return report, errors.New("encrypted package length is not aligned to AES block size")
			}
			plain := make([]byte, count)
			for offset := 0; offset < count; offset += aes.BlockSize {
				block.Decrypt(plain[offset:offset+aes.BlockSize], chunk[offset:offset+aes.BlockSize])
			}
			if processed == stat.Size() {
				plain, err = removePKCS7(plain, aes.BlockSize)
				if err != nil {
					return report, err
				}
			}
			if _, err := destinationFile.Write(plain); err != nil {
				return report, err
			}
		}

		if errors.Is(readErr, io.EOF) || errors.Is(readErr, io.ErrUnexpectedEOF) {
			break
		}
	}

	report.EncryptedSHA256 = hex.EncodeToString(sha.Sum(nil))
	report.EncryptedCRC32 = crc.Sum32()
	report.IntegrityVerified = report.EncryptedCRC32 == expectedCRC
	if !report.IntegrityVerified {
		if temporaryPath != "" {
			_ = os.Remove(temporaryPath)
		}
		return report, fmt.Errorf("CRC32 mismatch: got %d, expected %d", report.EncryptedCRC32, expectedCRC)
	}
	if destinationFile != nil {
		if err := destinationFile.Sync(); err != nil {
			return report, err
		}
		if err := destinationFile.Close(); err != nil {
			return report, err
		}
		if o.force {
			_ = os.Remove(zipPath)
		}
		if err := os.Rename(temporaryPath, zipPath); err != nil {
			return report, err
		}
		report.ZIPFile = abs(zipPath)
	}
	if !o.keepEnc4 && !o.noDecrypt {
		if err := os.Remove(encryptedPath); err != nil {
			return report, err
		}
	}
	return report, nil
}

type cipherBlock interface {
	Decrypt(destination, source []byte)
}
