# Samsung FUS CLI

[![CI](https://github.com/ryuya0124/samsung-fus-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/ryuya0124/samsung-fus-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/ryuya0124/samsung-fus-cli)](https://github.com/ryuya0124/samsung-fus-cli/releases)
[![License](https://img.shields.io/badge/license-All%20rights%20reserved-lightgrey.svg)](LICENSE.txt)

[日本語](README.md) | [English](README.en.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md)

**検索用の機能名:** Samsung firmware downloader / Samsung FUS downloader / Galaxy firmware download / Samsungファームウェア ダウンロード / 삼성 펌웨어 다운로드 / 三星固件下载

Samsung公式FUS配信サーバーからファームウェアを取得する、Windows/macOS対応の非公式ターミナルアプリです。Samsungによる公式製品ではなく、Samsungとの提携・承認関係はありません。Go製の単一バイナリなので、実行時にPythonや追加ライブラリは必要ありません。

## 注意事項

このソフトウェアの利用は自己責任です。モデル、CSC、ビルド、保存先、空き容量を利用者自身で確認してください。作者・権利者は、利用によって生じたデータ消失、通信量、端末故障その他の損害について責任を負いません。

本アプリが行うのはPCへのダウンロード、復号、整合性確認までです。端末への書き込み、root化、ブートローダー操作は行いません。取得したSamsungファームウェアの著作権はSamsungなど各権利者に帰属します。取得物の再配布は各権利者の条件に従ってください。

IMEI、シリアル番号、電話番号、Samsungアカウント、ADB接続は使用しません。FUSにはモデル、CSC、ソフトウェア版とセッション用認証値だけを送ります。

## 対応環境

- Windows 10/11: x86-64
- macOS: Apple Silicon、Intel
- ビルド時のみGo 1.24以降

## 使い方

配信確認:

```text
samsung-fus-cli check --model SM-S948Q --csc SJP --build BZI9
```

ダウンロード、復号、整合性確認:

```text
samsung-fus-cli download --model SM-S948Q --csc SJP --build BZI9
```

暗号化された`.enc4`だけを保存:

```text
samsung-fus-cli download --model SM-S948Q --csc SJP --build BZI9 --no-decrypt
```

中断後は同じ`download`コマンドを実行すると、サーバーがRange要求を受け付ける場合に続きから再開します。既存ファイルが想定サイズを超えている場合や、サーバーが再開を拒否した場合は上書きせず停止します。

`--build`からのソフトウェア版自動生成は、現在SJP版Galaxy S26シリーズ向けです。それ以外ではFUSの完全な版文字列を指定します。

```text
samsung-fus-cli check --model SM-XXXXX --csc XXX --version AP/CSC/CP/DATA
```

## 保存内容

既定では次の構成で保存します。

```text
downloads/
└── SM-S948Q-SJP-S948QOPU1BZI9/
    ├── *.zip.enc4
    ├── *.zip
    └── download-report.json
```

`download-report.json`には、ファイルサイズ、SHA-256、実測CRC32、FUS通知CRC32、照合結果を記録します。復号中は一時ファイル`.part`へ書き、CRC32一致後に正式なZIP名へ変更します。

## ソースからビルド

Windows PowerShell:

```powershell
go test ./...
New-Item -ItemType Directory -Force dist | Out-Null
go build -trimpath -o dist\samsung-fus-cli-windows-amd64.exe .
```

配布用Windows版はレジストリやインストーラーを使わないポータブルZIPです。

```powershell
.\scripts\build-windows-portable.ps1 -Version 1.0.0
```

アンインストールは展開したフォルダーを削除するだけです。レジストリ、Windowsサービス、AppDataへの設定保存はありません。

ビルドスクリプトはZIPと同時にSHA-256ファイル（`.sha256`）も作成します。

macOS Apple Silicon:

```bash
go test ./...
mkdir -p dist
GOOS=darwin GOARCH=arm64 go build -trimpath -o dist/samsung-fus-cli-macos-arm64 .
chmod +x dist/samsung-fus-cli-macos-arm64
```

macOS Intel:

```bash
GOOS=darwin GOARCH=amd64 go build -trimpath -o dist/samsung-fus-cli-macos-amd64 .
chmod +x dist/samsung-fus-cli-macos-amd64
```

## macOS署名とApple公証

配布用macOS版はFastlaneでUniversal Binaryを作り、Developer ID Application署名、Hardened Runtime、Developer ID Installer署名付きPKG、Apple公証、staple、Gatekeeper検証まで実行します。

```bash
bundle install
bundle exec fastlane mac release version:1.0.0
```

証明書名と`notarytool`のKeychainプロファイルが必要です。認証情報の準備と環境変数は[fastlane/README.md](fastlane/README.md)を参照してください。

PKGには`/usr/local/bin/samsung-fus-cli-uninstall`も入り、コマンド実行で本体・文書・PKG receiptを削除します。ダウンロード済みファームウェアは削除しません。

公証済みPKGと同時にSHA-256ファイル（`.sha256`）も作成されます。

## ライセンス

本体コードは[LICENSE.txt](LICENSE.txt)を参照してください。FUS認証方式と同梱データはBifrost/SamloaderKotlinを参考にしており、MITライセンス表記を[THIRD_PARTY_NOTICES.txt](THIRD_PARTY_NOTICES.txt)へ収録しています。

コードの役割分担、データフロー、ファイル保護方針は[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)にまとめています。

機密情報を含めない不具合報告方法とリリース要件は[SECURITY.md](SECURITY.md)を参照してください。
