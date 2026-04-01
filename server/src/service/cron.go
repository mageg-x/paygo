package service

import (
	"fmt"
	"log"
	"paygo/src/config"
	"paygo/src/model"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// CronService 计划任务服务
type CronService struct {
	cron   *cron.Cron
	entries map[string]cron.EntryID
	mu     sync.RWMutex
}

var (
	cronService *CronService
	cronOnce    sync.Once
)

// GetCronService 获取单例
func GetCronService() *CronService {
	cronOnce.Do(func() {
		cronService = &CronService{
			cron:   cron.New(cron.WithSeconds()),
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

	// 清理7天前的日志
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	config.DB.Where("time < ?", sevenDaysAgo).Delete(&model.Log{})

	log.Printf("[Cron] 清理完成")
}
