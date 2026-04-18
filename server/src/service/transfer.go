package service

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"paygo/src/config"
	"paygo/src/model"
	"paygo/src/plugin"
)

type TransferService struct {
	authSvc *AuthService
}

func NewTransferService() *TransferService {
	return &TransferService{
		authSvc: NewAuthService(),
	}
}

func parseTransferFeeRate() float64 {
	raw := strings.TrimSpace(config.Get("transfer_fee"))
	if raw == "" {
		return 0
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0
	}
	if v < 0 {
		return 0
	}
	return v / 100
}

// 创建转账
func (s *TransferService) CreateTransfer(uid uint, transferType, account, username string, money float64, desc string) (*model.Transfer, error) {
	if money <= 0 {
		return nil, errors.New("转账金额必须大于0")
	}

	transferMin := config.Get("transfer_min")
	if transferMin != "" {
		if minVal, err := strconv.ParseFloat(transferMin, 64); err == nil && minVal > 0 && money < minVal {
			return nil, fmt.Errorf("最低转账金额为%.2f", minVal)
		}
	}

	transferMax := config.Get("transfer_max")
	if transferMax != "" {
		if maxVal, err := strconv.ParseFloat(transferMax, 64); err == nil && maxVal > 0 && money > maxVal {
			return nil, fmt.Errorf("最高转账金额为%.2f", maxVal)
		}
	}

	// 获取商户
	user, err := s.authSvc.GetUser(uid)
	if err != nil {
		log.Printf("[create_transfer_failed] uid=%d, transfer_type=%s, money=%.2f, reason=merchant not found, error=%s", uid, transferType, money, err.Error())
		return nil, errors.New("商户不存在")
	}

	if user.Transfer != 1 {
		log.Printf("[create_transfer_failed] uid=%d, transfer_type=%s, money=%.2f, reason=merchant no transfer permission")
		return nil, errors.New("商户没有转账权限")
	}

	// 检查转账方式是否开启
	cfgKey := "transfer_alipay"
	switch transferType {
	case "alipay":
		cfgKey = "transfer_alipay"
	case "wxpay":
		cfgKey = "transfer_wxpay"
	case "qqpay":
		cfgKey = "transfer_qqpay"
	case "bank":
		cfgKey = "transfer_bank"
	}

	enabled := s.authSvc.GetConfig(cfgKey)
	if enabled != "1" {
		log.Printf("[create_transfer_failed] uid=%d, transfer_type=%s, money=%.2f, reason=transfer type not enabled")
		return nil, errors.New("该转账方式未开启")
	}

	// 检查余额
	if user.Money < money {
		log.Printf("[create_transfer_failed] uid=%d, transfer_type=%s, money=%.2f, balance=%.2f, reason=insufficient balance")
		return nil, errors.New("余额不足")
	}

	channel, err := s.selectTransferChannel(transferType)
	if err != nil {
		log.Printf("[create_transfer_failed] uid=%d, transfer_type=%s, money=%.2f, reason=%s", uid, transferType, money, err.Error())
		return nil, err
	}

	// 生成转账单号
	bizNo := fmt.Sprintf("T%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)

	// 计算手续费（读取系统配置 transfer_fee，单位%）
	costRate := parseTransferFeeRate()
	costMoney := money * costRate

	tx := config.DB.Begin()

	// 扣除余额
	oldMoney := user.Money
	newMoney := oldMoney - money
	tx.Model(&user).Update("money", newMoney)

	// 创建转账记录
	transfer := &model.Transfer{
		BizNo:     bizNo,
		UID:       uid,
		Type:      transferType,
		Channel:   int(channel.ID),
		Account:   account,
		Username:  username,
		Money:     money,
		Costmoney: costMoney,
		Status:    0, // 处理中
		API:       1, // API发起
		Desc:      desc,
	}

	if err := tx.Create(transfer).Error; err != nil {
		tx.Rollback()
		log.Printf("[create_transfer_failed] uid=%d, biz_no=%s, reason=create transfer record failed, error=%s", uid, bizNo, err.Error())
		return nil, errors.New("创建转账记录失败")
	}

	// 记录资金变动
	record := &model.Record{
		UID:      uid,
		Action:   3, // 转账
		Money:    -money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "transfer",
		TradeNo:  bizNo,
		Date:     time.Now(),
	}
	tx.Create(record)

	tx.Commit()

	// 异步执行转账
	go s.executeTransferAsync(transfer.BizNo)

	return transfer, nil
}

// 转账查询
func (s *TransferService) QueryTransfer(bizNo string) (*model.Transfer, error) {
	var transfer model.Transfer
	result := config.DB.Where("biz_no = ?", bizNo).First(&transfer)
	if result.Error != nil {
		log.Printf("[query_transfer_failed] biz_no=%s, reason=transfer not found, error=%s", bizNo, result.Error.Error())
		return nil, errors.New("转账记录不存在")
	}

	// 若仍处理中，尝试实时查询通道结果
	if transfer.Status == 0 && transfer.Channel > 0 {
		if err := s.refreshTransferStatus(&transfer); err != nil {
			log.Printf("[query_transfer_status_refresh_failed] biz_no=%s, error=%s", bizNo, err.Error())
		}
		// 重新读取最新状态
		config.DB.Where("biz_no = ?", bizNo).First(&transfer)
	}

	return &transfer, nil
}

// 获取商户转账记录
func (s *TransferService) GetUserTransfers(uid uint, page, pageSize int) ([]model.Transfer, int64, error) {
	var transfers []model.Transfer
	var total int64

	query := config.DB.Model(&model.Transfer{}).Where("uid = ?", uid)
	query.Count(&total)

	result := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&transfers)
	if result.Error != nil {
		log.Printf("[get_user_transfers_failed] uid=%d, reason=query failed, error=%s", uid, result.Error.Error())
		return nil, 0, result.Error
	}

	return transfers, total, nil
}

// 更新转账状态
func (s *TransferService) UpdateTransferStatus(bizNo string, status int, result string) error {
	return config.DB.Model(&model.Transfer{}).Where("biz_no = ?", bizNo).
		Updates(map[string]interface{}{
			"status": status,
			"result": result,
		}).Error
}

// 余额查询
func (s *TransferService) QueryBalance(uid uint) (float64, error) {
	user, err := s.authSvc.GetUser(uid)
	if err != nil {
		log.Printf("[query_balance_failed] uid=%d, reason=merchant not found, error=%s", uid, err.Error())
		return 0, errors.New("商户不存在")
	}
	return user.Money, nil
}

// 转账退款（人工退款）
func (s *TransferService) RefundTransfer(bizNo string) error {
	var transfer model.Transfer
	if config.DB.First(&transfer, "biz_no = ?", bizNo).Error != nil {
		log.Printf("[refund_transfer_failed] biz_no=%s, reason=transfer not found")
		return errors.New("转账记录不存在")
	}

	if transfer.Status != 1 { // 只有成功的才能退款
		log.Printf("[refund_transfer_failed] biz_no=%s, status=%d, reason=invalid status for refund")
		return errors.New("状态不允许退款")
	}

	// 退还余额给商户
	tx := config.DB.Begin()

	var user model.User
	tx.First(&user, transfer.UID)

	oldMoney := user.Money
	newMoney := oldMoney + transfer.Money
	tx.Model(&user).Update("money", newMoney)

	// 更新转账状态
	tx.Model(&transfer).Updates(map[string]interface{}{
		"status": 3, // 已退款
	})

	// 记录资金变动
	record := &model.Record{
		UID:      transfer.UID,
		Action:   9, // 转账退款
		Money:    transfer.Money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "transfer_refund",
		TradeNo:  bizNo,
		Date:     time.Now(),
	}
	tx.Create(record)

	tx.Commit()

	return nil
}

func (s *TransferService) selectTransferChannel(transferType string) (*model.Channel, error) {
	transferType = strings.ToLower(strings.TrimSpace(transferType))
	cfgKey := ""
	switch transferType {
	case "alipay":
		cfgKey = "transfer_alipay"
	case "wxpay":
		cfgKey = "transfer_wxpay"
	}

	if cfgKey != "" {
		configured := strings.TrimSpace(config.Get(cfgKey))
		if configured != "" {
			if id, err := strconv.Atoi(configured); err == nil && id > 0 {
				var fixed model.Channel
				if err := config.DB.Where("id = ? AND status = 1", id).First(&fixed).Error; err == nil {
					handler := plugin.GetHandler(fixed.Plugin)
					if handler != nil && supportsTransferType(handler.GetInfo().Transtypes, transferType) {
						return &fixed, nil
					}
				}
			}
		}
	}

	var channels []model.Channel
	if err := config.DB.Where("status = 1").Order("id ASC").Find(&channels).Error; err != nil {
		return nil, errors.New("获取通道失败")
	}
	for _, ch := range channels {
		handler := plugin.GetHandler(ch.Plugin)
		if handler == nil {
			continue
		}
		if supportsTransferType(handler.GetInfo().Transtypes, transferType) {
			return &ch, nil
		}
	}
	return nil, errors.New("没有可用的转账通道")
}

func supportsTransferType(types []string, transferType string) bool {
	transferType = strings.ToLower(strings.TrimSpace(transferType))
	for _, t := range types {
		tv := strings.ToLower(strings.TrimSpace(t))
		if tv == transferType || tv == "all" {
			return true
		}
	}
	return false
}

func (s *TransferService) executeTransferAsync(bizNo string) {
	var transfer model.Transfer
	if err := config.DB.Where("biz_no = ?", bizNo).First(&transfer).Error; err != nil {
		log.Printf("[execute_transfer_async_failed] biz_no=%s, reason=transfer not found, error=%s", bizNo, err.Error())
		return
	}
	if transfer.Status != 0 {
		return
	}

	var channel model.Channel
	if err := config.DB.Where("id = ?", transfer.Channel).First(&channel).Error; err != nil {
		s.failTransferAndRefund(&transfer, "转账通道不存在")
		return
	}

	handler := plugin.GetHandler(channel.Plugin)
	if handler == nil {
		s.failTransferAndRefund(&transfer, "转账插件不存在")
		return
	}

	result, err := handler.Transfer(map[string]interface{}{
		"biz_no":   transfer.BizNo,
		"account":  transfer.Account,
		"username": transfer.Username,
		"money":    transfer.Money,
		"channel":  channel,
	})
	if err != nil {
		s.failTransferAndRefund(&transfer, err.Error())
		return
	}
	if result.Code != 0 {
		msg := result.ErrMsg
		if msg == "" {
			msg = "转账失败"
		}
		s.failTransferAndRefund(&transfer, msg)
		return
	}

	updates := map[string]interface{}{
		"status":       1,
		"result":       result.OrderID,
		"pay_order_no": result.OrderID,
		"paytime":      time.Now(),
	}
	if result.PayDate != "" {
		if payTime, err := time.Parse("2006-01-02 15:04:05", result.PayDate); err == nil {
			updates["paytime"] = payTime
		}
	}

	if err := config.DB.Model(&model.Transfer{}).Where("biz_no = ? AND status = 0", transfer.BizNo).Updates(updates).Error; err != nil {
		log.Printf("[execute_transfer_async_failed] biz_no=%s, reason=update success status failed, error=%s", bizNo, err.Error())
		return
	}

	log.Printf("[execute_transfer_async_success] biz_no=%s, order_id=%s", bizNo, result.OrderID)
}

func (s *TransferService) refreshTransferStatus(transfer *model.Transfer) error {
	var channel model.Channel
	if err := config.DB.Where("id = ?", transfer.Channel).First(&channel).Error; err != nil {
		return err
	}

	handler := plugin.GetHandler(channel.Plugin)
	if handler == nil {
		return errors.New("转账插件不存在")
	}

	queryResult, err := handler.TransferQuery(map[string]interface{}{
		"biz_no":  transfer.BizNo,
		"channel": channel,
	})
	if err != nil {
		return err
	}
	if queryResult.Code != 0 {
		return errors.New(queryResult.ErrMsg)
	}

	switch queryResult.Status {
	case 1:
		updates := map[string]interface{}{
			"status":  1,
			"paytime": time.Now(),
		}
		if queryResult.PayDate != "" {
			if payTime, err := time.Parse("2006-01-02 15:04:05", queryResult.PayDate); err == nil {
				updates["paytime"] = payTime
			}
		}
		config.DB.Model(&model.Transfer{}).Where("biz_no = ? AND status = 0", transfer.BizNo).Updates(updates)
	case 2:
		msg := queryResult.ErrMsg
		if msg == "" {
			msg = "转账失败"
		}
		s.failTransferAndRefund(transfer, msg)
	}

	return nil
}

func (s *TransferService) failTransferAndRefund(transfer *model.Transfer, reason string) {
	tx := config.DB.Begin()

	updateRes := tx.Model(&model.Transfer{}).Where("biz_no = ? AND status = 0", transfer.BizNo).Updates(map[string]interface{}{
		"status":  2,
		"result":  reason,
		"paytime": time.Now(),
	})
	if updateRes.Error != nil {
		tx.Rollback()
		log.Printf("[transfer_fail_refund_failed] biz_no=%s, reason=update transfer failed, error=%s", transfer.BizNo, updateRes.Error.Error())
		return
	}
	if updateRes.RowsAffected == 0 {
		tx.Rollback()
		return
	}

	var user model.User
	if err := tx.Where("uid = ?", transfer.UID).First(&user).Error; err != nil {
		tx.Rollback()
		log.Printf("[transfer_fail_refund_failed] biz_no=%s, reason=user not found, error=%s", transfer.BizNo, err.Error())
		return
	}

	oldMoney := user.Money
	newMoney := oldMoney + transfer.Money
	if err := tx.Model(&user).Update("money", newMoney).Error; err != nil {
		tx.Rollback()
		log.Printf("[transfer_fail_refund_failed] biz_no=%s, reason=refund money failed, error=%s", transfer.BizNo, err.Error())
		return
	}

	record := &model.Record{
		UID:      transfer.UID,
		Action:   10,
		Money:    transfer.Money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "transfer_fail_refund",
		TradeNo:  transfer.BizNo,
		Date:     time.Now(),
	}
	if err := tx.Create(record).Error; err != nil {
		tx.Rollback()
		log.Printf("[transfer_fail_refund_failed] biz_no=%s, reason=create record failed, error=%s", transfer.BizNo, err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("[transfer_fail_refund_failed] biz_no=%s, reason=commit failed, error=%s", transfer.BizNo, err.Error())
		return
	}

	log.Printf("[transfer_fail_refund_success] biz_no=%s, reason=%s", transfer.BizNo, reason)
}

// 获取转账记录详情
func (s *TransferService) GetTransferDetail(bizNo string) (map[string]interface{}, error) {
	var transfer model.Transfer
	if config.DB.First(&transfer, "biz_no = ?", bizNo).Error != nil {
		log.Printf("[get_transfer_detail_failed] biz_no=%s, reason=transfer not found")
		return nil, errors.New("转账记录不存在")
	}

	var user model.User
	config.DB.First(&user, transfer.UID)

	return map[string]interface{}{
		"transfer": transfer,
		"user":     user.Username,
	}, nil
}

// 获取商户资金记录
func (s *TransferService) GetUserRecords(uid uint, action int, page, pageSize int) ([]model.Record, int64, error) {
	var records []model.Record
	var total int64

	query := config.DB.Model(&model.Record{}).Where("uid = ?", uid)
	if action >= 0 {
		query = query.Where("action = ?", action)
	}

	query.Count(&total)

	result := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records)
	if result.Error != nil {
		log.Printf("[get_user_records_failed] uid=%d, action=%d, reason=query failed, error=%s", uid, action, result.Error.Error())
		return nil, 0, result.Error
	}

	return records, total, nil
}

// 管理员加款/扣款
func (s *TransferService) AdminChangeMoney(uid uint, money float64, typ string, remark string) error {
	var user model.User
	if config.DB.First(&user, uid).Error != nil {
		log.Printf("[admin_change_money_failed] uid=%d, money=%.2f, type=%s, reason=merchant not found")
		return errors.New("商户不存在")
	}

	oldMoney := user.Money
	newMoney := oldMoney + money

	if newMoney < 0 {
		log.Printf("[admin_change_money_failed] uid=%d, money=%.2f, type=%s, old_money=%.2f, reason=result would be negative")
		return errors.New("余额不能为负数")
	}

	tx := config.DB.Begin()

	tx.Model(&user).Update("money", newMoney)

	record := &model.Record{
		UID:      uid,
		Money:    money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     typ,
		Date:     time.Now(),
	}

	if typ == "admin_add" {
		record.Action = 5 // 后台加款
	} else if typ == "admin_sub" {
		record.Action = 6 // 后台扣款
	}

	tx.Create(record)

	tx.Commit()

	return nil
}
