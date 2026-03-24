package app

import (
	"fmt"
	"time"

	"bsphp_go_demo/user/internal/bsphp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

var regQuestions = []string{
	"你最喜欢的颜色？", "你母亲的名字？", "你父亲的名字？", "你的出生地？", "你最喜欢的食物？", "你的小学名称？", "自定义问题",
}

var feedbackTypes = []string{"建议反馈", "BUG", "使用问题"}

func (ui *UI) buildAllTabs() *container.AppTabs {
	return container.NewAppTabs(
		container.NewTabItem("密码登录", ui.tabPasswordLogin()),
		container.NewTabItem("短信登录", ui.tabSmsLogin()),
		container.NewTabItem("邮箱登录", ui.tabEmailLogin()),
		container.NewTabItem("账号注册", ui.tabRegister()),
		container.NewTabItem("短信注册", ui.tabSmsRegister()),
		container.NewTabItem("邮箱注册", ui.tabEmailRegister()),
		container.NewTabItem("解绑", ui.tabUnbind()),
		container.NewTabItem("充值", ui.tabRecharge()),
		container.NewTabItem("短信找回", ui.tabSmsRecover()),
		container.NewTabItem("邮箱找回", ui.tabEmailRecover()),
		container.NewTabItem("找回密码", ui.tabRecoverPwd()),
		container.NewTabItem("修改密码", ui.tabChangePwd()),
		container.NewTabItem("意见反馈", ui.tabFeedback()),
	)
}

func (ui *UI) scrollForm(inner fyne.CanvasObject) fyne.CanvasObject {
	s := container.NewScroll(inner)
	s.SetMinSize(fyne.NewSize(700, 560))
	return s
}

func formRowLbl(lbl string, w fyne.CanvasObject) fyne.CanvasObject {
	l := widget.NewLabel(lbl)
	l.Alignment = fyne.TextAlignTrailing
	return container.NewBorder(nil, nil, l, nil, w)
}

func (ui *UI) refreshCaptcha(img *canvas.Image) {
	if img == nil {
		return
	}
	base := ui.client.CodeImageURL()
	if base == "" {
		return
	}
	ui.mu.Lock()
	ui.codeRefresh = time.Now().Unix()
	ts := ui.codeRefresh
	ui.mu.Unlock()
	full := fmt.Sprintf("%s&_=%d", base, ts)
	u, err := storage.ParseURI(full)
	if err != nil {
		return
	}
	loaded := canvas.NewImageFromURI(u)
	if loaded != nil {
		img.File = loaded.File
		img.Resource = loaded.Resource
		img.Image = loaded.Image
	}
	img.Refresh()
}

func (ui *UI) captchaBlock(coode *widget.Entry) (*widget.Entry, *canvas.Image, fyne.CanvasObject) {
	img := canvas.NewImageFromFile("")
	img.SetMinSize(fyne.NewSize(120, 36))
	img.FillMode = canvas.ImageFillContain
	ui.refreshCaptcha(img)
	ref := widget.NewButton("刷新", func() { ui.refreshCaptcha(img) })
	row := container.NewHBox(coode, img, ref)
	return coode, img, row
}

func (ui *UI) finishAccountLogin(r bsphp.APIResult) bool {
	if r.Code == nil {
		return false
	}
	if *r.Code == 1011 {
		ui.mu.Lock()
		ui.isLoggedIn = true
		ui.mu.Unlock()
		ui.showAlert("BSPHP", "登录成功！")
		ui.openConsole()
		ui.redrawHeader()
		return true
	}
	if *r.Code == 9908 {
		ui.mu.Lock()
		ui.isLoggedIn = true
		ui.mu.Unlock()
		ui.showAlert("BSPHP", "使用已经到期！")
		ui.openConsole()
		ui.redrawHeader()
		return true
	}
	return false
}

func (ui *UI) tabPasswordLogin() fyne.CanvasObject {
	user := widget.NewEntry()
	user.SetText("admin")
	pass := widget.NewPasswordEntry()
	pass.SetText("admin")
	coode := widget.NewEntry()
	_, _, capRow := ui.captchaBlock(coode)

	var captchaRow fyne.CanvasObject
	if ui.codeEnabled(bsphp.CodeLogin) {
		captchaRow = formRowLbl("验  证  码：", capRow)
	}

	doLogin := func() {
		ui.withBusy(func() {
			c := ""
			if ui.codeEnabled(bsphp.CodeLogin) {
				c = coode.Text
			}
			r := ui.client.Login(user.Text, pass.Text, c, "", "")
			if ui.finishAccountLogin(r) {
				return
			}
			msg := r.Message()
			if msg == "" {
				msg = "登录失败"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	}

	btns := container.NewHBox(
		widget.NewButton("测试网络", func() {
			ui.withBusy(func() {
				ok := ui.client.Connect()
				t := "测试连接失败!"
				if ok {
					t = "测试连接成功!"
				}
				dialog.ShowInformation("提示", t, ui.mainWindow)
			})
		}),
		widget.NewButton("检测到期", func() {
			ui.withBusy(func() {
				r := ui.client.GetEndTime()
				msg := r.Message()
				if msg == "" {
					msg = "系统错误，取到期时间失败！"
				}
				ui.showAPIAlert("BSPHP", r.Code, msg)
			})
		}),
		widget.NewButton("获取版本", func() {
			ui.withBusy(func() {
				r := ui.client.GetVersion()
				msg := "获取版本失败"
				if s, ok := r.Data.(string); ok {
					msg = s
				}
				ui.showAPIAlert("BSPHP", r.Code, msg)
			})
		}),
		widget.NewButton("Web方式登陆", func() {
			if ui.Ready() {
				ui.startWebLoginPoll()
			}
		}),
		widget.NewButton("登录", doLogin),
	)

	box := container.NewVBox(
		formRowLbl("登录账号：", user),
		formRowLbl("登录密码：", pass),
	)
	if captchaRow != nil {
		box.Add(captchaRow)
	}
	box.Add(btns)
	return ui.scrollForm(box)
}

func (ui *UI) tabSmsLogin() fyne.CanvasObject {
	mobile := widget.NewEntry()
	area := widget.NewEntry()
	area.SetText("86")
	coode := widget.NewEntry()
	_, _, capRow := ui.captchaBlock(coode)
	smsCode := widget.NewEntry()
	smsCode.SetPlaceHolder("4位数字")
	key := widget.NewEntry()
	key.SetText(bsphp.MachineCode())
	maxor := widget.NewEntry()
	maxor.SetText(bsphp.MachineCode())
	sent := new(bool)
	sentLab := widget.NewLabel("")

	sendBtn := widget.NewButton("发送验证码", func() {
		if !ui.Ready() || mobile.Text == "" || coode.Text == "" {
			return
		}
		ui.withBusy(func() {
			a := area.Text
			if a == "" {
				a = "86"
			}
			r := ui.client.SendSmsCode("login", mobile.Text, a, coode.Text)
			*sent = r.Code != nil && *r.Code == 200
			if *sent {
				sentLab.SetText("已发送(code=200)")
			} else {
				sentLab.SetText("")
			}
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，发送短信验证码失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	loginBtn := widget.NewButton("短信登录", func() {
		if !ui.Ready() || !*sent || smsCode.Text == "" || coode.Text == "" {
			dialog.ShowInformation("提示", "请先成功发送验证码并填写短信码与图片验证码", ui.mainWindow)
			return
		}
		ui.withBusy(func() {
			a := area.Text
			if a == "" {
				a = "86"
			}
			r := ui.client.LoginSms(mobile.Text, a, smsCode.Text, key.Text, maxor.Text, coode.Text)
			if ui.finishAccountLogin(r) {
				return
			}
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，短信验证码登录失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	return ui.scrollForm(container.NewVBox(
		formRowLbl("手机号码：", mobile),
		formRowLbl("区  号：", area),
		formRowLbl("验  证  码：", container.NewVBox(capRow, container.NewHBox(sendBtn, sentLab))),
		formRowLbl("短信验证码：", smsCode),
		widget.NewLabel("OTP有效期：300秒"),
		formRowLbl("绑定特征key：", key),
		formRowLbl("maxoror：", maxor),
		loginBtn,
	))
}

func (ui *UI) tabEmailLogin() fyne.CanvasObject {
	email := widget.NewEntry()
	coode := widget.NewEntry()
	_, _, capRow := ui.captchaBlock(coode)
	emailCode := widget.NewEntry()
	emailCode.SetPlaceHolder("6位数字")
	key := widget.NewEntry()
	key.SetText(bsphp.MachineCode())
	maxor := widget.NewEntry()
	maxor.SetText(bsphp.MachineCode())
	sent := new(bool)
	sentLab := widget.NewLabel("")

	sendBtn := widget.NewButton("发送验证码", func() {
		if !ui.Ready() || email.Text == "" || coode.Text == "" {
			return
		}
		ui.withBusy(func() {
			r := ui.client.SendEmailCode("login", email.Text, coode.Text)
			*sent = r.Code != nil && *r.Code == 200
			if *sent {
				sentLab.SetText("已发送(code=200)")
			} else {
				sentLab.SetText("")
			}
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，发送邮箱验证码失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	loginBtn := widget.NewButton("邮箱登录", func() {
		if !ui.Ready() || !*sent || emailCode.Text == "" || coode.Text == "" {
			dialog.ShowInformation("提示", "请先成功发送验证码并填写邮箱验证码与图片验证码", ui.mainWindow)
			return
		}
		ui.withBusy(func() {
			r := ui.client.LoginEmail(email.Text, emailCode.Text, key.Text, maxor.Text, coode.Text)
			if ui.finishAccountLogin(r) {
				return
			}
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，邮箱验证码登录失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	return ui.scrollForm(container.NewVBox(
		formRowLbl("邮箱地址：", email),
		formRowLbl("验  证  码：", container.NewVBox(capRow, container.NewHBox(sendBtn, sentLab))),
		formRowLbl("邮箱验证码：", emailCode),
		widget.NewLabel("OTP有效期：300秒"),
		formRowLbl("绑定特征key：", key),
		formRowLbl("maxoror：", maxor),
		loginBtn,
	))
}

func (ui *UI) tabRegister() fyne.CanvasObject {
	user := widget.NewEntry()
	pwd := widget.NewPasswordEntry()
	pwd2 := widget.NewPasswordEntry()
	qq := widget.NewEntry()
	qq.SetPlaceHolder("QQ(可选)")
	mail := widget.NewEntry()
	mail.SetPlaceHolder("邮箱(可选)")
	mobile := widget.NewEntry()
	regQ := widget.NewSelect(regQuestions, nil)
	regQ.SetSelected(regQuestions[0])
	answer := widget.NewEntry()
	answer.SetPlaceHolder("答案")
	coode := widget.NewEntry()
	ext := widget.NewEntry()
	ext.SetPlaceHolder("可选")

	var captchaRow fyne.CanvasObject
	if ui.codeEnabled(bsphp.CodeReg) {
		_, _, capRow := ui.captchaBlock(coode)
		refAll := widget.NewButton("刷新", func() {
			_ = ui.ReBootstrap()
			ui.redrawHeader()
		})
		captchaRow = formRowLbl("验  证  码：", container.NewVBox(capRow, refAll))
	}

	regBtn := widget.NewButton("注册", func() {
		ui.withBusy(func() {
			c := ""
			if ui.codeEnabled(bsphp.CodeReg) {
				c = coode.Text
			}
			wenti := regQ.Selected
			if wenti == "" {
				wenti = regQuestions[0]
			}
			r := ui.client.Reg(user.Text, pwd.Text, pwd2.Text, c, mobile.Text, wenti, answer.Text, qq.Text, mail.Text, ext.Text)
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，注册失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	box := container.NewVBox(
		formRowLbl("注册账号：", user),
		formRowLbl("注册密码：", container.NewHBox(pwd, pwd2)),
		formRowLbl("QQ / 邮箱：", container.NewHBox(qq, mail)),
		formRowLbl("手机号码：", mobile),
		formRowLbl("密保问题：", container.NewHBox(regQ, answer)),
	)
	if captchaRow != nil {
		box.Add(captchaRow)
	}
	box.Add(formRowLbl("推  广  码：", ext))
	box.Add(regBtn)
	return ui.scrollForm(box)
}

func (ui *UI) tabSmsRegister() fyne.CanvasObject {
	mobile := widget.NewEntry()
	area := widget.NewEntry()
	area.SetText("86")
	coode := widget.NewEntry()
	_, _, capRow := ui.captchaBlock(coode)
	smsCode := widget.NewEntry()
	user := widget.NewEntry()
	pwd := widget.NewPasswordEntry()
	pwd2 := widget.NewPasswordEntry()
	key := widget.NewEntry()
	key.SetText(bsphp.MachineCode())
	sent := new(bool)
	sentLab := widget.NewLabel("")

	sendBtn := widget.NewButton("发送验证码", func() {
		if !ui.Ready() || mobile.Text == "" || coode.Text == "" {
			return
		}
		ui.withBusy(func() {
			a := area.Text
			if a == "" {
				a = "86"
			}
			r := ui.client.SendSmsCode("register", mobile.Text, a, coode.Text)
			*sent = r.Code != nil && *r.Code == 200
			if *sent {
				sentLab.SetText("已发送(code=200)")
			}
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，发送短信验证码失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	sub := widget.NewButton("短信注册", func() {
		if !ui.Ready() || !*sent {
			dialog.ShowInformation("提示", "请先成功发送验证码", ui.mainWindow)
			return
		}
		ui.withBusy(func() {
			a := area.Text
			if a == "" {
				a = "86"
			}
			r := ui.client.RegisterSms(user.Text, mobile.Text, a, smsCode.Text, pwd.Text, pwd2.Text, key.Text, coode.Text)
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，短信注册失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	return ui.scrollForm(container.NewVBox(
		formRowLbl("手机号码：", mobile),
		formRowLbl("区  号：", area),
		formRowLbl("验  证  码：", container.NewVBox(capRow, container.NewHBox(sendBtn, sentLab))),
		formRowLbl("短信验证码：", smsCode),
		widget.NewLabel("OTP有效期：300秒"),
		formRowLbl("账号：", user),
		formRowLbl("注册密码：", pwd),
		formRowLbl("确认密码：", pwd2),
		formRowLbl("绑定特征key：", key),
		sub,
	))
}

func (ui *UI) tabEmailRegister() fyne.CanvasObject {
	email := widget.NewEntry()
	coode := widget.NewEntry()
	_, _, capRow := ui.captchaBlock(coode)
	emailCode := widget.NewEntry()
	user := widget.NewEntry()
	pwd := widget.NewPasswordEntry()
	pwd2 := widget.NewPasswordEntry()
	key := widget.NewEntry()
	key.SetText(bsphp.MachineCode())
	sent := new(bool)
	sentLab := widget.NewLabel("")

	sendBtn := widget.NewButton("发送验证码", func() {
		if !ui.Ready() || email.Text == "" || coode.Text == "" {
			return
		}
		ui.withBusy(func() {
			r := ui.client.SendEmailCode("register", email.Text, coode.Text)
			*sent = r.Code != nil && *r.Code == 200
			if *sent {
				sentLab.SetText("已发送(code=200)")
			}
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，发送邮箱验证码失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	sub := widget.NewButton("邮箱注册", func() {
		if !ui.Ready() || !*sent || emailCode.Text == "" || user.Text == "" || pwd.Text == "" || pwd2.Text == "" || coode.Text == "" {
			dialog.ShowInformation("提示", "请填写完整并已发送验证码", ui.mainWindow)
			return
		}
		ui.withBusy(func() {
			r := ui.client.RegisterEmail(user.Text, email.Text, emailCode.Text, pwd.Text, pwd2.Text, key.Text, coode.Text)
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，邮箱验证码注册失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	return ui.scrollForm(container.NewVBox(
		formRowLbl("邮箱地址：", email),
		formRowLbl("验  证  码：", container.NewVBox(capRow, container.NewHBox(sendBtn, sentLab))),
		formRowLbl("邮箱验证码：", emailCode),
		widget.NewLabel("OTP有效期：300秒"),
		formRowLbl("账号：", user),
		formRowLbl("注册密码：", pwd),
		formRowLbl("确认密码：", pwd2),
		formRowLbl("绑定特征key：", key),
		sub,
	))
}

func (ui *UI) tabUnbind() fyne.CanvasObject {
	user := widget.NewEntry()
	pass := widget.NewPasswordEntry()
	btn := widget.NewButton("解绑", func() {
		ui.withBusy(func() {
			r := ui.client.Unbind(user.Text, pass.Text)
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，解绑失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})
	return ui.scrollForm(container.NewVBox(
		formRowLbl("登录账号：", user),
		formRowLbl("登录密码：", pass),
		btn,
	))
}

func (ui *UI) tabRecharge() fyne.CanvasObject {
	user := widget.NewEntry()
	pass := widget.NewPasswordEntry()
	card := widget.NewEntry()
	cardPwd := widget.NewPasswordEntry()
	verify := true
	verifyCheck := widget.NewCheck("是(1) 验证登录密码，防止充值错误给了别人", func(v bool) { verify = v })
	verifyCheck.SetChecked(true)

	btn := widget.NewButton("充值", func() {
		ui.withBusy(func() {
			r := ui.client.Pay(user.Text, pass.Text, verify, card.Text, cardPwd.Text)
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，充值失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	return ui.scrollForm(container.NewVBox(
		formRowLbl("充值账号：", user),
		formRowLbl("登录密码：", pass),
		formRowLbl("充值卡号：", card),
		formRowLbl("充值密码：", cardPwd),
		formRowLbl("是否需要验证密码：", container.NewVBox(verifyCheck, widget.NewLabel("否(0) 不验证登录密码即可充值（取消勾选）"))),
		btn,
	))
}

func (ui *UI) tabSmsRecover() fyne.CanvasObject {
	mobile := widget.NewEntry()
	area := widget.NewEntry()
	area.SetText("86")
	coode := widget.NewEntry()
	_, _, capRow := ui.captchaBlock(coode)
	smsCode := widget.NewEntry()
	pwd := widget.NewPasswordEntry()
	pwd2 := widget.NewPasswordEntry()
	sent := new(bool)
	sentLab := widget.NewLabel("")

	sendBtn := widget.NewButton("发送验证码", func() {
		if !ui.Ready() || mobile.Text == "" || coode.Text == "" {
			return
		}
		ui.withBusy(func() {
			a := area.Text
			if a == "" {
				a = "86"
			}
			r := ui.client.SendSmsCode("reset", mobile.Text, a, coode.Text)
			*sent = r.Code != nil && *r.Code == 200
			if *sent {
				sentLab.SetText("已发送(code=200)")
			}
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，发送短信验证码失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	sub := widget.NewButton("重置密码", func() {
		if !ui.Ready() || !*sent {
			return
		}
		ui.withBusy(func() {
			a := area.Text
			if a == "" {
				a = "86"
			}
			r := ui.client.ResetSmsPwd(mobile.Text, a, smsCode.Text, pwd.Text, pwd2.Text, coode.Text)
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，短信找回失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	return ui.scrollForm(container.NewVBox(
		formRowLbl("手机号码：", mobile),
		formRowLbl("区  号：", area),
		formRowLbl("验  证  码：", container.NewVBox(capRow, container.NewHBox(sendBtn, sentLab))),
		formRowLbl("短信验证码：", smsCode),
		formRowLbl("新密码：", pwd),
		formRowLbl("确认新密码：", pwd2),
		sub,
	))
}

func (ui *UI) tabEmailRecover() fyne.CanvasObject {
	email := widget.NewEntry()
	coode := widget.NewEntry()
	_, _, capRow := ui.captchaBlock(coode)
	emailCode := widget.NewEntry()
	pwd := widget.NewPasswordEntry()
	pwd2 := widget.NewPasswordEntry()
	sent := new(bool)
	sentLab := widget.NewLabel("")

	sendBtn := widget.NewButton("发送验证码", func() {
		if !ui.Ready() || email.Text == "" || coode.Text == "" {
			return
		}
		ui.withBusy(func() {
			r := ui.client.SendEmailCode("reset", email.Text, coode.Text)
			*sent = r.Code != nil && *r.Code == 200
			if *sent {
				sentLab.SetText("已发送(code=200)")
			}
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，发送邮箱验证码失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	sub := widget.NewButton("重置密码", func() {
		if !ui.Ready() || !*sent {
			return
		}
		ui.withBusy(func() {
			r := ui.client.ResetEmailPwd(email.Text, emailCode.Text, pwd.Text, pwd2.Text, coode.Text)
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，邮箱找回失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	return ui.scrollForm(container.NewVBox(
		formRowLbl("邮箱地址：", email),
		formRowLbl("验  证  码：", container.NewVBox(capRow, container.NewHBox(sendBtn, sentLab))),
		formRowLbl("邮箱验证码：", emailCode),
		formRowLbl("新密码：", pwd),
		formRowLbl("确认新密码：", pwd2),
		sub,
	))
}

func (ui *UI) tabRecoverPwd() fyne.CanvasObject {
	user := widget.NewEntry()
	regQ := widget.NewSelect(regQuestions, nil)
	regQ.SetSelected(regQuestions[0])
	answer := widget.NewEntry()
	np1 := widget.NewPasswordEntry()
	np2 := widget.NewPasswordEntry()
	coode := widget.NewEntry()

	var captchaRow fyne.CanvasObject
	if ui.codeEnabled(bsphp.CodeBackPwd) {
		_, _, capRow := ui.captchaBlock(coode)
		refAll := widget.NewButton("刷新", func() {
			_ = ui.ReBootstrap()
			ui.redrawHeader()
		})
		captchaRow = formRowLbl("验  证  码：", container.NewVBox(capRow, refAll))
	}

	sub := widget.NewButton("找回密码", func() {
		ui.withBusy(func() {
			c := ""
			if ui.codeEnabled(bsphp.CodeBackPwd) {
				c = coode.Text
			}
			w := regQ.Selected
			if w == "" {
				w = regQuestions[0]
			}
			r := ui.client.BackPass(user.Text, np1.Text, np2.Text, w, answer.Text, c)
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，找回密码失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	box := container.NewVBox(
		formRowLbl("登录账号：", user),
		formRowLbl("密保问题：", container.NewHBox(regQ, answer)),
		formRowLbl("新密码：", container.NewHBox(np1, np2)),
	)
	if captchaRow != nil {
		box.Add(captchaRow)
	}
	box.Add(sub)
	return ui.scrollForm(box)
}

func (ui *UI) tabChangePwd() fyne.CanvasObject {
	user := widget.NewEntry()
	oldp := widget.NewPasswordEntry()
	np1 := widget.NewPasswordEntry()
	np2 := widget.NewPasswordEntry()
	img := widget.NewEntry()
	img.SetPlaceHolder("img 参数（演示与原版一致为独立输入框）")

	sub := widget.NewButton("修改密码", func() {
		ui.withBusy(func() {
			r := ui.client.EditPass(user.Text, oldp.Text, np1.Text, np2.Text, img.Text)
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，修改密码失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	return ui.scrollForm(container.NewVBox(
		formRowLbl("登录账号：", user),
		formRowLbl("旧密码：", oldp),
		formRowLbl("新密码：", container.NewHBox(np1, np2)),
		formRowLbl("img：", img),
		sub,
	))
}

func (ui *UI) tabFeedback() fyne.CanvasObject {
	user := widget.NewEntry()
	pass := widget.NewPasswordEntry()
	title := widget.NewEntry()
	contact := widget.NewEntry()
	ft := widget.NewSelect(feedbackTypes, nil)
	ft.SetSelected(feedbackTypes[0])
	content := widget.NewMultiLineEntry()
	coode := widget.NewEntry()

	var captchaRow fyne.CanvasObject
	if ui.codeEnabled(bsphp.CodeSay) {
		_, _, capRow := ui.captchaBlock(coode)
		captchaRow = formRowLbl("验  证  码：", capRow)
	}

	sub := widget.NewButton("提交", func() {
		ui.withBusy(func() {
			c := ""
			if ui.codeEnabled(bsphp.CodeSay) {
				c = coode.Text
			}
			leix := ft.Selected
			if leix == "" {
				leix = feedbackTypes[0]
			}
			r := ui.client.Feedback(user.Text, pass.Text, title.Text, contact.Text, leix, content.Text, c)
			msg := r.Message()
			if msg == "" {
				msg = "系统错误，意见反馈失败！"
			}
			ui.showAPIAlert("BSPHP", r.Code, msg)
		})
	})

	box := container.NewVBox(
		formRowLbl("账号：", user),
		formRowLbl("密码：", pass),
		formRowLbl("标题：", title),
		formRowLbl("联系：", contact),
		formRowLbl("类型：", ft),
		formRowLbl("内容：", content),
	)
	if captchaRow != nil {
		box.Add(captchaRow)
	}
	box.Add(sub)
	return ui.scrollForm(box)
}
