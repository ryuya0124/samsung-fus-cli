# Samsung FUS CLI

[日本語](README.md) | [English](README.en.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md)

Samsung FUS CLI 是适用于 Windows 和 macOS 的非官方独立命令行三星固件下载工具。它不是 Samsung 官方产品，也未获得 Samsung 的关联或认可。它可以从 Samsung FUS 下载官方 Galaxy 固件、解密 `.enc4` 文件、续传中断的下载，并校验文件大小、CRC32 和 SHA-256。程序使用 Go 编写，运行时不需要 Python。

当前 Samsung Smart Download API 不使用 IMEI 或序列号。只有旧设备的传统 FUS 要求标识符时，才可通过 `--device-id-file` 读取您自己设备的 IMEI 或序列号。该值不会显示，也不会写入 JSON 或设置。本程序只把固件下载到电脑，不会刷写设备、解锁引导加载程序或获取 root 权限。

## 风险说明

请自行确认型号、CSC、版本、保存位置和磁盘空间。因使用本软件造成的数据丢失、网络流量、设备损坏或其他损失由使用者自行承担。下载固件的权利归 Samsung 等相关权利人所有，使用和再分发时请遵守其条款。

```text
samsung-fus-cli check --model SM-S948Q --csc SJP --build BZI9
samsung-fus-cli download --model SM-S948Q --csc SJP --build BZI9
```

自动查询最新版本：`samsung-fus-cli check --model SC-51Aa --csc DCM --latest`。如果传统 FUS 返回 `408`，请将您自己设备的标识符写入私有文件并使用 `--device-id-file PATH`。不要提交或附加该文件，也不要使用其他设备的标识符或生成的假 IMEI。

再次执行同一下载命令即可在服务器支持时续传。构建方法和许可证信息请参阅[日文 README](README.md)。

**搜索关键词：** 三星固件下载、Samsung FUS 下载器、Galaxy 固件下载工具、三星固件命令行、Windows macOS 三星固件下载。
