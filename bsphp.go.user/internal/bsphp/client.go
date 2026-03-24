// Package bsphp：AppEn 加密通讯与账号模式 .lg 接口封装。
package bsphp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func shortForLog(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// ClientConfig 连接参数（URL、mutualkey、RSA 等）。
type ClientConfig struct {
	URL               string
	MutualKey         string
	ServerPrivateKey  string
	ClientPublicKey   string
	CodeURLPrefix     string
}

// APIResult 单次接口返回的 data 与 code。
type APIResult struct {
	Data any
	Code *int
}

func (r APIResult) Message() string {
	if s, ok := r.Data.(string); ok {
		return s
	}
	if arr, ok := r.Data.([]any); ok {
		parts := make([]string, 0, len(arr))
		for _, v := range arr {
			parts = append(parts, fmt.Sprint(v))
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

// Client BSPHP 账号模式客户端
type Client struct {
	cfg         ClientConfig
	BsPhpSeSsL  string
	httpClient  *http.Client
	dateFmt     string // yyyy-MM-dd HH:mm:ss
	dateHashFmt string // yyyy-MM-dd#HH:mm:ss
}

func NewClient(cfg ClientConfig) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		dateFmt:     "2006-01-02 15:04:05",
		dateHashFmt: "2006-01-02#15:04:05",
	}
}

func (c *Client) CodeImageURL() string {
	if c.cfg.CodeURLPrefix == "" {
		return ""
	}
	return c.cfg.CodeURLPrefix + c.BsPhpSeSsL
}

// EncodeParameter 请求体拼接用：仅 A–Z a–z 0–9 及 -._~ 原样，其余百分号编码（与服务端约定一致）。
func EncodeParameter(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '.' || r == '_' || r == '~' {
			b.WriteRune(r)
		} else {
			for _, x := range []byte(string(r)) {
				b.WriteString(fmt.Sprintf("%%%02X", x))
			}
		}
	}
	return b.String()
}

func (c *Client) Send(api string, params map[string]string) (map[string]any, error) {
	now := time.Now()
	appsafecode := MD5Hex(now.Format(c.dateFmt))

	param := map[string]string{
		"api":          api,
		"BSphpSeSsL":   c.BsPhpSeSsL,
		"date":         now.Format(c.dateHashFmt),
		"md5":          "",
		"mutualkey":    c.cfg.MutualKey,
		"appsafecode":  appsafecode,
	}
	for k, v := range params {
		param[k] = v
	}

	var pairs []string
	for k, v := range param {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, EncodeParameter(v)))
	}
	dataStr := strings.Join(pairs, "&")
	fmt.Println("[BSPHP] ========== 加密前 ==========")
	fmt.Printf("[BSPHP] api: %s\n", api)
	fmt.Printf("[BSPHP] request_data: %s\n", dataStr)

	aesKeyFull := MD5Hex(c.cfg.ServerPrivateKey + appsafecode)
	aesKey := aesKeyFull[:16]

	encB64, err := AES128CBCEncryptBase64(dataStr, aesKey)
	if err != nil {
		fmt.Println("[BSPHP] ========== 请求加密失败（故无「加密后」「解密前」）==========")
		fmt.Printf("[BSPHP] api: %s 错误: %v\n", api, err)
		fmt.Println("[BSPHP] 常见原因：clientPublicKey 须为后台「客户端公钥」；勿与 serverPrivateKey 对调；须与 mutualkey 对应应用一致")
		return nil, err
	}
	sigMd5 := MD5Hex(encB64)
	sigContent := fmt.Sprintf("0|AES-128-CBC|%s|%s|json", aesKey, sigMd5)
	rsaB64, err := RSAEncryptPKCS1Base64(sigContent, c.cfg.ClientPublicKey)
	if err != nil {
		fmt.Println("[BSPHP] ========== 请求加密失败（故无「加密后」「解密前」）==========")
		fmt.Printf("[BSPHP] api: %s 错误: %v\n", api, err)
		fmt.Println("[BSPHP] 常见原因：clientPublicKey 须为后台「客户端公钥」；勿与 serverPrivateKey 对调；须与 mutualkey 对应应用一致")
		return nil, err
	}

	payload := encB64 + "|" + rsaB64
	encoded := EncodeParameter(payload)
	fmt.Println("[BSPHP] ========== 加密后 ==========")
	fmt.Printf("[BSPHP] encrypted_b64: %s\n", shortForLog(encB64, 120))
	fmt.Printf("[BSPHP] rsa_b64: %s\n", shortForLog(rsaB64, 120))
	fmt.Printf("[BSPHP] payload 总长度: %d, encoded 总长度: %d\n", len(payload), len(encoded))
	body := "parameter=" + encoded

	req, err := http.NewRequest(http.MethodPost, c.cfg.URL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 || len(rawBytes) == 0 {
		fmt.Printf("[BSPHP] send 失败: HTTP %d，body 长度 %d\n", resp.StatusCode, len(rawBytes))
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}

	raw := string(rawBytes)
	if u, err := url.QueryUnescape(raw); err == nil {
		raw = u
	}
	fmt.Println("[BSPHP] ========== 解密前(服务器原始响应) ==========")
	fmt.Printf("[BSPHP] raw 长度: %d\n", len(raw))
	fmt.Println("[BSPHP] raw (完整，便于复制整段控制台):")
	fmt.Println(raw)

	parts := strings.Split(raw, "|")
	if len(parts) < 3 {
		fmt.Printf("[BSPHP] send 失败: 正文需至少 3 段 pipe 分隔密文（当前 %d 段），常见于 PHP 报错/HTML 或非加密直出\n", len(parts))
		return nil, errorsNew("bad response pipe count")
	}
	respEnc := strings.TrimSpace(parts[1])
	respRsa := strings.TrimSpace(parts[2])

	sigDecrypted, err := RSADecryptPKCS1Base64(respRsa, c.cfg.ServerPrivateKey)
	if err != nil {
		fmt.Printf("[BSPHP] send 失败: 响应 RSA 段解密异常（检查 serverPrivateKey 是否与后台「服务器私钥」一致）— %v\n", err)
		return nil, err
	}
	sigParts := strings.Split(sigDecrypted, "|")
	if len(sigParts) < 4 {
		fmt.Printf("[BSPHP] send 失败: 签名串分段不足 4 段: %s\n", shortForLog(sigDecrypted, 80))
		return nil, errorsNew("bad sig parts")
	}
	respAesKey := sigParts[2]
	if len(respAesKey) < 16 {
		fmt.Println("[BSPHP] send 失败: bad aes key")
		return nil, errorsNew("bad aes key")
	}
	key16 := respAesKey[:16]

	decrypted, err := AES128CBCDecryptBase64ToString(respEnc, key16)
	if err != nil {
		fmt.Printf("[BSPHP] send 失败: AES 解密异常 — %v\n", err)
		return nil, err
	}
	fmt.Println("[BSPHP] ========== 解密后 ==========")
	fmt.Printf("[BSPHP] decrypted: %s\n", decrypted)

	var root map[string]any
	if err := json.Unmarshal([]byte(decrypted), &root); err != nil {
		fmt.Println("[BSPHP] send 失败: 解密结果不是含 response 的 JSON（若 raw 是 HTML/Notice 说明 URL 或 PHP 报错）")
		return nil, err
	}
	respObj, ok := root["response"].(map[string]any)
	if !ok {
		return nil, errorsNew("no response object")
	}

	if ac, ok := respObj["appsafecode"].(string); ok && ac != appsafecode {
		respObj["data"] = "appsafecode 安全参数验证不通过"
	}
	return respObj, nil
}

func errorsNew(s string) error { return fmt.Errorf("%s", s) }

func intFromAny(v any) *int {
	switch x := v.(type) {
	case float64:
		i := int(x)
		return &i
	case int:
		return &x
	case json.Number:
		i, _ := x.Int64()
		ii := int(i)
		return &ii
	default:
		return nil
	}
}

func (c *Client) apiResult(api string, params map[string]string) APIResult {
	r, err := c.Send(api, params)
	if err != nil || r == nil {
		return APIResult{}
	}
	code := intFromAny(r["code"])
	return APIResult{Data: r["data"], Code: code}
}

func (c *Client) Connect() bool {
	r, err := c.Send("internet.in", nil)
	if err != nil || r == nil {
		return false
	}
	s, _ := r["data"].(string)
	return s == "1"
}

func (c *Client) GetSeSsL() error {
	r, err := c.Send("BSphpSeSsL.in", nil)
	if err != nil {
		return err
	}
	s, _ := r["data"].(string)
	if s == "" {
		return errorsNew("empty BSphpSeSsL")
	}
	c.BsPhpSeSsL = s
	return nil
}

func (c *Client) Bootstrap() error {
	if !c.Connect() {
		return errorsNew("连接失败")
	}
	if err := c.GetSeSsL(); err != nil {
		return errorsNew("获取 BSphpSeSsL 失败")
	}
	return nil
}

func (c *Client) Logout() APIResult {
	r, _ := c.Send("cancellation.lg", nil)
	var res APIResult
	if r != nil {
		res.Data = r["data"]
		res.Code = intFromAny(r["code"])
	}
	c.BsPhpSeSsL = ""
	_ = c.GetSeSsL()
	return res
}

func (c *Client) GetNotice() APIResult       { return c.apiResult("gg.in", nil) }
func (c *Client) GetVersion() APIResult      { return c.apiResult("v.in", nil) }
func (c *Client) GetSoftInfo() APIResult     { return c.apiResult("miao.in", nil) }
func (c *Client) GetServerDate() APIResult   { return c.apiResult("date.in", nil) }
func (c *Client) GetPresetURL() APIResult    { return c.apiResult("url.in", nil) }
func (c *Client) GetWebURL() APIResult       { return c.apiResult("weburl.in", nil) }
func (c *Client) GetGlobalInfo() APIResult   { return c.apiResult("globalinfo.in", nil) }
func (c *Client) GetLogicA() APIResult       { return c.apiResult("logica.in", nil) }
func (c *Client) GetLogicB() APIResult       { return c.apiResult("logicb.in", nil) }
func (c *Client) GetLogicInfoA() APIResult   { return c.apiResult("logicinfoa.in", nil) }
func (c *Client) GetLogicInfoB() APIResult   { return c.apiResult("logicinfob.in", nil) }
func (c *Client) GetEndTime() APIResult      { return c.apiResult("vipdate.lg", nil) }
func (c *Client) GetUserKey() APIResult      { return c.apiResult("userkey.lg", nil) }
func (c *Client) Heartbeat() APIResult       { return c.apiResult("timeout.lg", nil) }

func (c *Client) GetAppCustom(info string) APIResult {
	return c.apiResult("appcustom.in", map[string]string{"info": info})
}

func (c *Client) GetCodeEnabled(typeStr string) APIResult {
	p := map[string]string{}
	if typeStr != "" {
		p["type"] = typeStr
	}
	return c.apiResult("getsetimag.in", p)
}

func (c *Client) GetUserInfo(info string) APIResult {
	p := map[string]string{}
	if info != "" {
		p["info"] = info
	}
	return c.apiResult("getuserinfo.lg", p)
}

func (c *Client) loginUpdateSeSsL(r map[string]any) {
	code := intFromAny(r["code"])
	if code == nil {
		return
	}
	if (*code == 1011 || *code == 9908) {
		if ssl, ok := r["SeSsL"].(string); ok && ssl != "" {
			c.BsPhpSeSsL = ssl
		}
	}
}

func (c *Client) Login(user, password, code, key, maxoror string) APIResult {
	mc := key
	if mc == "" {
		mc = MachineCode()
	}
	mx := maxoror
	if mx == "" {
		mx = mc
	}
	r, err := c.Send("login.lg", map[string]string{
		"user": user, "pwd": password, "coode": code,
		"key": mc, "maxoror": mx,
	})
	if err != nil || r == nil {
		return APIResult{Data: "系统错误，登录失败！"}
	}
	c.loginUpdateSeSsL(r)
	return APIResult{Data: r["data"], Code: intFromAny(r["code"])}
}

func (c *Client) SendEmailCode(scene, email, coode string) APIResult {
	return c.apiResult("send_email.lg", map[string]string{"scene": scene, "email": email, "coode": coode})
}

func (c *Client) RegisterEmail(user, email, emailCode, pwd, pwdb, key, coode string) APIResult {
	return c.apiResult("register_email.lg", map[string]string{
		"user": user, "email": email, "email_code": emailCode,
		"pwd": pwd, "pwdb": pwdb, "key": key, "coode": coode,
	})
}

func (c *Client) LoginEmail(email, emailCode, key, maxoror, coode string) APIResult {
	r, err := c.Send("login_email.lg", map[string]string{
		"email": email, "email_code": emailCode, "key": key, "maxoror": maxoror, "coode": coode,
	})
	if err != nil || r == nil {
		return APIResult{Data: "系统错误，邮箱验证码登录失败！"}
	}
	c.loginUpdateSeSsL(r)
	return APIResult{Data: r["data"], Code: intFromAny(r["code"])}
}

func (c *Client) ResetEmailPwd(email, emailCode, pwd, pwdb, coode string) APIResult {
	return c.apiResult("resetpwd_email.lg", map[string]string{
		"email": email, "email_code": emailCode, "pwd": pwd, "pwdb": pwdb, "coode": coode,
	})
}

func (c *Client) SendSmsCode(scene, mobile, area, coode string) APIResult {
	return c.apiResult("send_sms.lg", map[string]string{
		"scene": scene, "mobile": mobile, "area": area, "coode": coode,
	})
}

func (c *Client) RegisterSms(user, mobile, area, smsCode, pwd, pwdb, key, coode string) APIResult {
	return c.apiResult("register_sms.lg", map[string]string{
		"user": user, "mobile": mobile, "area": area, "sms_code": smsCode,
		"pwd": pwd, "pwdb": pwdb, "key": key, "coode": coode,
	})
}

func (c *Client) LoginSms(mobile, area, smsCode, key, maxoror, coode string) APIResult {
	r, err := c.Send("login_sms.lg", map[string]string{
		"mobile": mobile, "area": area, "sms_code": smsCode,
		"key": key, "maxoror": maxoror, "coode": coode,
	})
	if err != nil || r == nil {
		return APIResult{Data: "系统错误，短信验证码登录失败！"}
	}
	c.loginUpdateSeSsL(r)
	return APIResult{Data: r["data"], Code: intFromAny(r["code"])}
}

func (c *Client) ResetSmsPwd(mobile, area, smsCode, pwd, pwdb, coode string) APIResult {
	return c.apiResult("resetpwd_sms.lg", map[string]string{
		"mobile": mobile, "area": area, "sms_code": smsCode,
		"pwd": pwd, "pwdb": pwdb, "coode": coode,
	})
}

func (c *Client) Reg(user, pwd, pwdb, coode, mobile, mibaoWenti, mibaoDaan, qq, mail, extensionCode string) APIResult {
	return c.apiResult("registration.lg", map[string]string{
		"user": user, "pwd": pwd, "pwdb": pwdb, "coode": coode,
		"mobile": mobile, "mibao_wenti": mibaoWenti, "mibao_daan": mibaoDaan,
		"qq": qq, "mail": mail, "extension": extensionCode,
	})
}

func (c *Client) Unbind(user, pwd string) APIResult {
	return c.apiResult("jiekey.lg", map[string]string{"user": user, "pwd": pwd})
}

func (c *Client) Pay(user, userpwd string, userset bool, ka, pwd string) APIResult {
	us := "0"
	if userset {
		us = "1"
	}
	return c.apiResult("chong.lg", map[string]string{
		"user": user, "userpwd": userpwd, "userset": us, "ka": ka, "pwd": pwd,
	})
}

func (c *Client) BackPass(user, pwd, pwdb, wenti, daan, coode string) APIResult {
	return c.apiResult("backto.lg", map[string]string{
		"user": user, "pwd": pwd, "pwdb": pwdb,
		"wenti": wenti, "daan": daan, "coode": coode,
	})
}

func (c *Client) EditPass(user, pwd, pwda, pwdb, img string) APIResult {
	return c.apiResult("password.lg", map[string]string{
		"user": user, "pwd": pwd, "pwda": pwda, "pwdb": pwdb, "img": img,
	})
}

func (c *Client) Feedback(user, pwd, table, qq, leix, text, coode string) APIResult {
	return c.apiResult("liuyan.in", map[string]string{
		"user": user, "pwd": pwd, "table": table, "qq": qq, "leix": leix, "txt": text, "coode": coode,
	})
}

// UserInfoField getuserinfo.lg 的 info 字段名。
type UserInfoField string

const (
	UserName               UserInfoField = "UserName"
	UserUID                UserInfoField = "UserUID"
	UserReDate             UserInfoField = "UserReDate"
	UserReIp               UserInfoField = "UserReIp"
	UserIsLock             UserInfoField = "UserIsLock"
	UserLogInDate          UserInfoField = "UserLogInDate"
	UserLogInIp            UserInfoField = "UserLogInIp"
	UserVipDate            UserInfoField = "UserVipDate"
	UserKeyField           UserInfoField = "UserKey"
	ClassNane              UserInfoField = "Class_Nane"
	ClassMark              UserInfoField = "Class_Mark"
	UserQQ                 UserInfoField = "UserQQ"
	UserMAIL               UserInfoField = "UserMAIL"
	UserPayZhe             UserInfoField = "UserPayZhe"
	UserTreasury           UserInfoField = "UserTreasury"
	UserMobile             UserInfoField = "UserMobile"
	UserRMB                UserInfoField = "UserRMB"
	UserPoint              UserInfoField = "UserPoint"
	UsermibaoWenti         UserInfoField = "Usermibao_wenti"
	UserVipWhether         UserInfoField = "UserVipWhether"
	UserVipDateSurplusDAY  UserInfoField = "UserVipDateSurplus_DAY"
	UserVipDateSurplusH    UserInfoField = "UserVipDateSurplus_H"
	UserVipDateSurplusI    UserInfoField = "UserVipDateSurplus_I"
	UserVipDateSurplusS    UserInfoField = "UserVipDateSurplus_S"
)

var AllUserInfoFields = []UserInfoField{
	UserName, UserUID, UserReDate, UserReIp, UserIsLock, UserLogInDate, UserLogInIp,
	UserVipDate, UserKeyField, ClassNane, ClassMark, UserQQ, UserMAIL, UserPayZhe,
	UserTreasury, UserMobile, UserRMB, UserPoint, UsermibaoWenti, UserVipWhether,
	UserVipDateSurplusDAY, UserVipDateSurplusH, UserVipDateSurplusI, UserVipDateSurplusS,
}

func UserInfoFieldDisplay(f UserInfoField) string {
	m := map[UserInfoField]string{
		UserName: "用户名称", UserUID: "用户UID", UserReDate: "激活时间", UserReIp: "激活时Ip",
		UserIsLock: "用户状态", UserLogInDate: "登录时间", UserLogInIp: "登录Ip", UserVipDate: "到期时",
		UserKeyField: "绑定特征", ClassNane: "用户分组名称", ClassMark: "用户分组别名", UserQQ: "用户QQ",
		UserMAIL: "用户邮箱", UserPayZhe: "购卡折扣", UserTreasury: "是否代理", UserMobile: "电话",
		UserRMB: "帐号金额", UserPoint: "帐号积分", UsermibaoWenti: "密保问题", UserVipWhether: "vip是否到期",
		UserVipDateSurplusDAY: "到期倒计时-天", UserVipDateSurplusH: "到期倒计时-时",
		UserVipDateSurplusI: "到期倒计时-分", UserVipDateSurplusS: "到期倒计时-秒",
	}
	if s, ok := m[f]; ok {
		return s
	}
	return string(f)
}

func JoinUserInfoFields(fs []UserInfoField) string {
	var parts []string
	for _, f := range fs {
		parts = append(parts, string(f))
	}
	return strings.Join(parts, ",")
}

// CodeType 验证码开关
type CodeType string

const (
	CodeLogin   CodeType = "INGES_LOGIN"
	CodeReg     CodeType = "INGES_RE"
	CodeBackPwd CodeType = "INGES_MACK"
	CodeSay     CodeType = "INGES_SAY"
)

var AllCodeTypes = []CodeType{CodeLogin, CodeReg, CodeBackPwd, CodeSay}

func JoinCodeTypes(ts []CodeType) string {
	var parts []string
	for _, t := range ts {
		parts = append(parts, string(t))
	}
	return strings.Join(parts, "|")
}

// FetchCodeEnabledMap 解析 getsetimag.in 多类型 | 分隔
func (c *Client) FetchCodeEnabledMap() map[CodeType]bool {
	types := AllCodeTypes
	r := c.GetCodeEnabled(JoinCodeTypes(types))
	s, ok := r.Data.(string)
	if !ok {
		return nil
	}
	parts := strings.Split(s, "|")
	out := make(map[CodeType]bool)
	for i, t := range types {
		if i < len(parts) {
			out[t] = strings.EqualFold(strings.TrimSpace(parts[i]), "checked")
		}
	}
	return out
}

func CodeEnabled(m map[CodeType]bool, t CodeType) bool {
	if m == nil {
		return true
	}
	v, ok := m[t]
	if !ok {
		return true
	}
	return v
}

// ParseUserInfoValue 解析 "field=value" 或纯字符串
func ParseUserInfoValue(s string) string {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "=") {
		parts := strings.SplitN(s, "=", 2)
		return strings.TrimSpace(parts[len(parts)-1])
	}
	return s
}
