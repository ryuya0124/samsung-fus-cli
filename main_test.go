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

func TestRemovePKCS7RejectsInvalidPadding(t *testing.T) {
	input := append(bytes.Repeat([]byte{'x'}, 14), 2, 3)
	if _, err := removePKCS7(input, 16); err == nil {
		t.Fatal("expected invalid padding to be rejected")
	}
}
