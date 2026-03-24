// Package bsphp：AppEn 加密通讯与卡模式 .ic、公共 .in 接口。
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

type ClientConfig struct {
	URL              string
	MutualKey        string
	ServerPrivateKey string
	ClientPublicKey  string
}

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

type Client struct {
	cfg         ClientConfig
	BsPhpSeSsL  string
	httpClient  *http.Client
	dateFmt     string
	dateHashFmt string
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
		"api":         api,
		"BSphpSeSsL":  c.BsPhpSeSsL,
		"date":        now.Format(c.dateHashFmt),
		"md5":         "",
		"mutualkey":   c.cfg.MutualKey,
		"appsafecode": appsafecode,
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
		return nil, fmt.Errorf("bad response pipe count")
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
		return nil, fmt.Errorf("bad sig parts")
	}
	respAesKey := sigParts[2]
	if len(respAesKey) < 16 {
		fmt.Println("[BSPHP] send 失败: bad aes key")
		return nil, fmt.Errorf("bad aes key")
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
		return nil, fmt.Errorf("no response object")
	}
	if ac, ok := respObj["appsafecode"].(string); ok && ac != appsafecode {
		respObj["data"] = "appsafecode 安全参数验证不通过"
	}
	return respObj, nil
}

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
	return APIResult{Data: r["data"], Code: intFromAny(r["code"])}
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
		return fmt.Errorf("empty BSphpSeSsL")
	}
	c.BsPhpSeSsL = s
	return nil
}

func (c *Client) Bootstrap() error {
	if !c.Connect() {
		return fmt.Errorf("连接失败")
	}
	if err := c.GetSeSsL(); err != nil {
		return fmt.Errorf("获取 BSphpSeSsL 失败")
	}
	return nil
}

func (c *Client) Logout() APIResult {
	r, _ := c.Send("cancellation.ic", nil)
	var res APIResult
	if r != nil {
		res.Data = r["data"]
		res.Code = intFromAny(r["code"])
	}
	c.BsPhpSeSsL = ""
	_ = c.GetSeSsL()
	return res
}

func (c *Client) AddCardFeatures(carid, key, maxoror string) APIResult {
	return c.apiResult("AddCardFeatures.key.ic", map[string]string{"carid": carid, "key": key, "maxoror": maxoror})
}

func (c *Client) CallRemote(datas string) APIResult {
	return c.apiResult("CallRemote.in", map[string]string{"datas": datas})
}

func (c *Client) GetMyData(keys string) APIResult {
	return c.apiResult("GetMyData.in", map[string]string{"keys": keys})
}

func (c *Client) SetAppRemarks(icid, icpwd, remarks string) APIResult {
	return c.apiResult("SetAppRemarks.ic", map[string]string{"icid": icid, "icpwd": icpwd, "remarks": remarks})
}

func (c *Client) SetMyData(keys, datas string) APIResult {
	return c.apiResult("SetMysData.in", map[string]string{"keys": keys, "datas": datas})
}

func (c *Client) AppBadPush(table string) APIResult {
	return c.apiResult("appbadpush.in", map[string]string{"table": table})
}

func (c *Client) GetAppCustom(info, getType, user, pwd, icid, icpwd string) APIResult {
	p := map[string]string{"info": info}
	if getType != "" {
		p["get_type"] = getType
	}
	if user != "" {
		p["user"] = user
	}
	if pwd != "" {
		p["pwd"] = pwd
	}
	if icid != "" {
		p["icid"] = icid
	}
	if icpwd != "" {
		p["icpwd"] = icpwd
	}
	return c.apiResult("appcustom.in", p)
}

func (c *Client) RechargeCard(icid, ka, pwd string) APIResult {
	return c.apiResult("chong.ic", map[string]string{"icid": icid, "ka": ka, "pwd": pwd})
}

func (c *Client) GetServerDate(dateFormatM string) APIResult {
	p := map[string]string{}
	if dateFormatM != "" {
		p["m"] = dateFormatM
	}
	return c.apiResult("date.in", p)
}

func (c *Client) GetData(key string) APIResult {
	return c.apiResult("getdata.ic", map[string]string{"key": key})
}

func (c *Client) GetDateIC() APIResult {
	return c.apiResult("getdate.ic", nil)
}

func (c *Client) GetCardInfo(icCarid, icPwd, info, typ string) APIResult {
	p := map[string]string{"ic_carid": icCarid, "ic_pwd": icPwd, "info": info}
	if typ != "" {
		p["type"] = typ
	}
	return c.apiResult("getinfo.ic", p)
}

func (c *Client) GetLoginInfo() APIResult {
	return c.apiResult("getlkinfo.ic", nil)
}

func (c *Client) GetGlobalInfo(info string) APIResult {
	p := map[string]string{}
	if info != "" {
		p["info"] = info
	}
	return c.apiResult("globalinfo.in", p)
}

func (c *Client) GetCaptchaImage() APIResult {
	return c.apiResult("imga.in", nil)
}

func (c *Client) PushAddMoney(user, ka string) APIResult {
	return c.apiResult("pushaddmoney.in", map[string]string{"user": user, "ka": ka})
}

func (c *Client) PushLog(user, log string) APIResult {
	return c.apiResult("pushlog.in", map[string]string{"user": user, "log": log})
}

func (c *Client) RemoteCancellation(icid, icpwd, typ, biaoji string) APIResult {
	p := map[string]string{"icid": icid, "icpwd": icpwd, "type": typ}
	if biaoji != "" {
		p["biaoji"] = biaoji
	}
	return c.apiResult("remotecancellation.ic", p)
}

func (c *Client) UnbindCard(icid, icpwd string) APIResult {
	return c.apiResult("setcarnot.ic", map[string]string{"icid": icid, "icpwd": icpwd})
}

func (c *Client) BindCard(key, icid, icpwd string) APIResult {
	return c.apiResult("setcaron.ic", map[string]string{"key": key, "icid": icid, "icpwd": icpwd})
}

func (c *Client) QueryCard(cardid string) APIResult {
	return c.apiResult("socard.in", map[string]string{"cardid": cardid})
}

func (c *Client) GetNotice() APIResult    { return c.apiResult("gg.in", nil) }
func (c *Client) GetVersion() APIResult   { return c.apiResult("v.in", nil) }
func (c *Client) GetSoftInfo() APIResult  { return c.apiResult("miao.in", nil) }
func (c *Client) GetPresetURL() APIResult { return c.apiResult("url.in", nil) }
func (c *Client) GetWebURL() APIResult    { return c.apiResult("weburl.in", nil) }

func (c *Client) GetCodeEnabled(typeStr string) APIResult {
	p := map[string]string{}
	if typeStr != "" {
		p["type"] = typeStr
	}
	return c.apiResult("getsetimag.in", p)
}

func (c *Client) GetLogicA() APIResult     { return c.apiResult("logica.in", nil) }
func (c *Client) GetLogicB() APIResult     { return c.apiResult("logicb.in", nil) }
func (c *Client) GetLogicInfoA() APIResult { return c.apiResult("logicinfoa.in", nil) }
func (c *Client) GetLogicInfoB() APIResult { return c.apiResult("logicinfob.in", nil) }
func (c *Client) Heartbeat() APIResult     { return c.apiResult("timeout.ic", nil) }

func (c *Client) Feedback(table, leix, qq, txt, img, user, pwd string) APIResult {
	p := map[string]string{"table": table, "leix": leix, "qq": qq, "txt": txt}
	if img != "" {
		p["img"] = img
	}
	if user != "" {
		p["user"] = user
	}
	if pwd != "" {
		p["pwd"] = pwd
	}
	return c.apiResult("liuyan.in", p)
}

func (c *Client) LoginIC(icid, icpwd, key, maxoror string) APIResult {
	k := key
	if k == "" {
		k = MachineCode()
	}
	m := maxoror
	if m == "" {
		m = k
	}
	r, err := c.Send("login.ic", map[string]string{
		"icid": icid, "icpwd": icpwd, "key": k, "maxoror": m,
	})
	if err != nil || r == nil {
		return APIResult{Data: "系统错误，登录失败！"}
	}
	if code := intFromAny(r["code"]); code != nil {
		if (*code == 1011 || *code == 9908 || *code == 1081) {
			if ssl, ok := r["SeSsL"].(string); ok && ssl != "" {
				c.BsPhpSeSsL = ssl
			}
		}
	}
	return APIResult{Data: r["data"], Code: intFromAny(r["code"])}
}

// LoginOK1081 卡密/机器码验证成功（返回信息含 1081 或 code==1081）
func LoginOK1081(r APIResult) bool {
	if r.Code != nil && *r.Code == 1081 {
		return true
	}
	return strings.Contains(r.Message(), "1081")
}
