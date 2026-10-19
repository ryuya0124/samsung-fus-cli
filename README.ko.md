# Samsung FUS CLI

[日本語](README.md) | [English](README.en.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md)

Samsung FUS CLI는 Windows 및 macOS에서 실행되는 비공식 독립 명령줄 삼성 펌웨어 다운로드 도구입니다. Samsung의 공식 제품이 아니며 Samsung과 제휴하거나 승인을 받은 제품이 아닙니다. Samsung FUS에서 공식 Galaxy 펌웨어를 다운로드하고, `.enc4` 파일을 복호화하며, 중단된 다운로드를 재개하고, 파일 크기·CRC32·SHA-256을 확인합니다. Go 단일 실행 파일이므로 실행할 때 Python이 필요하지 않습니다.

IMEI, 일련번호, 전화번호, Samsung 계정 또는 ADB 연결을 사용하지 않습니다. PC로 파일을 받는 기능만 제공하며 기기 플래시, 부트로더 잠금 해제 또는 루팅은 수행하지 않습니다.

## 사용 책임

모델, CSC, 빌드, 저장 위치 및 디스크 여유 공간은 사용자가 직접 확인해야 합니다. 이 소프트웨어 사용으로 발생한 데이터 손실, 통신량, 기기 손상 또는 기타 손해는 사용자 책임입니다. 다운로드한 펌웨어의 권리는 Samsung 등 각 권리자에게 있으며, 사용 및 재배포 시 해당 조건을 따라야 합니다.

```text
samsung-fus-cli check --model SM-S948Q --csc SJP --build BZI9
samsung-fus-cli download --model SM-S948Q --csc SJP --build BZI9
```

같은 명령을 다시 실행하면 서버가 지원하는 경우 이어받습니다. 자세한 빌드 방법과 라이선스는 [일본어 README](README.md)를 참조하세요.

**검색어:** 삼성 펌웨어 다운로드, Samsung FUS 다운로드, Galaxy 펌웨어 다운로드, 삼성 펌웨어 CLI, Windows macOS 삼성 펌웨어 도구.
