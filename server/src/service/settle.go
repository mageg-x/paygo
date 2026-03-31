package service

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"paygo/src/config"
	"paygo/src/model"
	"paygo/src/plugin"
)

type SettleService struct {
	authSvc *AuthService
}

func NewSettleService() *SettleService {
	return &SettleService{
		authSvc: NewAuthService(),
	}
}

// 申请结算
func (s *SettleService) ApplySettle(uid uint, account, username string, money float64, settleType int) (*model.Settle, error) {
	// 获取商户
	user, err := s.authSvc.GetUser(uid)
	if err != nil {
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, settle_type=%d, reason=merchant not found, error=%s", uid, money, settleType, err.Error())
		return nil, errors.New("商户不存在")
	}

	if user.Settle != 1 {
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, settle_type=%d, reason=merchant no settle permission")
		return nil, errors.New("商户没有结算权限")
	}

	// 检查结算方式是否开启
	cfgKey := "settle_alipay"
	switch settleType {
	case 1:
		cfgKey = "settle_alipay"
	case 2:
		cfgKey = "settle_wxpay"
	case 3:
		cfgKey = "settle_qqpay"
	case 4:
		cfgKey = "settle_bank"
	}

	enabled := s.authSvc.GetConfig(cfgKey)
	if enabled != "1" {
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, settle_type=%d, reason=settle type not enabled")
		return nil, errors.New("该结算方式未开启")
	}

	// 检查最低结算限额
	minMoney, _ := strconv.ParseFloat(s.authSvc.GetConfig("settle_money"), 10)
	if money < minMoney {
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, min_money=%.2f, reason=below minimum settle amount")
		return nil, fmt.Errorf("最低结算金额%.2f元", minMoney)
	}

	// 检查余额
	if user.Money < money {
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, balance=%.2f, reason=insufficient balance")
		return nil, errors.New("余额不足")
	}

	// 计算实际到账金额（扣除手续费）
	rate := 0.0 // 默认0手续费
	settleRateStr := s.authSvc.GetConfig("settle_rate_" + strconv.Itoa(settleType))
	if settleRateStr != "" {
		rate, _ = strconv.ParseFloat(settleRateStr, 10)
	}
	realMoney := money * (1 - rate/100)

	tx := config.DB.Begin()

	// 扣除余额
	oldMoney := user.Money
	newMoney := oldMoney - money
	tx.Model(&user).Update("money", newMoney)

	// 创建结算记录
	settle := &model.Settle{
		UID:       uid,
		Auto:      1,
		Type:      settleType,
		Account:   account,
		Username:  username,
		Money:     money,
		Realmoney: realMoney,
		Addtime:   time.Now(),
		Status:    model.SettleStatusPending,
	}

	if err := tx.Create(settle).Error; err != nil {
		tx.Rollback()
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, reason=create settle record failed, error=%s", uid, money, err.Error())
		return nil, errors.New("创建结算记录失败")
	}

	// 记录资金变动
	record := &model.Record{
		UID:      uid,
		Action:   2, // 结算扣款
		Money:    -money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "settle",
		TradeNo:  fmt.Sprintf("%d", settle.ID),
		Date:     time.Now(),
	}
	tx.Create(record)

	tx.Commit()

	return settle, nil
}

// 获取商户结算记录
func (s *SettleService) GetUserSettles(uid uint, page, pageSize int) ([]model.Settle, int64, error) {
	var settles []model.Settle
	var total int64

	query := config.DB.Model(&model.Settle{}).Where("uid = ?", uid)
	query.Count(&total)

	result := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&settles)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return settles, total, nil
}

// 获取结算记录详情
func (s *SettleService) GetSettle(id uint) (*model.Settle, error) {
	var settle model.Settle
	result := config.DB.First(&settle, id)
	if result.Error != nil {
		log.Printf("[get_settle_failed] id=%d, reason=settle not found, error=%s", id, result.Error.Error())
		return nil, errors.New("结算记录不存在")
	}
	return &settle, nil
}

// 同意结算
func (s *SettleService) ApproveSettle(id uint) error {
	var settle model.Settle
	result := config.DB.First(&settle, id)
	if result.Error != nil {
		log.Printf("[approve_settle_failed] id=%d, reason=settle not found, error=%s", id, result.Error.Error())
		return errors.New("结算记录不存在")
	}

	if settle.Status != model.SettleStatusPending {
		log.Printf("[approve_settle_failed] id=%d, status=%d, reason=invalid status for approval")
		return errors.New("状态不允许操作")
	}

	// 获取商户信息
	user, err := s.authSvc.GetUser(settle.UID)
	if err != nil {
		log.Printf("[approve_settle_failed] id=%d, reason=get user failed, error=%s", id, err.Error())
		return errors.New("获取商户信息失败")
	}

	// 调用支付通道进行转账
	transferResult, err := s.executeTransfer(settle, user)
	if err != nil {
		log.Printf("[approve_settle_failed] id=%d, reason=transfer failed, error=%s", id, err.Error())
		return err
	}

	tx := config.DB.Begin()

	// 更新状态
	updateData := map[string]interface{}{
		"status":    model.SettleStatusCompleted,
		"endtime":   time.Now(),
		"result":    "已同意",
		"transfer_status": 1,
		"transfer_date":   time.Now(),
	}

	if transferResult != nil {
		updateData["transfer_no"] = transferResult.OrderID
	}

	tx.Model(&settle).Updates(updateData)

	tx.Commit()

	log.Printf("[approve_settle_success] id=%d, transfer_no=%s", id, transferResult.OrderID)
	return nil
}

// 执行转账
func (s *SettleService) executeTransfer(settle model.Settle, user *model.User) (*TransferResult, error) {
	// 根据结算类型选择通道（用于日志记录）
	switch settle.Type {
	case 1:
		log.Printf("[execute_transfer] settle_id=%d, type=alipay", settle.ID)
	case 2:
		log.Printf("[execute_transfer] settle_id=%d, type=wxpay", settle.ID)
	default:
		log.Printf("[execute_transfer] settle_id=%d, type=unknown", settle.ID)
	}

	// 获取商户的结算通道配置
	var channel model.Channel
	result := config.DB.Where("type = ? AND status = 1", settle.Type).First(&channel)
	if result.Error != nil {
		log.Printf("[execute_transfer_failed] settle_id=%d, reason=no available channel for type=%d", settle.ID, settle.Type)
		return nil, errors.New("没有可用的结算通道")
	}

	// 获取通道插件
	pluginHandler := plugin.GetHandler(channel.Plugin)
	if pluginHandler == nil {
		log.Printf("[execute_transfer_failed] settle_id=%d, plugin=%s, reason=plugin not found", settle.ID, channel.Plugin)
		return nil, errors.New("支付通道插件不存在")
	}

	// 构造转账参数
	bizNo := fmt.Sprintf("S%d%d", settle.ID, time.Now().Unix())
	params := map[string]interface{}{
		"biz_no":  bizNo,
		"account": settle.Account,
		"money":   settle.Realmoney,
		"channel": channel,
	}

	// 调用插件转账
	transferResult, err := pluginHandler.Transfer(params)
	if err != nil {
		log.Printf("[execute_transfer_failed] settle_id=%d, reason=plugin transfer failed, error=%s", settle.ID, err.Error())
		return nil, err
	}

	if transferResult.Code != 0 {
		log.Printf("[execute_transfer_failed] settle_id=%d, reason=transfer rejected, code=%d, msg=%s", settle.ID, transferResult.Code, transferResult.ErrMsg)
		return nil, errors.New(transferResult.ErrMsg)
	}

	return &TransferResult{
		OrderID: transferResult.OrderID,
		PayDate: transferResult.PayDate,
	}, nil
}

// 转账结果结构
type TransferResult struct {
	OrderID string
	PayDate string
}

// 拒绝结算
func (s *SettleService) RejectSettle(id uint, reason string) error {
	var settle model.Settle
	result := config.DB.First(&settle, id)
	if result.Error != nil {
		log.Printf("[reject_settle_failed] id=%d, reason=settle not found, error=%s", id, result.Error.Error())
		return errors.New("结算记录不存在")
	}

	if settle.Status != model.SettleStatusPending {
		log.Printf("[reject_settle_failed] id=%d, status=%d, reason=invalid status for rejection")
		return errors.New("状态不允许操作")
	}

	tx := config.DB.Begin()

	// 退还余额给商户
	var user model.User
	tx.First(&user, settle.UID)

	oldMoney := user.Money
	newMoney := oldMoney + settle.Money
	tx.Model(&user).Update("money", newMoney)

	// 更新结算状态
	tx.Model(&settle).Updates(map[string]interface{}{
		"status":  model.SettleStatusFailed,
		"endtime": time.Now(),
		"result":  "拒绝: " + reason,
	})

	// 记录资金变动（退还）
	record := &model.Record{
		UID:      settle.UID,
		Action:   8, // 结算失败返还
		Money:    settle.Money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "settle",
		TradeNo:  fmt.Sprintf("%d", settle.ID),
		Date:     time.Now(),
	}
	tx.Create(record)

	tx.Commit()

	return nil
}

// 生成批量结算批次
func (s *SettleService) CreateBatch(settleIDs []uint) (*model.Batch, []model.Settle, error) {
	var settles []model.Settle
	config.DB.Where("id IN ? AND status = ?", settleIDs, model.SettleStatusPending).Find(&settles)
	if len(settles) == 0 {
		log.Printf("[create_batch_failed] reason=no pending settle records, count=%d", len(settleIDs))
		return nil, nil, errors.New("没有待处理的结算记录")
	}

	// 计算总金额
	var totalMoney float64
	for _, s := range settles {
		totalMoney += s.Money
	}

	// 生成批次号
	batchNo := fmt.Sprintf("B%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)

	batch := &model.Batch{
		Batch:    batchNo,
		Allmoney: totalMoney,
		Count:    len(settles),
		Time:     time.Now(),
		Status:   0,
	}

	tx := config.DB.Begin()

	// 创建批次
	tx.Create(batch)

	// 更新结算记录的批次号
	for _, settle := range settles {
		tx.Model(&settle).Update("batch", batchNo)
	}

	tx.Commit()

	return batch, settles, nil
}

// 执行批量转账
func (s *SettleService) ExecuteBatchTransfer(batchNo string) error {
	var batch model.Batch
	if config.DB.First(&batch, "batch = ?", batchNo).Error != nil {
		log.Printf("[execute_batch_transfer_failed] batch_no=%s, reason=batch not found")
		return errors.New("批次不存在")
	}

	if batch.Status == 1 {
		log.Printf("[execute_batch_transfer_failed] batch_no=%s, status=%d, reason=batch already processed")
		return errors.New("批次已处理")
	}

	var settles []model.Settle
	config.DB.Where("batch = ?", batchNo).Find(&settles)

	// 逐个执行转账
	successCount := 0
	failCount := 0
	var lastError error

	for _, settle := range settles {
		// 获取商户信息
		user, err := s.authSvc.GetUser(settle.UID)
		if err != nil {
			log.Printf("[execute_batch_transfer_failed] settle_id=%d, reason=get user failed, error=%s", settle.ID, err.Error())
			failCount++
			lastError = err
			continue
		}

		// 执行单个转账
		transferResult, err := s.executeTransfer(settle, user)
		if err != nil {
			log.Printf("[execute_batch_transfer_failed] settle_id=%d, reason=transfer failed, error=%s", settle.ID, err.Error())
			failCount++
			lastError = err

			// 更新为失败状态
			config.DB.Model(&settle).Updates(map[string]interface{}{
				"status":          model.SettleStatusFailed,
				"endtime":         time.Now(),
				"result":          "转账失败: " + err.Error(),
				"transfer_status": 2,
			})
			continue
		}

		// 更新为成功状态
		config.DB.Model(&settle).Updates(map[string]interface{}{
			"status":          model.SettleStatusCompleted,
			"endtime":         time.Now(),
			"result":          "转账成功",
			"transfer_status": 1,
			"transfer_date":   time.Now(),
			"transfer_no":     transferResult.OrderID,
		})
		successCount++
	}

	// 更新批次状态
	if failCount == 0 {
		config.DB.Model(&batch).Update("status", 1) // 全部成功
	} else if successCount > 0 {
		config.DB.Model(&batch).Update("status", 2) // 部分成功
	} else {
		config.DB.Model(&batch).Update("status", 3) // 全部失败
	}

	log.Printf("[execute_batch_transfer_completed] batch_no=%s, success=%d, fail=%d", batchNo, successCount, failCount)

	if lastError != nil && failCount > 0 && successCount == 0 {
		return lastError
	}

	return nil
}

// 获取所有待结算记录
func (s *SettleService) GetPendingSettles() ([]model.Settle, error) {
	var settles []model.Settle
	result := config.DB.Where("status = ?", model.SettleStatusPending).Find(&settles)
	if result.Error != nil {
		log.Printf("[get_pending_settles_failed] reason=query failed, error=%s", result.Error.Error())
		return nil, result.Error
	}
	return settles, nil
}
