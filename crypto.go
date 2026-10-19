package main

import (
	"crypto/md5"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

// authParam is provided by Bifrost/SamloaderKotlin under the MIT License.
// See THIRD_PARTY_NOTICES.txt.
//
//go:embed auth_param.dat
var authParam []byte

func authenticate(nonce string) (string, error) {
	if len(authParam) < 56 {
		return "", errors.New("embedded auth_param.dat is invalid")
	}
	header := make([]uint32, 14)
	for i := range header {
		header[i] = binary.LittleEndian.Uint32(authParam[i*4 : i*4+4])
	}
	table := authParam[56:]
	shift := [16]int{0, 5, 10, 15, 4, 9, 14, 3, 8, 13, 2, 7, 12, 1, 6, 11}
	temp := make([]byte, 320)
	copy(temp[:16], []byte(nonce))
	state := make([]byte, 64)
	final := int(header[3])

	for round := 0; round < 9; round++ {
		start := round * 32
		for idx := 0; idx < 16; idx++ {
			temp[start+16+idx] = temp[start+shift[idx]]
		}
		for group := 0; group < 4; group++ {
			group4, group16 := group*4, group*16
			base := final + int(header[5]) + int(header[11]) + 6144*(group+round*4)
			for k := 0; k < 4; k++ {
				block := round*16 + group4 + k
				index := int(temp[start+16+group4+k])
				sourceAt := block*4096 + index*16
				selectAt := final + int(header[5]) + block*32
				if sourceAt+16 > len(table) || selectAt+32 > len(table) {
					return "", errors.New("auth parameter index is out of range")
				}
				source := table[sourceAt : sourceAt+16]
				selector := table[selectAt : selectAt+32]
				for output := 0; output < 4; output++ {
					var accumulator byte
					for bit := 0; bit < 8; bit++ {
						s := selector[output*8+bit]
						sourceIndex := int((s >> 3) & 31)
						var sourceByte byte
						if sourceIndex < 16 {
							sourceByte = source[sourceIndex]
						}
						accumulator |= ((sourceByte >> (7 - (s & 7))) & 1) << (7 - byte(bit))
					}
					state[group16+k*4+output] = accumulator
				}
			}

			for k := 0; k < 4; k++ {
				a, b := state[group16+k], state[group16+k+4]
				c, d := state[group16+k+8], state[group16+k+12]
				if base+1536*(k+1) > len(table) {
					return "", errors.New("auth parameter table is out of range")
				}
				lookup := table[base+1536*k : base+1536*(k+1)]
				left := (16*lookup[int(a&0xf0)|int(b>>4)] ^ lookup[256+int((a&0x0f)<<4|(b&0x0f))]) & 0xff
				right := (16*lookup[512+(int(c&0xf0)|int(d>>4))] ^ lookup[768+int((c&0x0f)<<4|(d&0x0f))]) & 0xff
				temp[(round+1)*32+group4+k] = (16*lookup[1024+int((left&0xf0)|(right>>4))] ^ lookup[1280+int((left&0x0f)<<4|(right&0x0f))]) & 0xff
			}
		}
	}

	result := make([]byte, 16)
	for idx := 0; idx < 16; idx++ {
		at := final + idx*256 + int(temp[288+shift[idx]])
		if at >= len(table) {
			return "", errors.New("auth result index is out of range")
		}
		result[idx] = table[at]
	}
	return hex.EncodeToString(result), nil
}

func logicCheck(source, nonce string) (string, error) {
	if len(source) < 16 {
		return "", errors.New("logic source must contain at least 16 characters")
	}
	result := make([]byte, 0, len(nonce))
	for _, char := range []byte(nonce) {
		result = append(result, source[int(char)&15])
	}
	return string(result), nil
}

func decryptionKey(version, logic string) ([]byte, error) {
	selected, err := logicCheck(version, logic)
	if err != nil {
		return nil, fmt.Errorf("derive decryption key: %w", err)
	}
	sum := md5.Sum([]byte(selected))
	return sum[:], nil
}

func removePKCS7(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("invalid decrypted data length")
	}
	padding := int(data[len(data)-1])
	if padding < 1 || padding > blockSize || padding > len(data) {
		return nil, errors.New("invalid decryption padding")
	}
	for _, value := range data[len(data)-padding:] {
		if int(value) != padding {
			return nil, errors.New("invalid decryption padding")
		}
	}
	return data[:len(data)-padding], nil
}
