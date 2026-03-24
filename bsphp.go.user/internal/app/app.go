// Package app 账号模式图形界面（登录、注册、控制台等）。
package app

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"bsphp_go_demo/user/internal/bsphp"
	"bsphp_go_demo/user/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// UI 持有 API 客户端与窗口状态。
type UI struct {
	mu sync.Mutex

	fyneApp fyne.App
	client  *bsphp.Client

	mainWindow    fyne.Window
	consoleWindow fyne.Window
	webPollStop   chan struct{}

	isReady     bool
	isLoggedIn  bool
	initErr     string
	loginEnd    string
	noticeText  string
	codeMap     map[bsphp.CodeType]bool
	codeRefresh int64
	isBusy      bool

	noticeLabel *widget.Label
	statusLabel *widget.Label
	tabs        *container.AppTabs
}

func Run() {
	cfg := bsphp.ClientConfig{
		URL:              config.BSPHPURL,
		MutualKey:        config.BSPHPMutualKey,
		ServerPrivateKey: config.BSPHPServerPrivateKey,
		ClientPublicKey:  config.BSPHPClientPublicKey,
		CodeURLPrefix:    config.BSPHPCodeURL,
	}
	ui := &UI{
		client: bsphp.NewClient(cfg),
		fyneApp: app.NewWithID("bsphp.go.user"),
	}
	ui.mainWindow = ui.fyneApp.NewWindow("BSPHP 账号模式")
	ui.mainWindow.Resize(fyne.NewSize(780, 840))
	ui.mainWindow.SetContent(ui.buildMain())
	go ui.bootstrapAsync()
	ui.mainWindow.ShowAndRun()
}

func (ui *UI) bootstrapAsync() {
	err := ui.client.Bootstrap()
	ui.mu.Lock()
	if err != nil {
		ui.initErr = err.Error()
		ui.isReady = false
	} else {
		ui.initErr = ""
		ui.isReady = true
		ui.codeMap = ui.client.FetchCodeEnabledMap()
		r := ui.client.GetNotice()
		if r.Message() != "" {
			ui.noticeText = r.Message()
		} else {
			ui.noticeText = "公告获取失败"
		}
	}
	ui.mu.Unlock()
	ui.RebuildMain()
	ui.refreshFromMainThread()
	if ui.initErr != "" {
		dialog.ShowError(fmt.Errorf("初始化失败: %s", ui.initErr), ui.mainWindow)
	}
}

func (ui *UI) refreshFromMainThread() {
	ui.redrawHeader()
	if ui.tabs != nil {
		ui.tabs.Refresh()
	}
}

func (ui *UI) redrawHeader() {
	if ui.noticeLabel != nil {
		ui.mu.Lock()
		t := ui.noticeText
		if ui.initErr != "" && !ui.isReady {
			t = ui.initErr
		}
		ui.mu.Unlock()
		ui.noticeLabel.SetText(t)
	}
	if ui.statusLabel != nil {
		ui.mu.Lock()
		s := "服务未连接"
		if ui.isReady {
			s = "服务已连接"
		}
		if ui.isLoggedIn {
			s += "  |  已登录"
		}
		ui.mu.Unlock()
		ui.statusLabel.SetText(s)
	}
}

func (ui *UI) buildMain() fyne.CanvasObject {
	ui.noticeLabel = widget.NewLabel("加载中...")
	ui.noticeLabel.Wrapping = fyne.TextWrapWord
	noticeScroll := container.NewScroll(ui.noticeLabel)
	noticeScroll.SetMinSize(fyne.NewSize(0, 90))
	noticeBox := widget.NewCard("公告", "", noticeScroll)

	ui.statusLabel = widget.NewLabel("…")
	statusBar := container.NewBorder(nil, nil, nil, nil, ui.statusLabel)

	ui.tabs = ui.buildAllTabs()
	body := container.NewVBox(noticeBox, statusBar, ui.tabs)
	return container.NewPadded(body)
}

// RebuildMain 初始化完成后刷新布局（验证码开关等依赖服务端）
func (ui *UI) RebuildMain() {
	ui.mainWindow.SetContent(ui.buildMain())
	ui.redrawHeader()
}

func (ui *UI) showAlert(title, msg string) {
	dialog.ShowInformation(title, msg, ui.mainWindow)
}

func (ui *UI) showAPIAlert(title string, code *int, msg string) {
	t := title
	if code != nil {
		t = fmt.Sprintf("%s (code=%d)", title, *code)
	}
	dialog.ShowInformation(t, msg, ui.mainWindow)
}

func (ui *UI) Ready() bool {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	return ui.isReady
}

func (ui *UI) codeEnabled(t bsphp.CodeType) bool {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	return bsphp.CodeEnabled(ui.codeMap, t)
}

func (ui *UI) withBusy(fn func()) {
	ui.mu.Lock()
	if ui.isBusy {
		ui.mu.Unlock()
		return
	}
	ui.isBusy = true
	ui.mu.Unlock()
	go func() {
		defer func() {
			ui.mu.Lock()
			ui.isBusy = false
			ui.mu.Unlock()
			ui.refreshFromMainThread()
		}()
		fn()
	}()
}

func (ui *UI) openInBrowser(urlStr string) {
	var err error
	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", urlStr).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", urlStr).Start()
	default:
		err = exec.Command("xdg-open", urlStr).Start()
	}
	if err != nil {
		dialog.ShowError(err, ui.mainWindow)
	}
}

func (ui *UI) fetchLoginEndTime() {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	r := ui.client.GetUserInfo(string(bsphp.UserVipDate))
	if s, ok := r.Data.(string); ok && s != "" {
		ui.loginEnd = bsphp.ParseUserInfoValue(s)
		if ui.loginEnd != "" {
			return
		}
	}
	r2 := ui.client.GetEndTime()
	if s, ok := r2.Data.(string); ok {
		ui.loginEnd = strings.TrimSpace(s)
	}
}

// ReBootstrap 注册/找回页「刷新」：重连、刷新验证码开关与公告
func (ui *UI) ReBootstrap() error {
	if err := ui.client.Bootstrap(); err != nil {
		return err
	}
	ui.mu.Lock()
	ui.codeMap = ui.client.FetchCodeEnabledMap()
	r := ui.client.GetNotice()
	if r.Message() != "" {
		ui.noticeText = r.Message()
	}
	ui.mu.Unlock()
	ui.refreshFromMainThread()
	return nil
}

func (ui *UI) openConsole() {
	if ui.consoleWindow == nil {
		ui.consoleWindow = ui.fyneApp.NewWindow("控制台")
		ui.consoleWindow.Resize(fyne.NewSize(820, 480))
		ui.consoleWindow.SetContent(ui.buildConsole())
		ui.consoleWindow.SetOnClosed(func() {
			ui.consoleWindow = nil
		})
	}
	ui.fetchLoginEndTime()
	ui.consoleWindow.Show()
	ui.consoleWindow.RequestFocus()
}

// 开始 Web 登录：打开系统浏览器，并轮询心跳直至返回 5031。
func (ui *UI) startWebLoginPoll() {
	ui.mu.Lock()
	if ui.webPollStop != nil {
		close(ui.webPollStop)
	}
	ui.webPollStop = make(chan struct{})
	stop := ui.webPollStop
	ui.mu.Unlock()

	urlStr := config.BSPHPWebLoginURL + ui.client.BsPhpSeSsL
	ui.openInBrowser(urlStr)
	dialog.ShowInformation("Web 登录",
		"已在系统浏览器打开登录页。完成登录后本程序将自动检测（心跳 5031）；也可关闭本提示等待。",
		ui.mainWindow)

	go func() {
		t := time.NewTicker(2 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				r := ui.client.Heartbeat()
				if r.Code != nil && *r.Code == 5031 {
					ui.mu.Lock()
					ui.isLoggedIn = true
					ui.mu.Unlock()
					ui.openConsole()
					ui.redrawHeader()
					return
				}
			}
		}
	}()
}
