# Samsung FUS CLI

[日本語](README.md) | [English](README.en.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md)

Samsung FUS CLI is an unofficial, independent command-line Samsung firmware downloader for Windows and macOS. It is not an official Samsung product and is not affiliated with or endorsed by Samsung. It downloads official Galaxy firmware from Samsung FUS, decrypts the `.enc4` package, resumes interrupted downloads, and verifies the size, CRC32, and SHA-256. It is a single Go binary and does not require Python at runtime.

The current Samsung Smart Download API does not use an IMEI, serial number, phone number, Samsung account, or ADB connection. For older devices whose legacy FUS endpoint requires an identifier, `--device-id-file` can read an IMEI or serial locally. The value is sent only to Samsung and is never printed, written to JSON, or saved in settings. The application only downloads files to the computer. It does not flash a device, unlock a bootloader, or provide root access.

## Use at your own risk

You are responsible for checking the model, CSC, build, destination, and available disk space. The authors and copyright holders are not liable for data loss, network usage, device damage, or other loss caused by use of this software. Downloaded firmware remains the property of Samsung and other applicable rights holders; follow their terms when using or redistributing it.

## Commands

Check availability:

```text
samsung-fus-cli check --model SM-S948Q --csc SJP --build BZI9
```

Download, decrypt, and verify:

```text
samsung-fus-cli download --model SM-S948Q --csc SJP --build BZI9
```

Resume by running the same command again. To keep only the encrypted package, add `--no-decrypt`. Automatic expansion of `--build` currently targets Japanese SJP Galaxy S26 firmware. For other models or regions, provide the full FUS value with `--version AP/CSC/CP/DATA`.

Discover the current version through Samsung FOTA:

```text
samsung-fus-cli check --model SC-51Aa --csc DCM --latest
```

If legacy FUS returns status `408`, put your own device IMEI or serial in a private local file and pass `--device-id-file PATH`. Do not commit or attach that file, and do not use another device's identifier or a generated IMEI.

See the [Japanese README](README.md) for build instructions, output layout, licensing, and implementation details.

**Related search terms:** Samsung firmware downloader, Samsung FUS downloader, Galaxy firmware download tool, Samsung firmware CLI, Windows Samsung firmware downloader, macOS Samsung firmware downloader.
