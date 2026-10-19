# Samsung FUS CLI

[日本語](README.md) | [English](README.en.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md)

Samsung FUS CLI 是適用於 Windows 與 macOS 的非官方獨立命令列三星韌體下載工具。它不是 Samsung 官方產品，也未獲 Samsung 關聯或認可。它可以從 Samsung FUS 下載官方 Galaxy 韌體、解密 `.enc4` 檔案、接續中斷的下載，並驗證檔案大小、CRC32 與 SHA-256。程式以 Go 編寫，執行時不需要 Python。

目前 Samsung Smart Download API 不使用 IMEI 或序號。只有舊裝置的傳統 FUS 要求識別碼時，才可透過 `--device-id-file` 讀取您自己裝置的 IMEI 或序號。該值不會顯示，也不會寫入 JSON 或設定。本程式只將韌體下載到電腦，不會刷寫裝置、解鎖開機載入程式或取得 root 權限。

## 風險說明

請自行確認型號、CSC、版本、儲存位置與磁碟空間。因使用本軟體造成的資料遺失、網路流量、裝置損壞或其他損失由使用者自行承擔。下載韌體的權利歸 Samsung 等相關權利人所有，使用與再散布時請遵守其條款。

```text
samsung-fus-cli check --model SM-S948Q --csc SJP --build BZI9
samsung-fus-cli download --model SM-S948Q --csc SJP --build BZI9
```

自動查詢最新版本：`samsung-fus-cli check --model SC-51Aa --csc DCM --latest`。若傳統 FUS 回傳 `408`，請將您自己裝置的識別碼寫入私人檔案並使用 `--device-id-file PATH`。請勿提交或附加該檔案，也不要使用其他裝置的識別碼或產生的假 IMEI。

再次執行相同下載命令，即可在伺服器支援時續傳。建置方法與授權資訊請參閱[日文 README](README.md)。

**搜尋關鍵字：** 三星韌體下載、Samsung FUS 下載器、Galaxy 韌體下載工具、三星韌體命令列、Windows macOS 三星韌體下載。
