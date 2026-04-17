package api

import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"paygo/src/middleware"
	"paygo/src/model"
	"paygo/src/plugin"
	"paygo/src/service"
)

// 支付API Handler
type PayHandler struct {
	paymentSvc *service.PaymentService
	orderSvc   *service.OrderService
	authSvc    *service.AuthService
}

func NewPayHandler() *PayHandler {
	return &PayHandler{
		paymentSvc: service.NewPaymentService(),
		orderSvc:   service.NewOrderService(),
		authSvc:    service.NewAuthService(),
	}
}

func payStringParam(c *gin.Context, key string) string {
	if v := strings.TrimSpace(c.Query(key)); v != "" {
		return v
	}
	return strings.TrimSpace(c.PostForm(key))
}

func payIntParam(c *gin.Context, key string, defaultValue int) int {
	value := payStringParam(c, key)
	if value == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}

func formatSignAmount(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func inferDeviceByUA(ua string) string {
	u := strings.ToLower(strings.TrimSpace(ua))
	if u == "" {
		return "pc"
	}
	mobileKeywords := []string{
		"iphone", "ipod", "ipad", "android", "mobile", "windows phone",
		"blackberry", "opera mini", "opera mobi", "micromessenger",
	}
	for _, kw := range mobileKeywords {
		if strings.Contains(u, kw) {
			return "mobile"
		}
	}
	return "pc"
}

func signPayloadWithoutKey(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		if k == "sign" || k == "sign_type" || params[k] == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(params[k])
	}
	return b.String()
}

func (h *PayHandler) verifyOpenAPISign(pid uint, signType string, sign string, signParams map[string]string) error {
	if pid == 0 {
		return strconv.ErrSyntax
	}
	typ := strings.ToUpper(strings.TrimSpace(signType))
	if typ == "" {
		typ = "MD5"
	}
	if typ != "MD5" {
		return strconv.ErrRange
	}
	if strings.TrimSpace(sign) == "" {
		return strconv.ErrSyntax
	}

	user, err := h.authSvc.GetUser(pid)
	if err != nil {
		return err
	}

	providedSign := strings.ToLower(strings.TrimSpace(sign))
	expectedSign := h.authSvc.MakeSign(signParams, user.Key)
	if providedSign != expectedSign {
		log.Printf("[openapi_sign_mismatch] pid=%d, sign_type=%s, provided=%s, expected=%s, payload=%s",
			pid, typ, providedSign, expectedSign, signPayloadWithoutKey(signParams))
		return http.ErrNoCookie
	}
	return nil
}

// 支付提交 (POST /api/pay/submit)
func (h *PayHandler) Submit(c *gin.Context) {
	pid, _ := strconv.Atoi(c.PostForm("pid"))
	payType, _ := strconv.Atoi(c.PostForm("type"))
	channelID, _ := strconv.Atoi(c.PostForm("channel"))
	outTradeNo := c.PostForm("out_trade_no")
	name := c.PostForm("name")
	moneyStr := c.PostForm("money")
	notifyURL := c.PostForm("notify_url")
	returnURL := c.PostForm("return_url")
	param := c.PostForm("param")
	device := strings.TrimSpace(c.PostForm("device"))

	money, _ := strconv.ParseFloat(moneyStr, 10)
	if money <= 0 {
		log.Printf("[pay_submit_failed] pid=%d, out_trade_no=%s, money=%s, reason=invalid amount")
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "金额必须大于0"})
		return
	}

	ip := middleware.GetRealIP(c)
	if device == "" {
		device = inferDeviceByUA(c.GetHeader("User-Agent"))
	}

	params := service.SubmitParams{
		UID:        uint(pid),
		OutTradeNo: outTradeNo,
		Type:       payType,
		ChannelID:  channelID,
		Name:       name,
		Money:      money,
		NotifyURL:  notifyURL,
		ReturnURL:  returnURL,
		Param:      param,
		IP:         ip,
		Device:     device,
	}

	result, err := h.paymentSvc.SubmitPayment(params)
	if err != nil {
		log.Printf("[pay_submit_failed] pid=%d, out_trade_no=%s, money=%.2f, reason=%s", pid, outTradeNo, money, err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	submitResult := result["result"].(plugin.SubmitResult)

	c.JSON(http.StatusOK, gin.H{
		"code":     0,
		"trade_no": result["trade_no"],
		"result":   submitResult,
	})
}

// JSON API创建订单 (POST /api/pay/create)
func (h *PayHandler) Create(c *gin.Context) {
	var req struct {
		PID        uint    `json:"pid"`
		Type       int     `json:"type"`
		OutTradeNo string  `json:"out_trade_no"`
		Name       string  `json:"name"`
		Money      float64 `json:"money"`
		NotifyURL  string  `json:"notify_url"`
		ReturnURL  string  `json:"return_url"`
		ClientIP   string  `json:"clientip"`
		Device     string  `json:"device"`
		Param      string  `json:"param"`
		Sign       string  `json:"sign"`
		SignType   string  `json:"sign_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[pay_create_failed] reason=invalid json params, error=%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	if req.PID == 0 || req.Type <= 0 || strings.TrimSpace(req.OutTradeNo) == "" || strings.TrimSpace(req.Name) == "" || req.Money <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	signParams := map[string]string{
		"pid":          strconv.FormatUint(uint64(req.PID), 10),
		"type":         strconv.Itoa(req.Type),
		"out_trade_no": req.OutTradeNo,
		"name":         req.Name,
		"money":        formatSignAmount(req.Money),
		"notify_url":   req.NotifyURL,
		"return_url":   req.ReturnURL,
		"clientip":     req.ClientIP,
		"device":       req.Device,
		"param":        req.Param,
	}
	if err := h.verifyOpenAPISign(req.PID, req.SignType, req.Sign, signParams); err != nil {
		log.Printf("[pay_create_sign_failed] pid=%d, out_trade_no=%s, sign_type=%s, reason=%s", req.PID, req.OutTradeNo, req.SignType, err.Error())
		msg := "签名错误"
		if err == strconv.ErrRange {
			msg = "sign_type不支持"
		} else if err == strconv.ErrSyntax {
			msg = "签名不能为空"
		} else if strings.Contains(err.Error(), "record not found") {
			msg = "商户不存在"
		}
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": msg})
		return
	}

	clientIP := strings.TrimSpace(req.ClientIP)
	if clientIP == "" {
		clientIP = middleware.GetRealIP(c)
	}
	device := strings.TrimSpace(req.Device)
	if device == "" {
		device = inferDeviceByUA(c.GetHeader("User-Agent"))
	}

	params := service.SubmitParams{
		UID:        req.PID,
		OutTradeNo: req.OutTradeNo,
		Type:       req.Type,
		Name:       req.Name,
		Money:      req.Money,
		NotifyURL:  req.NotifyURL,
		ReturnURL:  req.ReturnURL,
		Param:      req.Param,
		IP:         clientIP,
		Device:     device,
	}

	result, err := h.paymentSvc.SubmitPayment(params)
	if err != nil {
		log.Printf("[pay_create_failed] pid=%d, out_trade_no=%s, money=%.2f, reason=%s", req.PID, req.OutTradeNo, req.Money, err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	submitResult := result["result"].(plugin.SubmitResult)

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"trade_no":  result["trade_no"],
		"pay_type":  submitResult.Type,
		"pay_info":  submitResult.URL,
		"pay_data":  submitResult.Data,
		"result":    submitResult,
		"timestamp": time.Now().Unix(),
	})
}

// 订单查询 (GET/POST /api/pay/query)
func (h *PayHandler) Query(c *gin.Context) {
	pid := payIntParam(c, "pid", 0)
	tradeNo := payStringParam(c, "trade_no")
	outTradeNo := payStringParam(c, "out_trade_no")
	sign := payStringParam(c, "sign")
	signType := payStringParam(c, "sign_type")

	if pid <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	signParams := map[string]string{
		"pid":          strconv.Itoa(pid),
		"trade_no":     tradeNo,
		"out_trade_no": outTradeNo,
	}
	if err := h.verifyOpenAPISign(uint(pid), signType, sign, signParams); err != nil {
		log.Printf("[pay_query_sign_failed] pid=%d, trade_no=%s, out_trade_no=%s, sign_type=%s, reason=%s", pid, tradeNo, outTradeNo, signType, err.Error())
		msg := "签名错误"
		if err == strconv.ErrRange {
			msg = "sign_type不支持"
		} else if err == strconv.ErrSyntax {
			msg = "签名不能为空"
		} else if strings.Contains(err.Error(), "record not found") {
			msg = "商户不存在"
		}
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": msg})
		return
	}

	var order *model.Order
	var err error

	if tradeNo != "" {
		order, err = h.orderSvc.GetOrder(tradeNo)
	} else if outTradeNo != "" {
		order, err = h.orderSvc.GetOrderByOutTradeNo(outTradeNo, uint(pid))
	} else {
		log.Printf("[pay_query_failed] pid=%d, trade_no=%s, out_trade_no=%s, reason=missing order number")
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "缺少订单号"})
		return
	}

	if err != nil {
		log.Printf("[pay_query_failed] pid=%d, trade_no=%s, out_trade_no=%s, reason=%s", pid, tradeNo, outTradeNo, err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "订单不存在"})
		return
	}

	// 验证订单所属
	if order.UID != uint(pid) {
		log.Printf("[pay_query_failed] pid=%d, trade_no=%s, order_uid=%d, reason=order does not belong to merchant")
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "订单不属于该商户"})
		return
	}

	typename := h.orderSvc.GetTypeName(order.Type)

	c.JSON(http.StatusOK, gin.H{
		"code":         0,
		"trade_no":     order.TradeNo,
		"out_trade_no": order.OutTradeNo,
		"api_trade_no": order.ApiTradeNo,
		"type":         typename,
		"pid":          order.UID,
		"addtime":      order.Addtime.Format("2006-01-02 15:04:05"),
		"endtime":      order.Endtime.Format("2006-01-02 15:04:05"),
		"name":         order.Name,
		"money":        order.Money,
		"status":       order.Status,
		"buyer":        order.Buyer,
	})
}

// 退款 (POST /api/pay/refund)
func (h *PayHandler) Refund(c *gin.Context) {
	pid := payIntParam(c, "pid", 0)
	tradeNo := payStringParam(c, "trade_no")
	moneyStr := payStringParam(c, "money")
	sign := payStringParam(c, "sign")
	signType := payStringParam(c, "sign_type")

	money, _ := strconv.ParseFloat(moneyStr, 10)
	if pid <= 0 || tradeNo == "" || money <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	signParams := map[string]string{
		"pid":      strconv.Itoa(pid),
		"trade_no": tradeNo,
		"money":    formatSignAmount(money),
	}
	if err := h.verifyOpenAPISign(uint(pid), signType, sign, signParams); err != nil {
		log.Printf("[pay_refund_sign_failed] pid=%d, trade_no=%s, sign_type=%s, reason=%s", pid, tradeNo, signType, err.Error())
		msg := "签名错误"
		if err == strconv.ErrRange {
			msg = "sign_type不支持"
		} else if err == strconv.ErrSyntax {
			msg = "签名不能为空"
		} else if strings.Contains(err.Error(), "record not found") {
			msg = "商户不存在"
		}
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": msg})
		return
	}

	order, err := h.orderSvc.GetOrder(tradeNo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "订单不存在"})
		return
	}
	if order.UID != uint(pid) {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "订单不属于该商户"})
		return
	}

	err = h.paymentSvc.Refund(tradeNo, money)
	if err != nil {
		log.Printf("[pay_refund_failed] trade_no=%s, money=%.2f, reason=%s", tradeNo, money, err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "退款成功"})
}

// 异步回调 (POST /api/pay/notify/:trade_no)
func (h *PayHandler) Notify(c *gin.Context) {
	tradeNo := c.Param("trade_no")

	// 获取订单对应的通道插件
	order, err := h.orderSvc.GetOrder(tradeNo)
	if err != nil {
		log.Printf("[pay_notify_failed] trade_no=%s, reason=get order failed, error=%s", tradeNo, err.Error())
		c.String(http.StatusOK, "fail")
		return
	}

	// 获取通道信息
	channel, err := h.paymentSvc.GetChannelConfig(order.Channel)
	if err != nil {
		log.Printf("[pay_notify_failed] trade_no=%s, reason=get channel failed, error=%s", tradeNo, err.Error())
		c.String(http.StatusOK, "fail")
		return
	}

	result, err := h.paymentSvc.HandleNotify(tradeNo, channel.Plugin, c)
	if err != nil || !result["success"].(bool) {
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		log.Printf("[pay_notify_failed] trade_no=%s, plugin=%s, reason=handle notify failed, error=%s", tradeNo, channel.Plugin, errMsg)
		if channel.Plugin == "wxpay" {
			c.JSON(http.StatusOK, gin.H{"code": "FAIL", "message": "handle failed"})
		} else {
			c.String(http.StatusOK, "fail")
		}
		return
	}

	if channel.Plugin == "wxpay" {
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "success"})
	} else {
		c.String(http.StatusOK, "success")
	}
}

// 同步回调 (GET /api/pay/return/:trade_no)
func (h *PayHandler) Return(c *gin.Context) {
	tradeNo := c.Param("trade_no")

	order, err := h.orderSvc.GetOrder(tradeNo)
	if err != nil {
		log.Printf("[pay_return_failed] trade_no=%s, reason=get order failed, error=%s", tradeNo, err.Error())
		c.Redirect(http.StatusFound, "/?error=订单不存在")
		return
	}

	// 获取通道信息，调用插件处理同步回调
	channel, err := h.paymentSvc.GetChannelConfig(order.Channel)
	if err == nil {
		result, err := h.paymentSvc.HandleReturn(tradeNo, channel.Plugin, c)
		if err == nil && result.Success {
			if result.URL != "" {
				c.Redirect(http.StatusFound, result.URL)
				return
			}
		}
	}

	if order.Status == 1 {
		c.Redirect(http.StatusFound, "/?success=支付成功")
		return
	}

	c.Redirect(http.StatusFound, "/?error=支付失败")
}

// 获取可用支付方式 (GET /api/pay/types)
func (h *PayHandler) GetTypes(c *gin.Context) {
	pid, _ := strconv.Atoi(c.Query("pid"))

	types, err := h.paymentSvc.GetAvailableTypes(uint(pid))
	if err != nil {
		log.Printf("[get_types_failed] pid=%d, reason=%s", pid, err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": types})
}

// 获取可用通道 (GET /api/pay/channels)
func (h *PayHandler) GetChannels(c *gin.Context) {
	pid, _ := strconv.Atoi(c.Query("pid"))
	typeID, _ := strconv.Atoi(c.Query("type"))

	channels, err := h.paymentSvc.GetAvailableChannels(uint(pid), typeID)
	if err != nil {
		log.Printf("[get_channels_failed] pid=%d, type_id=%d, reason=%s", pid, typeID, err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": channels})
}
