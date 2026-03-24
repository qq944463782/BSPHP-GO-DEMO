# BSPHP-GO-DEMO
[www.bsphp.com](https://www.bsphp.com)

[BSPHP](https://www.bsphp.com) 为软件会员/订阅与授权管理方案，支持账号密码注册、充值卡激活等多种方式。

本仓库是 BSPHP 的 Go 图形化演示，包含两个独立模块：

- `bsphp.go.user`：账号模式（密码/短信/邮箱登录、注册、找回、反馈、控制台）
- `bsphp.go.car`：卡模式（卡串登录、机器码账号、充值续费、控制面板）

## 快速运行

先安装 Go 1.22+，然后在子模块目录运行：

```sh
# 账号模式
cd bsphp.go.user
go run .

# 卡模式
cd ../bsphp.go.car
go run .
```

> 首次运行会下载依赖并生成 `go.sum`。

## 配置说明

两个模块的服务端地址、密钥与 Web 链接在各自配置文件中：

- `bsphp.go.user/internal/config/config.go`
- `bsphp.go.car/internal/config/config.go`

切换测试/正式环境时，修改对应常量后重新运行即可。

## 文档

- 运行说明：`run.md`
- 子项目说明：
  - `bsphp.go.user/README.md`
  - `bsphp.go.car/README.md`

## 目录结构

```text
BSPHP-GO-DEMO/
├── README.md
├── run.md
├── bsphp.go.user/
│   ├── README.md
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   ├── internal/
│   │   ├── app/
│   │   ├── bsphp/
│   │   └── config/
│   └── 配置说明/
└── bsphp.go.car/
    ├── README.md
    ├── main.go
    ├── go.mod
    ├── go.sum
    ├── internal/
    │   ├── app/
    │   ├── bsphp/
    │   └── config/
    └── 配置说明/
```
