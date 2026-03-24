# bsphp.go.car

`bsphp.go.car` 是 BSPHP 卡模式 Go 演示工程（Fyne GUI）。

## 功能概览

- 卡串登录验证
- 机器码账号验证（`AddCardFeatures.key.ic` + `login.ic`）
- 机器码账号充值续费（`chong.ic`）
- 网络测试、版本检测、公告
- 登录后主控制面板（心跳、到期、绑定/解绑、配置读取等）

## 运行

```sh
go run .
```

或先编译：

```sh
go build ./...
```

## 主要配置

编辑 `internal/config/config.go`：

- `BSPHPURL`
- `MutualKey`
- `ServerPrivateKey`
- `ClientPublicKey`
- `RenewBase` / `GenURL` / `StockURL`

以上参数需要与后台同一应用一致（含 `daihao`）。

## 目录结构

```text
bsphp.go.car/
├── README.md
├── main.go
├── go.mod
├── go.sum
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   └── panel.go
│   ├── bsphp/
│   │   ├── client.go
│   │   ├── crypto.go
│   │   └── machine.go
│   └── config/
│       └── config.go
└── 配置说明/
```
