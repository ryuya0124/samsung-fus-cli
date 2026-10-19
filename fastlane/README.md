# macOS署名・公証

Fastlaneの`mac release`レーンは次を順番に実行します。

1. Goテスト
2. IntelとApple Silicon向けバイナリのビルド
3. `lipo`によるUniversal Binary作成
4. Developer ID Application署名とHardened Runtime有効化
5. Developer ID Installer署名付き`.pkg`作成
6. `notarytool`によるApple公証
7. 公証チケットのstaple
8. `codesign`、`pkgutil`、`stapler`、`spctl`による検証

## 初回準備

Macのキーチェーンへ次の証明書を登録します。

- `Developer ID Application`
- `Developer ID Installer`

Apple公証情報は平文ファイルへ保存せず、`notarytool`のKeychainプロファイルへ登録します。

```bash
xcrun notarytool store-credentials "samsung-fus-cli-notary" \
  --apple-id "APPLE_ID" \
  --team-id "TEAM_ID"
```

コマンド実行後、アプリ用パスワードを対話入力します。

## 実行

```bash
bundle install

export MAC_APPLICATION_IDENTITY="Developer ID Application: NAME (TEAMID)"
export MAC_INSTALLER_IDENTITY="Developer ID Installer: NAME (TEAMID)"
export APPLE_NOTARY_KEYCHAIN_PROFILE="samsung-fus-cli-notary"
export MAC_PACKAGE_IDENTIFIER="dev.samsungfus.cli"

bundle exec fastlane mac release version:1.0.0
```

完成品は`dist/samsung-fus-cli-1.0.0-macos-universal.pkg`です。証明書、Apple ID、アプリ用パスワードはリポジトリへ保存しません。

## アンインストール

PKGにはアンインストーラーも入ります。

```bash
samsung-fus-cli-uninstall
```

必要に応じて`sudo`へ切り替え、`/usr/local/bin/samsung-fus-cli`、同梱文書、PKG receipt、アンインストーラー自身を削除します。利用者が取得したファームウェアは削除しません。
