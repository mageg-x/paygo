package alipay

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"paygo/src/config"
	"paygo/src/model"
	"paygo/src/plugin"

	"github.com/gin-gonic/gin"
)

// 支付宝插件
type AlipayPlugin struct {
	plugin.BasePlugin
}

// 配置结构
type AlipayConfig struct {
	AppID      string `json:"appid"`
	AppKey     string `json:"appkey"`    // 支付宝公钥
	AppSecret  string `json:"appsecret"` // 应用私钥
	GatewayURL string `json:"appurl"`    // 自定义网关
}

var gatewayPrecheckCooldown sync.Map // key: appid|gateway, value: time.Time

func gatewayPrecheckKey(cfg *AlipayConfig) string {
	return strings.TrimSpace(cfg.AppID) + "|" + strings.TrimSpace(cfg.GatewayURL)
}

func shouldSkipGatewayPrecheck(cfg *AlipayConfig) bool {
	key := gatewayPrecheckKey(cfg)
	v, ok := gatewayPrecheckCooldown.Load(key)
	if !ok {
		return false
	}
	until, ok := v.(time.Time)
	if !ok {
		return false
	}
	if time.Now().Before(until) {
		return true
	}
	gatewayPrecheckCooldown.Delete(key)
	return false
}

func setGatewayPrecheckCooldown(cfg *AlipayConfig, reason string) {
	key := gatewayPrecheckKey(cfg)
	until := time.Now().Add(10 * time.Minute)
	gatewayPrecheckCooldown.Store(key, until)
	log.Printf("[alipay_gateway_precheck_skip] appid=%s, gateway=%s, cooldown_until=%s, reason=%s",
		cfg.AppID, strings.TrimSpace(cfg.GatewayURL), until.Format("2006-01-02 15:04:05"), reason)
}

func stringifyAny(v interface{}) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case fmt.Stringer:
		return strings.TrimSpace(t.String())
	default:
		if v == nil {
			return ""
		}
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func (p *AlipayPlugin) extractGatewayError(result map[string]interface{}) map[string]interface{} {
	if result == nil {
		return nil
	}
	if er, ok := result["error_response"].(map[string]interface{}); ok {
		return er
	}
	for k, v := range result {
		if !strings.HasSuffix(k, "_response") {
			continue
		}
		if m, ok := v.(map[string]interface{}); ok {
			code := stringifyAny(m["code"])
			if code != "" && code != "10000" {
				return m
			}
		}
	}
	return nil
}

func normalizeGatewaySubCode(subCode string) string {
	s := strings.ToLower(strings.TrimSpace(subCode))
	s = strings.TrimPrefix(s, "isv.")
	return s
}

func (p *AlipayPlugin) precheckGatewaySignatureConfig(cfg *AlipayConfig) error {
	if shouldSkipGatewayPrecheck(cfg) {
		return nil
	}

	precheckParams := map[string]interface{}{
		"app_id":    cfg.AppID,
		"method":    "alipay.test.request",
		"format":    "JSON",
		"charset":   "UTF-8",
		"sign_type": "RSA2",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"version":   "1.0",
	}

	sign, err := p.signParams(precheckParams, cfg.AppSecret)
	if err != nil {
		return fmt.Errorf("应用私钥签名失败: %v", err)
	}
	precheckParams["sign"] = sign

	formData := url.Values{}
	for k, v := range precheckParams {
		if v == nil {
			continue
		}
		switch vt := v.(type) {
		case string:
			formData.Set(k, vt)
		default:
			b, _ := json.Marshal(v)
			formData.Set(k, string(b))
		}
	}

	gatewayURL := attachGatewayCharset(p.getGatewayURL(cfg), "UTF-8")
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.PostForm(gatewayURL, formData)
	if err != nil {
		setGatewayPrecheckCooldown(cfg, "gateway unreachable")
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		setGatewayPrecheckCooldown(cfg, "read response failed")
		return nil
	}
	bodyText := strings.TrimSpace(string(body))
	if strings.HasPrefix(bodyText, "<") {
		setGatewayPrecheckCooldown(cfg, "html response")
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		setGatewayPrecheckCooldown(cfg, "json parse failed")
		return nil
	}

	errResp := p.extractGatewayError(result)
	if errResp == nil {
		return nil
	}

	code := stringifyAny(errResp["code"])
	subCode := stringifyAny(errResp["sub_code"])
	msg := stringifyAny(errResp["msg"])
	subMsg := stringifyAny(errResp["sub_msg"])
	normSubCode := normalizeGatewaySubCode(subCode)

	switch {
	case strings.Contains(normSubCode, "missing-signature-config"):
		return fmt.Errorf("支付宝应用未配置签名公钥（missing-signature-config），请在开放平台应用“接口加签方式”完成公钥配置")
	case strings.Contains(normSubCode, "invalid-app-id"):
		return fmt.Errorf("无效AppID（invalid-app-id），请检查appid与网关环境（正式/沙箱）是否一致")
	case strings.Contains(normSubCode, "invalid-signature"):
		return fmt.Errorf("支付宝验签失败（invalid-signature），请检查应用私钥与开放平台应用公钥是否匹配")
	}

	// 对未知业务错误不拦截下单，避免因为网关临时策略影响正常支付链路
	log.Printf("[alipay_gateway_precheck_warn] appid=%s, code=%s, sub_code=%s, msg=%s, sub_msg=%s", cfg.AppID, code, subCode, msg, subMsg)
	return nil
}

func mergeConfigJSON(baseRaw, overrideRaw string) (map[string]interface{}, error) {
	merged := make(map[string]interface{})

	mergeOne := func(raw string, isOverride bool) error {
		raw = strings.TrimSpace(raw)
		if raw == "" || raw == "{}" {
			return nil
		}
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			return err
		}
		for k, v := range m {
			if isOverride {
				if v == nil {
					continue
				}
				if s, ok := v.(string); ok && strings.TrimSpace(s) == "" {
					continue
				}
			}
			merged[k] = v
		}
		return nil
	}

	if err := mergeOne(baseRaw, false); err != nil {
		return nil, err
	}
	if err := mergeOne(overrideRaw, true); err != nil {
		return nil, err
	}
	return merged, nil
}

func New() plugin.Plugin {
	return &AlipayPlugin{}
}

func init() {
	plugin.Register("alipay", New)
}

func (p *AlipayPlugin) GetInfo() plugin.PluginInfo {
	return plugin.PluginInfo{
		Name:       "alipay",
		Showname:   "支付宝官方支付",
		Author:     "支付宝",
		Link:       "https://www.alipay.com/",
		Types:      []string{"alipay"},
		Transtypes: []string{"alipay", "bank"},
		Inputs: map[string]plugin.InputConfig{
			"appid":     {Name: "应用APPID", Type: "input"},
			"appkey":    {Name: "支付宝公钥", Type: "textarea"},
			"appsecret": {Name: "应用私钥", Type: "textarea"},
			"appurl":    {Name: "网关地址", Type: "input", Note: "留空使用默认网关"},
		},
		Select: map[string]string{
			"1": "电脑网站支付",
			"2": "手机网站支付",
			"3": "当面付",
			"4": "订单码支付",
			"5": "APP支付",
			"6": "JSAPI支付",
		},
		Note: `<p>支付宝官方支付接口，支持多种支付方式</p>
<h4 class="mt-3 font-medium">配置示例：</h4>
<pre class="bg-gray-50 p-2 rounded text-xs mt-1 overflow-x-auto">
{
  "appid": "2021001234567890",
  "appkey": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ...\n-----END PUBLIC KEY-----",
  "appsecret": "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA0Z3VS5...\n-----END RSA PRIVATE KEY-----",
  "appurl": ""
}
</pre>
<p class="text-xs text-gray-500 mt-2">* appkey: 支付宝公钥（RSA2）<br>* appsecret: 应用私钥（RSA2）<br>* appurl: 留空使用默认网关 https://openapi.alipay.com/gateway.do</p>`,
	}
}

// 获取配置
func (p *AlipayPlugin) getConfig(channel model.Channel) (*AlipayConfig, error) {
	cfg := &AlipayConfig{}

	var dbPlugin model.Plugin
	pluginConfig := ""
	result := config.DB.Where("name = ?", channel.Plugin).Limit(1).Find(&dbPlugin)
	if result.Error != nil {
		log.Printf("[alipay_get_config_failed] channel_id=%d, reason=query plugin config failed, error=%s", channel.ID, result.Error.Error())
		return nil, result.Error
	}
	if result.RowsAffected > 0 {
		pluginConfig = dbPlugin.Config
	}

	merged, err := mergeConfigJSON(pluginConfig, channel.Config)
	if err != nil {
		log.Printf("[alipay_get_config_failed] channel_id=%d, reason=merge config failed, error=%s", channel.ID, err.Error())
		return nil, err
	}
	if len(merged) == 0 {
		return cfg, nil
	}

	b, _ := json.Marshal(merged)
	if err := json.Unmarshal(b, cfg); err != nil {
		log.Printf("[alipay_get_config_failed] channel_id=%d, reason=parse merged config failed, error=%s", channel.ID, err.Error())
		return nil, err
	}

	return cfg, nil
}

// 获取网关地址
func (p *AlipayPlugin) getGatewayURL(cfg *AlipayConfig) string {
	url := strings.TrimSpace(cfg.GatewayURL)
	url = strings.TrimRight(url, ",;，；")
	if url != "" {
		return url
	}
	return "https://openapi.alipay.com/gateway.do"
}

func attachGatewayCharset(gatewayURL, charset string) string {
	gatewayURL = strings.TrimSpace(gatewayURL)
	if gatewayURL == "" {
		return gatewayURL
	}
	if charset == "" {
		charset = "UTF-8"
	}
	parsed, err := url.Parse(gatewayURL)
	if err != nil {
		return gatewayURL
	}
	q := parsed.Query()
	if q.Get("charset") == "" {
		q.Set("charset", charset)
	}
	parsed.RawQuery = q.Encode()
	return parsed.String()
}

// 提交支付
func (p *AlipayPlugin) Submit(params map[string]interface{}) (plugin.SubmitResult, error) {
	method := params["method"].(string)
	channel := params["channel"].(model.Channel)
	tradeNo := params["trade_no"].(string)

	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[alipay_submit_failed] trade_no=%s, channel_id=%d, reason=get config failed, error=%s", tradeNo, channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	if err := p.precheckGatewaySignatureConfig(cfg); err != nil {
		log.Printf("[alipay_submit_failed] trade_no=%s, channel_id=%d, reason=gateway precheck failed, error=%s", tradeNo, channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	switch method {
	case "1", "web", "jump":
		return p.submitWeb(params, channel)
	case "2", "wap":
		return p.submitWap(params, channel)
	case "3", "scan":
		return p.submitScan(params, channel)
	case "4":
		return p.submitOrderCode(params, channel)
	case "5", "app":
		return p.submitApp(params, channel)
	case "6", "jsapi":
		return p.submitJSAPI(params, channel)
	default:
		log.Printf("[alipay_submit] trade_no=%s, method=%s, reason=unknown method, using web", tradeNo, method)
		return p.submitWeb(params, channel)
	}
}

// 电脑网站支付
func (p *AlipayPlugin) submitWeb(params map[string]interface{}, channel model.Channel) (plugin.SubmitResult, error) {
	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[alipay_submit_web_failed] channel_id=%d, reason=%s", channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	returnURL := params["return_url"].(string)

	bizContent := map[string]interface{}{
		"out_trade_no": tradeNo,
		"product_code": "FAST_INSTANT_TRADE_PAY",
		"total_amount": strconv.FormatFloat(money, 'f', 2, 64),
		"subject":      name,
	}

	p2 := map[string]interface{}{
		"app_id":      cfg.AppID,
		"method":      "alipay.trade.page.pay",
		"format":      "JSON",
		"charset":     "UTF-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  notifyURL,
		"return_url":  returnURL,
		"biz_content": bizContent,
	}

	sign, err := p.signParams(p2, cfg.AppSecret)
	if err != nil {
		log.Printf("[alipay_submit_web_failed] trade_no=%s, reason=sign failed, error=%s", params["trade_no"], err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}
	p2["sign"] = sign

	// 构建表单
	html := p.buildForm(p2, p.getGatewayURL(cfg))

	log.Printf("[alipay_submit_web_success] trade_no=%s", params["trade_no"])
	return plugin.SubmitResult{
		Type: "html",
		Data: html,
	}, nil
}

// 扫码支付
func (p *AlipayPlugin) submitScan(params map[string]interface{}, channel model.Channel) (plugin.SubmitResult, error) {
	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[alipay_submit_scan_failed] channel_id=%d, reason=%s", channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)

	bizContent := map[string]interface{}{
		"out_trade_no": tradeNo,
		"product_code": "FACE_TO_FACE_PAYMENT",
		"total_amount": strconv.FormatFloat(money, 'f', 2, 64),
		"subject":      name,
	}

	p2 := map[string]interface{}{
		"app_id":      cfg.AppID,
		"method":      "alipay.trade.pay",
		"format":      "JSON",
		"charset":     "UTF-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  notifyURL,
		"biz_content": bizContent,
	}

	// 添加 auth_code（扫码枪扫的码）
	if authCode, ok := params["auth_code"].(string); ok && authCode != "" {
		bizContent["auth_code"] = authCode
	}

	sign, err := p.signParams(p2, cfg.AppSecret)
	if err != nil {
		log.Printf("[alipay_submit_scan_failed] trade_no=%s, reason=sign failed, error=%s", params["trade_no"], err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}
	p2["sign"] = sign

	resp, err := p.httpPost(p.getGatewayURL(cfg), p2)
	if err != nil {
		log.Printf("[alipay_submit_scan_failed] trade_no=%s, reason=http post failed, error=%s", params["trade_no"], err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(resp), &result)

	response := result["alipay_trade_pay_response"].(map[string]interface{})
	if response["code"] != "10000" {
		log.Printf("[alipay_submit_scan_failed] trade_no=%s, reason=alipay error, code=%s, msg=%s", params["trade_no"], response["code"], response["msg"])
		return plugin.SubmitResult{Msg: fmt.Sprintf("%s:%s", response["code"], response["msg"])}, nil
	}

	// 当面付返回二维码
	if qrCode, ok := response["qr_code"].(string); ok && qrCode != "" {
		return plugin.SubmitResult{
			Type: "qrcode",
			URL:  qrCode,
		}, nil
	}

	return plugin.SubmitResult{
		Type: "scan",
		Data: map[string]string{"trade_no": response["trade_no"].(string)},
	}, nil
}

// JSAPI支付
func (p *AlipayPlugin) submitJSAPI(params map[string]interface{}, channel model.Channel) (plugin.SubmitResult, error) {
	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[alipay_submit_jsapi_failed] channel_id=%d, reason=%s", channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	openid := params["openid"].(string)

	bizContent := map[string]interface{}{
		"out_trade_no": tradeNo,
		"product_code": "JSAPI_PAY",
		"total_amount": strconv.FormatFloat(money, 'f', 2, 64),
		"subject":      name,
		"buyer_id":     openid,
	}

	p2 := map[string]interface{}{
		"app_id":      cfg.AppID,
		"method":      "alipay.trade.create",
		"format":      "JSON",
		"charset":     "UTF-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  notifyURL,
		"biz_content": bizContent,
	}

	sign, err := p.signParams(p2, cfg.AppSecret)
	if err != nil {
		log.Printf("[alipay_submit_jsapi_failed] trade_no=%s, reason=sign failed, error=%s", params["trade_no"], err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}
	p2["sign"] = sign

	resp, err := p.httpPost(p.getGatewayURL(cfg), p2)
	if err != nil {
		log.Printf("[alipay_submit_jsapi_failed] trade_no=%s, reason=http post failed, error=%s", params["trade_no"], err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(resp), &result)

	response := result["alipay_trade_create_response"].(map[string]interface{})
	if response["code"] != "10000" {
		log.Printf("[alipay_submit_jsapi_failed] trade_no=%s, reason=alipay error, code=%s, msg=%s", params["trade_no"], response["code"], response["msg"])
		return plugin.SubmitResult{Msg: fmt.Sprintf("%s:%s", response["code"], response["msg"])}, nil
	}

	tradeNo2 := response["trade_no"].(string)

	// 返回 JSAPI 调起参数
	jsApiParams := map[string]interface{}{
		"appId":     cfg.AppID,
		"timestamp": strconv.FormatInt(time.Now().Unix(), 10),
		"nonceStr":  p.generateNonceStr(32),
		"package":   fmt.Sprintf("prepay_id=%s", tradeNo2),
		"signType":  "RSA2",
	}

	paySign, _ := p.signParams(jsApiParams, cfg.AppSecret)
	jsApiParams["paySign"] = paySign

	return plugin.SubmitResult{
		Type: "jsapi",
		Data: jsApiParams,
	}, nil
}

// APP支付
func (p *AlipayPlugin) submitApp(params map[string]interface{}, channel model.Channel) (plugin.SubmitResult, error) {
	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[alipay_submit_app_failed] channel_id=%d, reason=%s", channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)

	bizContent := map[string]interface{}{
		"out_trade_no": tradeNo,
		"product_code": "QUICK_MSECURITY_PAY",
		"total_amount": strconv.FormatFloat(money, 'f', 2, 64),
		"subject":      name,
	}

	p2 := map[string]interface{}{
		"app_id":      cfg.AppID,
		"method":      "alipay.trade.app.pay",
		"format":      "JSON",
		"charset":     "UTF-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  notifyURL,
		"biz_content": bizContent,
	}

	sign, err := p.signParams(p2, cfg.AppSecret)
	if err != nil {
		log.Printf("[alipay_submit_app_failed] trade_no=%s, reason=sign failed, error=%s", params["trade_no"], err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}
	p2["sign"] = sign

	// 返回 orderString
	bizContentStr, _ := json.Marshal(bizContent)
	orderString := fmt.Sprintf("app_id=%s&biz_content=%s&charset=UTF-8&format=JSON&method=alipay.trade.app.pay&notify_url=%s&sign_type=RSA2&timestamp=%s&version=1.0",
		cfg.AppID, url.QueryEscape(string(bizContentStr)), url.QueryEscape(notifyURL), url.QueryEscape(time.Now().Format("2006-01-02 15:04:05")))

	// 重新签名
	signStr, _ := p.signParams(map[string]interface{}{
		"app_id":      cfg.AppID,
		"biz_content": string(bizContentStr),
		"charset":     "UTF-8",
		"format":      "JSON",
		"method":      "alipay.trade.app.pay",
		"notify_url":  notifyURL,
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
	}, cfg.AppSecret)

	orderString += "&sign=" + url.QueryEscape(signStr)

	log.Printf("[alipay_submit_app_success] trade_no=%s", params["trade_no"])
	return plugin.SubmitResult{
		Type: "app",
		Data: map[string]string{"orderString": orderString},
	}, nil
}

// H5支付
func (p *AlipayPlugin) submitWap(params map[string]interface{}, channel model.Channel) (plugin.SubmitResult, error) {
	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[alipay_submit_wap_failed] channel_id=%d, reason=%s", channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	returnURL := params["return_url"].(string)

	bizContent := map[string]interface{}{
		"out_trade_no": tradeNo,
		"product_code": "QUICK_WAP_PAY",
		"total_amount": strconv.FormatFloat(money, 'f', 2, 64),
		"subject":      name,
	}

	p2 := map[string]interface{}{
		"app_id":      cfg.AppID,
		"method":      "alipay.trade.wap.pay",
		"format":      "JSON",
		"charset":     "UTF-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  notifyURL,
		"return_url":  returnURL,
		"biz_content": bizContent,
	}

	sign, err := p.signParams(p2, cfg.AppSecret)
	if err != nil {
		log.Printf("[alipay_submit_wap_failed] trade_no=%s, reason=sign failed, error=%s", params["trade_no"], err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}
	p2["sign"] = sign

	html := p.buildForm(p2, p.getGatewayURL(cfg))

	log.Printf("[alipay_submit_wap_success] trade_no=%s", params["trade_no"])
	return plugin.SubmitResult{
		Type: "html",
		Data: html,
	}, nil
}

// 订单码支付
func (p *AlipayPlugin) submitOrderCode(params map[string]interface{}, channel model.Channel) (plugin.SubmitResult, error) {
	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[alipay_submit_order_code_failed] channel_id=%d, reason=%s", channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)

	bizContent := map[string]interface{}{
		"out_trade_no": tradeNo,
		"product_code": "FACE_TO_FACE_PAYMENT",
		"total_amount": strconv.FormatFloat(money, 'f', 2, 64),
		"subject":      name,
	}

	p2 := map[string]interface{}{
		"app_id":      cfg.AppID,
		"method":      "alipay.trade.precreate",
		"format":      "JSON",
		"charset":     "UTF-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  notifyURL,
		"biz_content": bizContent,
	}

	sign, err := p.signParams(p2, cfg.AppSecret)
	if err != nil {
		log.Printf("[alipay_submit_order_code_failed] trade_no=%s, reason=sign failed, error=%s", params["trade_no"], err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}
	p2["sign"] = sign

	resp, err := p.httpPost(p.getGatewayURL(cfg), p2)
	if err != nil {
		log.Printf("[alipay_submit_order_code_failed] trade_no=%s, reason=http post failed, error=%s", params["trade_no"], err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		log.Printf("[alipay_submit_order_code_failed] trade_no=%s, reason=parse response failed, error=%s", params["trade_no"], err.Error())
		return plugin.SubmitResult{Msg: "解析响应失败"}, err
	}

	response, ok := result["alipay_trade_precreate_response"].(map[string]interface{})
	if !ok {
		log.Printf("[alipay_submit_order_code_failed] trade_no=%s, reason=invalid response format")
		return plugin.SubmitResult{Msg: "响应格式错误"}, nil
	}

	qrCode, _ := response["qr_code"].(string)
	if qrCode == "" {
		log.Printf("[alipay_submit_order_code_failed] trade_no=%s, reason=no qr_code returned")
		return plugin.SubmitResult{Msg: "获取二维码失败"}, nil
	}

	log.Printf("[alipay_submit_order_code_success] trade_no=%s", params["trade_no"])
	return plugin.SubmitResult{
		Type: "qrcode",
		URL:  qrCode,
	}, nil
}

// 移动端提交
func (p *AlipayPlugin) Mapi(params map[string]interface{}) (plugin.SubmitResult, error) {
	return p.submitWap(params, params["channel"].(model.Channel))
}

// 异步回调
func (p *AlipayPlugin) Notify(tradeNo string, c *gin.Context) (plugin.NotifyResult, error) {
	if err := c.Request.ParseForm(); err != nil {
		log.Printf("[alipay_notify_failed] trade_no=%s, reason=parse form failed, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: err.Error()}, err
	}

	// 获取订单
	var order model.Order
	if err := config.DB.Where("trade_no = ?", tradeNo).First(&order).Error; err != nil {
		log.Printf("[alipay_notify_failed] trade_no=%s, reason=order not found, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: "订单不存在"}, err
	}

	// 获取通道
	var channel model.Channel
	if err := config.DB.First(&channel, order.Channel).Error; err != nil {
		log.Printf("[alipay_notify_failed] trade_no=%s, reason=channel not found, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: "通道不存在"}, err
	}

	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[alipay_notify_failed] trade_no=%s, reason=get config failed, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: err.Error()}, err
	}

	// 过滤sign参数后验签
	paramsToVerify := make(map[string]string)
	for k, v := range c.Request.PostForm {
		if k == "sign" || k == "sign_type" {
			continue
		}
		paramsToVerify[k] = v[0]
	}

	sign := c.Request.PostForm.Get("sign")
	signType := c.Request.PostForm.Get("sign_type")

	if !p.verifySign(paramsToVerify, sign, signType, cfg.AppKey) {
		log.Printf("[alipay_notify_failed] trade_no=%s, reason=sign verification failed")
		return plugin.NotifyResult{Success: false, Message: "验签失败"}, nil
	}

	// 获取回调数据
	outTradeNo := c.Request.PostForm.Get("out_trade_no")
	tradeStatus := c.Request.PostForm.Get("trade_status")
	buyerID := c.Request.PostForm.Get("buyer_id")
	totalAmount := c.Request.PostForm.Get("total_amount")

	if outTradeNo != tradeNo {
		log.Printf("[alipay_notify_failed] trade_no=%s, out_trade_no=%s, reason=trade no mismatch")
		return plugin.NotifyResult{Success: false, Message: "订单号不匹配"}, nil
	}

	// 验证金额
	realMoney, _ := strconv.ParseFloat(totalAmount, 64)
	if order.Realmoney != realMoney {
		log.Printf("[alipay_notify_failed] trade_no=%s, expected=%.2f, got=%.2f, reason=amount mismatch")
		return plugin.NotifyResult{Success: false, Message: "金额不匹配"}, nil
	}

	if tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED" {
		log.Printf("[alipay_notify_success] trade_no=%s, api_trade_no=%s, amount=%.2f", tradeNo, c.Request.PostForm.Get("trade_no"), realMoney)
		return plugin.NotifyResult{
			Success:    true,
			TradeNo:    tradeNo,
			APITradeNo: c.Request.PostForm.Get("trade_no"),
			Amount:     realMoney,
			Buyer:      buyerID,
			Message:    "成功",
		}, nil
	}

	log.Printf("[alipay_notify_failed] trade_no=%s, status=%s, reason=trade not success")
	return plugin.NotifyResult{Success: false, Message: "交易未成功"}, nil
}

// 同步回调
func (p *AlipayPlugin) Return(tradeNo string, c *gin.Context) (plugin.ReturnResult, error) {
	return plugin.ReturnResult{
		Success: true,
		TradeNo: tradeNo,
		Message: "支付成功",
		URL:     "/user/order",
	}, nil
}

// 支付成功页面
func (p *AlipayPlugin) OK(tradeNo string) (string, error) {
	return "订单支付成功", nil
}

// 退款
func (p *AlipayPlugin) Refund(params map[string]interface{}) (plugin.RefundResult, error) {
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	channel := params["channel"].(model.Channel)

	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[alipay_refund_failed] trade_no=%s, money=%.2f, reason=get config failed, error=%s", tradeNo, money, err.Error())
		return plugin.RefundResult{Code: -1, ErrMsg: err.Error()}, err
	}

	order, err := p.getOrder(tradeNo)
	if err != nil {
		log.Printf("[alipay_refund_failed] trade_no=%s, money=%.2f, reason=order not found, error=%s", tradeNo, money, err.Error())
		return plugin.RefundResult{Code: -1, ErrMsg: "订单不存在"}, err
	}

	bizContent := map[string]interface{}{
		"trade_no":       order.ApiTradeNo,
		"refund_amount":  strconv.FormatFloat(money, 'f', 2, 64),
		"out_request_no": tradeNo,
		"refund_reason":  "用户请求退款",
	}

	p2 := map[string]interface{}{
		"app_id":      cfg.AppID,
		"method":      "alipay.trade.refund",
		"format":      "JSON",
		"charset":     "UTF-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": bizContent,
	}

	sign, err := p.signParams(p2, cfg.AppSecret)
	if err != nil {
		log.Printf("[alipay_refund_failed] trade_no=%s, money=%.2f, reason=sign failed, error=%s", tradeNo, money, err.Error())
		return plugin.RefundResult{Code: -1, ErrMsg: err.Error()}, err
	}
	p2["sign"] = sign

	resp, err := p.httpPost(p.getGatewayURL(cfg), p2)
	if err != nil {
		log.Printf("[alipay_refund_failed] trade_no=%s, money=%.2f, reason=http post failed, error=%s", tradeNo, money, err.Error())
		return plugin.RefundResult{Code: -1, ErrMsg: err.Error()}, err
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(resp), &result)

	response := result["alipay_trade_refund_response"].(map[string]interface{})
	if response["code"] != "10000" {
		log.Printf("[alipay_refund_failed] trade_no=%s, money=%.2f, reason=alipay error, code=%s, msg=%s", tradeNo, money, response["code"], response["msg"])
		return plugin.RefundResult{Code: -1, ErrMsg: fmt.Sprintf("%s:%s", response["code"], response["msg"])}, nil
	}

	return plugin.RefundResult{
		Code:    0,
		TradeNo: tradeNo,
		Fee:     money,
		Time:    time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// 转账
func (p *AlipayPlugin) Transfer(params map[string]interface{}) (plugin.TransferResult, error) {
	bizNo := params["biz_no"].(string)
	account := params["account"].(string)
	money := params["money"].(float64)
	channel := params["channel"].(model.Channel)

	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[alipay_transfer_failed] biz_no=%s, account=%s, money=%.2f, reason=get config failed, error=%s", bizNo, account, money, err.Error())
		return plugin.TransferResult{Code: -1, ErrMsg: err.Error()}, err
	}

	bizContent := map[string]interface{}{
		"out_biz_no":    bizNo,
		"payee_type":    "ALIPAY_LOGONID",
		"payee_account": account,
		"amount":        strconv.FormatFloat(money, 'f', 2, 64),
		"remark":        "商户转账",
	}

	p2 := map[string]interface{}{
		"app_id":      cfg.AppID,
		"method":      "alipay.fund.trans.toaccount.transfer",
		"format":      "JSON",
		"charset":     "UTF-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": bizContent,
	}

	sign, err := p.signParams(p2, cfg.AppSecret)
	if err != nil {
		log.Printf("[alipay_transfer_failed] biz_no=%s, account=%s, money=%.2f, reason=sign failed, error=%s", bizNo, account, money, err.Error())
		return plugin.TransferResult{Code: -1, ErrMsg: err.Error()}, err
	}
	p2["sign"] = sign

	resp, err := p.httpPost(p.getGatewayURL(cfg), p2)
	if err != nil {
		log.Printf("[alipay_transfer_failed] biz_no=%s, account=%s, money=%.2f, reason=http post failed, error=%s", bizNo, account, money, err.Error())
		return plugin.TransferResult{Code: -1, ErrMsg: err.Error()}, err
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(resp), &result)

	response := result["alipay_fund_trans_toaccount_transfer_response"].(map[string]interface{})
	if response["code"] != "10000" {
		log.Printf("[alipay_transfer_failed] biz_no=%s, account=%s, money=%.2f, reason=alipay error, code=%s, msg=%s", bizNo, account, money, response["code"], response["msg"])
		return plugin.TransferResult{Code: -1, ErrMsg: fmt.Sprintf("%s:%s", response["code"], response["msg"])}, nil
	}

	return plugin.TransferResult{
		Code:    0,
		OrderID: response["order_id"].(string),
		PayDate: response["pay_date"].(string),
	}, nil
}

// 转账查询
func (p *AlipayPlugin) TransferQuery(params map[string]interface{}) (plugin.TransferQueryResult, error) {
	bizNo := params["biz_no"].(string)
	channel := params["channel"].(model.Channel)

	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[alipay_transfer_query_failed] biz_no=%s, reason=get config failed, error=%s", bizNo, err.Error())
		return plugin.TransferQueryResult{Code: -1, ErrMsg: err.Error()}, err
	}

	bizContent := map[string]interface{}{
		"out_biz_no": bizNo,
	}

	p2 := map[string]interface{}{
		"app_id":      cfg.AppID,
		"method":      "alipay.fund.trans.order.query",
		"format":      "JSON",
		"charset":     "UTF-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": bizContent,
	}

	sign, err := p.signParams(p2, cfg.AppSecret)
	if err != nil {
		log.Printf("[alipay_transfer_query_failed] biz_no=%s, reason=sign failed, error=%s", bizNo, err.Error())
		return plugin.TransferQueryResult{Code: -1, ErrMsg: err.Error()}, err
	}
	p2["sign"] = sign

	resp, err := p.httpPost(p.getGatewayURL(cfg), p2)
	if err != nil {
		log.Printf("[alipay_transfer_query_failed] biz_no=%s, reason=http post failed, error=%s", bizNo, err.Error())
		return plugin.TransferQueryResult{Code: -1, ErrMsg: err.Error()}, err
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(resp), &result)

	response := result["alipay_fund_trans_order_query_response"].(map[string]interface{})
	if response["code"] != "10000" {
		log.Printf("[alipay_transfer_query_failed] biz_no=%s, reason=alipay error, code=%s, msg=%s", bizNo, response["code"], response["msg"])
		return plugin.TransferQueryResult{Code: -1, ErrMsg: fmt.Sprintf("%s:%s", response["code"], response["msg"])}, nil
	}

	status := 0
	switch response["status"].(string) {
	case "SUCCESS":
		status = 1
	case "FAIL":
		status = 2
	}

	return plugin.TransferQueryResult{
		Code:    0,
		Status:  status,
		Amount:  0,
		PayDate: response["pay_date"].(string),
	}, nil
}

// 订单查单（仅查询上游状态）
func (p *AlipayPlugin) QueryOrder(params map[string]interface{}) (map[string]interface{}, error) {
	tradeNo := stringifyAny(params["trade_no"])
	channel := params["channel"].(model.Channel)

	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[alipay_query_order_failed] trade_no=%s, channel_id=%d, reason=get config failed, error=%s", tradeNo, channel.ID, err.Error())
		return nil, err
	}

	bizContent := map[string]interface{}{
		"out_trade_no": tradeNo,
	}

	p2 := map[string]interface{}{
		"app_id":      cfg.AppID,
		"method":      "alipay.trade.query",
		"format":      "JSON",
		"charset":     "UTF-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": bizContent,
	}

	sign, err := p.signParams(p2, cfg.AppSecret)
	if err != nil {
		log.Printf("[alipay_query_order_failed] trade_no=%s, reason=sign failed, error=%s", tradeNo, err.Error())
		return nil, err
	}
	p2["sign"] = sign

	resp, err := p.httpPost(p.getGatewayURL(cfg), p2)
	if err != nil {
		log.Printf("[alipay_query_order_failed] trade_no=%s, reason=http post failed, error=%s", tradeNo, err.Error())
		return nil, err
	}

	bodyText := strings.TrimSpace(resp)
	if strings.HasPrefix(bodyText, "<") {
		log.Printf("[alipay_query_order_failed] trade_no=%s, reason=html response from gateway", tradeNo)
		return nil, fmt.Errorf("网关返回HTML页面，无法完成查单")
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		log.Printf("[alipay_query_order_failed] trade_no=%s, reason=parse response failed, error=%s", tradeNo, err.Error())
		return nil, err
	}

	response, ok := result["alipay_trade_query_response"].(map[string]interface{})
	if !ok {
		log.Printf("[alipay_query_order_failed] trade_no=%s, reason=missing alipay_trade_query_response", tradeNo)
		return nil, fmt.Errorf("支付宝响应格式异常")
	}

	code := stringifyAny(response["code"])
	msg := stringifyAny(response["msg"])
	subCode := stringifyAny(response["sub_code"])
	subMsg := stringifyAny(response["sub_msg"])

	if code != "10000" {
		normSubCode := normalizeGatewaySubCode(subCode)
		if code == "40004" && strings.Contains(normSubCode, "trade_not_exist") {
			log.Printf("[alipay_query_order_not_found] trade_no=%s", tradeNo)
			return map[string]interface{}{
				"exists": false,
				"paid":   false,
				"status": "NOT_FOUND",
			}, nil
		}
		errMsg := msg
		if subMsg != "" {
			errMsg = subMsg
		}
		log.Printf("[alipay_query_order_failed] trade_no=%s, code=%s, sub_code=%s, msg=%s, sub_msg=%s", tradeNo, code, subCode, msg, subMsg)
		return nil, fmt.Errorf("支付宝查单失败: %s", errMsg)
	}

	tradeStatus := stringifyAny(response["trade_status"])
	paid := tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED"
	amount, _ := strconv.ParseFloat(stringifyAny(response["total_amount"]), 64)

	resultMap := map[string]interface{}{
		"exists":       true,
		"paid":         paid,
		"status":       tradeStatus,
		"api_trade_no": stringifyAny(response["trade_no"]),
		"amount":       amount,
		"buyer":        stringifyAny(response["buyer_logon_id"]),
		"pay_time":     stringifyAny(response["send_pay_date"]),
	}

	log.Printf("[alipay_query_order_success] trade_no=%s, status=%s, paid=%v", tradeNo, tradeStatus, paid)
	return resultMap, nil
}

// ==================== 辅助方法 ====================

func (p *AlipayPlugin) getOrder(tradeNo string) (*model.Order, error) {
	var order model.Order
	if err := config.DB.Where("trade_no = ?", tradeNo).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// 签名参数
func (p *AlipayPlugin) signParams(params map[string]interface{}, privateKey string) (string, error) {
	// 构建签名字符串
	keys := make([]string, 0)
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var signData string
	for _, k := range keys {
		v := params[k]
		if v == nil || v == "" {
			continue
		}
		var val string
		switch vt := v.(type) {
		case string:
			val = vt
		default:
			b, _ := json.Marshal(v)
			val = string(b)
		}
		signData += k + "=" + val + "&"
	}
	signData = strings.TrimSuffix(signData, "&")

	// 解析私钥（支持带或不带 PEM 头尾）
	privateKeyParsed, err := parsePrivateKey(privateKey)
	if err != nil {
		log.Printf("[alipay_sign_failed] reason=parse private key failed, error=%s", err.Error())
		return "", err
	}

	h := sha256.New()
	h.Write([]byte(signData))

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKeyParsed, crypto.SHA256, h.Sum(nil))
	if err != nil {
		log.Printf("[alipay_sign_failed] reason=rsa sign failed, error=%s", err.Error())
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// 验签
func (p *AlipayPlugin) verifySign(params map[string]string, sign, signType, publicKey string) bool {
	keys := make([]string, 0)
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var signData string
	for _, k := range keys {
		signData += k + "=" + params[k] + "&"
	}
	signData = strings.TrimSuffix(signData, "&")

	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		log.Printf("[alipay_verify_sign_failed] reason=base64 decode sign failed, error=%s", err.Error())
		return false
	}

	pub, err := parsePublicKey(publicKey)
	if err != nil {
		log.Printf("[alipay_verify_sign_failed] reason=parse public key failed, error=%s", err.Error())
		return false
	}

	h := sha256.New()
	h.Write([]byte(signData))

	var err2 error
	switch signType {
	case "RSA":
		err2 = rsa.VerifyPKCS1v15(pub.(*rsa.PublicKey), crypto.SHA256, h.Sum(nil), signBytes)
	default: // RSA2
		err2 = rsa.VerifyPKCS1v15(pub.(*rsa.PublicKey), crypto.SHA256, h.Sum(nil), signBytes)
	}

	if err2 != nil {
		log.Printf("[alipay_verify_sign_failed] reason=rsa verify failed, error=%s", err2.Error())
	}

	return err2 == nil
}

// 发送 HTTP POST 请求
func (p *AlipayPlugin) httpPost(gatewayURL string, params map[string]interface{}) (string, error) {
	// 构建 form 数据
	formData := url.Values{}
	for k, v := range params {
		if v == nil {
			continue
		}
		switch vt := v.(type) {
		case string:
			formData.Set(k, vt)
		default:
			b, _ := json.Marshal(v)
			formData.Set(k, string(b))
		}
	}

	reqCharset := stringifyAny(params["charset"])
	resp, err := http.PostForm(attachGatewayCharset(gatewayURL, reqCharset), formData)
	if err != nil {
		log.Printf("[alipay_http_post_failed] gateway=%s, reason=http post failed, error=%s", gatewayURL, err.Error())
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[alipay_http_post_failed] gateway=%s, reason=non-200 status, status=%d", gatewayURL, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[alipay_http_post_failed] gateway=%s, reason=read body failed, error=%s", gatewayURL, err.Error())
		return "", err
	}

	return string(body), nil
}

// 构建 HTML 表单
func (p *AlipayPlugin) buildForm(params map[string]interface{}, gatewayURL string) string {
	reqCharset := stringifyAny(params["charset"])
	if reqCharset == "" {
		reqCharset = "UTF-8"
	}

	var buf bytes.Buffer
	buf.WriteString("<!DOCTYPE html><html><head><meta charset=\"UTF-8\"><meta http-equiv=\"Content-Type\" content=\"text/html; charset=UTF-8\"></head><body>")
	buf.WriteString(fmt.Sprintf("<form id=\"payform\" action=\"%s\" method=\"POST\" accept-charset=\"UTF-8\">", html.EscapeString(attachGatewayCharset(gatewayURL, reqCharset))))

	for k, v := range params {
		if v == nil {
			continue
		}
		var val string
		switch vt := v.(type) {
		case string:
			val = vt
		default:
			b, _ := json.Marshal(v)
			val = string(b)
		}
		buf.WriteString(fmt.Sprintf(
			"<input type=\"hidden\" name=\"%s\" value=\"%s\">",
			html.EscapeString(k),
			html.EscapeString(val),
		))
	}

	buf.WriteString("</form><script>var f=document.getElementById('payform');if(f){f.acceptCharset='UTF-8';f.submit();}</script></body></html>")
	return buf.String()
}

// 生成随机字符串
func (p *AlipayPlugin) generateNonceStr(length int) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(result)
}

// 解析私钥（支持带或不带 PEM 头尾）
func normalizeKeyMaterial(key string) string {
	key = strings.TrimSpace(key)
	key = strings.Trim(key, `"'`)
	key = strings.ReplaceAll(key, "\r\n", "\n")
	key = strings.ReplaceAll(key, "\r", "\n")
	key = strings.ReplaceAll(key, `\n`, "\n")
	key = strings.ReplaceAll(key, `\t`, "\t")
	return strings.TrimSpace(key)
}

func compactBase64(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func classifyPrivateKeyParseError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errMsg, "sequence truncated"), strings.Contains(errMsg, "asn1"):
		return fmt.Errorf("应用私钥内容不完整或格式错误，请检查是否完整粘贴（含 BEGIN/END）")
	case strings.Contains(errMsg, "not rsa"):
		return fmt.Errorf("应用私钥不是 RSA 格式，请使用支付宝应用私钥（RSA2）")
	default:
		return fmt.Errorf("应用私钥格式错误：%v", err)
	}
}

func parsePrivateKey(key string) (*rsa.PrivateKey, error) {
	key = normalizeKeyMaterial(key)
	if key == "" {
		return nil, fmt.Errorf("应用私钥不能为空")
	}

	if strings.Contains(key, "BEGIN ENCRYPTED PRIVATE KEY") {
		return nil, fmt.Errorf("不支持加密私钥，请使用未加密的 RSA 私钥")
	}

	// 如果包含 PEM 头尾，使用 pem.Decode
	if strings.Contains(key, "BEGIN") {
		block, _ := pem.Decode([]byte(key))
		if block == nil {
			log.Printf("[alipay_parse_private_key_failed] reason=pem decode failed")
			return nil, fmt.Errorf("应用私钥 PEM 格式错误，请粘贴完整的私钥内容")
		}
		pk, err := parseRSAPrivateKeyDER(block.Bytes)
		if err != nil {
			log.Printf("[alipay_parse_private_key_failed] reason=parse private key failed, error=%s", err.Error())
			return nil, classifyPrivateKeyParseError(err)
		}
		return pk, nil
	}

	// 否则直接 Base64 解码
	key = compactBase64(key)
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		log.Printf("[alipay_parse_private_key_failed] reason=base64 decode failed, error=%s", err.Error())
		return nil, fmt.Errorf("应用私钥不是有效的 Base64/PEM 内容")
	}

	pk, err := parseRSAPrivateKeyDER(keyBytes)
	if err != nil {
		log.Printf("[alipay_parse_private_key_failed] reason=parse private key failed, error=%s", err.Error())
		return nil, classifyPrivateKeyParseError(err)
	}

	return pk, nil
}

// 解析 RSA 私钥 DER（兼容 PKCS#1 与 PKCS#8）
func parseRSAPrivateKeyDER(der []byte) (*rsa.PrivateKey, error) {
	if pk, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return pk, nil
	}

	anyKey, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := anyKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}

	return rsaKey, nil
}

// 解析公钥（支持带或不带 PEM 头尾）
func parsePublicKey(key string) (crypto.PublicKey, error) {
	key = strings.TrimSpace(key)

	// 如果包含 PEM 头尾，使用 pem.Decode
	if strings.Contains(key, "BEGIN") {
		block, _ := pem.Decode([]byte(key))
		if block == nil {
			log.Printf("[alipay_parse_public_key_failed] reason=pem decode failed")
			return nil, fmt.Errorf("failed to decode public key")
		}
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			log.Printf("[alipay_parse_public_key_failed] reason=parse pkix failed, error=%s", err.Error())
			return nil, fmt.Errorf("failed to parse public key: %v", err)
		}
		return pub, nil
	}

	// 否则直接 Base64 解码
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		log.Printf("[alipay_parse_public_key_failed] reason=base64 decode failed, error=%s", err.Error())
		return nil, fmt.Errorf("failed to decode public key: %v", err)
	}

	pub, err := x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		log.Printf("[alipay_parse_public_key_failed] reason=parse pkix failed, error=%s", err.Error())
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	return pub, nil
}

// 测试配置
func (p *AlipayPlugin) TestConfig(config string) (bool, string) {
	var cfg AlipayConfig
	if err := json.Unmarshal([]byte(config), &cfg); err != nil {
		log.Printf("[alipay_test_config_failed] reason=parse config failed, error=%s", err.Error())
		return false, "配置格式错误: " + err.Error()
	}

	if cfg.AppID == "" {
		log.Printf("[alipay_test_config_failed] reason=missing appid")
		return false, "缺少应用APPID"
	}

	if cfg.AppSecret == "" {
		log.Printf("[alipay_test_config_failed] reason=missing appsecret")
		return false, "缺少应用私钥"
	}

	if cfg.AppKey == "" {
		log.Printf("[alipay_test_config_failed] reason=missing appkey")
		return false, "缺少支付宝公钥"
	}

	// 私钥格式预检
	if _, err := parsePrivateKey(cfg.AppSecret); err != nil {
		log.Printf("[alipay_test_config_failed] appid=%s, reason=invalid private key, error=%s", cfg.AppID, err.Error())
		return false, "应用私钥格式错误: " + err.Error()
	}

	// 公钥格式预检
	if _, err := parsePublicKey(cfg.AppKey); err != nil {
		log.Printf("[alipay_test_config_failed] appid=%s, reason=invalid public key, error=%s", cfg.AppID, err.Error())
		return false, "支付宝公钥格式错误: " + err.Error()
	}

	// 测试签名（使用 alipay.system.oauth.token 作为测试接口）
	testParams := map[string]interface{}{
		"app_id":     cfg.AppID,
		"method":     "alipay.system.oauth.token",
		"format":     "JSON",
		"charset":    "UTF-8",
		"sign_type":  "RSA2",
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"version":    "1.0",
		"grant_type": "authorization_code",
		"code":       "test_code",
	}

	sign, err := p.signParams(testParams, cfg.AppSecret)
	if err != nil {
		log.Printf("[alipay_test_config_failed] appid=%s, reason=sign failed, error=%s", cfg.AppID, err.Error())
		return false, "应用私钥签名失败: " + err.Error()
	}

	// 测试调用
	gatewayURL := attachGatewayCharset(strings.TrimSpace(p.getGatewayURL(&cfg)), "UTF-8")
	formData := url.Values{}
	for k, v := range testParams {
		if v == nil {
			continue
		}
		switch vt := v.(type) {
		case string:
			formData.Set(k, vt)
		default:
			b, _ := json.Marshal(v)
			formData.Set(k, string(b))
		}
	}
	formData.Set("sign", sign)

	resp, err := http.PostForm(gatewayURL, formData)
	if err != nil {
		log.Printf("[alipay_test_config_failed] appid=%s, gateway=%s, reason=http post failed, error=%s", cfg.AppID, gatewayURL, err.Error())
		return true, "密钥校验通过；网关连通性检查失败: " + err.Error()
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[alipay_test_config_failed] appid=%s, reason=read body failed, error=%s", cfg.AppID, err.Error())
		return false, "读取响应失败: " + err.Error()
	}

	bodyText := strings.TrimSpace(string(body))
	if strings.HasPrefix(bodyText, "<") {
		log.Printf("[alipay_test_config_failed] appid=%s, gateway=%s, reason=html response, status=%d, body=%s", cfg.AppID, gatewayURL, resp.StatusCode, bodyText)
		return true, "密钥校验通过；网关返回HTML页面，请检查网络代理/防火墙或网关策略"
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[alipay_test_config_failed] appid=%s, reason=parse response failed, error=%s, body=%s", cfg.AppID, err.Error(), bodyText)
		return true, "密钥校验通过；网关响应无法解析: " + err.Error()
	}

	if resp.StatusCode != 200 {
		log.Printf("[alipay_test_config_failed] appid=%s, reason=non-200 status, status=%d, body=%s", cfg.AppID, resp.StatusCode, string(body))
		return true, fmt.Sprintf("密钥校验通过；网关HTTP状态异常: %d", resp.StatusCode)
	}

	// 检查返回结果
	if methodResp, ok := result["alipay_system_oauth_token_response"].(map[string]interface{}); ok {
		if code, ok := methodResp["code"].(string); ok {
			if code == "10000" {
				log.Printf("[alipay_test_config_success] appid=%s", cfg.AppID)
				return true, "配置正确，密钥校验通过"
			}
			msg, _ := methodResp["msg"].(string)
			subCode, _ := methodResp["sub_code"].(string)
			subMsg, _ := methodResp["sub_msg"].(string)
			log.Printf("[alipay_test_config_failed] appid=%s, reason=alipay error, code=%s, sub_code=%s, msg=%s, sub_msg=%s", cfg.AppID, code, subCode, msg, subMsg)
			if subCode != "" {
				return false, fmt.Sprintf("支付宝返回错误: %s/%s - %s", code, subCode, subMsg)
			}
			return false, fmt.Sprintf("支付宝返回错误: %s - %s", code, msg)
		}
	}

	// 检查全局错误
	if code, ok := result["code"].(string); ok {
		if code != "10000" {
			msg := result["msg"]
			log.Printf("[alipay_test_config_failed] appid=%s, reason=global error, code=%s, msg=%s", cfg.AppID, code, msg)
			return false, fmt.Sprintf("测试失败: %s - %v", code, msg)
		}
	}

	log.Printf("[alipay_test_config_success] appid=%s", cfg.AppID)
	return true, "配置正确，测试成功"
}
