# 运行说明 / 執行說明 / Run Guide

## 项目概览 / 專案概覽 / Overview

- 简体中文：本仓库包含两个独立 Go 模块：`bsphp.go.user`（账号模式）与 `bsphp.go.car`（卡模式）。
- 繁體中文：本倉庫包含兩個獨立 Go 模組：`bsphp.go.user`（帳號模式）與 `bsphp.go.car`（卡密模式）。
- English: This repo contains two standalone Go modules: `bsphp.go.user` (account mode) and `bsphp.go.car` (card mode).

## 运行前准备 / 執行前準備 / Prerequisites

- 简体中文：安装 Go 1.22+，确保网络可访问 `demo.bsphp.com`，并支持图形界面。
- 繁體中文：安裝 Go 1.22+，確認網路可連線 `demo.bsphp.com`，且可開啟圖形介面。
- English: Install Go 1.22+, ensure network access to `demo.bsphp.com`, and run in a GUI-capable environment.

## 运行命令 / 執行指令 / Commands

### 账号模式 / 帳號模式 / Account Mode
```sh
cd bsphp.go.user
go run .
```

### 卡模式 / 卡密模式 / Card Mode
```sh
cd bsphp.go.car
go run .
```

- 简体中文：首次运行会下载依赖，请稍等。
- 繁體中文：首次執行會下載相依套件，請稍候。
- English: The first run will download dependencies.

## 启动行为 / 啟動行為 / Startup Behavior

### `bsphp.go.user`
- 简体中文：启动后执行 `client.Bootstrap()`，加载公告与验证码开关；Web 登录会轮询心跳，`5031` 视为登录成功。
- 繁體中文：啟動後執行 `client.Bootstrap()`，載入公告與驗證碼開關；Web 登入以心跳輪詢，`5031` 視為成功。
- English: On start it runs `client.Bootstrap()`, loads notice and captcha flags; Web login polls heartbeat and treats `5031` as success.

### `bsphp.go.car`
- 简体中文：启动后执行 `client.Bootstrap()`；通过卡串/机器码执行验证、充值、续费等接口。
- 繁體中文：啟動後執行 `client.Bootstrap()`；可透過卡密/機器碼執行驗證、儲值、續費等介面。
- English: On start it runs `client.Bootstrap()`; supports verify/recharge/renew flows with card string or machine code.

## 配置修改 / 設定修改 / Configuration

- 简体中文：请修改以下文件中的常量后重启：
  - `bsphp.go.user/internal/config/config.go`
  - `bsphp.go.car/internal/config/config.go`
- 繁體中文：請修改以下檔案中的常數後重新啟動：
  - `bsphp.go.user/internal/config/config.go`
  - `bsphp.go.car/internal/config/config.go`
- English: Update constants in the files below and restart:
  - `bsphp.go.user/internal/config/config.go`
  - `bsphp.go.car/internal/config/config.go`

## 常见问题 / 常見問題 / FAQ

- 简体中文：若提示网络异常，请先检查代理、防火墙与域名连通性。
- 繁體中文：若提示網路異常，請先檢查代理、防火牆與網域連通性。
- English: If you see network errors, check proxy/firewall/domain connectivity first.

