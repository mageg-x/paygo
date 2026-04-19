package wxpay

import (
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopay/src/config"
	"gopay/src/model"
	"gopay/src/plugin"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/consts"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
)

// 微信支付插件
type WxpayPlugin struct {
	plugin.BasePlugin
}

// 配置结构
type WxpayConfig struct {
	AppID        string `json:"appid"`
	AppKey       string `json:"appkey"`        // APIv3密钥
	MchID        string `json:"appmchid"`      // 商户号
	SerialNo     string `json:"serial_no"`     // 商户证书序列号（可留空自动解析）
	PrivateKey   string `json:"private_key"`   // 商户私钥内容PEM
	MerchantCert string `json:"merchant_cert"` // 商户证书内容PEM（用于自动解析序列号）
	// 兼容旧字段（仅兼容读取）
	AppSecret        string `json:"appsecret"`
	CertPath         string `json:"cert_path"`
	KeyPath          string `json:"key_path"`
	PrivateKeyPath   string `json:"private_key_path"`
	MerchantCertPath string `json:"merchant_cert_path"`
	PlatformCert     string `json:"platform_cert"`
	PlatformCertPath string `json:"platform_cert_path"`
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

func configHasAnyKey(raw string, keys ...string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return false
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return false
	}
	for _, k := range keys {
		v, ok := m[k]
		if !ok || v == nil {
			continue
		}
		if s, ok := v.(string); ok {
			if strings.TrimSpace(s) == "" {
				continue
			}
		}
		return true
	}
	return false
}

func configStringValue(raw string, key string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

func configAnyNonEmpty(raw string, keys ...string) bool {
	for _, k := range keys {
		if configStringValue(raw, k) != "" {
			return true
		}
	}
	return false
}

func protectWxpayCredentialFields(merged map[string]interface{}, pluginRaw string) {
	if merged == nil {
		return
	}
	pluginRaw = strings.TrimSpace(pluginRaw)
	if pluginRaw == "" || pluginRaw == "{}" {
		return
	}
	var pm map[string]interface{}
	if err := json.Unmarshal([]byte(pluginRaw), &pm); err != nil {
		return
	}
	keys := []string{
		"appid", "appmchid", "appkey", "serial_no",
		"private_key", "merchant_cert", "platform_cert",
		"appsecret", "cert_path", "key_path", "private_key_path", "merchant_cert_path", "platform_cert_path",
	}
	for _, k := range keys {
		v, ok := pm[k]
		if !ok || v == nil {
			continue
		}
		if s, ok := v.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		merged[k] = v
	}
}

func New() plugin.Plugin {
	return &WxpayPlugin{}
}

func init() {
	plugin.Register("wxpay", New)
}

func (p *WxpayPlugin) GetInfo() plugin.PluginInfo {
	return plugin.PluginInfo{
		Name:       "wxpay",
		Showname:   "微信支付官方",
		Author:     "微信支付",
		Link:       "https://pay.weixin.qq.com/",
		Types:      []string{"wxpay"},
		Transtypes: []string{"wxpay", "bank"},
		Inputs: map[string]plugin.InputConfig{
			"appid":         {Name: "应用ID(AppID)", Type: "input"},
			"appmchid":      {Name: "商户号(MCHID)", Type: "input"},
			"appkey":        {Name: "APIv3密钥", Type: "input"},
			"private_key":   {Name: "商户私钥内容(PEM)", Type: "textarea", Note: "必填，直接粘贴"},
			"merchant_cert": {Name: "商户证书内容(PEM)", Type: "textarea", Note: "建议填写，可自动解析 serial_no"},
			"platform_cert": {Name: "微信支付平台证书内容(PEM)", Type: "textarea", Note: "必填，用于异步回调验签（支持粘贴多段证书PEM）"},
			"serial_no":     {Name: "商户证书序列号", Type: "input", Note: "可选；不填将从 merchant_cert 自动计算"},
		},
		Select: map[string]string{
			"1": "扫码支付",
			"2": "公众号支付",
			"3": "H5支付",
			"4": "小程序支付",
			"5": "APP支付",
		},
		Note: `配置示例：
	{
	  "appid": "wx1234567890abcdef",
	  "appmchid": "1234567890",
	  "appkey": "your_api_v3_key_here",
	  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
	  "merchant_cert": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
	  "serial_no": "",
	  "platform_cert": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
	}

	回调说明：
	1) 上游回调（微信支付 -> 平台）由系统自动使用平台回调地址，不依赖商户传入 notify_url。
	2) 商户回调（平台 -> 商户）使用 OpenAPI 下单参数 notify_url；为空则不回调商户。`,
	}
}

// 获取配置
func (p *WxpayPlugin) getConfig(channel model.Channel) (*WxpayConfig, error) {
	cfg := &WxpayConfig{}

	var dbPlugin model.Plugin
	pluginConfig := ""
	result := config.DB.Where("name = ?", channel.Plugin).Limit(1).Find(&dbPlugin)
	if result.Error != nil {
		log.Printf("[wxpay_get_config_failed] channel_id=%d, reason=query plugin config failed, error=%s", channel.ID, result.Error.Error())
		return nil, result.Error
	}
	if result.RowsAffected > 0 {
		pluginConfig = dbPlugin.Config
	}

	pluginHasCred := configHasAnyKey(pluginConfig, "appid", "appmchid", "serial_no")
	channelHasCred := configHasAnyKey(channel.Config, "appid", "appmchid", "serial_no")
	log.Printf("[wxpay_config_source] channel_id=%d, plugin_cfg_len=%d, channel_cfg_len=%d, plugin_has_cred=%t, channel_has_cred=%t", channel.ID, len(strings.TrimSpace(pluginConfig)), len(strings.TrimSpace(channel.Config)), pluginHasCred, channelHasCred)

	pluginAppID := configStringValue(pluginConfig, "appid")
	pluginMchID := configStringValue(pluginConfig, "appmchid")
	channelAppID := configStringValue(channel.Config, "appid")
	channelMchID := configStringValue(channel.Config, "appmchid")

	// 仅当插件配置缺失凭据时，才允许依赖通道凭据。
	if channelAppID != "" && pluginAppID != "" && channelAppID != pluginAppID {
		log.Printf("[wxpay_config_override_ignored] channel_id=%d, field=appid, plugin=%s, channel=%s", channel.ID, pluginAppID, channelAppID)
	}
	if channelMchID != "" && pluginMchID != "" && channelMchID != pluginMchID {
		log.Printf("[wxpay_config_override_ignored] channel_id=%d, field=appmchid, plugin=%s, channel=%s", channel.ID, pluginMchID, channelMchID)
	}
	if (channelAppID == "" && channelMchID != "") || (channelAppID != "" && channelMchID == "") {
		pluginHasAppAndMch := pluginAppID != "" && pluginMchID != ""
		if !pluginHasAppAndMch {
			log.Printf("[wxpay_get_config_failed] channel_id=%d, reason=channel config has partial credentials, appid_set=%t, mchid_set=%t", channel.ID, channelAppID != "", channelMchID != "")
			return nil, fmt.Errorf("通道配置错误：微信配置中的 appid 与 appmchid 必须同时配置或同时留空")
		}
		log.Printf("[wxpay_config_partial_ignored] channel_id=%d, reason=channel has partial appid/appmchid but plugin has complete credentials", channel.ID)
	}

	merged, err := mergeConfigJSON(pluginConfig, channel.Config)
	if err != nil {
		log.Printf("[wxpay_get_config_failed] channel_id=%d, reason=merge config failed, error=%s", channel.ID, err.Error())
		return nil, err
	}
	if configAnyNonEmpty(pluginConfig, "appid", "appmchid", "appkey", "private_key", "merchant_cert", "serial_no", "platform_cert") {
		protectWxpayCredentialFields(merged, pluginConfig)
	}
	if len(merged) == 0 {
		return cfg, nil
	}

	b, _ := json.Marshal(merged)
	if err := json.Unmarshal(b, cfg); err != nil {
		log.Printf("[wxpay_get_config_failed] channel_id=%d, reason=parse merged config failed, error=%s", channel.ID, err.Error())
		return nil, err
	}

	return cfg, nil
}

func wrapWxpayConfigMismatchError(err error, cfg *WxpayConfig) error {
	if err == nil || cfg == nil {
		return err
	}
	if strings.Contains(err.Error(), "APPID_MCHID_NOT_MATCH") {
		return fmt.Errorf("微信配置不匹配：appid=%s, mchid=%s，请确认该 appid 已绑定该商户号", strings.TrimSpace(cfg.AppID), strings.TrimSpace(cfg.MchID))
	}
	return err
}

// 获取微信支付客户端
func (p *WxpayPlugin) getClient(channel model.Channel) (*core.Client, error) {
	client, _, err := p.getClientAndConfig(channel)
	return client, err
}

func (p *WxpayPlugin) getClientAndConfig(channel model.Channel) (*core.Client, *WxpayConfig, error) {
	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[wxpay_get_client_failed] channel_id=%d, reason=get config failed, error=%s", channel.ID, err.Error())
		return nil, nil, err
	}

	// 加载私钥（优先读取内容字段，兼容历史路径字段）
	privateKey, err := p.resolvePrivateKey(cfg)
	if err != nil {
		log.Printf("[wxpay_get_client_failed] channel_id=%d, reason=load private key failed, error=%s", channel.ID, err.Error())
		return nil, nil, fmt.Errorf("加载私钥失败: %v", err)
	}

	// 解析私钥
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		log.Printf("[wxpay_get_client_failed] channel_id=%d, reason=invalid private key format")
		return nil, nil, fmt.Errorf("私钥格式错误")
	}
	rsaKey, err := parseRSAPrivateKey(block.Bytes)
	if err != nil {
		log.Printf("[wxpay_get_client_failed] channel_id=%d, reason=parse private key failed, error=%s", channel.ID, err.Error())
		return nil, nil, fmt.Errorf("解析私钥失败: %v", err)
	}

	serialNo, err := p.resolveMerchantSerialNo(cfg)
	if err != nil {
		log.Printf("[wxpay_get_client_failed] channel_id=%d, reason=resolve serial no failed, error=%s", channel.ID, err.Error())
		return nil, nil, fmt.Errorf("缺少商户证书序列号(serial_no)，请填写或提供merchant_cert自动解析")
	}

	log.Printf("[wxpay_config_effective] channel_id=%d, appid=%s, mchid=%s, serial_no=%s", channel.ID, strings.TrimSpace(cfg.AppID), strings.TrimSpace(cfg.MchID), serialNo)

	// 创建客户端，不验签（验签在回调时单独处理）
	opts := []core.ClientOption{
		option.WithoutValidator(),
		option.WithMerchantCredential(cfg.MchID, serialNo, rsaKey),
	}

	client, err := core.NewClient(context.Background(), opts...)
	if err != nil {
		log.Printf("[wxpay_get_client_failed] channel_id=%d, reason=create client failed, error=%s", channel.ID, err.Error())
		return nil, nil, fmt.Errorf("创建客户端失败: %v", err)
	}
	return client, cfg, nil
}

func (p *WxpayPlugin) resolvePrivateKey(cfg *WxpayConfig) (string, error) {
	candidates := []string{
		strings.TrimSpace(cfg.PrivateKey),
		strings.TrimSpace(cfg.AppSecret), // 兼容旧字段
		strings.TrimSpace(cfg.KeyPath),   // 兼容旧字段
		strings.TrimSpace(cfg.PrivateKeyPath),
	}
	for _, v := range candidates {
		if v == "" {
			continue
		}
		// 兼容历史路径：如果不是PEM内容，按路径读取
		if strings.Contains(v, "BEGIN") {
			return v, nil
		}
		b, err := os.ReadFile(v)
		if err == nil {
			return string(b), nil
		}
	}
	return "", fmt.Errorf("未配置商户私钥(private_key)")
}

func (p *WxpayPlugin) resolveMerchantCert(cfg *WxpayConfig) (string, error) {
	candidates := []string{
		strings.TrimSpace(cfg.MerchantCert),
		strings.TrimSpace(cfg.CertPath), // 兼容旧字段
		strings.TrimSpace(cfg.MerchantCertPath),
	}
	for _, v := range candidates {
		if v == "" {
			continue
		}
		if strings.Contains(v, "BEGIN CERTIFICATE") {
			return v, nil
		}
		b, err := os.ReadFile(v)
		if err == nil {
			return string(b), nil
		}
	}
	return "", fmt.Errorf("未配置商户证书(merchant_cert)")
}

func normalizeSerialNo(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "SERIAL=")
	s = strings.ReplaceAll(s, ":", "")
	return s
}

func (p *WxpayPlugin) resolveMerchantSerialNo(cfg *WxpayConfig) (string, error) {
	if serial := normalizeSerialNo(cfg.SerialNo); serial != "" {
		return serial, nil
	}
	certPEM, err := p.resolveMerchantCert(cfg)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return "", fmt.Errorf("merchant_cert格式错误")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	if cert.SerialNumber == nil || len(cert.SerialNumber.Bytes()) == 0 {
		return "", fmt.Errorf("证书序列号为空")
	}
	serial := strings.ToUpper(hex.EncodeToString(cert.SerialNumber.Bytes()))
	return serial, nil
}

func wxpayAPIURL(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return consts.WechatPayAPIServer
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return consts.WechatPayAPIServer + path
}

// 提交支付
func (p *WxpayPlugin) Submit(params map[string]interface{}) (plugin.SubmitResult, error) {
	method := params["method"].(string)
	channel := params["channel"].(model.Channel)
	tradeNo, _ := params["trade_no"].(string)
	log.Printf("[wxpay_submit_route] channel_id=%d, trade_no=%s, method=%s", channel.ID, tradeNo, method)

	switch method {
	case "scan":
		return p.submitScan(params, channel)
	case "jsapi", "miniprogram":
		return p.submitJSAPI(params, channel)
	case "app":
		return p.submitApp(params, channel)
	case "wap", "h5":
		return p.submitH5(params, channel)
	default:
		return p.submitScan(params, channel)
	}
}

type WxpayAmount struct {
	Total    int    `json:"total"`
	Currency string `json:"currency"`
}

type WxpayNativeOrderRequest struct {
	Appid       string      `json:"appid"`
	Mchid       string      `json:"mchid"`
	Description string      `json:"description"`
	OutTradeNo  string      `json:"out_trade_no"`
	NotifyUrl   string      `json:"notify_url,omitempty"`
	Amount      WxpayAmount `json:"amount"`
}

type WxpayJSAPIOrderRequest struct {
	Appid       string      `json:"appid"`
	Mchid       string      `json:"mchid"`
	Description string      `json:"description"`
	OutTradeNo  string      `json:"out_trade_no"`
	NotifyUrl   string      `json:"notify_url,omitempty"`
	Amount      WxpayAmount `json:"amount"`
	Payer       struct {
		Openid string `json:"openid"`
	} `json:"payer"`
}

type WxpayH5OrderRequest struct {
	Appid       string      `json:"appid"`
	Mchid       string      `json:"mchid"`
	Description string      `json:"description"`
	OutTradeNo  string      `json:"out_trade_no"`
	NotifyUrl   string      `json:"notify_url,omitempty"`
	Amount      WxpayAmount `json:"amount"`
	SceneInfo   struct {
		PayerClientIp string `json:"payer_client_ip"`
		H5Info        struct {
			Type   string `json:"type"`
			AppUrl string `json:"app_url,omitempty"`
		} `json:"h5_info"`
	} `json:"scene_info"`
}

// 扫码支付
func (p *WxpayPlugin) submitScan(params map[string]interface{}, channel model.Channel) (plugin.SubmitResult, error) {
	client, cfg, err := p.getClientAndConfig(channel)
	if err != nil {
		log.Printf("[wxpay_submit_scan_failed] channel_id=%d, reason=%s", channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)

	req := WxpayNativeOrderRequest{
		Appid:       cfg.AppID,
		Mchid:       cfg.MchID,
		Description: name,
		OutTradeNo:  tradeNo,
		NotifyUrl:   notifyURL,
		Amount: WxpayAmount{
			Total:    int(money * 100),
			Currency: "CNY",
		},
	}

	result, err := client.Post(context.Background(), wxpayAPIURL("/v3/pay/transactions/native"), req)
	if err != nil {
		err = wrapWxpayConfigMismatchError(err, cfg)
		log.Printf("[wxpay_submit_scan_failed] trade_no=%s, reason=http post failed, error=%s", tradeNo, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	body, _ := io.ReadAll(result.Response.Body)
	var resp map[string]interface{}
	json.Unmarshal(body, &resp)

	codeUrl, _ := resp["code_url"].(string)
	if codeUrl != "" {
		log.Printf("[wxpay_submit_scan_success] trade_no=%s", tradeNo)
		return plugin.SubmitResult{
			Type: "qrcode",
			URL:  codeUrl,
		}, nil
	}

	log.Printf("[wxpay_submit_scan_failed] trade_no=%s, reason=no code_url returned")
	return plugin.SubmitResult{Msg: "获取二维码失败"}, nil
}

// JSAPI支付
func (p *WxpayPlugin) submitJSAPI(params map[string]interface{}, channel model.Channel) (plugin.SubmitResult, error) {
	client, cfg, err := p.getClientAndConfig(channel)
	if err != nil {
		log.Printf("[wxpay_submit_jsapi_failed] channel_id=%d, reason=%s", channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	openid, _ := params["openid"].(string)

	if openid == "" {
		log.Printf("[wxpay_submit_jsapi_failed] trade_no=%s, reason=openid is empty")
		return plugin.SubmitResult{Msg: "openid不能为空"}, nil
	}

	req := WxpayJSAPIOrderRequest{
		Appid:       cfg.AppID,
		Mchid:       cfg.MchID,
		Description: name,
		OutTradeNo:  tradeNo,
		NotifyUrl:   notifyURL,
		Amount: WxpayAmount{
			Total:    int(money * 100),
			Currency: "CNY",
		},
	}
	req.Payer.Openid = openid

	result, err := client.Post(context.Background(), wxpayAPIURL("/v3/pay/transactions/jsapi"), req)
	if err != nil {
		err = wrapWxpayConfigMismatchError(err, cfg)
		log.Printf("[wxpay_submit_jsapi_failed] trade_no=%s, reason=http post failed, error=%s", tradeNo, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	body, _ := io.ReadAll(result.Response.Body)
	var resp map[string]interface{}
	json.Unmarshal(body, &resp)

	prepayId, _ := resp["prepay_id"].(string)
	if prepayId == "" {
		log.Printf("[wxpay_submit_jsapi_failed] trade_no=%s, reason=no prepay_id returned")
		return plugin.SubmitResult{Msg: "prepay_id获取失败"}, nil
	}

	// 构造调起支付参数
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonceStr := p.generateNonceStr(32)

	// JSAPI/小程序调起签名第4行必须是 package 值: prepay_id=xxx
	signStr := fmt.Sprintf("%s\n%s\n%s\nprepay_id=%s\n", cfg.AppID, timestamp, nonceStr, prepayId)
	privateKey, _ := p.resolvePrivateKey(cfg)
	signature, _ := p.sign(signStr, privateKey)

	jsApiParams := map[string]interface{}{
		"appId":     cfg.AppID,
		"timeStamp": timestamp,
		"nonceStr":  nonceStr,
		"package":   fmt.Sprintf("prepay_id=%s", prepayId),
		"signType":  "RSA",
		"paySign":   signature,
	}

	return plugin.SubmitResult{
		Type: "jsapi",
		Data: jsApiParams,
	}, nil
}

// APP支付
func (p *WxpayPlugin) submitApp(params map[string]interface{}, channel model.Channel) (plugin.SubmitResult, error) {
	client, cfg, err := p.getClientAndConfig(channel)
	if err != nil {
		log.Printf("[wxpay_submit_app_failed] channel_id=%d, reason=%s", channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)

	req := WxpayNativeOrderRequest{
		Appid:       cfg.AppID,
		Mchid:       cfg.MchID,
		Description: name,
		OutTradeNo:  tradeNo,
		NotifyUrl:   notifyURL,
		Amount: WxpayAmount{
			Total:    int(money * 100),
			Currency: "CNY",
		},
	}

	result, err := client.Post(context.Background(), wxpayAPIURL("/v3/pay/transactions/app"), req)
	if err != nil {
		err = wrapWxpayConfigMismatchError(err, cfg)
		log.Printf("[wxpay_submit_app_failed] trade_no=%s, reason=http post failed, error=%s", tradeNo, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	body, _ := io.ReadAll(result.Response.Body)
	var resp map[string]interface{}
	json.Unmarshal(body, &resp)

	prepayId, _ := resp["prepay_id"].(string)
	if prepayId == "" {
		log.Printf("[wxpay_submit_app_failed] trade_no=%s, reason=no prepay_id returned")
		return plugin.SubmitResult{Msg: "prepay_id获取失败"}, nil
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonceStr := p.generateNonceStr(32)

	signStr := fmt.Sprintf("%s\n%s\n%s\n%s\n", cfg.AppID, timestamp, nonceStr, prepayId)
	privateKey, _ := p.resolvePrivateKey(cfg)
	signature, _ := p.sign(signStr, privateKey)

	appParams := map[string]interface{}{
		"appid":     cfg.AppID,
		"partnerid": cfg.MchID,
		"prepayid":  prepayId,
		"package":   "Sign=WXPay",
		"timestamp": timestamp,
		"noncestr":  nonceStr,
		"sign":      signature,
	}

	return plugin.SubmitResult{
		Type: "app",
		Data: appParams,
	}, nil
}

// H5支付
func (p *WxpayPlugin) submitH5(params map[string]interface{}, channel model.Channel) (plugin.SubmitResult, error) {
	client, cfg, err := p.getClientAndConfig(channel)
	if err != nil {
		log.Printf("[wxpay_submit_h5_failed] channel_id=%d, reason=%s", channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	ip := params["ip"].(string)
	if ip == "" {
		ip = "127.0.0.1"
	}

	req := WxpayH5OrderRequest{
		Appid:       cfg.AppID,
		Mchid:       cfg.MchID,
		Description: name,
		OutTradeNo:  tradeNo,
		NotifyUrl:   notifyURL,
		Amount: WxpayAmount{
			Total:    int(money * 100),
			Currency: "CNY",
		},
	}
	req.SceneInfo.PayerClientIp = ip
	req.SceneInfo.H5Info.Type = "Wap"

	result, err := client.Post(context.Background(), wxpayAPIURL("/v3/pay/transactions/h5"), req)
	if err != nil {
		err = wrapWxpayConfigMismatchError(err, cfg)
		log.Printf("[wxpay_submit_h5_failed] trade_no=%s, reason=http post failed, error=%s", tradeNo, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	body, _ := io.ReadAll(result.Response.Body)
	var resp map[string]interface{}
	json.Unmarshal(body, &resp)

	h5Url, _ := resp["h5_url"].(string)
	if h5Url == "" {
		log.Printf("[wxpay_submit_h5_failed] trade_no=%s, reason=no h5_url returned")
		return plugin.SubmitResult{Msg: "h5_url获取失败"}, nil
	}

	returnURL, _ := params["return_url"].(string)
	returnURL = strings.TrimSpace(returnURL)
	if returnURL == "" {
		localURL := strings.TrimRight(strings.TrimSpace(config.Get("localurl")), "/")
		if localURL != "" {
			returnURL = fmt.Sprintf("%s/api/pay/return/%s", localURL, tradeNo)
		}
	}
	if returnURL != "" {
		sep := "?"
		if strings.Contains(h5Url, "?") {
			sep = "&"
		}
		h5Url = h5Url + sep + "redirect_url=" + url.QueryEscape(returnURL)
	}

	log.Printf("[wxpay_submit_h5_success] trade_no=%s", tradeNo)
	return plugin.SubmitResult{
		Type: "jump",
		URL:  h5Url,
	}, nil
}

// 移动端提交
func (p *WxpayPlugin) Mapi(params map[string]interface{}) (plugin.SubmitResult, error) {
	return p.submitH5(params, params["channel"].(model.Channel))
}

// 异步回调
func (p *WxpayPlugin) Notify(tradeNo string, c *gin.Context) (plugin.NotifyResult, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=read body failed, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: err.Error()}, err
	}

	var order model.Order
	if err := config.DB.Where("trade_no = ?", tradeNo).First(&order).Error; err != nil {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=order not found, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: "order not found"}, err
	}

	var channel model.Channel
	if err := config.DB.First(&channel, order.Channel).Error; err != nil {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=channel not found, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: "channel not found"}, err
	}

	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=get config failed, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: err.Error()}, err
	}

	// 验证回调签名
	wxTimestamp := c.GetHeader("Wechatpay-Timestamp")
	wxNonce := c.GetHeader("Wechatpay-Nonce")
	wxSignature := c.GetHeader("Wechatpay-Signature")
	wxSerial := c.GetHeader("Wechatpay-Serial")
	if wxTimestamp == "" || wxNonce == "" || wxSignature == "" || wxSerial == "" {
		err := fmt.Errorf("missing wechatpay signature headers")
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=missing wechatpay headers", tradeNo)
		return plugin.NotifyResult{Success: false, Message: "missing wechatpay signature headers"}, err
	}
	if err := p.verifyNotifyTimestamp(wxTimestamp); err != nil {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=invalid timestamp, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: "invalid wechatpay timestamp"}, err
	}
	if err := p.verifyNotifySignature(wxTimestamp, wxNonce, string(body), wxSignature, wxSerial, cfg); err != nil {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=signature verification failed, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: "signature verification failed"}, err
	}

	var notifyData map[string]interface{}
	if err := json.Unmarshal(body, &notifyData); err != nil {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=parse notify data failed, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: "parse notify data failed"}, err
	}

	resource, ok := notifyData["resource"].(map[string]interface{})
	if !ok {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=missing resource field", tradeNo)
		return plugin.NotifyResult{Success: false, Message: "missing resource field"}, nil
	}

	ciphertext, _ := resource["ciphertext"].(string)
	nonce, _ := resource["nonce"].(string)
	associatedData, _ := resource["associated_data"].(string)

	plaintext, err := p.decryptCiphertext(ciphertext, nonce, associatedData, cfg.AppKey)
	if err != nil {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=decrypt failed, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: "decrypt failed: " + err.Error()}, err
	}

	var result struct {
		OutTradeNo     string `json:"out_trade_no"`
		TransactionId  string `json:"transaction_id"`
		TradeState     string `json:"trade_state"`
		TradeStateDesc string `json:"trade_state_desc"`
		Amount         struct {
			Total int `json:"total"`
		} `json:"amount"`
		Payer struct {
			Openid string `json:"openid"`
		} `json:"payer"`
	}

	if err := json.Unmarshal([]byte(plaintext), &result); err != nil {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=parse decrypted data failed, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: "parse decrypted data failed"}, err
	}

	if result.OutTradeNo != tradeNo {
		log.Printf("[wxpay_notify_failed] trade_no=%s, expected=%s, got=%s, reason=trade no mismatch", tradeNo, tradeNo, result.OutTradeNo)
		return plugin.NotifyResult{Success: false, Message: "trade no mismatch"}, nil
	}

	if float64(result.Amount.Total)/100 != order.Money {
		log.Printf("[wxpay_notify_failed] trade_no=%s, expected=%.2f, got=%.2f, reason=amount mismatch", tradeNo, order.Money, float64(result.Amount.Total)/100)
		return plugin.NotifyResult{Success: false, Message: "amount mismatch"}, nil
	}

	if result.TradeState == "SUCCESS" {
		log.Printf("[wxpay_notify_success] trade_no=%s, transaction_id=%s, amount=%.2f", tradeNo, result.TransactionId, float64(result.Amount.Total)/100)
		return plugin.NotifyResult{
			Success:    true,
			TradeNo:    tradeNo,
			APITradeNo: result.TransactionId,
			Amount:     float64(result.Amount.Total) / 100,
			Buyer:      result.Payer.Openid,
			Message:    "success",
		}, nil
	}

	log.Printf("[wxpay_notify_failed] trade_no=%s, trade_state=%s, desc=%s", tradeNo, result.TradeState, result.TradeStateDesc)
	return plugin.NotifyResult{Success: false, Message: result.TradeStateDesc}, nil
}

func (p *WxpayPlugin) verifyNotifySignature(timestamp, nonce, body, signature, serial string, cfg *WxpayConfig) error {
	message := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, body)

	platformCert, err := p.resolvePlatformCert(cfg)
	if err != nil {
		return fmt.Errorf("get platform cert failed: %v", err)
	}
	pubKey, err := p.pickPlatformPublicKeyBySerial(platformCert, serial)
	if err != nil {
		return err
	}

	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("decode signature failed: %v", err)
	}

	h := sha256.New()
	h.Write([]byte(message))
	hashed := h.Sum(nil)

	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed, sigBytes); err != nil {
		return fmt.Errorf("signature verification failed: %v", err)
	}
	return nil
}

func (p *WxpayPlugin) verifyNotifyTimestamp(ts string) error {
	ts = strings.TrimSpace(ts)
	if ts == "" {
		return fmt.Errorf("empty timestamp")
	}
	got, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp format")
	}
	now := time.Now().Unix()
	if got <= 0 {
		return fmt.Errorf("invalid timestamp value")
	}
	// 微信支付回调验签推荐校验时间戳新鲜度，避免重放攻击。
	if now-got > 300 || got-now > 300 {
		return fmt.Errorf("timestamp expired")
	}
	return nil
}

func (p *WxpayPlugin) pickPlatformPublicKeyBySerial(certPEM, serial string) (*rsa.PublicKey, error) {
	serial = normalizeSerialNo(serial)
	if serial == "" {
		return nil, fmt.Errorf("empty wechatpay serial")
	}

	rest := []byte(certPEM)
	for len(rest) > 0 {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		certSerial := normalizeSerialNo(strings.ToUpper(hex.EncodeToString(cert.SerialNumber.Bytes())))
		if certSerial != serial {
			continue
		}
		pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("platform cert public key is not RSA")
		}
		return pubKey, nil
	}
	return nil, fmt.Errorf("platform cert serial mismatch: serial=%s", serial)
}

func (p *WxpayPlugin) resolvePlatformCert(cfg *WxpayConfig) (string, error) {
	candidates := []string{
		strings.TrimSpace(cfg.PlatformCert),
		strings.TrimSpace(cfg.PlatformCertPath),
	}
	for _, v := range candidates {
		if v == "" {
			continue
		}
		if strings.Contains(v, "BEGIN CERTIFICATE") {
			return v, nil
		}
		b, err := os.ReadFile(v)
		if err == nil {
			return string(b), nil
		}
	}
	return "", fmt.Errorf("platform cert not configured")
}

// 解密
func (p *WxpayPlugin) decryptCiphertext(ciphertext, nonce, associatedData, apiKey string) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		log.Printf("[wxpay_decrypt_failed] reason=base64 decode failed, error=%s", err.Error())
		return "", err
	}

	key := []byte(apiKey)
	if len(key) != 32 {
		log.Printf("[wxpay_decrypt_failed] reason=invalid api key length, expected=32, got=%d", len(key))
		return "", fmt.Errorf("invalid api key length: %d, expected 32", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("[wxpay_decrypt_failed] reason=create cipher failed, error=%s", err.Error())
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("[wxpay_decrypt_failed] reason=create gcm failed, error=%s", err.Error())
		return "", err
	}

	plaintext, err := gcm.Open(nil, []byte(nonce), cipherBytes, []byte(associatedData))
	if err != nil {
		log.Printf("[wxpay_decrypt_failed] reason=gcm open failed, error=%s", err.Error())
		return "", err
	}

	return string(plaintext), nil
}

// 同步回调
func (p *WxpayPlugin) Return(tradeNo string, c *gin.Context) (plugin.ReturnResult, error) {
	return plugin.ReturnResult{
		Success: true,
		TradeNo: tradeNo,
		Message: "支付成功",
		URL:     "/user/order",
	}, nil
}

// 支付成功页面
func (p *WxpayPlugin) OK(tradeNo string) (string, error) {
	return "订单支付成功", nil
}

// 退款
func (p *WxpayPlugin) Refund(params map[string]interface{}) (plugin.RefundResult, error) {
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	channel := params["channel"].(model.Channel)

	client, err := p.getClient(channel)
	if err != nil {
		log.Printf("[wxpay_refund_failed] trade_no=%s, money=%.2f, reason=get client failed, error=%s", tradeNo, money, err.Error())
		return plugin.RefundResult{Code: -1, ErrMsg: err.Error()}, err
	}

	order, err := p.getOrder(tradeNo)
	if err != nil {
		log.Printf("[wxpay_refund_failed] trade_no=%s, money=%.2f, reason=order not found, error=%s", tradeNo, money, err.Error())
		return plugin.RefundResult{Code: -1, ErrMsg: "订单不存在"}, err
	}

	req := map[string]interface{}{
		"transaction_id": order.ApiTradeNo,
		"out_refund_no":  fmt.Sprintf("R%s", tradeNo),
		"amount": map[string]interface{}{
			"refund":   int(money * 100),
			"total":    int(order.Money * 100),
			"currency": "CNY",
		},
	}

	result, err := client.Post(context.Background(), wxpayAPIURL("/v3/refund/domestic/refunds"), req)
	if err != nil {
		log.Printf("[wxpay_refund_failed] trade_no=%s, money=%.2f, reason=http post failed, error=%s", tradeNo, money, err.Error())
		return plugin.RefundResult{Code: -1, ErrMsg: err.Error()}, err
	}

	body, _ := io.ReadAll(result.Response.Body)
	var resp map[string]interface{}
	json.Unmarshal(body, &resp)

	log.Printf("[wxpay_refund_success] trade_no=%s, money=%.2f", tradeNo, money)
	return plugin.RefundResult{
		Code:    0,
		TradeNo: tradeNo,
		Fee:     money,
		Time:    time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// 转账
func (p *WxpayPlugin) Transfer(params map[string]interface{}) (plugin.TransferResult, error) {
	bizNo := params["biz_no"].(string)
	account := params["account"].(string)
	money := params["money"].(float64)
	channel := params["channel"].(model.Channel)

	client, err := p.getClient(channel)
	if err != nil {
		log.Printf("[wxpay_transfer_failed] biz_no=%s, account=%s, money=%.2f, reason=get client failed, error=%s", bizNo, account, money, err.Error())
		return plugin.TransferResult{Code: -1, ErrMsg: err.Error()}, err
	}

	cfg, _ := p.getConfig(channel)

	req := map[string]interface{}{
		"appid":        cfg.AppID,
		"mchid":        cfg.MchID,
		"out_batch_no": bizNo,
		"batch_name":   "商户转账",
		"batch_reason": "商户转账",
		"total_amount": int(money * 100),
		"total_num":    1,
		"transfer_detail_list": []map[string]interface{}{
			{
				"out_detail_no":   bizNo,
				"transfer_amount": int(money * 100),
				"transfer_remark": "商户转账",
				"openid":          account,
			},
		},
	}

	result, err := client.Post(context.Background(), wxpayAPIURL("/v3/transfer/batches"), req)
	if err != nil {
		log.Printf("[wxpay_transfer_failed] biz_no=%s, account=%s, money=%.2f, reason=http post failed, error=%s", bizNo, account, money, err.Error())
		return plugin.TransferResult{Code: -1, ErrMsg: err.Error()}, err
	}

	body, _ := io.ReadAll(result.Response.Body)
	var resp map[string]interface{}
	json.Unmarshal(body, &resp)

	batchId, _ := resp["batch_id"].(string)

	log.Printf("[wxpay_transfer_success] biz_no=%s, batch_id=%s", bizNo, batchId)
	return plugin.TransferResult{
		Code:    0,
		OrderID: batchId,
		PayDate: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// 转账查询
func (p *WxpayPlugin) TransferQuery(params map[string]interface{}) (plugin.TransferQueryResult, error) {
	bizNo := params["biz_no"].(string)
	channel := params["channel"].(model.Channel)

	client, err := p.getClient(channel)
	if err != nil {
		log.Printf("[wxpay_transfer_query_failed] biz_no=%s, reason=get client failed, error=%s", bizNo, err.Error())
		return plugin.TransferQueryResult{Code: -1, ErrMsg: err.Error()}, err
	}

	url := fmt.Sprintf("/v3/transfer/batches/out_batch_no/%s?need_query_detail=true", bizNo)
	result, err := client.Get(context.Background(), wxpayAPIURL(url))
	if err != nil {
		log.Printf("[wxpay_transfer_query_failed] biz_no=%s, reason=http get failed, error=%s", bizNo, err.Error())
		return plugin.TransferQueryResult{Code: -1, ErrMsg: err.Error()}, err
	}

	body, _ := io.ReadAll(result.Response.Body)
	var resp map[string]interface{}
	json.Unmarshal(body, &resp)

	status := 0
	switch resp["batch_status"] {
	case "SUCCESS":
		status = 1
	case "FAIL":
		status = 2
	}

	return plugin.TransferQueryResult{
		Code:    0,
		Status:  status,
		Amount:  0,
		PayDate: "",
	}, nil
}

// 辅助方法
func (p *WxpayPlugin) getOrder(tradeNo string) (*model.Order, error) {
	var order model.Order
	if err := config.DB.Where("trade_no = ?", tradeNo).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// RSA 签名
func (p *WxpayPlugin) sign(message, privateKey string) (string, error) {
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		log.Printf("[wxpay_sign_failed] reason=pem decode failed")
		return "", fmt.Errorf("failed to decode private key")
	}

	privateKeyParsed, err := parseRSAPrivateKey(block.Bytes)
	if err != nil {
		log.Printf("[wxpay_sign_failed] reason=parse private key failed, error=%s", err.Error())
		return "", err
	}

	h := sha256.New()
	h.Write([]byte(message))

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKeyParsed, crypto.SHA256, h.Sum(nil))
	if err != nil {
		log.Printf("[wxpay_sign_failed] reason=rsa sign failed, error=%s", err.Error())
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func parseRSAPrivateKey(keyBytes []byte) (*rsa.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(keyBytes); err == nil {
		return key, nil
	}
	pkcs8Key, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := pkcs8Key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}
	return rsaKey, nil
}

// 生成随机字符串
func (p *WxpayPlugin) generateNonceStr(length int) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		randBytes := make([]byte, 1)
		rand.Read(randBytes)
		result[i] = chars[int(randBytes[0])%len(chars)]
	}
	return string(result)
}

// 订单查询
func (p *WxpayPlugin) QueryOrder(params map[string]interface{}) (map[string]interface{}, error) {
	tradeNo, _ := params["trade_no"].(string)
	channel, ok := params["channel"].(model.Channel)
	if !ok {
		log.Printf("[wxpay_query_order_failed] trade_no=%s, reason=missing channel param", tradeNo)
		return nil, fmt.Errorf("missing channel param")
	}

	client, err := p.getClient(channel)
	if err != nil {
		log.Printf("[wxpay_query_order_failed] trade_no=%s, channel_id=%d, reason=get client failed, error=%s", tradeNo, channel.ID, err.Error())
		return nil, err
	}

	cfg, _ := p.getConfig(channel)
	url := fmt.Sprintf("/v3/pay/transactions/out-trade-no/%s?mchid=%s", tradeNo, cfg.MchID)

	result, err := client.Get(context.Background(), wxpayAPIURL(url))
	if err != nil {
		log.Printf("[wxpay_query_order_failed] trade_no=%s, reason=http get failed, error=%s", tradeNo, err.Error())
		return nil, err
	}

	body, _ := io.ReadAll(result.Response.Body)
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Printf("[wxpay_query_order_failed] trade_no=%s, reason=parse response failed, error=%s", tradeNo, err.Error())
		return nil, err
	}

	tradeState, _ := resp["trade_state"].(string)
	if tradeState == "" {
		log.Printf("[wxpay_query_order_failed] trade_no=%s, reason=missing trade_state in response", tradeNo)
		return nil, fmt.Errorf("missing trade_state in response")
	}

	paid := tradeState == "SUCCESS"

	amount := 0.0
	if amountVal, ok := resp["amount"].(map[string]interface{}); ok {
		if total, ok := amountVal["total"].(float64); ok {
			amount = total / 100
		}
	}

	apiTradeNo, _ := resp["transaction_id"].(string)
	buyer := ""
	if payer, ok := resp["payer"].(map[string]interface{}); ok {
		buyer, _ = payer["openid"].(string)
	}

	successTime, _ := resp["success_time"].(string)

	log.Printf("[wxpay_query_order_success] trade_no=%s, trade_state=%s, paid=%v", tradeNo, tradeState, paid)
	return map[string]interface{}{
		"exists":       true,
		"paid":         paid,
		"status":       tradeState,
		"api_trade_no": apiTradeNo,
		"amount":       amount,
		"buyer":        buyer,
		"pay_time":     successTime,
	}, nil
}

// 测试配置
func (p *WxpayPlugin) TestConfig(config string) (bool, string) {
	var cfg WxpayConfig
	if err := json.Unmarshal([]byte(config), &cfg); err != nil {
		log.Printf("[wxpay_test_config_failed] reason=parse config failed, error=%s", err.Error())
		return false, "配置格式错误: " + err.Error()
	}

	if cfg.AppID == "" {
		log.Printf("[wxpay_test_config_failed] reason=missing appid")
		return false, "缺少应用APPID"
	}

	if cfg.MchID == "" {
		log.Printf("[wxpay_test_config_failed] reason=missing mchid")
		return false, "缺少商户号MCHID"
	}

	if cfg.AppKey == "" {
		log.Printf("[wxpay_test_config_failed] reason=missing appkey")
		return false, "缺少API密钥"
	}

	privateKey, err := p.resolvePrivateKey(&cfg)
	if err != nil {
		log.Printf("[wxpay_test_config_failed] reason=missing private key, error=%s", err.Error())
		return false, "缺少商户私钥(private_key)"
	}
	serialNo, err := p.resolveMerchantSerialNo(&cfg)
	if err != nil {
		log.Printf("[wxpay_test_config_failed] reason=resolve serial_no failed, error=%s", err.Error())
		return false, "缺少商户证书序列号(serial_no)，请填写或提供merchant_cert自动解析"
	}
	if _, err := p.resolvePlatformCert(&cfg); err != nil {
		log.Printf("[wxpay_test_config_failed] reason=missing platform cert, error=%s", err.Error())
		return false, "缺少微信支付平台证书(platform_cert)，无法验证回调签名"
	}

	// 测试签名
	testMessage := fmt.Sprintf("%s|%s|%s", cfg.AppID, cfg.MchID, p.generateNonceStr(32))

	sign, err := p.sign(testMessage, privateKey)
	if err != nil {
		log.Printf("[wxpay_test_config_failed] appid=%s, reason=sign failed, error=%s", cfg.AppID, err.Error())
		return false, "签名失败: " + err.Error()
	}

	log.Printf("[wxpay_test_config_success] appid=%s", cfg.AppID)
	return true, "配置正确，序列号=" + serialNo + "，签名测试成功: " + sign[:20] + "..."
}
