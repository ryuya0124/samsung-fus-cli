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
	metadataBaseURL       = "https://neofussvr.sslcs.cdngc.net/"
	downloadBaseURL       = "https://cloud-neofussvr.samsungmobile.com/NF_SmartDownloadBinaryForMass.do"
	legacyDownloadBaseURL = "https://cloud-neofussvr.samsungmobile.com/NF_DownloadBinaryForMass.do"
	firmwareBaseURL       = "https://fota-cloud-dn.ospserver.net/firmware"
	userAgent             = "SMART 2.0"
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
	latest    bool
	deviceID  string
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
  --version A/B/C[/D]    Full FUS version (overrides --build derivation)
  --latest               Discover the latest version from Samsung FOTA
  --device-id-file FILE  Read an IMEI/serial for legacy FUS only; never logged or saved
  --out DIR              Download root (default: downloads)
  --no-decrypt           Keep only the .enc4 package
  --keep-enc4=true       Keep encrypted package after decryption
  --force                Replace an existing completed ZIP

Examples:
  samsung-fus-cli check --model SM-S948Q --csc SJP --build BZI9
  samsung-fus-cli download --model SM-S948Q --csc SJP --build BZI9
  samsung-fus-cli download --model SC-51Aa --csc DCM --latest
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
	fs.BoolVar(&o.latest, "latest", false, "discover latest version from Samsung FOTA")
	deviceIDFile := fs.String("device-id-file", "", "read an IMEI/serial for legacy FUS")
	fs.BoolVar(&o.noDecrypt, "no-decrypt", false, "do not decrypt the enc4 file")
	fs.BoolVar(&o.keepEnc4, "keep-enc4", true, "keep encrypted package after decryption")
	fs.BoolVar(&o.force, "force", false, "replace an existing completed ZIP")
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	o.model = canonicalModel(strings.TrimSpace(o.model))
	o.csc = strings.ToUpper(strings.TrimSpace(o.csc))
	o.build = strings.ToUpper(strings.TrimSpace(o.build))
	o.version = strings.TrimSpace(o.version)
	if *deviceIDFile != "" {
		stat, err := os.Stat(*deviceIDFile)
		if err != nil {
			return o, fmt.Errorf("read --device-id-file: %w", err)
		}
		if stat.Size() < 1 || stat.Size() > 4096 {
			return o, errors.New("--device-id-file must contain a short IMEI or serial value")
		}
		value, err := os.ReadFile(*deviceIDFile)
		if err != nil {
			return o, fmt.Errorf("read --device-id-file: %w", err)
		}
		o.deviceID = strings.TrimSpace(string(value))
		if !regexp.MustCompile(`^[A-Za-z0-9]{8,32}$`).MatchString(o.deviceID) {
			return o, errors.New("--device-id-file must contain 8-32 letters or digits")
		}
	}
	upperModel := strings.ToUpper(o.model)
	if !strings.HasPrefix(upperModel, "SM-") && !strings.HasPrefix(upperModel, "SC") {
		return o, errors.New("--model must be a Samsung SM- or Japanese carrier SC model")
	}
	validModel := regexp.MustCompile(`^[A-Za-z0-9-]+$`)
	validID := regexp.MustCompile(`^[A-Z0-9-]+$`)
	if !validModel.MatchString(o.model) || !validID.MatchString(o.csc) || !validID.MatchString(o.build) {
		return o, errors.New("model, CSC, and build may contain only letters, digits, and hyphens")
	}
	if len(o.csc) != 3 {
		return o, errors.New("--csc must be a three-character code")
	}
	if o.latest {
		if o.version != "" {
			return o, errors.New("--latest and --version cannot be used together")
		}
		return o, nil
	}
	if o.version == "" {
		var err error
		o.version, err = deriveSJPVersion(o.model, o.csc, o.build)
		if err != nil {
			return o, err
		}
	}
	if err := validateVersion(o.version); err != nil {
		return o, err
	}
	return o, nil
}

func canonicalModel(model string) string {
	if len(model) >= 3 && strings.EqualFold(model[:3], "SC-") {
		// ahamo variants such as SC-51Aa use SC51Aa in Samsung FOTA/FUS,
		// while ordinary Japanese carrier IDs such as SC-51A keep the hyphen.
		if regexp.MustCompile(`^SC-[0-9]{2}[A-Z]a$`).MatchString(model) {
			return "SC" + model[3:]
		}
		return "SC-" + strings.ToUpper(model[3:])
	}
	if len(model) >= 3 && strings.EqualFold(model[:3], "SM-") {
		return "SM-" + strings.ToUpper(model[3:])
	}
	return model
}

func validateVersion(value string) error {
	validVersion := regexp.MustCompile(`^[A-Za-z0-9]+(?:/[A-Za-z0-9]+){2,3}$`)
	if !validVersion.MatchString(value) {
		return errors.New("--version must contain three or four slash-separated alphanumeric components")
	}
	return nil
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
	if o.latest {
		o.version, err = c.latestVersion(ctx, o)
		if err != nil {
			return err
		}
		fmt.Println("Latest version:", o.version)
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
	if o.latest {
		o.version, err = c.latestVersion(ctx, o)
		if err != nil {
			return err
		}
		fmt.Println("Latest version:", o.version)
	}
	m, err := c.query(ctx, o)
	if err != nil {
		return err
	}
	if m.Results["Status"] != "S00" && m.Results["Status"] != "200" {
		_ = printJSON(m)
		if m.Results["Status"] == "408" && o.deviceID == "" {
			return errors.New("this legacy firmware requires an IMEI or serial; supply it with --device-id-file")
		}
		return fmt.Errorf("FUS did not accept this version (status %q)", m.Results["Status"])
	}
	report, err := c.download(ctx, o, m)
	if err != nil {
		return err
	}
	return printJSON(report)
}
