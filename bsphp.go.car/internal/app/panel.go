package app

import (
	"fmt"

	"bsphp_go_demo/car/internal/bsphp"
	"bsphp_go_demo/car/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (ui *CarUI) openPanel(loggedID, initialVIP string) {
	if ui.panelWindow != nil {
		ui.panelWindow.Close()
	}
	ui.panelWindow = ui.fyneApp.NewWindow("主控制面板")
	ui.panelWindow.Resize(fyne.NewSize(720, 620))
	ui.panelWindow.SetContent(ui.buildPanel(loggedID, initialVIP))
	ui.panelWindow.SetOnClosed(func() { ui.panelWindow = nil })
	ui.panelWindow.Show()
}

func (ui *CarUI) buildPanel(loggedID, initialVIP string) fyne.CanvasObject {
	vipLbl := widget.NewLabel(initialVIP)
	if vipLbl.Text == "" {
		vipLbl.SetText("-")
	}
	detail := widget.NewLabel("")
	detail.Wrapping = fyne.TextWrapWord
	dScroll := container.NewScroll(detail)
	dScroll.SetMinSize(fyne.NewSize(680, 120))
	auxPwd := widget.NewPasswordEntry()
	auxPwd.SetPlaceHolder("可选：卡密码")

	run := func(title string, fn func() bsphp.APIResult) {
		r := fn()
		body := r.Message()
		if body == "" {
			body = "（无 data 文本）"
		}
		c := "nil"
		if r.Code != nil {
			c = fmt.Sprint(*r.Code)
		}
		detail.SetText(fmt.Sprintf("[%s] code=%s\n%s", title, c, body))
	}

	head := widget.NewLabel("当前卡号：" + loggedID)
	head2 := widget.NewLabel("VIP 到期：")
	row1 := container.NewHBox(
		widget.NewButton("刷新到期", func() {
			r := ui.client.GetDateIC()
			t := r.Message()
			if t != "" {
				vipLbl.SetText(t)
			}
			run("刷新到期", func() bsphp.APIResult { return r })
		}),
		widget.NewButton("登录状态", func() { run("登录状态", func() bsphp.APIResult { return ui.client.GetLoginInfo() }) }),
		widget.NewButton("心跳", func() { run("心跳", func() bsphp.APIResult { return ui.client.Heartbeat() }) }),
		widget.NewButton("公告", func() { run("公告", func() bsphp.APIResult { return ui.client.GetNotice() }) }),
	)
	row2 := container.NewHBox(
		widget.NewButton("服务器时间", func() { run("服务器时间", func() bsphp.APIResult { return ui.client.GetServerDate("") }) }),
		widget.NewButton("版本", func() { run("版本", func() bsphp.APIResult { return ui.client.GetVersion() }) }),
		widget.NewButton("软件描述", func() { run("软件描述", func() bsphp.APIResult { return ui.client.GetSoftInfo() }) }),
		widget.NewButton("预设URL", func() { run("预设URL", func() bsphp.APIResult { return ui.client.GetPresetURL() }) }),
		widget.NewButton("Web地址", func() { run("Web地址", func() bsphp.APIResult { return ui.client.GetWebURL() }) }),
	)
	cust := widget.NewLabel("自定义配置模型")
	cust.TextStyle = fyne.TextStyle{Bold: true}
	row3 := container.NewHBox(
		widget.NewButton("软件配置", func() { run("软件配置", func() bsphp.APIResult { return ui.client.GetAppCustom("myapp", "", "", "", "", "") }) }),
		widget.NewButton("VIP配置", func() { run("VIP配置", func() bsphp.APIResult { return ui.client.GetAppCustom("myvip", "", "", "", "", "") }) }),
		widget.NewButton("登录配置", func() { run("登录配置", func() bsphp.APIResult { return ui.client.GetAppCustom("mylogin", "", "", "", "", "") }) }),
	)
	pub := widget.NewLabel("公共函数")
	pub.TextStyle = fyne.TextStyle{Bold: true}
	row4 := container.NewHBox(
		widget.NewButton("全局配置", func() { run("全局配置", func() bsphp.APIResult { return ui.client.GetGlobalInfo("") }) }),
		widget.NewButton("逻辑A", func() { run("逻辑A", func() bsphp.APIResult { return ui.client.GetLogicA() }) }),
		widget.NewButton("逻辑B", func() { run("逻辑B", func() bsphp.APIResult { return ui.client.GetLogicB() }) }),
		widget.NewButton("激活查询", func() { run("激活查询", func() bsphp.APIResult { return ui.client.QueryCard(loggedID) }) }),
	)
	row5 := container.NewHBox(
		widget.NewButton("卡信息示例", func() {
			run("卡信息示例", func() bsphp.APIResult {
				return ui.client.GetCardInfo(loggedID, auxPwd.Text, "UserName", "")
			})
		}),
	)
	bindRow := container.NewHBox(
		widget.NewButton("绑定本机", func() {
			run("绑定本机", func() bsphp.APIResult {
				return ui.client.BindCard(bsphp.MachineCode(), loggedID, auxPwd.Text)
			})
		}),
		widget.NewButton("解除绑定", func() {
			run("解除绑定", func() bsphp.APIResult {
				return ui.client.UnbindCard(loggedID, auxPwd.Text)
			})
		}),
	)
	webRow := container.NewHBox(
		widget.NewButton("续费充值", func() { ui.openBrowser(config.RenewURLForUser(loggedID)) }),
		widget.NewButton("购买充值卡", func() { ui.openBrowser(config.GenURL) }),
		widget.NewButton("购买库存卡", func() { ui.openBrowser(config.StockURL) }),
	)
	logout := widget.NewButton("注销并返回登录", func() {
		_ = ui.client.Logout()
		if ui.panelWindow != nil {
			ui.panelWindow.Close()
		}
		ui.setStatus("已注销")
	})

	box := container.NewVBox(
		widget.NewCard("主控制面板", "", container.NewVBox(
			head, container.NewHBox(head2, vipLbl),
			row1, row2,
			cust, row3,
			pub, row4, row5,
			widget.NewSeparator(),
			widget.NewLabel("卡密（解绑、绑定本机、卡信息 等需要时填写）"),
			auxPwd,
			bindRow,
			widget.NewSeparator(),
			widget.NewLabel("后台页面"),
			webRow,
			logout,
			widget.NewLabel("接口返回"),
			dScroll,
		)),
	)
	return container.NewScroll(box)
}
