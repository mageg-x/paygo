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

	"gorm.io/gorm"
)

type SettleService struct {
	authSvc *AuthService
}

func NewSettleService() *SettleService {
	return &SettleService{
		authSvc: NewAuthService(),
	}
}

// 申请结算（手动）
func (s *SettleService) ApplySettle(uid uint, account, username string, money float64, settleType int) (*model.Settle, error) {
	return s.applySettle(uid, account, username, money, settleType, 0)
}

// 申请结算（自动）
func (s *SettleService) ApplyAutoSettle(uid uint, account, username string, money float64, settleType int) (*model.Settle, error) {
	return s.applySettle(uid, account, username, money, settleType, 1)
}

func (s *SettleService) applySettle(uid uint, account, username string, money float64, settleType int, auto int) (*model.Settle, error) {
	account = strings.TrimSpace(account)
	username = strings.TrimSpace(username)
	if username == "" {
		username = account
	}
	if account == "" || username == "" {
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, settle_type=%d, reason=missing account or username", uid, money, settleType)
		return nil, errors.New("结算账号或姓名不能为空")
	}
	if money <= 0 {
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, settle_type=%d, reason=invalid amount", uid, money, settleType)
		return nil, errors.New("结算金额必须大于0")
	}

	user, err := s.authSvc.GetUser(uid)
	if err != nil {
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, settle_type=%d, reason=merchant not found, error=%s", uid, money, settleType, err.Error())
		return nil, errors.New("商户不存在")
	}
	if user.Status != 1 {
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, settle_type=%d, reason=merchant disabled, status=%d", uid, money, settleType, user.Status)
		return nil, errors.New("商户已被禁用")
	}
	if user.Settle != 1 {
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, settle_type=%d, reason=merchant no settle permission", uid, money, settleType)
		return nil, errors.New("商户没有结算权限")
	}

	cfgKey, err := settleTypeConfigKey(settleType)
	if err != nil {
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, settle_type=%d, reason=invalid settle type", uid, money, settleType)
		return nil, err
	}
	if s.authSvc.GetConfig(cfgKey) != "1" {
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, settle_type=%d, reason=settle type not enabled", uid, money, settleType)
		return nil, errors.New("该结算方式未开启")
	}

	minMoney := parseFloatOrDefault(s.authSvc.GetConfig("settle_money"), 0)
	if money < minMoney {
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, min_money=%.2f, reason=below minimum settle amount", uid, money, minMoney)
		return nil, fmt.Errorf("最低结算金额%.2f元", minMoney)
	}

	rate := s.getSettleRate(user.GID, settleType)
	realMoney := money * (1 - rate/100)
	if realMoney < 0 {
		realMoney = 0
	}

	now := time.Now()
	tx := config.DB.Begin()
	if tx.Error != nil {
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, reason=begin tx failed, error=%s", uid, money, tx.Error.Error())
		return nil, errors.New("系统繁忙，请稍后重试")
	}

	// 条件扣款，防止并发超扣。
	deductRes := tx.Model(&model.User{}).
		Where("uid = ? AND status = 1 AND settle = 1 AND money >= ?", uid, money).
		Update("money", gorm.Expr("money - ?", money))
	if deductRes.Error != nil {
		tx.Rollback()
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, reason=deduct failed, error=%s", uid, money, deductRes.Error.Error())
		return nil, errors.New("结算申请失败")
	}
	if deductRes.RowsAffected == 0 {
		tx.Rollback()
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, reason=insufficient balance or disabled", uid, money)
		return nil, errors.New("余额不足")
	}

	var updatedUser model.User
	if err := tx.Select("uid", "money").Where("uid = ?", uid).First(&updatedUser).Error; err != nil {
		tx.Rollback()
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, reason=load updated balance failed, error=%s", uid, money, err.Error())
		return nil, errors.New("结算申请失败")
	}

	settlerecord := &model.Settle{
		UID:       uid,
		Auto:      auto,
		Type:      settleType,
		Account:   account,
		Username:  username,
		Money:     money,
		Realmoney: realMoney,
		Addtime:   now,
		Status:    model.SettleStatusPending,
		Result:    "待处理",
	}
	if err := tx.Create(settlerecord).Error; err != nil {
		tx.Rollback()
		log.Printf("[apply_settle_failed] uid=%d, money=%.2f, reason=create settle record failed, error=%s", uid, money, err.Error())
		return nil, errors.New("创建结算记录失败")
	}

	newMoney := updatedUser.Money
	oldMoney := newMoney + money
	record := &model.Record{
		UID:      uid,
		Action:   2,
		Money:    -money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "settle",
		TradeNo:  fmt.Sprintf("%d", settlerecord.ID),
		Date:     now,
	}
	if err := tx.Create(record).Error; err != nil {
		tx.Rollback()
		log.Printf("[apply_settle_failed] uid=%d, settle_id=%d, reason=create record failed, error=%s", uid, settlerecord.ID, err.Error())
		return nil, errors.New("结算申请失败")
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("[apply_settle_failed] uid=%d, settle_id=%d, reason=commit failed, error=%s", uid, settlerecord.ID, err.Error())
		return nil, errors.New("结算申请失败")
	}

	log.Printf("[apply_settle_success] uid=%d, settle_id=%d, money=%.2f, rate=%.2f%%, real_money=%.2f, auto=%d", uid, settlerecord.ID, money, rate, realMoney, auto)
	return settlerecord, nil
}

func settleTypeConfigKey(settleType int) (string, error) {
	switch settleType {
	case 1:
		return "settle_alipay", nil
	case 2:
		return "settle_wxpay", nil
	case 3:
		return "settle_qqpay", nil
	case 4:
		return "settle_bank", nil
	default:
		return "", errors.New("结算方式不支持")
	}
}

func parseFloatOrDefault(raw string, def float64) float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return def
	}
	return v
}

func mapSettleTypeToTransferType(settleType int) string {
	switch settleType {
	case 1:
		return "alipay"
	case 2:
		return "wxpay"
	case 3:
		return "qqpay"
	case 4:
		return "bank"
	default:
		return ""
	}
}

func resolveTransferChannelByType(settleType int) (*model.Channel, error) {
	transferType := mapSettleTypeToTransferType(settleType)
	if transferType == "" {
		return nil, errors.New("结算方式不支持")
	}

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
			if cid, err := strconv.Atoi(configured); err == nil && cid > 0 {
				var fixed model.Channel
				if err := config.DB.Where("id = ? AND status = 1", cid).First(&fixed).Error; err == nil {
					h := plugin.GetHandler(fixed.Plugin)
					if h != nil && supportsTransferType(h.GetInfo().Transtypes, transferType) {
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
		h := plugin.GetHandler(ch.Plugin)
		if h == nil {
			continue
		}
		if supportsTransferType(h.GetInfo().Transtypes, transferType) {
			return &ch, nil
		}
	}
	return nil, errors.New("没有可用的结算通道")
}

func (s *SettleService) getSettleRate(gid uint, settleType int) float64 {
	// 优先使用用户组费率
	var group model.Group
	if err := config.DB.First(&group, gid).Error; err == nil {
		r := parseFloatOrDefault(group.SettleRate, 0)
		if r > 0 {
			return r
		}
	}

	// 兼容旧配置 settle_rate_x，并支持统一费率 settle_fee_rate
	for _, key := range []string{fmt.Sprintf("settle_rate_%d", settleType), "settle_fee_rate"} {
		r := parseFloatOrDefault(s.authSvc.GetConfig(key), 0)
		if r > 0 {
			return r
		}
	}

	return 0
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
	// 先抢占状态，避免并发审批/拒绝导致重复处理。
	lockRes := config.DB.Model(&model.Settle{}).
		Where("id = ? AND status = ?", id, model.SettleStatusPending).
		Updates(map[string]interface{}{
			"status":          model.SettleStatusProcessing,
			"result":          "转账处理中",
			"transfer_status": 0,
		})
	if lockRes.Error != nil {
		log.Printf("[approve_settle_failed] id=%d, reason=lock settle failed, error=%s", id, lockRes.Error.Error())
		return errors.New("结算处理失败")
	}
	if lockRes.RowsAffected == 0 {
		log.Printf("[approve_settle_failed] id=%d, reason=invalid status for approval", id)
		return errors.New("状态不允许操作")
	}

	var settle model.Settle
	if err := config.DB.First(&settle, id).Error; err != nil {
		log.Printf("[approve_settle_failed] id=%d, reason=settle not found after lock, error=%s", id, err.Error())
		return errors.New("结算记录不存在")
	}

	user, err := s.authSvc.GetUser(settle.UID)
	if err != nil {
		_ = s.markSettleFailedAndRefund(settle.ID, true, "转账失败: 获取商户信息失败", "获取商户信息失败")
		log.Printf("[approve_settle_failed] id=%d, reason=get user failed, error=%s", id, err.Error())
		return errors.New("获取商户信息失败")
	}

	transferResult, err := s.executeTransfer(settle, user)
	if err != nil {
		if refundErr := s.markSettleFailedAndRefund(settle.ID, true, "转账失败: "+err.Error(), err.Error()); refundErr != nil {
			log.Printf("[approve_settle_refund_failed] id=%d, reason=refund after transfer failed failed, error=%s", id, refundErr.Error())
			return errors.New("转账失败且退款补偿失败，请人工处理")
		}
		log.Printf("[approve_settle_failed] id=%d, reason=transfer failed, error=%s", id, err.Error())
		return err
	}

	updateData := map[string]interface{}{
		"status":          model.SettleStatusCompleted,
		"endtime":         time.Now(),
		"result":          "已同意",
		"transfer_status": 1,
		"transfer_date":   time.Now(),
		"transfer_result": "SUCCESS",
	}
	if transferResult != nil {
		updateData["transfer_no"] = transferResult.OrderID
	}

	finishRes := config.DB.Model(&model.Settle{}).
		Where("id = ? AND status = ?", settle.ID, model.SettleStatusProcessing).
		Updates(updateData)
	if finishRes.Error != nil {
		log.Printf("[approve_settle_failed] id=%d, reason=update success status failed, error=%s", id, finishRes.Error.Error())
		return errors.New("结算状态更新失败，请人工核对")
	}
	if finishRes.RowsAffected == 0 {
		log.Printf("[approve_settle_failed] id=%d, reason=settle status changed before success update", id)
		return errors.New("结算状态已变化，请刷新后重试")
	}

	transferNo := ""
	if transferResult != nil {
		transferNo = transferResult.OrderID
	}
	log.Printf("[approve_settle_success] id=%d, transfer_no=%s", id, transferNo)
	return nil
}

// 执行转账
func (s *SettleService) executeTransfer(settle model.Settle, user *model.User) (*TransferResult, error) {
	switch settle.Type {
	case 1:
		log.Printf("[execute_transfer] settle_id=%d, type=alipay", settle.ID)
	case 2:
		log.Printf("[execute_transfer] settle_id=%d, type=wxpay", settle.ID)
	default:
		log.Printf("[execute_transfer] settle_id=%d, type=unknown", settle.ID)
	}

	channel, err := resolveTransferChannelByType(settle.Type)
	if err != nil {
		log.Printf("[execute_transfer_failed] settle_id=%d, reason=resolve channel failed, type=%d, error=%s", settle.ID, settle.Type, err.Error())
		return nil, err
	}

	pluginHandler := plugin.GetHandler(channel.Plugin)
	if pluginHandler == nil {
		log.Printf("[execute_transfer_failed] settle_id=%d, plugin=%s, reason=plugin not found", settle.ID, channel.Plugin)
		return nil, errors.New("支付通道插件不存在")
	}

	bizNo := fmt.Sprintf("S%d%d", settle.ID, time.Now().Unix())
	params := map[string]interface{}{
		"biz_no":   bizNo,
		"account":  settle.Account,
		"username": settle.Username,
		"money":    settle.Realmoney,
		"channel":  *channel,
	}

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

func (s *SettleService) markSettleFailedAndRefund(settleID uint, allowProcessing bool, resultText, transferResult string) error {
	allowed := []int{model.SettleStatusPending}
	if allowProcessing {
		allowed = append(allowed, model.SettleStatusProcessing)
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		return errors.New("系统繁忙，请稍后重试")
	}

	var settle model.Settle
	if err := tx.Where("id = ?", settleID).First(&settle).Error; err != nil {
		tx.Rollback()
		log.Printf("[settle_refund_failed] settle_id=%d, reason=settle not found, error=%s", settleID, err.Error())
		return errors.New("结算记录不存在")
	}

	// 先通过状态条件更新抢占，避免并发重复退款。
	updateRes := tx.Model(&model.Settle{}).
		Where("id = ? AND status IN ?", settleID, allowed).
		Updates(map[string]interface{}{
			"status":          model.SettleStatusFailed,
			"endtime":         time.Now(),
			"result":          resultText,
			"transfer_status": 2,
			"transfer_result": transferResult,
		})
	if updateRes.Error != nil {
		tx.Rollback()
		log.Printf("[settle_refund_failed] settle_id=%d, reason=update status failed, error=%s", settleID, updateRes.Error.Error())
		return errors.New("更新结算状态失败")
	}
	if updateRes.RowsAffected == 0 {
		tx.Rollback()
		log.Printf("[settle_refund_skipped] settle_id=%d, reason=status changed", settleID)
		return errors.New("状态不允许操作")
	}

	if err := tx.Model(&model.User{}).
		Where("uid = ?", settle.UID).
		Update("money", gorm.Expr("money + ?", settle.Money)).Error; err != nil {
		tx.Rollback()
		log.Printf("[settle_refund_failed] settle_id=%d, uid=%d, reason=refund money failed, error=%s", settleID, settle.UID, err.Error())
		return errors.New("退款失败")
	}

	var user model.User
	if err := tx.Select("uid", "money").Where("uid = ?", settle.UID).First(&user).Error; err != nil {
		tx.Rollback()
		log.Printf("[settle_refund_failed] settle_id=%d, uid=%d, reason=load user balance failed, error=%s", settleID, settle.UID, err.Error())
		return errors.New("退款失败")
	}

	newMoney := user.Money
	oldMoney := newMoney - settle.Money
	record := &model.Record{
		UID:      settle.UID,
		Action:   8,
		Money:    settle.Money,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "settle",
		TradeNo:  fmt.Sprintf("%d", settle.ID),
		Date:     time.Now(),
	}
	if err := tx.Create(record).Error; err != nil {
		tx.Rollback()
		log.Printf("[settle_refund_failed] settle_id=%d, uid=%d, reason=create record failed, error=%s", settleID, settle.UID, err.Error())
		return errors.New("退款失败")
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("[settle_refund_failed] settle_id=%d, reason=commit failed, error=%s", settleID, err.Error())
		return errors.New("退款失败")
	}

	log.Printf("[settle_refund_success] settle_id=%d, uid=%d, money=%.2f, reason=%s", settleID, settle.UID, settle.Money, transferResult)
	return nil
}

// 拒绝结算
func (s *SettleService) RejectSettle(id uint, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "管理员拒绝"
	}
	return s.markSettleFailedAndRefund(id, false, "拒绝: "+reason, reason)
}

// 管理员补发差额（给商户余额加款）
func (s *SettleService) AdjustSettleCompensate(id uint, amount float64, reason string) error {
	if amount <= 0 {
		return errors.New("补发金额必须大于0")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "管理员补发差额"
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		return errors.New("系统繁忙，请稍后重试")
	}

	var settle model.Settle
	if err := tx.Where("id = ?", id).First(&settle).Error; err != nil {
		tx.Rollback()
		return errors.New("结算记录不存在")
	}

	var user model.User
	if err := tx.Where("uid = ?", settle.UID).First(&user).Error; err != nil {
		tx.Rollback()
		return errors.New("商户不存在")
	}

	oldMoney := user.Money
	newMoney := oldMoney + amount

	if err := tx.Model(&model.User{}).Where("uid = ?", user.UID).Update("money", newMoney).Error; err != nil {
		tx.Rollback()
		return errors.New("补发失败")
	}

	record := &model.Record{
		UID:      user.UID,
		Action:   5,
		Money:    amount,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "settle_compensate",
		TradeNo:  fmt.Sprintf("%d", settle.ID),
		Date:     time.Now(),
	}
	if err := tx.Create(record).Error; err != nil {
		tx.Rollback()
		return errors.New("补发失败")
	}

	remark := fmt.Sprintf("补发差额 %.2f 元", amount)
	if reason != "" {
		remark = fmt.Sprintf("%s；原因：%s", remark, reason)
	}
	if err := tx.Model(&model.Settle{}).Where("id = ?", settle.ID).Update("result", remark).Error; err != nil {
		tx.Rollback()
		return errors.New("补发失败")
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("补发失败")
	}

	log.Printf("[settle_compensate_success] settle_id=%d, uid=%d, amount=%.2f, reason=%s", settle.ID, user.UID, amount, reason)
	return nil
}

// 管理员冲正扣回（从商户余额扣款）
func (s *SettleService) AdjustSettleDeduct(id uint, amount float64, reason string) error {
	if amount <= 0 {
		return errors.New("扣回金额必须大于0")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "管理员冲正扣回"
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		return errors.New("系统繁忙，请稍后重试")
	}

	var settle model.Settle
	if err := tx.Where("id = ?", id).First(&settle).Error; err != nil {
		tx.Rollback()
		return errors.New("结算记录不存在")
	}

	var user model.User
	if err := tx.Where("uid = ?", settle.UID).First(&user).Error; err != nil {
		tx.Rollback()
		return errors.New("商户不存在")
	}

	if user.Money < amount {
		tx.Rollback()
		return errors.New("商户余额不足，无法扣回")
	}

	oldMoney := user.Money
	newMoney := oldMoney - amount

	if err := tx.Model(&model.User{}).Where("uid = ?", user.UID).Update("money", newMoney).Error; err != nil {
		tx.Rollback()
		return errors.New("扣回失败")
	}

	record := &model.Record{
		UID:      user.UID,
		Action:   6,
		Money:    -amount,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "settle_deduct",
		TradeNo:  fmt.Sprintf("%d", settle.ID),
		Date:     time.Now(),
	}
	if err := tx.Create(record).Error; err != nil {
		tx.Rollback()
		return errors.New("扣回失败")
	}

	remark := fmt.Sprintf("冲正扣回 %.2f 元", amount)
	if reason != "" {
		remark = fmt.Sprintf("%s；原因：%s", remark, reason)
	}
	if err := tx.Model(&model.Settle{}).Where("id = ?", settle.ID).Update("result", remark).Error; err != nil {
		tx.Rollback()
		return errors.New("扣回失败")
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("扣回失败")
	}

	log.Printf("[settle_deduct_success] settle_id=%d, uid=%d, amount=%.2f, reason=%s", settle.ID, user.UID, amount, reason)
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

	var totalMoney float64
	for _, st := range settles {
		totalMoney += st.Money
	}

	batchNo := fmt.Sprintf("B%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
	batch := &model.Batch{
		Batch:    batchNo,
		Allmoney: totalMoney,
		Count:    len(settles),
		Time:     time.Now(),
		Status:   0,
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		return nil, nil, errors.New("创建批次失败")
	}

	if err := tx.Create(batch).Error; err != nil {
		tx.Rollback()
		return nil, nil, errors.New("创建批次失败")
	}

	for _, st := range settles {
		if err := tx.Model(&model.Settle{}).Where("id = ?", st.ID).Update("batch", batchNo).Error; err != nil {
			tx.Rollback()
			return nil, nil, errors.New("创建批次失败")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, nil, errors.New("创建批次失败")
	}

	return batch, settles, nil
}

// 执行批量转账
func (s *SettleService) ExecuteBatchTransfer(batchNo string) error {
	var batch model.Batch
	if config.DB.First(&batch, "batch = ?", batchNo).Error != nil {
		log.Printf("[execute_batch_transfer_failed] batch_no=%s, reason=batch not found", batchNo)
		return errors.New("批次不存在")
	}

	if batch.Status == 1 {
		log.Printf("[execute_batch_transfer_failed] batch_no=%s, status=%d, reason=batch already processed", batchNo, batch.Status)
		return errors.New("批次已处理")
	}

	var settles []model.Settle
	config.DB.Where("batch = ?", batchNo).Find(&settles)

	successCount := 0
	failCount := 0
	var lastError error

	for _, item := range settles {
		lockRes := config.DB.Model(&model.Settle{}).
			Where("id = ? AND status = ?", item.ID, model.SettleStatusPending).
			Updates(map[string]interface{}{
				"status":          model.SettleStatusProcessing,
				"result":          "批量转账处理中",
				"transfer_status": 0,
			})
		if lockRes.Error != nil {
			failCount++
			lastError = lockRes.Error
			continue
		}
		if lockRes.RowsAffected == 0 {
			continue
		}

		var settle model.Settle
		if err := config.DB.First(&settle, item.ID).Error; err != nil {
			failCount++
			lastError = err
			continue
		}

		user, err := s.authSvc.GetUser(settle.UID)
		if err != nil {
			failCount++
			lastError = err
			if refundErr := s.markSettleFailedAndRefund(settle.ID, true, "转账失败: 获取商户信息失败", "获取商户信息失败"); refundErr != nil {
				log.Printf("[execute_batch_transfer_refund_failed] settle_id=%d, error=%s", settle.ID, refundErr.Error())
			}
			continue
		}

		transferResult, err := s.executeTransfer(settle, user)
		if err != nil {
			failCount++
			lastError = err
			if refundErr := s.markSettleFailedAndRefund(settle.ID, true, "转账失败: "+err.Error(), err.Error()); refundErr != nil {
				log.Printf("[execute_batch_transfer_refund_failed] settle_id=%d, error=%s", settle.ID, refundErr.Error())
			}
			continue
		}

		finishRes := config.DB.Model(&model.Settle{}).
			Where("id = ? AND status = ?", settle.ID, model.SettleStatusProcessing).
			Updates(map[string]interface{}{
				"status":          model.SettleStatusCompleted,
				"endtime":         time.Now(),
				"result":          "转账成功",
				"transfer_status": 1,
				"transfer_date":   time.Now(),
				"transfer_no":     transferResult.OrderID,
				"transfer_result": "SUCCESS",
			})
		if finishRes.Error != nil {
			failCount++
			lastError = finishRes.Error
			continue
		}
		if finishRes.RowsAffected == 0 {
			failCount++
			lastError = errors.New("结算状态已变化")
			continue
		}

		successCount++
	}

	switch {
	case failCount == 0:
		config.DB.Model(&batch).Update("status", 1)
	case successCount > 0:
		config.DB.Model(&batch).Update("status", 2)
	default:
		config.DB.Model(&batch).Update("status", 3)
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
