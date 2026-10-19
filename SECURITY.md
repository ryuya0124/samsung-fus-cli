# Security policy

## Sensitive information

Samsung FUS CLI does not need an IMEI or serial for the current Smart Download API. Some legacy FUS devices require one through `--device-id-file`; the value is never logged or saved by the application. Never add identifiers to bug reports, screenshots, terminal transcripts, test fixtures, or sample commands.

Before sharing `download-report.json`, review local file paths because they can contain an operating-system account name. Firmware files are large copyrighted binaries and should not be attached to ordinary bug reports.

## Reporting a problem

If this project is hosted on a service that supports private security advisories, use that private channel for vulnerabilities involving credential exposure, arbitrary file writes, package verification bypass, or unsafe update behavior. Include the application version, operating system, command with private values removed, and the smallest reproducible error output.

## Release expectations

- Windows releases are portable ZIP files with a SHA-256 sidecar.
- macOS releases are Universal PKG files signed with Developer ID Application and Developer ID Installer certificates, submitted through Apple notarytool, stapled, and checked with Gatekeeper.
- Release packages include the project license, third-party MIT notice, disclaimer, and architecture document.
- A release must pass `go test ./...` and `go vet ./...` before packaging.
