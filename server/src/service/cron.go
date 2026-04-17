package service

import (
	"errors"
	"fmt"
	"log"
	"paygo/src/config"
	"paygo/src/model"
	"paygo/src/plugin"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// CronService 计划任务服务
type CronService struct {
	cron    *cron.Cron
	entries map[string]cron.EntryID
	mu      sync.RWMutex
}

var (
	cronService *CronService
	cronOnce    sync.Once
)

// GetCronService 获取单例
func GetCronService() *CronService {
	cronOnce.Do(func() {
		cronService = &CronService{
			cron:    cron.New(cron.WithSeconds()),
			entries: make(map[string]cron.EntryID),
		}
	})
	return cronService
}

// AddTask 添加任务
func (s *CronService) AddTask(name, spec string, task func()) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.entries[name]; ok {
		s.cron.Remove(id)
	}

	id, err := s.cron.AddFunc(spec, task)
	if err != nil {
		return err
	}

	s.entries[name] = id
	log.Printf("[Cron] 添加任务: %s, 执行周期: %s", name, spec)
	return nil
}

// RemoveTask 移除任务
func (s *CronService) RemoveTask(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.entries[name]; ok {
		s.cron.Remove(id)
		delete(s.entries, name)
		log.Printf("[Cron] 移除任务: %s", name)
	}
}

// Start 启动
func (s *CronService) Start() {
	s.cron.Start()
	log.Println("[Cron] 计划任务服务已启动")
}

// Stop 停止
func (s *CronService) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("[Cron] 计划任务服务已停止")
}

// ListTasks 列出所有任务状态
func (s *CronService) ListTasks() []map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]map[string]interface{}, 0)
	for name, id := range s.entries {
		entry := s.cron.Entry(id)
		next := ""
		if entry.Next.After(time.Time{}) {
			next = entry.Next.Format("2006-01-02 15:04:05")
		}
		result = append(result, map[string]interface{}{
			"name": name,
			"id":   id,
			"next": next,
		})
	}
	return result
}

// InitSystemCrons 初始化系统计划任务
func InitSystemCrons() {
	cron := GetCronService()

	// 自动结算
	if config.Get("cron_auto_settle") == "1" {
		spec := config.Get("cron_auto_settle_spec")
		if spec == "" {
			spec = "0 0 * * * ?"
		}
		cron.AddTask("auto_settle", spec, AutoSettleTask)
	}

	// 回调重试
	if config.Get("cron_retry_notify") == "1" {
		spec := config.Get("cron_retry_notify_spec")
		if spec == "" {
			spec = "0 */5 * * * ?"
		}
		cron.AddTask("retry_notify", spec, RetryNotifyTask)
	}

	// 订单查单（查单并自动补单）
	if config.Get("cron_order_query") == "1" {
		spec := config.Get("cron_order_query_spec")
		if spec == "" {
			spec = "0 */3 * * * ?"
		}
		cron.AddTask("order_query", spec, OrderQueryTask)
	}

	// 风控检查
	if config.Get("cron_risk_check") == "1" {
		spec := config.Get("cron_risk_check_spec")
		if spec == "" {
			spec = "0 */30 * * * ?"
		}
		cron.AddTask("risk_check", spec, RiskCheckTask)
	}

	// 清理过期数据
	if config.Get("cron_cleanup") == "1" {
		spec := config.Get("cron_cleanup_spec")
		if spec == "" {
			spec = "0 0 0 * * ?"
		}
		cron.AddTask("cleanup", spec, CleanupTask)
	}

	cron.Start()
}

// 自动结算任务
func AutoSettleTask() {
	log.Println("[Cron] 执行自动结算任务")

	// 获取已开启自动结算且余额足够的商户
	var users []model.User
	config.DB.Where("money >= ?", 100).Find(&users)

	for _, user := range users {
		// 计算待结算金额
		if user.Money >= 100 {
			settleMoney := user.Money * 0.97 // 扣除3%手续费

			// 创建结算单
			settle := &model.Settle{
				UID:      user.UID,
				Money:    settleMoney,
				Account:  user.Account,
				Username: user.Username,
				Status:   0, // 待处理
				Addtime:  time.Now(),
			}
			if err := config.DB.Create(settle).Error; err == nil {
				log.Printf("[Cron] 商户 %d 创建结算单: %.2f", user.UID, settleMoney)
			}
		}
	}
}

// 回调重试任务
// 回调重试任务
func RetryNotifyTask() {
	log.Println("[Cron] 执行回调重试任务")

	var orders []model.Order
	// 查找已支付但未通知的订单，或通知失败的订单
	config.DB.Where("status = 1 AND (notify = 0 OR notify = 2) AND notifytime < ?",
		time.Now().Add(-5*time.Minute)).Limit(100).Find(&orders)

	for _, order := range orders {
		log.Printf("[Cron] 重试通知订单: %s", order.TradeNo)
		// 实际通知逻辑由payment服务处理
		// 这里仅标记为重试中
		config.DB.Model(&order).Updates(map[string]interface{}{
			"notify":     2,
			"notifytime": time.Now(),
		})
	}
}

type orderQueryPlugin interface {
	QueryOrder(params map[string]interface{}) (map[string]interface{}, error)
}

type OrderQueryOutcome struct {
	TradeNo    string
	ChannelID  int
	Plugin     string
	Exists     bool
	Paid       bool
	Status     string
	APITradeNo string
	Buyer      string
	Amount     float64
	PayTime    string
	Filled     bool
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case fmt.Stringer:
		return strings.TrimSpace(t.String())
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func toBool(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "1" || strings.EqualFold(t, "true")
	case float64:
		return t != 0
	case int:
		return t != 0
	default:
		return false
	}
}

func toFloat(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f
	default:
		return 0
	}
}

func refreshPendingOrder(order model.Order, orderSvc *OrderService) (*OrderQueryOutcome, error) {
	var channel model.Channel
	if err := config.DB.First(&channel, order.Channel).Error; err != nil {
		return nil, fmt.Errorf("通道不存在: %w", err)
	}

	outcome := &OrderQueryOutcome{
		TradeNo:   order.TradeNo,
		ChannelID: int(channel.ID),
		Plugin:    channel.Plugin,
	}

	handler := plugin.GetHandler(channel.Plugin)
	if handler == nil {
		return outcome, errors.New("支付插件不存在")
	}

	queryHandler, ok := handler.(orderQueryPlugin)
	if !ok {
		return outcome, errors.New("插件不支持查单")
	}

	queryResult, err := queryHandler.QueryOrder(map[string]interface{}{
		"trade_no": order.TradeNo,
		"channel":  channel,
	})
	if err != nil {
		return outcome, err
	}

	outcome.Exists = toBool(queryResult["exists"])
	outcome.Paid = toBool(queryResult["paid"])
	outcome.Status = toString(queryResult["status"])
	outcome.APITradeNo = toString(queryResult["api_trade_no"])
	outcome.Buyer = toString(queryResult["buyer"])
	outcome.Amount = toFloat(queryResult["amount"])
	outcome.PayTime = toString(queryResult["pay_time"])

	if !outcome.Paid {
		return outcome, nil
	}

	// 金额校验：上游金额与本地订单金额不一致时不自动补单，避免误补。
	if outcome.Amount > 0 && abs(outcome.Amount-order.Money) > 0.01 {
		return outcome, fmt.Errorf("上游已支付但金额不一致（本地%.2f，上游%.2f）", order.Money, outcome.Amount)
	}

	if err := orderSvc.OrderPaid(order.TradeNo, outcome.APITradeNo, outcome.Buyer); err != nil {
		return outcome, err
	}
	outcome.Filled = true
	return outcome, nil
}

// RefreshOrderStatus 手动刷新单笔订单状态（仅支持待支付订单）
func RefreshOrderStatus(tradeNo string) (*OrderQueryOutcome, error) {
	tradeNo = strings.TrimSpace(tradeNo)
	if tradeNo == "" {
		return nil, errors.New("缺少订单号")
	}

	var order model.Order
	if err := config.DB.Where("trade_no = ?", tradeNo).First(&order).Error; err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPending {
		return nil, fmt.Errorf("订单当前状态不是待支付（status=%d）", order.Status)
	}

	return refreshPendingOrder(order, NewOrderService())
}

// 订单查单任务（查单并自动补单）
func OrderQueryTask() {
	log.Println("[Cron] 执行订单查单任务")

	var orders []model.Order
	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)
	oldest := now.Add(-24 * time.Hour)
	config.DB.
		Select("trade_no", "channel", "money", "status", "addtime").
		Where("status = ? AND addtime >= ? AND addtime < ?", model.OrderStatusPending, oldest, cutoff).
		Order("addtime ASC").
		Limit(100).
		Find(&orders)
	if len(orders) == 0 {
		log.Println("[Cron] 订单查单任务: 无待查订单")
		return
	}

	orderSvc := NewOrderService()

	for _, order := range orders {
		outcome, err := refreshPendingOrder(order, orderSvc)
		if err != nil {
			if outcome != nil {
				log.Printf("[order_query_failed] trade_no=%s, channel_id=%d, plugin=%s, status=%s, reason=%s",
					outcome.TradeNo, outcome.ChannelID, outcome.Plugin, outcome.Status, err.Error())
			} else {
				log.Printf("[order_query_failed] trade_no=%s, channel_id=%d, reason=%s", order.TradeNo, order.Channel, err.Error())
			}
			continue
		}

		if outcome.Filled {
			log.Printf("[order_query_fill_success] trade_no=%s, channel_id=%d, plugin=%s, status=%s, api_trade_no=%s, amount=%.2f, buyer=%s, pay_time=%s",
				outcome.TradeNo, outcome.ChannelID, outcome.Plugin, outcome.Status, outcome.APITradeNo, outcome.Amount, outcome.Buyer, outcome.PayTime)
			continue
		}

		log.Printf("[order_query_result] trade_no=%s, channel_id=%d, plugin=%s, exists=%v, paid=%v, status=%s",
			outcome.TradeNo, outcome.ChannelID, outcome.Plugin, outcome.Exists, outcome.Paid, outcome.Status)
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// 风控检查任务
// 风控检查任务
func RiskCheckTask() {
	log.Println("[Cron] 执行风控检查任务")

	var users []model.User
	config.DB.Find(&users)

	today := time.Now().Format("2006-01-02")

	for _, user := range users {
		var totalCount, successCount int64
		config.DB.Model(&model.Order{}).Where("uid = ? AND date = ?", user.UID, today).Count(&totalCount)
		config.DB.Model(&model.Order{}).Where("uid = ? AND date = ? AND status = 1", user.UID, today).Count(&successCount)

		if totalCount >= 10 {
			rate := float64(successCount) / float64(totalCount)
			if rate < 0.5 {
				risk := &model.Risk{
					UID:     user.UID,
					Type:    1,
					Content: fmt.Sprintf("订单成功率过低: %.1f%%", rate*100),
					Status:  0,
					Date:    time.Now(),
				}
				config.DB.Create(risk)
				log.Printf("[Cron] 商户 %d 触发风控: 订单成功率 %.1f%%", user.UID, rate*100)
			}
		}
	}
}

// 清理过期数据任务
// 清理过期数据任务
func CleanupTask() {
	log.Println("[Cron] 执行清理过期数据任务")

	// 硬删除24小时前未支付订单，避免无效订单长期膨胀
	stalePendingBefore := time.Now().Add(-24 * time.Hour)
	orderCleanup := config.DB.Where("status = ? AND addtime < ?", model.OrderStatusPending, stalePendingBefore).Delete(&model.Order{})
	if orderCleanup.Error != nil {
		log.Printf("[Cron] 清理超时未支付订单失败: %s", orderCleanup.Error.Error())
	} else if orderCleanup.RowsAffected > 0 {
		log.Printf("[Cron] 清理超时未支付订单: %d", orderCleanup.RowsAffected)
	}

	// 清理7天前的日志
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	config.DB.Where("date < ?", sevenDaysAgo).Delete(&model.Log{})

	log.Printf("[Cron] 清理完成")
}
