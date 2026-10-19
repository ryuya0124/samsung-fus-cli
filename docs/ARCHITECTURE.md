# Architecture

The program uses only the Go standard library. Runtime behavior is split by responsibility:

- `main.go`: command parsing, validation, process signals, and command dispatch
- `fus.go`: FUS XML messages, session nonce handling, metadata queries, and download initialization
- `download.go`: resumable HTTP transfer, progress display, hashing, CRC32 verification, and AES-ECB package decryption
- `crypto.go`: FUS authentication transform, logic checks, key derivation, and PKCS#7 validation
- `util.go`: JSON output and path helpers
- `main_test.go`: deterministic tests for version expansion, XML parsing, logic selection, and padding checks

## Data flow

1. The CLI validates the model, CSC, build, and full version before network access.
2. The FUS client requests a short-lived nonce and derives the session signature locally.
3. The metadata request sends an empty `DEVICE_SN_NUMBER`; no IMEI or account data is collected.
4. The server returns the package name, path, size, CRC32, and decryption logic value.
5. The downloader resumes an existing `.enc4` only when the server accepts the byte range.
6. Verification streams the package through SHA-256 and CRC32 without loading it into memory.
7. Decryption writes to `.part`. The temporary file becomes the final ZIP only after CRC32 and PKCS#7 checks succeed.

## File safety

- Server-supplied local filenames must be plain base names without path separators.
- Existing partial downloads are preserved if resume is rejected.
- Existing completed ZIP files require the explicit `--force` flag before replacement.
- Downloaded firmware is never deleted by the macOS uninstaller.
- The Windows build is portable and does not access the registry, services, or AppData.

## Dependencies and licensing

The executable has no third-party Go module dependencies. `auth_param.dat` and the authentication approach come from Bifrost/SamloaderKotlin under the MIT License; its complete notice is included in `THIRD_PARTY_NOTICES.txt` and in release packages.
