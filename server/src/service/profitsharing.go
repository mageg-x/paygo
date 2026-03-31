package service

import (
	"errors"
	"log"
	"strconv"
	"time"

	"paygo/src/config"
	"paygo/src/model"
	"paygo/src/plugin"
)

// 分账服务
type ProfitService struct {
	orderSvc *OrderService
	authSvc  *AuthService
}

func NewProfitService() *ProfitService {
	return &ProfitService{
		orderSvc: NewOrderService(),
		authSvc:  NewAuthService(),
	}
}

// 分账状态常量
const (
	PsStatusPending   = 0 // 待分账
	PsStatusCompleted = 1 // 已分账
	PsStatusFailed    = 2 // 分账失败
)

// 执行订单分账
func (s *ProfitService) ProcessProfitSharing(tradeNo string) error {
	var order model.Order
	if config.DB.Where("trade_no = ?", tradeNo).First(&order).Error != nil {
		log.Printf("[profit_sharing_skipped] trade_no=%s, reason=order not found", tradeNo)
		return errors.New("订单不存在")
	}

	// 检查是否已分账
	if order.Profits == 1 {
		log.Printf("[profit_sharing_skipped] trade_no=%s, reason=already processed", tradeNo)
		return nil
	}

	// 检查订单状态
	if order.Status != model.OrderStatusPaid {
		log.Printf("[profit_sharing_skipped] trade_no=%s, reason=order not paid, status=%d", tradeNo, order.Status)
		return errors.New("订单未支付")
	}

	// 获取商户的分账接收人
	var receivers []model.PsReceiver
	config.DB.Where("uid = ? AND status = 1", order.UID).Find(&receivers)

	if len(receivers) == 0 {
		log.Printf("[profit_sharing_skipped] trade_no=%s, reason=no receivers configured", tradeNo)
		return nil
	}

	// 获取通道信息
	var channel model.Channel
	if config.DB.First(&channel, order.Channel).Error != nil {
		log.Printf("[profit_sharing_failed] trade_no=%s, reason=channel not found", tradeNo)
		return errors.New("通道不存在")
	}

	// 获取通道插件
	pluginHandler := plugin.GetHandler(channel.Plugin)
	if pluginHandler == nil {
		log.Printf("[profit_sharing_failed] trade_no=%s, plugin=%s, reason=plugin not found", tradeNo, channel.Plugin)
		return errors.New("支付通道插件不存在")
	}

	// 检查插件是否支持分账
	pluginInfo := pluginHandler.GetInfo()
	if !containsString(pluginInfo.Transtypes, "profitsharing") && !containsString(pluginInfo.Transtypes, "ps") {
		log.Printf("[profit_sharing_skipped] trade_no=%s, plugin=%s, reason=plugin not support profitsharing", tradeNo, channel.Plugin)
		return nil
	}

	// 执行分账
	successCount := 0
	failCount := 0

	for _, receiver := range receivers {
		result, err := s.executeProfitSharing(order, channel, receiver, pluginHandler)
		if err != nil {
			log.Printf("[profit_sharing_failed] trade_no=%s, receiver=%d, reason=%s", tradeNo, receiver.ID, err.Error())
			failCount++
			continue
		}

		if result {
			successCount++
		}
	}

	// 更新订单分账状态
	if failCount == 0 && successCount > 0 {
		config.DB.Model(&order).Update("profits", 1)
		log.Printf("[profit_sharing_success] trade_no=%s, count=%d", tradeNo, successCount)
	} else if successCount > 0 {
		config.DB.Model(&order).Update("profits", 2) // 部分分账
		log.Printf("[profit_sharing_partial] trade_no=%s, success=%d, fail=%d", tradeNo, successCount, failCount)
	} else {
		log.Printf("[profit_sharing_all_failed] trade_no=%s, count=%d", tradeNo, failCount)
	}

	return nil
}

// 执行单笔分账
func (s *ProfitService) executeProfitSharing(order model.Order, channel model.Channel, receiver model.PsReceiver, pluginHandler plugin.Plugin) (bool, error) {
	// 计算分账金额
	rate, _ := strconv.ParseFloat(receiver.Rate, 10)
	if rate <= 0 || rate > 100 {
		log.Printf("[execute_profit_sharing_failed] receiver=%d, reason=invalid rate=%s", receiver.ID, receiver.Rate)
		return false, errors.New("分账比例无效")
	}

	// 检查最低分账金额
	if receiver.Minmoney != "" {
		minMoney, _ := strconv.ParseFloat(receiver.Minmoney, 10)
		if order.Realmoney < minMoney {
			log.Printf("[execute_profit_sharing_skipped] receiver=%d, reason=below min_money", receiver.ID)
			return false, nil
		}
	}

	profitMoney := order.Realmoney * rate / 100

	// 构造分账参数
	params := map[string]interface{}{
		"trade_no":     order.TradeNo,
		"api_trade_no": order.ApiTradeNo,
		"money":        profitMoney,
		"account":      receiver.Account,
		"name":         receiver.Name,
		"channel":      channel,
	}

	// 调用插件分账接口
	// 注意：部分插件可能没有实现分账接口，需要根据实际情况处理
	result, err := pluginHandler.Transfer(params)
	if err != nil {
		log.Printf("[execute_profit_sharing_failed] receiver=%d, reason=plugin error, error=%s", receiver.ID, err.Error())
		return false, err
	}

	if result.Code != 0 {
		log.Printf("[execute_profit_sharing_failed] receiver=%d, reason=plugin rejected, code=%d, msg=%s", receiver.ID, result.Code, result.ErrMsg)
		return false, errors.New(result.ErrMsg)
	}

	// 记录分账订单
	psOrder := &model.PsOrder{
		RID:        int(receiver.ID),
		TradeNo:    order.TradeNo,
		ApiTradeNo: order.ApiTradeNo,
		Money:      profitMoney,
		Status:     PsStatusCompleted,
		Result:     result.OrderID,
		Addtime:    time.Now(),
	}
	config.DB.Create(psOrder)

	// 给接收人加款
	var receiverUser model.User
	if config.DB.Where("uid = ?", receiver.UID).First(&receiverUser).Error == nil {
		oldMoney := receiverUser.Money
		newMoney := oldMoney + profitMoney
		config.DB.Model(&receiverUser).Update("money", newMoney)

		// 记录资金变动
		record := &model.Record{
			UID:      receiver.UID,
			Action:   9, // 分账收入
			Money:    profitMoney,
			Oldmoney: oldMoney,
			Newmoney: newMoney,
			Type:     "profitsharing",
			TradeNo:  order.TradeNo,
			Date:     time.Now(),
		}
		config.DB.Create(record)
	}

	log.Printf("[execute_profit_sharing_success] receiver=%d, money=%.2f, order_id=%s", receiver.ID, profitMoney, result.OrderID)
	return true, nil
}

// 获取商户分账接收人列表
func (s *ProfitService) GetReceivers(uid uint) ([]model.PsReceiver, error) {
	var receivers []model.PsReceiver
	config.DB.Where("uid = ?", uid).Find(&receivers)
	return receivers, nil
}

// 添加分账接收人
func (s *ProfitService) AddReceiver(uid uint, account, name, rate, minmoney string, channelID int) (*model.PsReceiver, error) {
	rateVal, _ := strconv.ParseFloat(rate, 10)
	if rateVal <= 0 || rateVal > 100 {
		return nil, errors.New("分账比例无效")
	}

	receiver := &model.PsReceiver{
		Channel:  channelID,
		UID:      uid,
		Account:  account,
		Name:     name,
		Rate:     rate,
		Minmoney: minmoney,
		Status:   1,
		Addtime:  time.Now(),
	}

	if err := config.DB.Create(receiver).Error; err != nil {
		log.Printf("[add_receiver_failed] uid=%d, account=%s, error=%s", uid, account, err.Error())
		return nil, err
	}

	log.Printf("[add_receiver_success] id=%d, uid=%d, account=%s, rate=%s", receiver.ID, uid, account, rate)
	return receiver, nil
}

// 删除分账接收人
func (s *ProfitService) DeleteReceiver(id, uid uint) error {
	result := config.DB.Where("id = ? AND uid = ?", id, uid).Delete(&model.PsReceiver{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("接收人不存在或无权删除")
	}
	log.Printf("[delete_receiver_success] id=%d, uid=%d", id, uid)
	return nil
}

// 辅助函数：检查字符串数组是否包含某字符串
func containsString(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

// 获取分账订单列表
func (s *ProfitService) GetPsOrders(tradeNo string) ([]model.PsOrder, error) {
	var orders []model.PsOrder
	config.DB.Where("trade_no = ?", tradeNo).Find(&orders)
	return orders, nil
}