package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"paygo/src/config"
	"paygo/src/model"
	"paygo/src/plugin"

	"github.com/gin-gonic/gin"
)

var errOrderAlreadyProcessed = errors.New("订单不存在或已处理")

type PaymentService struct {
	orderSvc *OrderService
	authSvc  *AuthService
}

func NewPaymentService() *PaymentService {
	return &PaymentService{
		orderSvc: NewOrderService(),
		authSvc:  NewAuthService(),
	}
}

// 支付参数
type SubmitParams struct {
	UID        uint
	OutTradeNo string
	Type       int // 支付类型ID
	ChannelID  int // 通道ID（submit2指定通道）
	Name       string
	Money      float64
	NotifyURL  string
	ReturnURL  string
	Param      string
	Openid     string
	IP         string
	Device     string // pc/mobile
	Method     string // web/jump/jsapi/scan
}

func parsePaymethodCodes(paymethod string) []string {
	if paymethod == "" {
		return nil
	}
	parts := strings.Split(paymethod, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		code := strings.TrimSpace(p)
		if code == "" {
			continue
		}
		result = append(result, code)
	}
	return result
}

func pickByPriority(enabledCodes []string, priority []string) string {
	if len(enabledCodes) == 0 || len(priority) == 0 {
		return ""
	}
	enabled := make(map[string]struct{}, len(enabledCodes))
	for _, code := range enabledCodes {
		enabled[code] = struct{}{}
	}
	for _, code := range priority {
		if _, ok := enabled[code]; ok {
			return code
		}
	}
	return ""
}

func mapAlipayPaymethodToMethod(code string) string {
	switch code {
	case "1":
		return "web"
	case "2":
		return "wap"
	case "3":
		return "scan"
	case "4":
		return "4"
	case "5", "7":
		// 兼容历史编码: 7(旧) 与 5(现) 均表示 APP 支付
		return "app"
	case "6":
		return "jsapi"
	default:
		return ""
	}
}

func mapWxpayPaymethodToMethod(code string) string {
	switch code {
	case "1":
		return "scan"
	case "2", "4":
		return "jsapi"
	case "3":
		return "h5"
	case "5":
		return "app"
	default:
		return ""
	}
}

func deviceClass(device string) string {
	d := strings.ToLower(strings.TrimSpace(device))
	switch d {
	case "app", "ios", "android":
		return "app"
	case "mobile", "h5", "wap":
		return "mobile"
	default:
		return "pc"
	}
}

// 约定：仅当 param 使用 "domain|biz_param" 格式时，才启用域名授权校验。
func parseDomainFromParam(param string) string {
	p := strings.TrimSpace(param)
	if p == "" {
		return ""
	}
	parts := strings.SplitN(p, "|", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func resolveSubmitMethod(channel model.Channel, params SubmitParams) (string, error) {
	if strings.TrimSpace(params.Method) != "" {
		return params.Method, nil
	}

	codes := parsePaymethodCodes(channel.Paymethod)
	dc := deviceClass(params.Device)

	switch channel.Plugin {
	case "alipay":
		if len(codes) == 0 {
			if dc == "app" || dc == "mobile" {
				return "wap", nil
			}
			return "web", nil
		}

		var priority []string
		switch dc {
		case "app":
			priority = []string{"5", "7", "6", "2", "1", "3", "4"}
		case "mobile":
			priority = []string{"2", "6", "5", "7", "1", "3", "4"}
		default:
			priority = []string{"1", "3", "2", "6", "5", "7", "4"}
		}

		code := pickByPriority(codes, priority)
		if code == "" {
			return "", errors.New("通道未配置可用的支付宝支付方式")
		}
		method := mapAlipayPaymethodToMethod(code)
		if method == "" {
			return "", errors.New("当前支付宝支付方式暂未实现，请在通道中改用 1/2/3/4/5/6")
		}
		return method, nil

	case "wxpay":
		if len(codes) == 0 {
			if dc == "app" || dc == "mobile" {
				return "h5", nil
			}
			return "scan", nil
		}

		var priority []string
		switch dc {
		case "app":
			priority = []string{"5", "3", "2", "4", "1"}
		case "mobile":
			priority = []string{"3", "2", "4", "5", "1"}
		default:
			priority = []string{"1", "3", "2", "4", "5"}
		}

		code := pickByPriority(codes, priority)
		if code == "" {
			return "", errors.New("通道未配置可用的微信支付方式")
		}
		method := mapWxpayPaymethodToMethod(code)
		if method == "" {
			return "", errors.New("当前微信支付方式暂未实现")
		}
		return method, nil
	default:
		return "", nil
	}
}

func resolveRechargeSubmitMethod(channel model.Channel) (string, error) {
	switch strings.ToLower(strings.TrimSpace(channel.Plugin)) {
	case "alipay":
		codes := parsePaymethodCodes(channel.Paymethod)
		if len(codes) == 0 {
			return "web", nil
		}
		code := pickByPriority(codes, []string{"1", "2", "3", "6", "5", "7", "4"})
		if code == "" {
			return "", errors.New("通道未配置可用的支付宝支付方式")
		}
		method := mapAlipayPaymethodToMethod(code)
		if method == "" {
			return "", errors.New("当前支付宝支付方式暂未实现，请在通道中改用 1/2/3/4/5/6")
		}
		return method, nil
	case "wxpay":
		codes := parsePaymethodCodes(channel.Paymethod)
		if len(codes) == 0 {
			return "scan", nil
		}
		code := pickByPriority(codes, []string{"1", "3", "2", "4", "5"})
		if code == "" {
			return "", errors.New("通道未配置可用的微信支付方式")
		}
		method := mapWxpayPaymethodToMethod(code)
		if method == "" {
			return "", errors.New("当前微信支付方式暂未实现")
		}
		return method, nil
	default:
		return "", errors.New("该通道不支持充值下单")
	}
}

// 获取可用支付方式
func (s *PaymentService) GetAvailableTypes(uid uint) ([]model.PayType, error) {
	// 获取商户的用户组
	user, err := s.authSvc.GetUser(uid)
	if err != nil {
		log.Printf("[get_available_types_failed] uid=%d, reason=get user failed, error=%s", uid, err.Error())
		return nil, err
	}

	// 获取用户组配置
	var group model.Group
	config.DB.First(&group, user.GID)

	var availableTypes []model.PayType

	// 如果有自定义通道配置
	if group.Settings != "" {
		var settings map[string]interface{}
		json.Unmarshal([]byte(group.Settings), &settings)
		if types, ok := settings["types"]; ok {
			typeIDs := types.([]interface{})
			for _, t := range typeIDs {
				var payType model.PayType
				if config.DB.First(&payType, int(t.(float64))).Error == nil {
					availableTypes = append(availableTypes, payType)
				}
			}
			return availableTypes, nil
		}
	}

	// 默认返回所有开启的支付类型
	config.DB.Where("status = 1").Find(&availableTypes)
	return availableTypes, nil
}

// 获取可用通道
func (s *PaymentService) GetAvailableChannels(uid uint, typeID int) ([]model.Channel, error) {
	user, err := s.authSvc.GetUser(uid)
	if err != nil {
		log.Printf("[get_available_channels_failed] uid=%d, type_id=%d, reason=get user failed, error=%s", uid, typeID, err.Error())
		return nil, err
	}

	// 获取用户组
	var group model.Group
	config.DB.First(&group, user.GID)

	// 获取通道列表
	var channels []model.Channel
	query := config.DB.Where("type = ? AND status = 1", typeID)

	// 如果有自定义配置
	if group.Config != "" {
		var groupConfig map[string]interface{}
		json.Unmarshal([]byte(group.Config), &groupConfig)
		if channelIDs, ok := groupConfig["channels"]; ok {
			ids := channelIDs.([]interface{})
			idStrs := make([]string, len(ids))
			for i, v := range ids {
				idStrs[i] = strconv.Itoa(int(v.(float64)))
			}
			query = query.Where("id IN (?)", idStrs)
		}
	}

	query.Find(&channels)
	return channels, nil
}

// 选择通道（轮询或指定）
func (s *PaymentService) SelectChannel(uid uint, typeID int, channelID int) (*model.Channel, error) {
	if channelID > 0 {
		// 指定通道
		var channel model.Channel
		result := config.DB.Where("id = ? AND type = ? AND status = 1", channelID, typeID).First(&channel)
		if result.Error != nil {
			log.Printf("[select_channel_failed] uid=%d, type_id=%d, channel_id=%d, reason=specified channel not found or disabled", uid, typeID, channelID)
			return nil, errors.New("指定通道不存在或已关闭")
		}
		return &channel, nil
	}

	// 轮询选择
	var roll model.Roll
	result := config.DB.Where("type = ? AND status = 1", typeID).First(&roll)
	if result.Error != nil {
		// 没有轮询配置，直接查询可用通道
		var channel model.Channel
		r := config.DB.Where("type = ? AND status = 1", typeID).First(&channel)
		if r.Error != nil {
			log.Printf("[select_channel_failed] uid=%d, type_id=%d, reason=no available channel")
			return nil, errors.New("没有可用的支付通道")
		}
		return &channel, nil
	}

	// 解析轮询配置
	var rollInfo struct {
		Channels []int `json:"channels"`
		Weights  []int `json:"weights"`
	}
	json.Unmarshal([]byte(roll.Info), &rollInfo)

	if len(rollInfo.Channels) == 0 {
		log.Printf("[select_channel_failed] type_id=%d, reason=roll config empty")
		return nil, errors.New("轮询配置错误")
	}

	// 根据权重随机选择
	if roll.Kind == 1 && len(rollInfo.Weights) > 0 {
		totalWeight := 0
		for _, w := range rollInfo.Weights {
			totalWeight += w
		}
		rand := time.Now().UnixNano() % int64(totalWeight)
		curWeight := 0
		for i, w := range rollInfo.Weights {
			curWeight += w
			if int64(curWeight) >= rand {
				channelID = rollInfo.Channels[i]
				break
			}
		}
	} else {
		// 简单轮询
		currentIndex := roll.Index
		channelID = rollInfo.Channels[currentIndex%len(rollInfo.Channels)]

		// 更新索引
		config.DB.Model(&roll).Update("index", (currentIndex+1)%len(rollInfo.Channels))
	}

	var channel model.Channel
	if config.DB.First(&channel, channelID).Error != nil {
		log.Printf("[select_channel_failed] type_id=%d, channel_id=%d, reason=channel not found")
		return nil, errors.New("通道不存在")
	}

	return &channel, nil
}

// 提交支付
func (s *PaymentService) SubmitPayment(params SubmitParams) (map[string]interface{}, error) {
	// 获取商户
	user, err := s.authSvc.GetUser(params.UID)
	if err != nil {
		log.Printf("[submit_payment_failed] uid=%d, out_trade_no=%s, reason=merchant not found, error=%s", params.UID, params.OutTradeNo, err.Error())
		return nil, errors.New("商户不存在")
	}

	if user.Status != 1 {
		log.Printf("[submit_payment_failed] uid=%d, out_trade_no=%s, reason=merchant disabled, status=%d", params.UID, params.OutTradeNo, user.Status)
		return nil, errors.New("商户已被禁用")
	}

	if user.Pay != 1 {
		log.Printf("[submit_payment_failed] uid=%d, out_trade_no=%s, reason=merchant no pay permission")
		return nil, errors.New("商户没有支付权限")
	}

	// 执行风控检查
	riskSvc := NewRiskService()
	domain := parseDomainFromParam(params.Param)
	riskResult := riskSvc.CheckPaymentRisk(params.UID, params.IP, params.Name, params.Money)
	if !riskResult.Passed {
		log.Printf("[submit_payment_failed] uid=%d, out_trade_no=%s, ip=%s, risk_code=%d, reason=%s", params.UID, params.OutTradeNo, params.IP, riskResult.Code, riskResult.Msg)
		return nil, errors.New(riskResult.Msg)
	}

	// 选择通道
	channel, err := s.SelectChannel(params.UID, params.Type, params.ChannelID)
	if err != nil {
		log.Printf("[submit_payment_failed] uid=%d, out_trade_no=%s, reason=select channel failed, error=%s", params.UID, params.OutTradeNo, err.Error())
		return nil, err
	}

	// 检查金额限制
	if channel.Paymin != "" {
		minMoney, _ := strconv.ParseFloat(channel.Paymin, 10)
		if params.Money < minMoney {
			log.Printf("[submit_payment_failed] uid=%d, out_trade_no=%s, money=%.2f, min_money=%s, reason=below minimum", params.UID, params.OutTradeNo, params.Money, channel.Paymin)
			return nil, errors.New("最低支付金额" + channel.Paymin)
		}
	}
	if channel.Paymax != "" {
		maxMoney, _ := strconv.ParseFloat(channel.Paymax, 10)
		if params.Money > maxMoney {
			log.Printf("[submit_payment_failed] uid=%d, out_trade_no=%s, money=%.2f, max_money=%s, reason=exceeds maximum", params.UID, params.OutTradeNo, params.Money, channel.Paymax)
			return nil, errors.New("最高支付金额" + channel.Paymax)
		}
	}

	// 检查域名授权
	if domain != "" && !s.orderSvc.CheckDomainAuth(params.UID, domain) {
		log.Printf("[submit_payment_failed] uid=%d, out_trade_no=%s, domain=%s, reason=domain not authorized")
		return nil, errors.New("域名未授权")
	}

	// 计算费率
	rate := channel.Rate
	if user.Mode == 1 {
		rate = rate + (100-rate)*0.5
	} else if user.Mode == 2 {
		rate = rate * 0.5
	}
	rate = clampRate(rate)

	// 计算金额
	getmoney := params.Money * (1 - rate/100)
	costrate := channel.Costrate
	if costrate == 0 {
		costrate = rate
	}
	costrate = clampRate(costrate)
	realmoney := params.Money * (1 - costrate/100)
	profitmoney := realmoney - getmoney

	// 创建订单
	tradeNo := s.orderSvc.GenTradeNo()

	order := &model.Order{
		TradeNo:     tradeNo,
		OutTradeNo:  params.OutTradeNo,
		UID:         params.UID,
		Type:        params.Type,
		Channel:     int(channel.ID),
		Name:        params.Name,
		Money:       params.Money,
		Realmoney:   realmoney,
		Getmoney:    getmoney,
		Profitmoney: profitmoney,
		NotifyURL:   params.NotifyURL,
		ReturnURL:   params.ReturnURL,
		Param:       params.Param,
		Addtime:     time.Now(),
		Date:        time.Now().Format("2006-01-02"),
		IP:          params.IP,
		Status:      model.OrderStatusPending,
		Notify:      0,
	}

	if err := config.DB.Create(order).Error; err != nil {
		log.Printf("[submit_payment_failed] uid=%d, out_trade_no=%s, reason=create order failed, error=%s", params.UID, params.OutTradeNo, err.Error())
		return nil, errors.New("创建订单失败")
	}

	// 加载插件并提交
	pluginName := channel.Plugin
	pluginHandler := plugin.GetHandler(pluginName)
	if pluginHandler == nil {
		log.Printf("[submit_payment_failed] uid=%d, out_trade_no=%s, plugin=%s, reason=plugin not found")
		return nil, errors.New("支付通道插件不存在")
	}

	method, err := resolveSubmitMethod(*channel, params)
	if err != nil {
		log.Printf("[submit_payment_failed] uid=%d, out_trade_no=%s, channel_id=%d, plugin=%s, paymethod=%s, reason=resolve method failed, error=%s", params.UID, params.OutTradeNo, channel.ID, channel.Plugin, channel.Paymethod, err.Error())
		return nil, err
	}

	// 构造插件参数
	pluginParams := map[string]interface{}{
		"trade_no":     tradeNo,
		"out_trade_no": params.OutTradeNo,
		"money":        params.Money,
		"name":         params.Name,
		"notify_url":   params.NotifyURL,
		"return_url":   params.ReturnURL,
		"param":        params.Param,
		"openid":       strings.TrimSpace(params.Openid),
		"ip":           params.IP,
		"device":       params.Device,
		"method":       method,
		"channel":      *channel,
	}

	// 调用插件提交
	result, err := pluginHandler.Submit(pluginParams)
	if err != nil {
		log.Printf("[submit_payment_failed] uid=%d, out_trade_no=%s, plugin=%s, reason=plugin submit failed, error=%s", params.UID, params.OutTradeNo, pluginName, err.Error())
		return nil, err
	}

	return map[string]interface{}{
		"trade_no": tradeNo,
		"result":   result,
		"order":    order,
	}, nil
}

// 扫码支付提交
func (s *PaymentService) SubmitScanPayment(params SubmitParams) (map[string]interface{}, error) {
	params.Method = "scan"
	return s.SubmitPayment(params)
}

// JSAPI支付
func (s *PaymentService) SubmitJSAPIPayment(params SubmitParams) (map[string]interface{}, error) {
	params.Method = "jsapi"
	return s.SubmitPayment(params)
}

// APP支付
func (s *PaymentService) SubmitAppPayment(params SubmitParams) (map[string]interface{}, error) {
	params.Method = "app"
	return s.SubmitPayment(params)
}

// H5支付
func (s *PaymentService) SubmitH5Payment(params SubmitParams) (map[string]interface{}, error) {
	params.Method = "wap"
	return s.SubmitPayment(params)
}

// 余额充值下单（创建 tid=2 订单）
func (s *PaymentService) SubmitRechargePayment(params SubmitParams) (map[string]interface{}, error) {
	user, err := s.authSvc.GetUser(params.UID)
	if err != nil {
		return nil, errors.New("商户不存在")
	}
	if user.Status != 1 {
		return nil, errors.New("商户已被禁用")
	}
	if user.Pay != 1 {
		return nil, errors.New("商户没有支付权限")
	}

	channel, err := s.SelectChannel(params.UID, params.Type, params.ChannelID)
	if err != nil {
		return nil, err
	}

	method, err := resolveRechargeSubmitMethod(*channel)
	if err != nil {
		return nil, err
	}

	outTradeNo := strings.TrimSpace(params.OutTradeNo)
	if outTradeNo == "" {
		outTradeNo = fmt.Sprintf("RECHARGE_%d_%d", params.UID, time.Now().UnixNano())
	}

	order, err := s.orderSvc.CreateOrder(
		params.UID,
		outTradeNo,
		params.Name,
		params.NotifyURL,
		params.ReturnURL,
		params.Param,
		params.Money,
		params.Type,
		int(channel.ID),
		params.IP,
	)
	if err != nil {
		return nil, err
	}

	// 标记为充值订单
	if err := config.DB.Model(&model.Order{}).Where("trade_no = ?", order.TradeNo).Update("tid", 2).Error; err != nil {
		return nil, errors.New("创建订单失败")
	}
	order.Tid = 2

	pluginHandler := plugin.GetHandler(channel.Plugin)
	if pluginHandler == nil {
		return nil, errors.New("支付通道插件不存在")
	}

	result, err := pluginHandler.Submit(map[string]interface{}{
		"trade_no":     order.TradeNo,
		"out_trade_no": order.OutTradeNo,
		"money":        order.Money,
		"name":         order.Name,
		"notify_url":   order.NotifyURL,
		"return_url":   order.ReturnURL,
		"param":        order.Param,
		"ip":           params.IP,
		"device":       params.Device,
		"method":       method,
		"channel":      *channel,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"trade_no": order.TradeNo,
		"result":   result,
		"order":    order,
	}, nil
}

// 支付回调处理
func (s *PaymentService) HandleNotify(tradeNo string, pluginName string, c *gin.Context) (map[string]interface{}, error) {
	pluginHandler := plugin.GetHandler(pluginName)
	if pluginHandler == nil {
		log.Printf("[handle_notify_failed] trade_no=%s, plugin=%s, reason=plugin not found", tradeNo, pluginName)
		return nil, errors.New("插件不存在")
	}

	// 调用插件处理回调
	result, err := pluginHandler.Notify(tradeNo, c)
	if err != nil {
		log.Printf("[handle_notify_failed] trade_no=%s, plugin=%s, reason=plugin notify failed, error=%s", tradeNo, pluginName, err.Error())
		return nil, err
	}

	if result.Success {
		// 更新订单状态
		err = s.orderSvc.OrderPaid(result.TradeNo, result.APITradeNo, result.Buyer)
		if err != nil {
			// 回调可能重复推送，若订单已处理则按幂等成功响应，避免微信持续重试。
			if strings.Contains(err.Error(), errOrderAlreadyProcessed.Error()) {
				return map[string]interface{}{
					"success": true,
					"message": "already processed",
				}, nil
			}
			log.Printf("[handle_notify_failed] trade_no=%s, plugin=%s, reason=order paid update failed, error=%s", tradeNo, pluginName, err.Error())
			return nil, err
		}
	}

	return map[string]interface{}{
		"success": result.Success,
		"message": result.Message,
	}, nil
}

// 同步回调处理
func (s *PaymentService) HandleReturn(tradeNo string, pluginName string, c *gin.Context) (plugin.ReturnResult, error) {
	pluginHandler := plugin.GetHandler(pluginName)
	if pluginHandler == nil {
		log.Printf("[handle_return_failed] trade_no=%s, plugin=%s, reason=plugin not found", tradeNo, pluginName)
		return plugin.ReturnResult{}, errors.New("插件不存在")
	}

	return pluginHandler.Return(tradeNo, c)
}

// 退款
func (s *PaymentService) Refund(tradeNo string, money float64) error {
	if err := s.orderSvc.Refund(tradeNo, money); err != nil {
		log.Printf("[refund_failed] trade_no=%s, money=%.2f, reason=order refund failed, error=%s", tradeNo, money, err.Error())
		return err
	}
	return nil
}

// 获取通道配置
func (s *PaymentService) GetChannelConfig(channelID int) (*model.Channel, error) {
	var channel model.Channel
	result := config.DB.First(&channel, channelID)
	if result.Error != nil {
		log.Printf("[get_channel_config_failed] channel_id=%d, reason=channel not found, error=%s", channelID, result.Error.Error())
		return nil, errors.New("通道不存在")
	}
	return &channel, nil
}

// 获取插件列表
func (s *PaymentService) GetPluginList() ([]model.Plugin, error) {
	var plugins []model.Plugin
	config.DB.Find(&plugins)
	return plugins, nil
}

// 获取通道列表
func (s *PaymentService) GetChannelList(typeID int) ([]model.Channel, error) {
	var channels []model.Channel
	query := config.DB.Where("status = 1")
	if typeID > 0 {
		query = query.Where("type = ?", typeID)
	}
	query.Find(&channels)
	return channels, nil
}

// 手动补单
func (s *PaymentService) ManualFillOrder(tradeNo, apiTradeNo, buyer string) error {
	return s.orderSvc.OrderPaid(tradeNo, apiTradeNo, buyer)
}

// config.Now() 需要定义
