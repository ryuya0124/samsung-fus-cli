package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type fusClient struct {
	http  *http.Client
	nonce string
	auth  string
}

type fieldNode struct {
	XMLName xml.Name
	Data    string `xml:"Data"`
	Text    string `xml:",chardata"`
}

type fusResponse struct {
	Body struct {
		Put struct {
			Nodes []fieldNode `xml:",any"`
		} `xml:"Put"`
		Results struct {
			Nodes []fieldNode `xml:",any"`
		} `xml:"Results"`
	} `xml:"FUSBody"`
}

type metadata struct {
	Model            string            `json:"model"`
	CSC              string            `json:"csc"`
	RequestedVersion string            `json:"requested_version"`
	Fields           map[string]string `json:"fields"`
	Results          map[string]string `json:"results"`
}

type downloadReport struct {
	Model             string `json:"model"`
	CSC               string `json:"csc"`
	Version           string `json:"version"`
	EncryptedFile     string `json:"encrypted_file"`
	ZIPFile           string `json:"zip_file,omitempty"`
	Bytes             int64  `json:"bytes"`
	EncryptedSHA256   string `json:"encrypted_sha256"`
	EncryptedCRC32    uint32 `json:"encrypted_crc32"`
	ExpectedCRC32     uint32 `json:"expected_crc32"`
	IntegrityVerified bool   `json:"integrity_verified"`
}

type dataValue struct {
	Data string `xml:"Data"`
}

type fusHeader struct {
	ProtoVer  string `xml:"ProtoVer"`
	SessionID string `xml:"SessionID"`
	MsgID     string `xml:"MsgID"`
}

type informPut struct {
	XMLName         xml.Name  `xml:"Put"`
	CmdID           string    `xml:"CmdID"`
	RequestType     dataValue `xml:"REQUEST_TYPE"`
	BinaryVersion   dataValue `xml:"BINARY_SW_VERSION"`
	DeviceSerial    dataValue `xml:"DEVICE_SN_NUMBER"`
	BinaryLocalCode dataValue `xml:"BINARY_LOCAL_CODE"`
	BinaryModelName dataValue `xml:"BINARY_MODEL_NAME"`
	AccessMode      dataValue `xml:"ACCESS_MODE"`
	BinaryNature    dataValue `xml:"BINARY_NATURE"`
	LogicCheck      dataValue `xml:"LOGIC_CHECK"`
	ClientLanguage  dataValue `xml:"CLIENT_LANGUAGE"`
}

type informMessage struct {
	XMLName xml.Name  `xml:"FUSMsg"`
	Header  fusHeader `xml:"FUSHdr"`
	Body    struct {
		Put informPut `xml:"Put"`
		Get struct {
			CmdID         string `xml:"CmdID"`
			BinaryVersion string `xml:"BINARY_SW_VERSION"`
		} `xml:"Get"`
	} `xml:"FUSBody"`
}

type initMessage struct {
	XMLName xml.Name  `xml:"FUSMsg"`
	Header  fusHeader `xml:"FUSHdr"`
	Body    struct {
		Put struct {
			XMLName       xml.Name  `xml:"Put"`
			BinaryName    dataValue `xml:"BINARY_NAME"`
			BinaryVersion dataValue `xml:"BINARY_SW_VERSION"`
			DeviceLocal   dataValue `xml:"DEVICE_LOCAL_CODE"`
			DeviceModel   dataValue `xml:"DEVICE_MODEL_TYPE"`
			LogicCheck    dataValue `xml:"LOGIC_CHECK"`
		} `xml:"Put"`
	} `xml:"FUSBody"`
}

func (c *fusClient) query(ctx context.Context, o options) (metadata, error) {
	m := metadata{Model: o.model, CSC: o.csc, RequestedVersion: o.version}
	if _, err := c.request(ctx, "NF_SmartDownloadGenerateNonce.do", nil, true); err != nil {
		return m, fmt.Errorf("generate FUS nonce: %w", err)
	}
	logic, err := logicCheck(o.version, c.nonce)
	if err != nil {
		return m, err
	}
	msg := informMessage{Header: fusHeader{ProtoVer: "1", SessionID: "0", MsgID: "1"}}
	msg.Body.Put = informPut{
		CmdID: "1", RequestType: dv("2"), BinaryVersion: dv(o.version),
		DeviceSerial: dv(""), BinaryLocalCode: dv(o.csc), BinaryModelName: dv(o.model),
		AccessMode: dv("1"), BinaryNature: dv("1"), LogicCheck: dv(logic), ClientLanguage: dv("1033"),
	}
	msg.Body.Get.CmdID = "2"
	payload, err := xml.Marshal(msg)
	if err != nil {
		return m, err
	}
	body, err := c.request(ctx, "NF_SmartDownloadBinaryInform.do", append([]byte(xml.Header), payload...), false)
	if err != nil {
		return m, fmt.Errorf("query firmware metadata: %w", err)
	}
	m.Fields, m.Results, err = parseFUSResponse(body)
	return m, err
}

func (c *fusClient) initializeDownload(ctx context.Context, o options, m metadata) error {
	name := m.Fields["BINARY_NAME"]
	if len(name) < 25 {
		return errors.New("FUS returned an invalid binary name")
	}
	special := name[len(name)-25 : len(name)-9]
	check, err := logicCheck(special, c.nonce)
	if err != nil {
		return err
	}
	msg := initMessage{Header: fusHeader{ProtoVer: "1", SessionID: "0", MsgID: "1"}}
	msg.Body.Put.BinaryName = dv(name)
	msg.Body.Put.BinaryVersion = dv(o.version)
	msg.Body.Put.DeviceLocal = dv(o.csc)
	msg.Body.Put.DeviceModel = dv(m.Fields["DEVICE_MODEL_TYPE"])
	msg.Body.Put.LogicCheck = dv(check)
	payload, err := xml.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = c.request(ctx, "NF_SmartDownloadBinaryInitForMass.do", append([]byte(xml.Header), payload...), false)
	return err
}

func (c *fusClient) request(ctx context.Context, endpoint string, payload []byte, nonceRequest bool) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, metadataBaseURL+endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.authorization())
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if value := resp.Header.Get("NONCE"); value != "" {
		if len(value) > 16 {
			value = value[:16]
		}
		c.nonce = value
		c.auth, err = authenticate(value)
		if err != nil {
			return nil, err
		}
	}
	if nonceRequest && c.nonce == "" {
		return nil, errors.New("FUS did not return a nonce")
	}
	return body, nil
}

func (c *fusClient) authorization() string {
	return fmt.Sprintf(`FUS nonce="%s", signature="%s", nc="", type="", realm=""`, c.nonce, c.auth)
}

func parseFUSResponse(data []byte) (map[string]string, map[string]string, error) {
	var parsed fusResponse
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, nil, fmt.Errorf("parse FUS XML: %w", err)
	}
	fields := nodesToMap(parsed.Body.Put.Nodes)
	results := nodesToMap(parsed.Body.Results.Nodes)
	if len(fields) == 0 && len(results) == 0 {
		return nil, nil, errors.New("FUS response contained no fields")
	}
	return fields, results, nil
}

func nodesToMap(nodes []fieldNode) map[string]string {
	out := make(map[string]string, len(nodes))
	for _, n := range nodes {
		value := n.Data
		if value == "" {
			value = strings.TrimSpace(n.Text)
		}
		out[n.XMLName.Local] = value
	}
	return out
}

func dv(value string) dataValue { return dataValue{Data: value} }
