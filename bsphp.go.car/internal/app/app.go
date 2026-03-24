// Package app 卡模式主界面与控制面板。
package app

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"bsphp_go_demo/car/internal/bsphp"
	"bsphp_go_demo/car/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type CarUI struct {
	fyneApp fyne.App
	client  *bsphp.Client

	mainWindow  fyne.Window
	panelWindow fyne.Window

	noticeLabel *widget.Label
	statusLabel *widget.Label
}

func Run() {
	ui := &CarUI{
		client: bsphp.NewClient(bsphp.ClientConfig{
			URL:              config.BSPHPURL,
			MutualKey:        config.MutualKey,
			ServerPrivateKey: config.ServerPrivateKey,
			ClientPublicKey:  config.ClientPublicKey,
		}),
		fyneApp: app.NewWithID("bsphp.go.car"),
	}
	ui.mainWindow = ui.fyneApp.NewWindow("BSPHP 卡模式")
	ui.mainWindow.Resize(fyne.NewSize(560, 520))
	ui.mainWindow.SetContent(ui.buildMain())
	go ui.bootstrapAsync()
	ui.mainWindow.ShowAndRun()
}

func (ui *CarUI) bootstrapAsync() {
	err := ui.client.Bootstrap()
	msg := "加载中..."
	if err != nil {
		msg = "初始化失败：" + err.Error()
	} else {
		r := ui.client.GetNotice()
		if m := r.Message(); m != "" {
			msg = m
		} else {
			msg = "暂无公告"
		}
	}
	if ui.noticeLabel != nil {
		ui.noticeLabel.SetText(msg)
	}
	if err != nil {
		dialog.ShowError(err, ui.mainWindow)
	}
}

func (ui *CarUI) setStatus(s string) {
	if ui.statusLabel != nil {
		ui.statusLabel.SetText(s)
	}
}

func (ui *CarUI) openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		dialog.ShowError(err, ui.mainWindow)
	}
}

func (ui *CarUI) buildMain() fyne.CanvasObject {
	ui.noticeLabel = widget.NewLabel("加载中...")
	ui.noticeLabel.Wrapping = fyne.TextWrapWord
	ns := container.NewScroll(ui.noticeLabel)
	ns.SetMinSize(fyne.NewSize(0, 60))
	noticeCard := widget.NewCard("公告", "", ns)

	ui.statusLabel = widget.NewLabel("待操作")

	cardInput := widget.NewEntry()
	cardInput.SetPlaceHolder("请输入卡串")
	cardPwd := widget.NewEntry()
	cardPwd.SetPlaceHolder("无密码可留空")

	machineInput := widget.NewEntry()
	machineInput.SetText(bsphp.MachineCode())
	mcKa := widget.NewEntry()
	mcKa.SetPlaceHolder("充值卡号")
	mcPwd := widget.NewEntry()
	mcPwd.SetPlaceHolder("充值密码,没有留空")

	tabCard := container.NewVBox(
		widget.NewLabel("制作的卡密直接登录"),
		formRow("卡串：", cardInput),
		formRow("密码：", cardPwd),
		container.NewHBox(
			widget.NewButton("验证使用", func() { ui.verifyCard(cardInput.Text, cardPwd.Text) }),
			widget.NewButton("网络测试", func() { ui.testNet() }),
			widget.NewButton("版本检测", func() { ui.checkVersion() }),
			widget.NewButton("续费充值", func() { ui.openBrowser(config.RenewURLForUser(cardInput.Text)) }),
			widget.NewButton("购买充值卡", func() { ui.openBrowser(config.GenURL) }),
			widget.NewButton("购买库存卡", func() { ui.openBrowser(config.StockURL) }),
		),
	)

	verifyBox := container.NewVBox(
		formRow("机器码：", machineInput),
		container.NewHBox(
			widget.NewButton("验证使用", func() { ui.verifyMachine(machineInput.Text) }),
			widget.NewButton("网络测试", func() { ui.testNet() }),
			widget.NewButton("版本检测", func() { ui.checkVersion() }),
		),
	)
	renewBox := container.NewVBox(
		formRow("机器码：", machineInput),
		formRow("充值卡号：", mcKa),
		formRow("充值密码：", mcPwd),
		container.NewHBox(
			widget.NewButton("确认充值", func() { ui.activate(machineInput.Text, mcKa.Text, mcPwd.Text) }),
			widget.NewButton("一键支付续费充值", func() { ui.openBrowser(config.RenewURLForUser(machineInput.Text)) }),
			widget.NewButton("购买充值卡", func() { ui.openBrowser(config.GenURL) }),
			widget.NewButton("购买库存卡", func() { ui.openBrowser(config.StockURL) }),
		),
	)
	machineTabs := container.NewAppTabs(
		container.NewTabItem("机器码验证使用", verifyBox),
		container.NewTabItem("机器码充值续费", renewBox),
	)

	tabMachine := container.NewVBox(
		widget.NewLabel("机器码直接注册做卡号模式（账号就是机器码）"),
		machineTabs,
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("制作卡密登陆模式", tabCard),
		container.NewTabItem("一键注册机器码账号", tabMachine),
	)

	body := container.NewVBox(noticeCard, ui.statusLabel, tabs)
	return container.NewPadded(body)
}

func formRow(lbl string, w fyne.CanvasObject) fyne.CanvasObject {
	l := widget.NewLabel(lbl)
	return container.NewBorder(nil, nil, l, nil, w)
}

func (ui *CarUI) verifyCard(icid, icpwd string) {
	icid = strings.TrimSpace(icid)
	if icid == "" {
		ui.setStatus("请输入卡串")
		return
	}
	go func() {
		r := ui.client.LoginIC(icid, icpwd, "", "")
		msg := r.Message()
		if msg == "" {
			msg = "验证失败"
		}
		ui.setStatus(msg)
		if bsphp.LoginOK1081(r) {
			exp := ui.client.GetDateIC().Message()
			if exp == "" {
				exp = "-"
			}
			ui.setStatus("验证成功，主控制面板已在新窗口打开")
			ui.openPanel(icid, exp)
		}
	}()
}

func (ui *CarUI) verifyMachine(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		ui.setStatus("【机器码】请输入机器码（账号）")
		return
	}
	g := bsphp.MachineCode()
	go func() {
		feat := ui.client.AddCardFeatures(id, g, g)
		featMsg := feat.Message()
		if featMsg == "" {
			featMsg = "（无 data）"
		}
		// 与 mac 演示保持一致：仅 1011/1081 或成功文案 视为成功，避免把失败码误判为成功。
		ok := (feat.Code != nil && (*feat.Code == 1011 || *feat.Code == 1081)) ||
			strings.Contains(featMsg, "1081") || strings.Contains(featMsg, "成功")
		code := "nil"
		if feat.Code != nil {
			code = fmt.Sprint(*feat.Code)
		}
		line := fmt.Sprintf("[AddCardFeatures.key.ic] code=%s %s", code, featMsg)
		ui.setStatus(line)
		if !ok {
			return
		}
		r := ui.client.LoginIC(id, "", g, g)
		msg := r.Message()
		if msg == "" {
			msg = "验证失败"
		}
		ui.setStatus("[login.ic] " + msg)
		if bsphp.LoginOK1081(r) {
			exp := ui.client.GetDateIC().Message()
			if exp == "" {
				exp = "-"
			}
			ui.setStatus("验证成功（机器码账号），主控制面板已在新窗口打开")
			ui.openPanel(id, exp)
		}
	}()
}

func (ui *CarUI) testNet() {
	go func() {
		ok := ui.client.Connect()
		t := "网络连接异常"
		if ok {
			t = "网络连接正常"
		}
		ui.setStatus(t)
	}()
}

func (ui *CarUI) checkVersion() {
	go func() {
		v := ui.client.GetVersion().Message()
		if v == "" {
			v = "版本获取失败"
		} else {
			v = "当前版本：" + v
		}
		ui.setStatus(v)
	}()
}

func (ui *CarUI) activate(icid, ka, pwd string) {
	icid = strings.TrimSpace(icid)
	if icid == "" {
		ui.setStatus("【机器码】请输入机器码（账号）")
		return
	}
	ka = strings.TrimSpace(ka)
	if ka == "" {
		ui.setStatus("请输入充值卡号")
		return
	}
	go func() {
		r := ui.client.RechargeCard(icid, ka, strings.TrimSpace(pwd))
		msg := r.Message()
		if msg == "" {
			msg = "（无 data）"
		}
		line := fmt.Sprintf("[chong.ic] code=%v %s", r.Code, msg)
		ui.setStatus(line)
	}()
}
