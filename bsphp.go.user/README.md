# bsphp.go.user

`bsphp.go.user` 是 BSPHP 账号模式 Go 演示工程（Fyne GUI）。

## 功能概览

- 密码登录、短信登录、邮箱登录
- 账号注册、短信注册、邮箱注册
- 短信找回、邮箱找回、密保找回、修改密码
- 解绑、充值、意见反馈
- Web 登录轮询（心跳 5031 判定）
- 登录后控制台接口调试

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
- `BSPHPMutualKey`
- `BSPHPServerPrivateKey`
- `BSPHPClientPublicKey`
- `BSPHPCodeURL`
- `BSPHPWebLoginURL`

以上参数需与后台同一应用保持一致。

## 目录结构

```text
bsphp.go.user/
├── README.md
├── main.go
├── go.mod
├── go.sum
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   ├── build_tabs.go
│   │   └── build_console.go
│   ├── bsphp/
│   │   ├── client.go
│   │   ├── crypto.go
│   │   └── machine.go
│   └── config/
│       └── config.go
└── 配置说明/
```
