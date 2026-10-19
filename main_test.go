package main

import (
	"bytes"
	"testing"
)

func TestDeriveSJPVersion(t *testing.T) {
	got, err := deriveSJPVersion("SM-S948Q", "SJP", "BZI9")
	if err != nil {
		t.Fatal(err)
	}
	want := "S948QOPU1BZI9/S948QSJP1BZI9/S948QOPU1BZI9/S948QOPU1BZI9"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestLogicCheck(t *testing.T) {
	got, err := logicCheck("0123456789abcdef", "A0f")
	if err != nil {
		t.Fatal(err)
	}
	if got != "106" {
		t.Fatalf("got %q", got)
	}
}

func TestRemovePKCS7(t *testing.T) {
	input := append(bytes.Repeat([]byte{'x'}, 14), 2, 2)
	got, err := removePKCS7(input, 16)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(bytes.Repeat([]byte{'x'}, 14)) {
		t.Fatal("unexpected unpadded value")
	}
}

func TestParseFUSResponse(t *testing.T) {
	x := []byte(`<FUSMsg><FUSBody><Put><BINARY_NAME><Data>a.enc4</Data></BINARY_NAME></Put><Results><Status><Data>S00</Data></Status></Results></FUSBody></FUSMsg>`)
	fields, results, err := parseFUSResponse(x)
	if err != nil {
		t.Fatal(err)
	}
	if fields["BINARY_NAME"] != "a.enc4" || results["Status"] != "S00" {
		t.Fatalf("unexpected parse: %#v %#v", fields, results)
	}
}

func TestCommonFlagsRejectsUnsafeModel(t *testing.T) {
	_, err := commonFlags("test", []string{"--model", "SM-../../escape"})
	if err == nil {
		t.Fatal("expected unsafe model to be rejected")
	}
}

func TestCommonFlagsAcceptsExplicitVersion(t *testing.T) {
	version := "A123/B123/C123/D123"
	opts, err := commonFlags("test", []string{"--model", "SM-TEST", "--csc", "ABC", "--version", version})
	if err != nil {
		t.Fatal(err)
	}
	if opts.version != version {
		t.Fatalf("got %q, want %q", opts.version, version)
	}
}

func TestCommonFlagsAcceptsJapaneseCarrierModel(t *testing.T) {
	opts, err := commonFlags("test", []string{"--model", "SC-51Aa", "--csc", "DCM", "--latest"})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.latest || opts.model != "SC51Aa" {
		t.Fatalf("unexpected options: %#v", opts)
	}
}

func TestCommonFlagsPreservesFUSCarrierModelCase(t *testing.T) {
	opts, err := commonFlags("test", []string{"--model", "SC51Aa", "--csc", "DCM", "--latest"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.model != "SC51Aa" {
		t.Fatalf("model = %q, want SC51Aa", opts.model)
	}
}

func TestCommonFlagsKeepsStandardCarrierHyphen(t *testing.T) {
	opts, err := commonFlags("test", []string{"--model", "SC-51A", "--csc", "DCM", "--latest"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.model != "SC-51A" {
		t.Fatalf("model = %q, want SC-51A", opts.model)
	}
}

func TestValidateThreePartVersion(t *testing.T) {
	if err := validateVersion("SC51AaOMS1DWL1/SC51AaDCL1DWL1/SC51AaOMS1DWL1"); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeLatestVersionPreservesCarrierModelCase(t *testing.T) {
	input := "SC51AaOMS1DWL1/SC51AaDCL1DWL1/SC51AaOMS1DWL1"
	want := input + "/SC51AaOMS1DWL1"
	if got := normalizeLatestVersion(input); got != want {
		t.Fatalf("normalizeLatestVersion() = %q, want %q", got, want)
	}
}

func TestRemovePKCS7RejectsInvalidPadding(t *testing.T) {
	input := append(bytes.Repeat([]byte{'x'}, 14), 2, 3)
	if _, err := removePKCS7(input, 16); err == nil {
		t.Fatal("expected invalid padding to be rejected")
	}
}
