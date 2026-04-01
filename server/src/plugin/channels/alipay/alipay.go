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
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"paygo/src/config"
	"paygo/src/model"
	"paygo/src/plugin"
)

// 支付宝插件
type AlipayPlugin struct {
	plugin.BasePlugin
}

// 配置结构
type AlipayConfig struct {
	AppID      string `json:"appid"`
	AppKey     string `json:"appkey"`     // 支付宝公钥
	AppSecret  string `json:"appsecret"`  // 应用私钥
	GatewayURL string `json:"appurl"`      // 自定义网关
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
			"3": "当面付扫码",
			"4": "当面付JS",
			"5": "预授权支付",
			"6": "APP支付",
			"7": "JSAPI支付",
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
	if channel.Config != "" {
		if err := json.Unmarshal([]byte(channel.Config), cfg); err != nil {
			log.Printf("[alipay_get_config_failed] channel_id=%d, reason=parse config failed, error=%s", channel.ID, err.Error())
			return nil, err
		}
	}
	return cfg, nil
}

// 获取网关地址
func (p *AlipayPlugin) getGatewayURL(cfg *AlipayConfig) string {
	if cfg.GatewayURL != "" {
		return cfg.GatewayURL
	}
	return "https://openapi.alipay.com/gateway.do"
}

// 提交支付
func (p *AlipayPlugin) Submit(params map[string]interface{}) (plugin.SubmitResult, error) {
	method := params["method"].(string)
	channel := params["channel"].(model.Channel)
	tradeNo := params["trade_no"].(string)

	switch method {
	case "web", "jump":
		return p.submitWeb(params, channel)
	case "scan":
		return p.submitScan(params, channel)
	case "jsapi":
		return p.submitJSAPI(params, channel)
	case "app":
		return p.submitApp(params, channel)
	case "wap":
		return p.submitWap(params, channel)
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
		"charset":     "utf-8",
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
		"charset":     "utf-8",
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
		"charset":     "utf-8",
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
		"charset":     "utf-8",
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
	orderString := fmt.Sprintf("app_id=%s&biz_content=%s&charset=utf-8&format=JSON&method=alipay.trade.app.pay&notify_url=%s&sign_type=RSA2&timestamp=%s&version=1.0",
		cfg.AppID, url.QueryEscape(string(bizContentStr)), url.QueryEscape(notifyURL), url.QueryEscape(time.Now().Format("2006-01-02 15:04:05")))

	// 重新签名
	signStr, _ := p.signParams(map[string]interface{}{
		"app_id":      cfg.AppID,
		"biz_content": string(bizContentStr),
		"charset":     "utf-8",
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
		"charset":     "utf-8",
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
		"charset":     "utf-8",
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
		Code:  0,
		TradeNo: tradeNo,
		Fee:   money,
		Time:  time.Now().Format("2006-01-02 15:04:05"),
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
		"charset":     "utf-8",
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
		"charset":     "utf-8",
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

	// RSA2 签名
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return "", fmt.Errorf("failed to decode private key")
	}

	h := sha256.New()
	h.Write([]byte(signData))

	privateKeyParsed, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKeyParsed, crypto.SHA256, h.Sum(nil))
	if err != nil {
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

	signBytes, _ := base64.StdEncoding.DecodeString(sign)

	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return false
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
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

	resp, err := http.PostForm(gatewayURL, formData)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// 构建 HTML 表单
func (p *AlipayPlugin) buildForm(params map[string]interface{}, gatewayURL string) string {
	var buf bytes.Buffer
	buf.WriteString("<!DOCTYPE html><html><head><meta charset=\"utf-8\"></head><body>")
	buf.WriteString(fmt.Sprintf("<form id=\"payform\" action=\"%s\" method=\"POST\">", gatewayURL))

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
		buf.WriteString(fmt.Sprintf("<input type=\"hidden\" name=\"%s\" value=\"%s\">", k, val))
	}

	buf.WriteString("</form><script>document.getElementById('payform').submit();</script></body></html>")
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
