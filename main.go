package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	metadataBaseURL = "https://neofussvr.sslcs.cdngc.net/"
	downloadBaseURL = "https://cloud-neofussvr.samsungmobile.com/NF_SmartDownloadBinaryForMass.do"
	userAgent       = "SMART 2.0"
)

var version = "dev"

type options struct {
	model     string
	csc       string
	build     string
	version   string
	outDir    string
	keepEnc4  bool
	noDecrypt bool
	force     bool
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var err error
	switch os.Args[1] {
	case "check":
		err = runCheck(ctx, os.Args[2:])
	case "download":
		err = runDownload(ctx, os.Args[2:])
	case "version", "--version", "-version":
		fmt.Printf("samsung-fus-cli %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		return
	case "help", "--help", "-h":
		usage()
		return
	default:
		usage()
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Print(`Samsung FUS CLI

Official Samsung firmware downloader for Windows and macOS.
No IMEI, serial number, ADB, or Samsung account is used.
Use at your own risk. Verify the model, CSC, build, and free disk space.

Usage:
  samsung-fus-cli check [options]
  samsung-fus-cli download [options]

Options:
  --model SM-S948Q       Samsung model
  --csc SJP              Region/CSC
  --build BZI9           Short build code for Japanese S26 firmware
  --version A/B/C/D      Full FUS version (overrides --build derivation)
  --out DIR              Download root (default: downloads)
  --no-decrypt           Keep only the .enc4 package
  --keep-enc4=true       Keep encrypted package after decryption
  --force                Replace an existing completed ZIP

Examples:
  samsung-fus-cli check --model SM-S948Q --csc SJP --build BZI9
  samsung-fus-cli download --model SM-S948Q --csc SJP --build BZI9
`)
}

func commonFlags(name string, args []string) (options, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var o options
	fs.StringVar(&o.model, "model", "SM-S948Q", "Samsung model")
	fs.StringVar(&o.csc, "csc", "SJP", "region/CSC")
	fs.StringVar(&o.build, "build", "BZI9", "short build code")
	fs.StringVar(&o.version, "version", "", "full FUS version")
	fs.StringVar(&o.outDir, "out", "downloads", "download root")
	fs.BoolVar(&o.noDecrypt, "no-decrypt", false, "do not decrypt the enc4 file")
	fs.BoolVar(&o.keepEnc4, "keep-enc4", true, "keep encrypted package after decryption")
	fs.BoolVar(&o.force, "force", false, "replace an existing completed ZIP")
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	o.model = strings.ToUpper(strings.TrimSpace(o.model))
	o.csc = strings.ToUpper(strings.TrimSpace(o.csc))
	o.build = strings.ToUpper(strings.TrimSpace(o.build))
	o.version = strings.ToUpper(strings.TrimSpace(o.version))
	if !strings.HasPrefix(o.model, "SM-") {
		return o, errors.New("--model must start with SM-")
	}
	validID := regexp.MustCompile(`^[A-Z0-9-]+$`)
	if !validID.MatchString(o.model) || !validID.MatchString(o.csc) || !validID.MatchString(o.build) {
		return o, errors.New("model, CSC, and build may contain only A-Z, 0-9, and hyphen")
	}
	if len(o.csc) != 3 {
		return o, errors.New("--csc must be a three-character code")
	}
	if o.version == "" {
		var err error
		o.version, err = deriveSJPVersion(o.model, o.csc, o.build)
		if err != nil {
			return o, err
		}
	}
	validVersion := regexp.MustCompile(`^[A-Z0-9]+(?:/[A-Z0-9]+){3}$`)
	if !validVersion.MatchString(o.version) {
		return o, errors.New("--version must contain four slash-separated alphanumeric components")
	}
	return o, nil
}

func deriveSJPVersion(model, csc, build string) (string, error) {
	if csc != "SJP" {
		return "", errors.New("automatic --build expansion currently supports SJP only; pass --version for other CSCs")
	}
	if build == "" || strings.Contains(build, "/") {
		return "", errors.New("invalid --build")
	}
	stem := strings.TrimPrefix(model, "SM-")
	apPrefix := "OPU1"
	if build == "AZH9" {
		apPrefix = "OPS1"
	}
	ap := stem + apPrefix + build
	cscPart := stem + csc + "1" + build
	return strings.Join([]string{ap, cscPart, ap, ap}, "/"), nil
}

func newFUSClient() (*fusClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &fusClient{http: &http.Client{Jar: jar, Timeout: 30 * time.Second}}, nil
}

func runCheck(ctx context.Context, args []string) error {
	o, err := commonFlags("check", args)
	if err != nil {
		return err
	}
	c, err := newFUSClient()
	if err != nil {
		return err
	}
	m, err := c.query(ctx, o)
	if err != nil {
		return err
	}
	return printJSON(m)
}

func runDownload(ctx context.Context, args []string) error {
	o, err := commonFlags("download", args)
	if err != nil {
		return err
	}
	c, err := newFUSClient()
	if err != nil {
		return err
	}
	m, err := c.query(ctx, o)
	if err != nil {
		return err
	}
	if m.Results["Status"] != "S00" {
		_ = printJSON(m)
		return fmt.Errorf("FUS did not accept this version (status %q)", m.Results["Status"])
	}
	report, err := c.download(ctx, o, m)
	if err != nil {
		return err
	}
	return printJSON(report)
}
