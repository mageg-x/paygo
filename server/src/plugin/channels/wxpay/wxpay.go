package wxpay

import (
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"paygo/src/config"
	"paygo/src/model"
	"paygo/src/plugin"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
)

// 微信支付插件
type WxpayPlugin struct {
	plugin.BasePlugin
}

// 配置结构
type WxpayConfig struct {
	AppID     string `json:"appid"`
	AppKey    string `json:"appkey"`   // APIv3密钥(平台证书密钥)
	MchID     string `json:"appmchid"` // 商户号
	AppSecret string `json:"appsecret"`
	CertPath  string `json:"cert_path"`
	KeyPath   string `json:"key_path"`
	SerialNo  string `json:"serial_no"` // 证书序列号
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

// 客户端缓存
var clientCache = make(map[string]*core.Client)

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
			"appid":    {Name: "应用ID(AppID)", Type: "input"},
			"appkey":   {Name: "APIv3密钥", Type: "input"},
			"appmchid": {Name: "商户号", Type: "input"},
		},
		Select: map[string]string{
			"1": "扫码支付",
			"2": "公众号支付",
			"3": "H5支付",
			"4": "小程序支付",
			"5": "APP支付",
		},
		Note: `<p>微信支付官方接口 V3版</p>
<h4 class="mt-3 font-medium">配置示例：</h4>
<pre class="bg-gray-50 p-2 rounded text-xs mt-1 overflow-x-auto">
{
  "appid": "wx1234567890abcdef",
  "appkey": "your_api_v3_key_here",
  "appmchid": "1234567890",
  "appsecret": "",
  "cert_path": "/path/to/apiclient_cert.pem",
  "key_path": "/path/to/apiclient_key.pem",
  "serial_no": "ABCDEF1234567890"
}
</pre>
<p class="text-xs text-gray-500 mt-2">* appkey: APIv3密钥（从微信支付平台获取）<br>* appsecret: 应用密钥（通常留空）<br>* cert_path: API证书路径（退款、转账需要）<br>* serial_no: 证书序列号</p>`,
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

	merged, err := mergeConfigJSON(pluginConfig, channel.Config)
	if err != nil {
		log.Printf("[wxpay_get_config_failed] channel_id=%d, reason=merge config failed, error=%s", channel.ID, err.Error())
		return nil, err
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

// 获取微信支付客户端
func (p *WxpayPlugin) getClient(channel model.Channel) (*core.Client, error) {
	key := fmt.Sprintf("%s_%d", channel.Plugin, channel.ID)
	if client, ok := clientCache[key]; ok {
		return client, nil
	}

	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[wxpay_get_client_failed] channel_id=%d, reason=get config failed, error=%s", channel.ID, err.Error())
		return nil, err
	}

	// 加载私钥
	privateKey, err := p.loadPrivateKey(cfg.KeyPath)
	if err != nil {
		log.Printf("[wxpay_get_client_failed] channel_id=%d, reason=load private key failed, error=%s", channel.ID, err.Error())
		return nil, fmt.Errorf("加载私钥失败: %v", err)
	}

	// 解析私钥
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		log.Printf("[wxpay_get_client_failed] channel_id=%d, reason=invalid private key format")
		return nil, fmt.Errorf("私钥格式错误")
	}
	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Printf("[wxpay_get_client_failed] channel_id=%d, reason=parse private key failed, error=%s", channel.ID, err.Error())
		return nil, fmt.Errorf("解析私钥失败: %v", err)
	}

	// 创建客户端，不验签（验签在回调时单独处理）
	opts := []core.ClientOption{
		option.WithoutValidator(),
	}
	if cfg.SerialNo != "" && rsaKey != nil {
		opts = append(opts, option.WithMerchantCredential(cfg.MchID, cfg.SerialNo, rsaKey))
	}

	client, err := core.NewClient(context.Background(), opts...)
	if err != nil {
		log.Printf("[wxpay_get_client_failed] channel_id=%d, reason=create client failed, error=%s", channel.ID, err.Error())
		return nil, fmt.Errorf("创建客户端失败: %v", err)
	}

	clientCache[key] = client
	return client, nil
}

// 加载私钥
func (p *WxpayPlugin) loadPrivateKey(keyPath string) (string, error) {
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// 提交支付
func (p *WxpayPlugin) Submit(params map[string]interface{}) (plugin.SubmitResult, error) {
	method := params["method"].(string)
	channel := params["channel"].(model.Channel)

	switch method {
	case "scan":
		return p.submitScan(params, channel)
	case "jsapi":
		return p.submitJSAPI(params, channel)
	case "app":
		return p.submitApp(params, channel)
	case "wap", "h5":
		return p.submitH5(params, channel)
	default:
		return p.submitScan(params, channel)
	}
}

// 统一下单请求结构
type UnifiedOrderRequest struct {
	Appid       string `json:"appid"`
	Mchid       string `json:"mchid"`
	Description string `json:"description"`
	OutTradeNo  string `json:"out_trade_no"`
	NotifyUrl   string `json:"notify_url"`
	Amount      struct {
		Total    int    `json:"total"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Payer struct {
		Openid string `json:"openid,omitempty"`
	} `json:"payer,omitempty"`
	SceneInfo struct {
		PayerClientIp string `json:"payer_client_ip"`
		H5Info        struct {
			Type string `json:"type"`
		} `json:"h5_info"`
	} `json:"scene_info,omitempty"`
}

// 扫码支付
func (p *WxpayPlugin) submitScan(params map[string]interface{}, channel model.Channel) (plugin.SubmitResult, error) {
	client, err := p.getClient(channel)
	if err != nil {
		log.Printf("[wxpay_submit_scan_failed] channel_id=%d, reason=%s", channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	cfg, _ := p.getConfig(channel)
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	ip := params["ip"].(string)
	if ip == "" {
		ip = "127.0.0.1"
	}

	req := UnifiedOrderRequest{
		Appid:       cfg.AppID,
		Mchid:       cfg.MchID,
		Description: name,
		OutTradeNo:  tradeNo,
		NotifyUrl:   notifyURL,
	}
	req.Amount.Total = int(money * 100)
	req.Amount.Currency = "CNY"

	result, err := client.Post(context.Background(), "/v3/pay/transactions/native", req)
	if err != nil {
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
	client, err := p.getClient(channel)
	if err != nil {
		log.Printf("[wxpay_submit_jsapi_failed] channel_id=%d, reason=%s", channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	cfg, _ := p.getConfig(channel)
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	openid := params["openid"].(string)

	if openid == "" {
		log.Printf("[wxpay_submit_jsapi_failed] trade_no=%s, reason=openid is empty")
		return plugin.SubmitResult{Msg: "openid不能为空"}, nil
	}

	req := UnifiedOrderRequest{
		Appid:       cfg.AppID,
		Mchid:       cfg.MchID,
		Description: name,
		OutTradeNo:  tradeNo,
		NotifyUrl:   notifyURL,
	}
	req.Amount.Total = int(money * 100)
	req.Amount.Currency = "CNY"
	req.Payer.Openid = openid

	result, err := client.Post(context.Background(), "/v3/pay/transactions/jsapi", req)
	if err != nil {
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

	signStr := fmt.Sprintf("%s\n%s\n%s\n%s\n", cfg.AppID, timestamp, nonceStr, prepayId)
	signature, _ := p.sign(signStr, cfg.AppKey)

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
	client, err := p.getClient(channel)
	if err != nil {
		log.Printf("[wxpay_submit_app_failed] channel_id=%d, reason=%s", channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	cfg, _ := p.getConfig(channel)
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)

	req := UnifiedOrderRequest{
		Appid:       cfg.AppID,
		Mchid:       cfg.MchID,
		Description: name,
		OutTradeNo:  tradeNo,
		NotifyUrl:   notifyURL,
	}
	req.Amount.Total = int(money * 100)
	req.Amount.Currency = "CNY"

	result, err := client.Post(context.Background(), "/v3/pay/transactions/app", req)
	if err != nil {
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
	signature, _ := p.sign(signStr, cfg.AppKey)

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
	client, err := p.getClient(channel)
	if err != nil {
		log.Printf("[wxpay_submit_h5_failed] channel_id=%d, reason=%s", channel.ID, err.Error())
		return plugin.SubmitResult{Msg: err.Error()}, err
	}

	cfg, _ := p.getConfig(channel)
	tradeNo := params["trade_no"].(string)
	money := params["money"].(float64)
	name := params["name"].(string)
	notifyURL := params["notify_url"].(string)
	ip := params["ip"].(string)
	if ip == "" {
		ip = "127.0.0.1"
	}

	req := UnifiedOrderRequest{
		Appid:       cfg.AppID,
		Mchid:       cfg.MchID,
		Description: name,
		OutTradeNo:  tradeNo,
		NotifyUrl:   notifyURL,
	}
	req.Amount.Total = int(money * 100)
	req.Amount.Currency = "CNY"
	req.SceneInfo.PayerClientIp = ip
	req.SceneInfo.H5Info.Type = "Wap"

	result, err := client.Post(context.Background(), "/v3/pay/transactions/h5", req)
	if err != nil {
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

	// 获取订单
	var order model.Order
	if err := config.DB.Where("trade_no = ?", tradeNo).First(&order).Error; err != nil {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=order not found, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: "订单不存在"}, err
	}

	// 获取通道
	var channel model.Channel
	if err := config.DB.First(&channel, order.Channel).Error; err != nil {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=channel not found, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: "通道不存在"}, err
	}

	cfg, err := p.getConfig(channel)
	if err != nil {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=get config failed, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: err.Error()}, err
	}

	// 解析回调数据
	var notifyData map[string]interface{}
	if err := json.Unmarshal(body, &notifyData); err != nil {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=parse notify data failed, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: "解析通知数据失败"}, err
	}

	resource, ok := notifyData["resource"].(map[string]interface{})
	if !ok {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=missing resource field")
		return plugin.NotifyResult{Success: false, Message: "缺少resource字段"}, nil
	}

	ciphertext, _ := resource["ciphertext"].(string)
	nonce, _ := resource["nonce"].(string)
	associatedData, _ := resource["associated_data"].(string)

	plaintext, err := p.decryptCiphertext(ciphertext, nonce, associatedData, cfg.AppKey)
	if err != nil {
		log.Printf("[wxpay_notify_failed] trade_no=%s, reason=decrypt failed, error=%s", tradeNo, err.Error())
		return plugin.NotifyResult{Success: false, Message: "解密失败: " + err.Error()}, err
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
		return plugin.NotifyResult{Success: false, Message: "解析解密数据失败"}, err
	}

	// 验证订单号
	if result.OutTradeNo != tradeNo {
		log.Printf("[wxpay_notify_failed] trade_no=%s, expected=%s, got=%s, reason=trade no mismatch")
		return plugin.NotifyResult{Success: false, Message: "订单号不匹配"}, nil
	}

	// 验证金额
	if float64(result.Amount.Total)/100 != order.Realmoney {
		log.Printf("[wxpay_notify_failed] trade_no=%s, expected=%.2f, got=%.2f, reason=amount mismatch")
		return plugin.NotifyResult{Success: false, Message: "金额不匹配"}, nil
	}

	// 交易状态
	if result.TradeState == "SUCCESS" {
		log.Printf("[wxpay_notify_success] trade_no=%s, transaction_id=%s, amount=%.2f", tradeNo, result.TransactionId, float64(result.Amount.Total)/100)
		return plugin.NotifyResult{
			Success:    true,
			TradeNo:    tradeNo,
			APITradeNo: result.TransactionId,
			Amount:     float64(result.Amount.Total) / 100,
			Buyer:      result.Payer.Openid,
			Message:    "成功",
		}, nil
	}

	log.Printf("[wxpay_notify_failed] trade_no=%s, trade_state=%s, desc=%s", tradeNo, result.TradeState, result.TradeStateDesc)
	return plugin.NotifyResult{Success: false, Message: result.TradeStateDesc}, nil
}

// 解密
func (p *WxpayPlugin) decryptCiphertext(ciphertext, nonce, associatedData, apiKey string) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		log.Printf("[wxpay_decrypt_failed] reason=base64 decode failed, error=%s", err.Error())
		return "", err
	}

	// 密钥 MD5
	h := md5.New()
	h.Write([]byte(apiKey))
	key := h.Sum(nil)

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

	if len(cipherBytes) < 12 {
		log.Printf("[wxpay_decrypt_failed] reason=ciphertext too short")
		return "", fmt.Errorf("ciphertext too short")
	}

	nonceBytes := []byte(nonce)
	cipherBytes = cipherBytes[12:]

	plaintext, err := gcm.Open(nil, nonceBytes, cipherBytes, []byte(associatedData))
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
			"total":    int(order.Realmoney * 100),
			"currency": "CNY",
		},
	}

	result, err := client.Post(context.Background(), "/v3/refund/domestic/refunds", req)
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

	result, err := client.Post(context.Background(), "/v3/transfer/batches", req)
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
	result, err := client.Get(context.Background(), url)
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

	privateKeyParsed, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Printf("[wxpay_sign_failed] reason=parse pkcs1 failed, error=%s", err.Error())
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

	// 测试签名
	testMessage := fmt.Sprintf("%s|%s|%s", cfg.AppID, cfg.MchID, p.generateNonceStr(32))

	sign, err := p.sign(testMessage, cfg.AppKey)
	if err != nil {
		log.Printf("[wxpay_test_config_failed] appid=%s, reason=sign failed, error=%s", cfg.AppID, err.Error())
		return false, "签名失败: " + err.Error()
	}

	log.Printf("[wxpay_test_config_success] appid=%s", cfg.AppID)
	return true, "配置正确，签名测试成功: " + sign[:20] + "..."
}
