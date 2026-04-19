package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gopay/src/config"
	"gopay/src/model"
	"gopay/src/plugin"
)

type OrderService struct {
	authSvc *AuthService
}

var merchantNotifyHTTPClient = &http.Client{
	Timeout: 8 * time.Second,
}

func clampRate(rate float64) float64 {
	if rate < 0 {
		return 0
	}
	if rate > 100 {
		return 100
	}
	return rate
}

func NewOrderService() *OrderService {
	return &OrderService{
		authSvc: NewAuthService(),
	}
}

func isBlockedCallbackHost(host string) bool {
	h := strings.TrimSpace(host)
	if h == "" {
		return true
	}
	if strings.EqualFold(h, "localhost") {
		return true
	}
	ip := net.ParseIP(h)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
		return true
	}
	return false
}

func validateNotifyURL(raw string) error {
	v := strings.TrimSpace(raw)
	if v == "" {
		return errors.New("empty notify url")
	}
	u, err := url.Parse(v)
	if err != nil || u == nil {
		return errors.New("invalid notify url")
	}
	scheme := strings.ToLower(strings.TrimSpace(u.Scheme))
	if scheme != "http" && scheme != "https" {
		return errors.New("notify url scheme must be http/https")
	}
	host := strings.TrimSpace(u.Hostname())
	if host == "" {
		return errors.New("notify url host empty")
	}
	if isBlockedCallbackHost(host) {
		return errors.New("notify url host is blocked")
	}
	return nil
}

// 生成订单号
func (s *OrderService) GenTradeNo() string {
	now := time.Now()
	// 格式: YYYYMMDDHHMMSS + 6位随机数
	randNum := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	return fmt.Sprintf("%s%s", now.Format("20060102150405"), randNum)
}

// 创建订单
func (s *OrderService) CreateOrder(uid uint, outTradeNo, name, notifyURL, returnURL, param string,
	money float64, payType int, channelID int, ip string) (*model.Order, error) {

	// 检查是否启用测试支付
	testOpen := s.authSvc.GetConfig("test_open")
	isTestPay := testOpen == "1"

	// 测试支付时使用指定的收款商户
	if isTestPay {
		testPayUid := s.authSvc.GetConfig("test_pay_uid")
		if testPayUid != "" {
			uidStr := strings.TrimSpace(testPayUid)
			if parsedUid, err := strconv.ParseUint(uidStr, 10, 32); err == nil {
				uid = uint(parsedUid)
				log.Printf("[test_pay] using test merchant uid=%d", uid)
			}
		}
	}

	// 获取商户信息
	user, err := s.authSvc.GetUser(uid)
	if err != nil {
		log.Printf("[create_order_failed] uid=%d, out_trade_no=%s, reason=merchant not found, error=%s", uid, outTradeNo, err.Error())
		return nil, errors.New("商户不存在")
	}

	if user.Status != 1 {
		log.Printf("[create_order_failed] uid=%d, out_trade_no=%s, reason=merchant disabled, status=%d", uid, outTradeNo, user.Status)
		return nil, errors.New("商户已被禁用")
	}

	if user.Pay != 1 {
		log.Printf("[create_order_failed] uid=%d, out_trade_no=%s, reason=merchant no pay permission")
		return nil, errors.New("商户没有支付权限")
	}

	// 获取通道信息
	var channel model.Channel
	result := config.DB.First(&channel, channelID)
	if result.Error != nil {
		log.Printf("[create_order_failed] uid=%d, out_trade_no=%s, channel_id=%d, reason=channel not found, error=%s", uid, outTradeNo, channelID, result.Error.Error())
		return nil, errors.New("通道不存在")
	}

	if channel.Status != 1 {
		log.Printf("[create_order_failed] uid=%d, out_trade_no=%s, channel_id=%d, reason=channel disabled", uid, outTradeNo, channelID)
		return nil, errors.New("通道已关闭")
	}

	// 检查金额限制
	if channel.Paymin != "" {
		minMoney, _ := strconv.ParseFloat(channel.Paymin, 10)
		if money < minMoney {
			log.Printf("[create_order_failed] uid=%d, out_trade_no=%s, money=%.2f, min_money=%.2f, reason=below minimum amount", uid, outTradeNo, money, minMoney)
			return nil, fmt.Errorf("最低支付金额%.2f", minMoney)
		}
	}
	if channel.Paymax != "" {
		maxMoney, _ := strconv.ParseFloat(channel.Paymax, 10)
		if money > maxMoney {
			log.Printf("[create_order_failed] uid=%d, out_trade_no=%s, money=%.2f, max_money=%.2f, reason=exceeds maximum amount", uid, outTradeNo, money, maxMoney)
			return nil, fmt.Errorf("最高支付金额%.2f", maxMoney)
		}
	}

	// 检查用户组费率
	var group model.Group
	config.DB.First(&group, user.GID)

	// 计算实际金额和商户可得
	// 费率 = 通道费率 + 用户组加成
	rate := channel.Rate
	if user.Mode == 1 {
		// 加费模式
		rate = rate + (100-rate)*0.5
	} else if user.Mode == 2 {
		// 减费模式
		rate = rate * 0.5
	}

	// 计算商户可得
	rate = clampRate(rate)
	getmoney := money * (1 - rate/100)

	// 平台实收（扣除上游成本费率后）
	costrate := channel.Costrate
	if costrate == 0 {
		costrate = rate
	}
	costrate = clampRate(costrate)
	realmoney := money * (1 - costrate/100)
	profitmoney := realmoney - getmoney

	order := &model.Order{
		TradeNo:     s.GenTradeNo(),
		OutTradeNo:  outTradeNo,
		UID:         uid,
		Tid:         0,
		Type:        payType,
		Channel:     channelID,
		Name:        name,
		Money:       money,
		Realmoney:   realmoney,
		Getmoney:    getmoney,
		Profitmoney: profitmoney,
		NotifyURL:   notifyURL,
		ReturnURL:   returnURL,
		Param:       param,
		Addtime:     time.Now(),
		Date:        time.Now().Format("2006-01-02"),
		IP:          ip,
		Status:      model.OrderStatusPending,
		Notify:      0,
		Invite:      user.Upid,
		Subchannel:  0,
		Version:     0,
	}

	// 如果有扩展信息
	if param != "" {
		order.Param = param
	}

	result = config.DB.Create(order)
	if result.Error != nil {
		log.Printf("[create_order_failed] uid=%d, out_trade_no=%s, reason=create order failed, error=%s", uid, outTradeNo, result.Error.Error())
		return nil, errors.New("创建订单失败")
	}

	return order, nil
}

// 订单支付成功回调
func (s *OrderService) OrderPaid(tradeNo, apiTradeNo, buyer string) error {
	var order model.Order
	result := config.DB.Where("trade_no = ? AND status = ?", tradeNo, model.OrderStatusPending).First(&order)
	if result.Error != nil {
		log.Printf("[order_paid_failed] trade_no=%s, reason=order not found or already processed, error=%s", tradeNo, result.Error.Error())
		return errors.New("订单不存在或已处理")
	}

	// 更新订单状态
	now := time.Now()
	updateRes := config.DB.Model(&order).Updates(map[string]interface{}{
		"status":       model.OrderStatusPaid,
		"api_trade_no": apiTradeNo,
		"buyer":        buyer,
		"endtime":      now,
		"notifytime":   now,
	})
	if updateRes.Error != nil {
		log.Printf("[order_paid_failed] trade_no=%s, reason=update order status failed, error=%s", tradeNo, updateRes.Error.Error())
		return errors.New("订单状态更新失败")
	}
	if updateRes.RowsAffected == 0 {
		log.Printf("[order_paid_failed] trade_no=%s, reason=update order status affected 0 rows", tradeNo)
		return errors.New("订单状态更新失败")
	}

	// 根据订单类型处理
	switch order.Tid {
	case 1:
		// 商户注册 - 给推荐人返现
		s.handleUserRegister(order, now)
	case 2:
		// 充值余额
		s.handleRecharge(order, now)
	case 3:
		// 聚合收款
		s.handleCombinePayment(order, now)
	case 4:
		// 购买用户组
		s.handleBuyGroup(order, now)
	default:
		// 普通订单 - 增加余额
		s.handleNormalOrder(order, now)
	}

	// 邀请人奖励（非充值订单）
	if order.Invite > 0 && order.Tid != 2 {
		s.addInviteMoney(order.Invite, order.UID, order.Money, tradeNo)
	}

	// 执行分账（异步）
	go s.executeProfitSharing(tradeNo)

	// 通知商户（异步）
	go s.notifyMerchant(order)

	return nil
}

// 执行分账
func (s *OrderService) executeProfitSharing(tradeNo string) {
	profitSvc := NewProfitService()
	if err := profitSvc.ProcessProfitSharing(tradeNo); err != nil {
		log.Printf("[execute_profit_sharing_error] trade_no=%s, error=%s", tradeNo, err.Error())
	}
}

// 处理商户注册订单 (tid=1)
func (s *OrderService) handleUserRegister(order model.Order, now time.Time) {
	var user model.User
	if config.DB.First(&user, order.UID).Error != nil {
		return
	}

	// 商户注册成功，给推荐人返现
	if order.Invite > 0 {
		s.addInviteMoney(order.Invite, order.UID, order.Money, order.TradeNo)
	}

	log.Printf("[order_paid_success] trade_no=%s, tid=1, uid=%d, type=user_register", order.TradeNo, order.UID)
}

// 处理充值订单 (tid=2)
func (s *OrderService) handleRecharge(order model.Order, now time.Time) {
	var user model.User
	if config.DB.First(&user, order.UID).Error != nil {
		return
	}

	oldMoney := user.Money
	newMoney := oldMoney + order.Money

	config.DB.Model(&user).Update("money", newMoney)

	// 记录资金变动
	record := &model.Record{
		UID:      order.UID,
		Action:   2, // 充值
		Money:    order.Money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "recharge",
		TradeNo:  order.TradeNo,
		Date:     now,
	}
	config.DB.Create(record)

	log.Printf("[order_paid_success] trade_no=%s, tid=2, uid=%d, money=%.2f", order.TradeNo, order.UID, order.Money)
}

// 处理聚合收款订单 (tid=3)
func (s *OrderService) handleCombinePayment(order model.Order, now time.Time) {
	var user model.User
	if config.DB.First(&user, order.UID).Error != nil {
		return
	}

	oldMoney := user.Money
	newMoney := oldMoney + order.Getmoney

	config.DB.Model(&user).Update("money", newMoney)

	// 记录资金变动
	record := &model.Record{
		UID:      order.UID,
		Action:   1, // 订单收入
		Money:    order.Getmoney,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "combine",
		TradeNo:  order.TradeNo,
		Date:     now,
	}
	config.DB.Create(record)

	log.Printf("[order_paid_success] trade_no=%s, tid=3, uid=%d, money=%.2f, getmoney=%.2f", order.TradeNo, order.UID, order.Money, order.Getmoney)
}

// 处理购买用户组订单 (tid=4)
func (s *OrderService) handleBuyGroup(order model.Order, now time.Time) {
	var user model.User
	if config.DB.First(&user, order.UID).Error != nil {
		return
	}

	// 从订单参数中解析目标用户组
	// param 格式: gid|days 或直接是 gid
	var newGID uint = 1
	var days int = 0

	if order.Param != "" {
		parts := strings.Split(order.Param, "|")
		if len(parts) >= 1 {
			gid, _ := strconv.ParseUint(parts[0], 10, 32)
			newGID = uint(gid)
		}
		if len(parts) >= 2 {
			days, _ = strconv.Atoi(parts[1])
		}
	}

	oldMoney := user.Money
	oldGID := user.GID

	// 扣除购买费用
	newMoney := oldMoney - order.Money
	config.DB.Model(&user).Updates(map[string]interface{}{
		"money": newMoney,
		"gid":   newGID,
	})

	// 记录资金变动
	record := &model.Record{
		UID:      order.UID,
		Action:   5, // 购买用户组
		Money:    -order.Money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "buygroup",
		TradeNo:  order.TradeNo,
		Date:     now,
	}
	config.DB.Create(record)

	log.Printf("[order_paid_success] trade_no=%s, tid=4, uid=%d, old_gid=%d, new_gid=%d, days=%d", order.TradeNo, order.UID, oldGID, newGID, days)
}

// 处理普通订单 (tid=0或其他)
func (s *OrderService) handleNormalOrder(order model.Order, now time.Time) {
	var user model.User
	if config.DB.First(&user, order.UID).Error != nil {
		return
	}

	oldMoney := user.Money
	newMoney := oldMoney + order.Getmoney

	config.DB.Model(&user).Update("money", newMoney)

	// 记录资金变动
	record := &model.Record{
		UID:      order.UID,
		Action:   1, // 订单收入
		Money:    order.Getmoney,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "order",
		TradeNo:  order.TradeNo,
		Date:     now,
	}
	config.DB.Create(record)

	log.Printf("[order_paid_success] trade_no=%s, tid=%d, uid=%d, money=%.2f, getmoney=%.2f", order.TradeNo, order.Tid, order.UID, order.Money, order.Getmoney)
}

// 添加邀请奖励
func (s *OrderService) addInviteMoney(inviteUID, uid uint, money float64, tradeNo string) {
	// 获取邀请人信息
	var inviteUser model.User
	if config.DB.First(&inviteUser, inviteUID).Error != nil {
		return
	}

	// 获取用户组配置
	var group model.Group
	if config.DB.First(&group, inviteUser.GID).Error != nil {
		return
	}

	// 计算奖励比例（默认1%）
	rate := 0.01
	if group.Settings != "" {
		var settings map[string]interface{}
		json.Unmarshal([]byte(group.Settings), &settings)
		if v, ok := settings["invite_rate"]; ok {
			rate, _ = strconv.ParseFloat(fmt.Sprintf("%v", v), 10)
		}
	}

	reward := money * rate

	oldMoney := inviteUser.Money
	newMoney := oldMoney + reward

	config.DB.Model(&inviteUser).Update("money", newMoney)

	// 记录
	record := &model.Record{
		UID:      inviteUID,
		Action:   7, // 邀请返现
		Money:    reward,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "invite",
		TradeNo:  tradeNo,
		Date:     time.Now(),
	}
	config.DB.Create(record)
}

// 通知商户
func (s *OrderService) notifyMerchant(order model.Order) {
	if order.NotifyURL == "" {
		log.Printf("[notify_merchant_skipped] trade_no=%s, reason=empty notify_url", order.TradeNo)
		return
	}
	if err := validateNotifyURL(order.NotifyURL); err != nil {
		log.Printf("[notify_merchant_failed] trade_no=%s, url=%s, reason=unsafe notify_url, error=%s", order.TradeNo, order.NotifyURL, err.Error())
		s.markNotifyFailed(order.TradeNo)
		return
	}

	// 构造通知数据
	params := map[string]string{
		"trade_no":     order.TradeNo,
		"out_trade_no": order.OutTradeNo,
		"type":         strconv.Itoa(order.Type),
		"status":       "1",
		"money":        strconv.FormatFloat(order.Money, 'f', 2, 64),
		"realmoney":    strconv.FormatFloat(order.Realmoney, 'f', 2, 64),
	}

	// 获取商户密钥进行签名
	var user model.User
	if config.DB.First(&user, order.UID).Error != nil {
		log.Printf("[notify_merchant_failed] trade_no=%s, reason=user not found", order.TradeNo)
		return
	}

	params["sign"] = s.authSvc.MakeSign(params, user.Key)

	// 发送HTTP POST通知
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}
	resp, err := merchantNotifyHTTPClient.PostForm(order.NotifyURL, formData)
	if err != nil {
		log.Printf("[notify_merchant_failed] trade_no=%s, url=%s, error=%s", order.TradeNo, order.NotifyURL, err.Error())
		// 记录失败，后续重试
		s.markNotifyFailed(order.TradeNo)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, _ := io.ReadAll(resp.Body)
	result := string(body)

	// 检查是否成功（需返回 success 或 SUCCESS）
	if strings.Contains(result, "success") || strings.Contains(result, "SUCCESS") {
		log.Printf("[notify_merchant_success] trade_no=%s, url=%s", order.TradeNo, order.NotifyURL)
		s.markNotifySuccess(order.TradeNo)
	} else {
		log.Printf("[notify_merchant_failed] trade_no=%s, url=%s, response=%s", order.TradeNo, order.NotifyURL, result)
		s.markNotifyFailed(order.TradeNo)
	}
}

// 手动重试回调通知
func (s *OrderService) RetryNotify(tradeNo string) error {
	order, err := s.GetOrder(tradeNo)
	if err != nil {
		return err
	}
	if order.Status != model.OrderStatusPaid && order.Status != model.OrderStatusRefunded {
		return errors.New("订单状态不支持通知")
	}

	config.DB.Model(&model.Order{}).Where("trade_no = ?", tradeNo).Updates(map[string]interface{}{
		"notify":     0,
		"notifytime": time.Now(),
	})

	go s.notifyMerchant(*order)
	return nil
}

// 标记通知成功
func (s *OrderService) markNotifySuccess(tradeNo string) {
	config.DB.Model(&model.Order{}).Where("trade_no = ?", tradeNo).Updates(map[string]interface{}{
		"notify":     1,
		"notifytime": time.Now(),
	})
}

// 标记通知失败
func (s *OrderService) markNotifyFailed(tradeNo string) {
	config.DB.Model(&model.Order{}).Where("trade_no = ?", tradeNo).Update("notify", 2)
}

// 订单退款
func (s *OrderService) Refund(tradeNo string, money float64) error {
	if money <= 0 {
		return errors.New("退款金额必须大于0")
	}

	var order model.Order
	result := config.DB.Where("trade_no = ?", tradeNo).First(&order)
	if result.Error != nil {
		log.Printf("[refund_failed] trade_no=%s, money=%.2f, reason=order not found, error=%s", tradeNo, money, result.Error.Error())
		return errors.New("订单不存在")
	}

	if order.Status != model.OrderStatusPaid {
		log.Printf("[refund_failed] trade_no=%s, money=%.2f, reason=invalid order status, status=%d", tradeNo, money, order.Status)
		return errors.New("订单状态不允许退款")
	}

	// 退款金额不能超过订单金额
	if order.Refundmoney+money > order.Money {
		log.Printf("[refund_failed] trade_no=%s, money=%.2f, reason=refund amount exceeds total, refundmoney=%.2f, total=%.2f", tradeNo, money, order.Refundmoney, order.Money)
		return errors.New("退款金额超过订单金额")
	}

	// 获取商户
	var user model.User
	config.DB.First(&user, order.UID)

	// 本地校验提前做，避免上游退款成功但本地记账失败
	if order.Tid != 2 {
		availableRefund := order.Money - order.Refundmoney
		if money > availableRefund {
			log.Printf("[refund_failed] trade_no=%s, money=%.2f, reason=refund exceeds available money, available=%.2f", tradeNo, money, availableRefund)
			return errors.New("退款金额超过可退金额")
		}

		if user.Money < money {
			log.Printf("[refund_failed] trade_no=%s, money=%.2f, reason=insufficient balance, balance=%.2f", tradeNo, money, user.Money)
			return errors.New("商户余额不足")
		}
	}

	// 先调用上游退款，只有上游成功才落本地账务
	var channel model.Channel
	if err := config.DB.First(&channel, order.Channel).Error; err != nil {
		log.Printf("[refund_failed] trade_no=%s, money=%.2f, reason=channel not found, channel_id=%d, error=%s", tradeNo, money, order.Channel, err.Error())
		return errors.New("通道不存在")
	}

	pluginHandler := plugin.GetHandler(channel.Plugin)
	if pluginHandler == nil {
		log.Printf("[refund_failed] trade_no=%s, money=%.2f, reason=plugin not found, plugin=%s", tradeNo, money, channel.Plugin)
		return errors.New("退款插件不存在")
	}

	refundResult, err := pluginHandler.Refund(map[string]interface{}{
		"trade_no": tradeNo,
		"money":    money,
		"channel":  channel,
	})
	if err != nil {
		log.Printf("[refund_failed] trade_no=%s, money=%.2f, reason=plugin refund failed, plugin=%s, error=%s", tradeNo, money, channel.Plugin, err.Error())
		return err
	}
	if refundResult.Code != 0 {
		msg := strings.TrimSpace(refundResult.ErrMsg)
		if msg == "" {
			msg = "上游退款失败"
		}
		log.Printf("[refund_failed] trade_no=%s, money=%.2f, reason=plugin refund rejected, plugin=%s, err_msg=%s", tradeNo, money, channel.Plugin, msg)
		return errors.New(msg)
	}

	// 扣除商户余额（普通订单和充值订单处理不同）
	oldMoney := user.Money
	var newMoney float64

	tx := config.DB.Begin()

	// 充值订单退款：原路退回，不从余额扣除
	// 普通订单退款：从商户可得金额中扣除
	if order.Tid == 2 {
		// 充值订单退款
		refundmoney := order.Refundmoney + money
		tx.Model(&order).Update("refundmoney", refundmoney)

		// 如果完全退款，更新状态
		if refundmoney >= order.Money {
			tx.Model(&order).Update("status", model.OrderStatusRefunded)
		}

		// 记录资金变动（退还充值）
		record := &model.Record{
			UID:      order.UID,
			Action:   8, // 充值退款
			Money:    -money,
			Oldmoney: oldMoney,
			Newmoney: oldMoney, // 余额不变，因为是原路退回
			Type:     "refund",
			TradeNo:  tradeNo,
			Date:     time.Now(),
		}
		tx.Create(record)
	} else {
		// 普通订单退款：按退款金额从商户余额扣除
		availableRefund := order.Money - order.Refundmoney
		if money > availableRefund {
			log.Printf("[refund_failed] trade_no=%s, money=%.2f, reason=refund exceeds available money, available=%.2f", tradeNo, money, availableRefund)
			tx.Rollback()
			return errors.New("退款金额超过可退金额")
		}

		if user.Money < money {
			log.Printf("[refund_failed] trade_no=%s, money=%.2f, reason=insufficient balance, balance=%.2f", tradeNo, money, user.Money)
			tx.Rollback()
			return errors.New("商户余额不足")
		}

		newMoney = oldMoney - money

		// 扣除余额
		tx.Model(&user).Update("money", newMoney)

		// 更新退款金额
		refundmoney := order.Refundmoney + money
		tx.Model(&order).Update("refundmoney", refundmoney)

		// 如果完全退款，更新状态
		if refundmoney >= order.Money {
			tx.Model(&order).Update("status", model.OrderStatusRefunded)
		}

		// 记录资金变动
		record := &model.Record{
			UID:      order.UID,
			Action:   4, // 退款
			Money:    -money,
			Oldmoney: oldMoney,
			Newmoney: newMoney,
			Type:     "refund",
			TradeNo:  tradeNo,
			Date:     time.Now(),
		}
		tx.Create(record)
	}

	// 创建退款订单记录
	refundNo := fmt.Sprintf("R%s", tradeNo)
	refundOrder := &model.RefundOrder{
		RefundNo:    refundNo,
		OutRefundNo: fmt.Sprintf("R%s_%d", tradeNo, int(time.Now().Unix())),
		TradeNo:     tradeNo,
		UID:         order.UID,
		Money:       money,
		Reducemoney: 0,
		Status:      1, // 退款成功
		Addtime:     time.Now(),
		Endtime:     time.Now(),
	}
	tx.Create(refundOrder)

	if err := tx.Commit().Error; err != nil {
		log.Printf("[refund_failed] trade_no=%s, money=%.2f, reason=commit failed, error=%s", tradeNo, money, err.Error())
		return errors.New("退款记账失败")
	}

	log.Printf("[refund_success] trade_no=%s, money=%.2f, refund_no=%s", tradeNo, money, refundNo)
	return nil
}

// 冻结订单
func (s *OrderService) Freeze(tradeNo string) error {
	return config.DB.Model(&model.Order{}).Where("trade_no = ?", tradeNo).
		Update("status", model.OrderStatusFrozen).Error
}

// 解冻订单
func (s *OrderService) Unfreeze(tradeNo string) error {
	return config.DB.Model(&model.Order{}).Where("trade_no = ? AND status = ?", tradeNo, model.OrderStatusFrozen).
		Update("status", model.OrderStatusPaid).Error
}

// 查询订单
func (s *OrderService) GetOrder(tradeNo string) (*model.Order, error) {
	var order model.Order
	result := config.DB.Where("trade_no = ?", tradeNo).First(&order)
	if result.Error != nil {
		log.Printf("[get_order_failed] trade_no=%s, reason=order not found, error=%s", tradeNo, result.Error.Error())
		return nil, errors.New("订单不存在")
	}
	return &order, nil
}

// 按商户订单号查询
func (s *OrderService) GetOrderByOutTradeNo(outTradeNo string, uid uint) (*model.Order, error) {
	var order model.Order
	result := config.DB.Where("out_trade_no = ? AND uid = ?", outTradeNo, uid).First(&order)
	if result.Error != nil {
		log.Printf("[get_order_failed] out_trade_no=%s, uid=%d, reason=order not found, error=%s", outTradeNo, uid, result.Error.Error())
		return nil, errors.New("订单不存在")
	}
	return &order, nil
}

// 获取商户订单列表
func (s *OrderService) GetUserOrders(uid uint, status int, page, pageSize int, tradeNo string) ([]model.Order, int64, error) {
	orders := make([]model.Order, 0)
	var total int64

	query := config.DB.Model(&model.Order{}).Where("uid = ?", uid)

	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if strings.TrimSpace(tradeNo) != "" {
		query = query.Where("trade_no LIKE ?", "%"+strings.TrimSpace(tradeNo)+"%")
	}

	query.Count(&total)

	result := query.Order("addtime DESC, trade_no DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return orders, total, nil
}

// 获取订单类型名称
func (s *OrderService) GetTypeName(typeID int) string {
	var payType model.PayType
	if config.DB.First(&payType, typeID).Error == nil {
		return payType.Showname
	}
	return "未知"
}

// 检查订单是否超时（未支付订单超过30分钟）
func (s *OrderService) IsOrderTimeout(order model.Order) bool {
	if order.Status != model.OrderStatusPending {
		return false
	}
	return time.Since(order.Addtime) > 30*time.Minute
}

// 删除超时未支付订单
func (s *OrderService) CleanTimeoutOrders() (int64, error) {
	timeout := time.Now().Add(-30 * time.Minute)
	result := config.DB.Where("status = ? AND addtime < ?", model.OrderStatusPending, timeout).Delete(&model.Order{})
	return result.RowsAffected, result.Error
}

// 获取订单统计
func (s *OrderService) GetOrderStats(uid uint, startDate, endDate string) (map[string]interface{}, error) {
	type Stats struct {
		Total      float64
		Count      int64
		Today      float64
		TodayCount int64
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	// 总计
	var totalMoney float64
	var totalCount int64

	config.DB.Model(&model.Order{}).Where("uid = ? AND status = ?", uid, model.OrderStatusPaid).Select("COALESCE(SUM(money), 0)").Scan(&totalMoney)
	config.DB.Model(&model.Order{}).Where("uid = ? AND status = ?", uid, model.OrderStatusPaid).Count(&totalCount)

	var todayMoney float64
	var todayCount int64
	config.DB.Model(&model.Order{}).Where("uid = ? AND status = ? AND date = ?", uid, model.OrderStatusPaid, today).Select("COALESCE(SUM(money), 0)").Scan(&todayMoney)
	config.DB.Model(&model.Order{}).Where("uid = ? AND status = ? AND date = ?", uid, model.OrderStatusPaid, today).Count(&todayCount)

	return map[string]interface{}{
		"total_money": totalMoney,
		"total_count": totalCount,
		"today_money": todayMoney,
		"today_count": todayCount,
	}, nil
}

// 检查黑名单
func (s *OrderService) IsBlacklisted(ip string) bool {
	var count int64
	config.DB.Model(&model.Blacklist{}).Where("type = 0 AND content = ?", ip).Count(&count)
	return count > 0
}

// 检查域名授权
func (s *OrderService) CheckDomainAuth(uid uint, domain string) bool {
	if domain == "" {
		return true
	}

	var count int64
	config.DB.Model(&model.Domain{}).Where("uid = ? AND domain = ? AND status = 1", uid, domain).Count(&count)
	return count > 0
}

// 获取订单详情（包含关联信息）
func (s *OrderService) GetOrderDetail(tradeNo string) (map[string]interface{}, error) {
	var order model.Order
	if config.DB.First(&order, tradeNo).Error != nil {
		log.Printf("[get_order_detail_failed] trade_no=%s, reason=order not found", tradeNo)
		return nil, errors.New("订单不存在")
	}

	// 获取商户信息
	var user model.User
	config.DB.First(&user, order.UID)

	// 获取通道信息
	var channel model.Channel
	config.DB.First(&channel, order.Channel)

	// 获取支付类型
	var payType model.PayType
	config.DB.First(&payType, order.Type)

	detail := map[string]interface{}{
		"order":    order,
		"user":     user.Username,
		"channel":  channel.Name,
		"typename": payType.Showname,
	}

	return detail, nil
}
