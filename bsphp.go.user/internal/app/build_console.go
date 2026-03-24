package app

import (
	"fmt"
	"net/url"

	"bsphp_go_demo/user/internal/bsphp"
	"bsphp_go_demo/user/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (ui *UI) buildConsole() fyne.CanvasObject {
	head := widget.NewLabel("")
	ui.updateConsoleHead(head)

	info := widget.NewLabel("")
	info.Wrapping = fyne.TextWrapWord
	infoScroll := container.NewScroll(info)
	infoScroll.SetMinSize(fyne.NewSize(760, 160))

	run := func(title string, fn func() bsphp.APIResult) {
		ui.withBusy(func() {
			r := fn()
			msg := r.Message()
			if msg == "" {
				msg = "（无 data 文本）"
			}
			c := "nil"
			if r.Code != nil {
				c = fmt.Sprint(*r.Code)
			}
			text := fmt.Sprintf("[%s] code=%s\n%s", title, c, msg)
			info.SetText(text)
		})
	}

	runStr := func(title string, fn func() string) {
		ui.withBusy(func() {
			info.SetText(fmt.Sprintf("[%s]\n%s", title, fn()))
		})
	}

	// 公用接口
	pub := widget.NewLabel("公用接口")
	pub.TextStyle = fyne.TextStyle{Bold: true}
	bPub := container.NewGridWithColumns(4,
		ui.cbtn("服务器时间", func() { run("服务器时间", func() bsphp.APIResult { return ui.client.GetServerDate() }) }),
		ui.cbtn("预设URL", func() { run("预设URL", func() bsphp.APIResult { return ui.client.GetPresetURL() }) }),
		ui.cbtn("Web地址", func() { run("Web地址", func() bsphp.APIResult { return ui.client.GetWebURL() }) }),
		ui.cbtn("全局配置", func() { run("全局配置", func() bsphp.APIResult { return ui.client.GetGlobalInfo() }) }),
		ui.cbtn("验证码开关(全部)", func() {
			run("验证码开关(全部)", func() bsphp.APIResult {
				return ui.client.GetCodeEnabled(bsphp.JoinCodeTypes(bsphp.AllCodeTypes))
			})
		}),
		ui.cbtn("登录验证码", func() { run("登录验证码", func() bsphp.APIResult { return ui.client.GetCodeEnabled(string(bsphp.CodeLogin)) }) }),
		ui.cbtn("注册验证码", func() { run("注册验证码", func() bsphp.APIResult { return ui.client.GetCodeEnabled(string(bsphp.CodeReg)) }) }),
		ui.cbtn("找回密码验证码", func() { run("找回密码验证码", func() bsphp.APIResult { return ui.client.GetCodeEnabled(string(bsphp.CodeBackPwd)) }) }),
		ui.cbtn("留言验证码", func() { run("留言验证码", func() bsphp.APIResult { return ui.client.GetCodeEnabled(string(bsphp.CodeSay)) }) }),
		ui.cbtn("逻辑值A", func() { run("逻辑值A", func() bsphp.APIResult { return ui.client.GetLogicA() }) }),
		ui.cbtn("逻辑值B", func() { run("逻辑值B", func() bsphp.APIResult { return ui.client.GetLogicB() }) }),
		ui.cbtn("逻辑值A内容", func() { run("逻辑值A内容", func() bsphp.APIResult { return ui.client.GetLogicInfoA() }) }),
		ui.cbtn("逻辑值B内容", func() { run("逻辑值B内容", func() bsphp.APIResult { return ui.client.GetLogicInfoB() }) }),
	)

	cust := widget.NewLabel("自定义配置模型")
	cust.TextStyle = fyne.TextStyle{Bold: true}
	bCust := container.NewGridWithColumns(3,
		ui.cbtn("软件配置", func() { run("软件配置", func() bsphp.APIResult { return ui.client.GetAppCustom("myapp") }) }),
		ui.cbtn("VIP配置", func() { run("VIP配置", func() bsphp.APIResult { return ui.client.GetAppCustom("myvip") }) }),
		ui.cbtn("登录配置", func() { run("登录配置", func() bsphp.APIResult { return ui.client.GetAppCustom("mylogin") }) }),
	)

	gen := widget.NewLabel("通用接口")
	gen.TextStyle = fyne.TextStyle{Bold: true}
	bGen := container.NewGridWithColumns(2,
		ui.cbtn("获取版本", func() {
			runStr("获取版本", func() string {
				r := ui.client.GetVersion()
				if s, ok := r.Data.(string); ok {
					return s
				}
				return "获取失败"
			})
		}),
		ui.cbtn("获取软件描述", func() { run("获取软件描述", func() bsphp.APIResult { return ui.client.GetSoftInfo() }) }),
	)

	login := widget.NewLabel("登录模式接口")
	login.TextStyle = fyne.TextStyle{Bold: true}
	bLogin := container.NewGridWithColumns(3,
		ui.cbtn("注销登陆", func() {
			ui.withBusy(func() {
				r := ui.client.Logout()
				ui.mu.Lock()
				ui.isLoggedIn = false
				ui.loginEnd = ""
				ui.mu.Unlock()
				ui.updateConsoleHead(head)
				msg := r.Message()
				if msg == "" {
					msg = "注销成功"
				}
				info.SetText(msg)
			})
		}),
		ui.cbtn("检测到期", func() {
			ui.withBusy(func() {
				ui.fetchLoginEndTime()
				ui.mu.Lock()
				le := ui.loginEnd
				ui.mu.Unlock()
				ui.updateConsoleHead(head)
				if le == "" {
					info.SetText("系统错误，取到期时间失败！")
				} else {
					info.SetText("到期时间：" + le)
				}
			})
		}),
		ui.cbtn("取用户信息(默认)", func() { run("取用户信息(默认)", func() bsphp.APIResult { return ui.client.GetUserInfo("") }) }),
		ui.cbtn("心跳包更新", func() { run("心跳包更新", func() bsphp.APIResult { return ui.client.Heartbeat() }) }),
		ui.cbtn("用户特征Key", func() { run("用户特征Key", func() bsphp.APIResult { return ui.client.GetUserKey() }) }),
	)

	infoLbl := widget.NewLabel("取用户信息 info 字段")
	infoLbl.TextStyle = fyne.TextStyle{Italic: true}
	var fieldBtns []fyne.CanvasObject
	for _, f := range bsphp.AllUserInfoFields {
		f := f
		fieldBtns = append(fieldBtns, ui.cbtn(bsphp.UserInfoFieldDisplay(f), func() {
			run(bsphp.UserInfoFieldDisplay(f), func() bsphp.APIResult {
				return ui.client.GetUserInfo(bsphp.JoinUserInfoFields([]bsphp.UserInfoField{f}))
			})
		}))
	}
	bFields := container.NewGridWithColumns(4, fieldBtns...)

	renew := widget.NewLabel("续费订阅推广")
	renew.TextStyle = fyne.TextStyle{Bold: true}
	bRenew := container.NewGridWithColumns(3,
		ui.cbtn("续费订阅(直接)", func() {
			ui.withBusy(func() {
				urlStr := config.BSPHPRenewURL
				r := ui.client.GetUserInfo(string(bsphp.UserName))
				if s, ok := r.Data.(string); ok && s != "" {
					u := bsphp.ParseUserInfoValue(s)
					if u != "" {
						urlStr = urlStr + "&user=" + url.QueryEscape(u)
					}
				}
				ui.openInBrowser(urlStr)
			})
		}),
		ui.cbtn("购买充值卡", func() { ui.openInBrowser(config.BSPHPRenewCardURL) }),
		ui.cbtn("购买库存卡", func() { ui.openInBrowser(config.BSPHPRenewStockCardURL) }),
	)

	all := container.NewVBox(
		head, pub, bPub, cust, bCust, gen, bGen, login, bLogin, infoLbl, bFields, renew, bRenew,
		widget.NewSeparator(), infoScroll,
	)
	scroll := container.NewScroll(all)
	ui.fetchLoginEndTime()
	ui.updateConsoleHead(head)
	return scroll
}

func (ui *UI) updateConsoleHead(lb *widget.Label) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	if ui.isLoggedIn {
		lb.SetText(fmt.Sprintf("已登录    到期时间：%s", ui.loginEnd))
	} else {
		lb.SetText("")
	}
}

func (ui *UI) cbtn(title string, fn func()) *widget.Button {
	b := widget.NewButton(title, func() {
		if !ui.Ready() {
			dialog.ShowInformation("提示", "服务未连接", ui.consoleWindow)
			return
		}
		fn()
	})
	return b
}
