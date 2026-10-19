# Changelog

## 0.2.0

- Add `--latest` to discover the current firmware version from Samsung FOTA.
- Accept Japanese carrier model names and map the ahamo retail name `SC-51Aa` to Samsung's internal `SC51Aa` identifier without losing its case.
- Fall back to legacy FUS when the current Smart Download API does not contain a package.
- Add `--device-id-file` for legacy FUS devices that require the owner's IMEI or serial. The value is never printed or written to reports.
- Document the difference between modern identifier-free downloads and legacy status `408`.

## 0.1.0

- Initial Windows portable release.
- Add Samsung FUS metadata checks, resumable downloads, ENC4 decryption, SHA-256 and CRC32 verification.
- Add Windows portable packaging and the Fastlane macOS signing, notarization, PKG, and uninstaller workflow.
