package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gopay/src/config"
	"gopay/src/model"
)

// 风控服务
type RiskService struct {
	authSvc *AuthService
}

func NewRiskService() *RiskService {
	return &RiskService{
		authSvc: NewAuthService(),
	}
}

// 风控检查结果
type RiskCheckResult struct {
	Passed bool
	Code   int // 0=通过, 1=IP黑名单, 2=用户黑名单, 3=IP每日限制, 4=用户每日限制, 5=商品名称过滤, 6=金额超限
	Msg    string
}

// 检查支付风控
func (s *RiskService) CheckPaymentRisk(uid uint, ip, name string, money float64) *RiskCheckResult {
	// 1. 检查IP黑名单
	if s.IsIPBlocked(ip) {
		log.Printf("[risk_check_failed] uid=%d, ip=%s, reason=ip blocked", uid, ip)
		return &RiskCheckResult{Passed: false, Code: 1, Msg: "IP禁止访问"}
	}

	// 2. 检查用户黑名单
	if uid > 0 && s.IsUserBlocked(uid) {
		log.Printf("[risk_check_failed] uid=%d, ip=%s, reason=user blocked", uid, ip)
		return &RiskCheckResult{Passed: false, Code: 2, Msg: "用户被禁止交易"}
	}

	// 3. 检查商品名称过滤
	blockNames := s.authSvc.GetConfig("blockname")
	if blockNames != "" && s.isBlockedName(name, blockNames) {
		log.Printf("[risk_check_failed] uid=%d, ip=%s, name=%s, reason=name blocked", uid, ip, name)
		return &RiskCheckResult{Passed: false, Code: 5, Msg: "商品名称禁止交易"}
	}

	// 4. 检查IP每日支付次数限制
	ipLimit := s.authSvc.GetConfig("pay_iplimit")
	if ipLimit != "" && !s.checkIPLimit(ip, ipLimit) {
		log.Printf("[risk_check_failed] uid=%d, ip=%s, reason=ip daily limit exceeded", uid, ip)
		return &RiskCheckResult{Passed: false, Code: 3, Msg: "IP每日支付次数超限"}
	}

	// 5. 检查用户每日支付次数限制
	if uid > 0 {
		userLimit := s.authSvc.GetConfig("pay_userlimit")
		if userLimit != "" && !s.checkUserLimit(uid, userLimit) {
			log.Printf("[risk_check_failed] uid=%d, ip=%s, reason=user daily limit exceeded", uid, ip)
			return &RiskCheckResult{Passed: false, Code: 4, Msg: "用户每日支付次数超限"}
		}
	}

	// 6. 检查金额限制
	minMoney := s.authSvc.GetConfig("pay_minmoney")
	if minMoney != "" {
		var min float64
		fmt.Sscanf(minMoney, "%f", &min)
		if min > 0 && money < min {
			log.Printf("[risk_check_failed] uid=%d, money=%.2f, min=%.2f, reason=below minimum", uid, money, min)
			return &RiskCheckResult{Passed: false, Code: 6, Msg: fmt.Sprintf("单笔金额不能低于%.2f元", min)}
		}
	}

	maxMoney := s.authSvc.GetConfig("pay_maxmoney")
	if maxMoney != "" {
		var max float64
		fmt.Sscanf(maxMoney, "%f", &max)
		if max > 0 && money > max {
			log.Printf("[risk_check_failed] uid=%d, money=%.2f, max=%.2f, reason=exceeds maximum", uid, money, max)
			return &RiskCheckResult{Passed: false, Code: 6, Msg: fmt.Sprintf("单笔金额不能超过%.2f元", max)}
		}
	}

	return &RiskCheckResult{Passed: true, Code: 0, Msg: "通过"}
}

// 检查IP是否在黑名单
func (s *RiskService) IsIPBlocked(ip string) bool {
	var count int64
	config.DB.Model(&model.Blacklist{}).Where("type = 1 AND content = ? AND (endtime IS NULL OR endtime > ?)", ip, time.Now()).Count(&count)
	return count > 0
}

// 检查用户是否在黑名单
func (s *RiskService) IsUserBlocked(uid uint) bool {
	var count int64
	config.DB.Model(&model.Blacklist{}).Where("type = 0 AND content = ? AND (endtime IS NULL OR endtime > ?)", fmt.Sprintf("%d", uid), time.Now()).Count(&count)
	return count > 0
}

// 检查商品名称是否被禁止
func (s *RiskService) isBlockedName(name, blockNames string) bool {
	name = strings.ToLower(name)
	blocks := strings.Split(blockNames, "|")
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block != "" && strings.Contains(name, strings.ToLower(block)) {
			return true
		}
	}
	return false
}

// 检查IP每日支付次数
func (s *RiskService) checkIPLimit(ip, limitStr string) bool {
	var limit int
	fmt.Sscanf(limitStr, "%d", &limit)
	if limit <= 0 {
		return true
	}

	today := time.Now().Format("2006-01-02")
	var count int64
	config.DB.Model(&model.Order{}).Where("ip = ? AND date = ? AND status IN (1,2,3)", ip, today).Count(&count)

	return count < int64(limit)
}

// 检查用户每日支付次数
func (s *RiskService) checkUserLimit(uid uint, limitStr string) bool {
	var limit int
	fmt.Sscanf(limitStr, "%d", &limit)
	if limit <= 0 {
		return true
	}

	today := time.Now().Format("2006-01-02")
	var count int64
	config.DB.Model(&model.Order{}).Where("uid = ? AND date = ? AND status IN (1,2,3)", uid, today).Count(&count)

	return count < int64(limit)
}

// 添加IP到黑名单
func (s *RiskService) AddIPToBlacklist(ip, remark string, duration time.Duration) error {
	blacklist := &model.Blacklist{
		Type:    1,
		Content: ip,
		Addtime: time.Now(),
		Remark:  remark,
	}

	if duration > 0 {
		endtime := time.Now().Add(duration)
		blacklist.Endtime = &endtime
	}

	if err := config.DB.Create(blacklist).Error; err != nil {
		log.Printf("[add_ip_blacklist_failed] ip=%s, error=%s", ip, err.Error())
		return err
	}

	log.Printf("[add_ip_blacklist_success] ip=%s, duration=%v", ip, duration)
	return nil
}

// 添加用户到黑名单
func (s *RiskService) AddUserToBlacklist(uid uint, remark string, duration time.Duration) error {
	blacklist := &model.Blacklist{
		Type:    0,
		Content: fmt.Sprintf("%d", uid),
		Addtime: time.Now(),
		Remark:  remark,
	}

	if duration > 0 {
		endtime := time.Now().Add(duration)
		blacklist.Endtime = &endtime
	}

	if err := config.DB.Create(blacklist).Error; err != nil {
		log.Printf("[add_user_blacklist_failed] uid=%d, error=%s", uid, err.Error())
		return err
	}

	log.Printf("[add_user_blacklist_success] uid=%d, duration=%v", uid, duration)
	return nil
}

// 从黑名单移除
func (s *RiskService) RemoveFromBlacklist(id uint) error {
	result := config.DB.Delete(&model.Blacklist{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("黑名单记录不存在")
	}
	return nil
}

// 获取黑名单列表
func (s *RiskService) GetBlacklist(page, pageSize int) ([]model.Blacklist, int64, error) {
	var blacklists []model.Blacklist
	var total int64

	query := config.DB.Model(&model.Blacklist{})
	query.Count(&total)

	result := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&blacklists)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return blacklists, total, nil
}

// 记录风控日志
func (s *RiskService) AddRiskLog(uid uint, riskType int, url, content string) {
	risk := &model.Risk{
		UID:     uid,
		Type:    riskType,
		URL:     url,
		Content: content,
		Date:    time.Now(),
		Status:  0,
	}
	config.DB.Create(risk)
}

// 获取风控日志
func (s *RiskService) GetRiskLogs(uid uint, page, pageSize int) ([]model.Risk, int64, error) {
	var logs []model.Risk
	var total int64

	query := config.DB.Model(&model.Risk{}).Where("uid = ?", uid)
	query.Count(&total)

	result := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return logs, total, nil
}

// 检查是否为蜘蛛/机器人（简单的User-Agent检测）
func (s *RiskService) IsSpider(userAgent string) bool {
	if userAgent == "" {
		return false
	}

	userAgent = strings.ToLower(userAgent)
	spiders := []string{
		"baiduspider",
		"googlebot",
		"bingbot",
		"slurp",
		"duckduckbot",
		"baidu",
		"360spider",
		"bytespider",
		"python",
		"requests",
		"scrapy",
		"curl",
		"wget",
		"go-http",
		"java/",
		"php",
		"libwww",
		"apache",
		"nginx",
	}

	for _, spider := range spiders {
		if strings.Contains(userAgent, spider) {
			return true
		}
	}

	return false
}

// 检查代理/VPN（基于IP数据库查询 - 简化实现）
func (s *RiskService) IsProxyIP(ip string) bool {
	// 简化实现：实际应该查询IP数据库
	// 这里仅作为占位符
	proxyRanges := []string{
		// 常见的代理IP段（仅作示例，实际需要IP库）
	}

	for _, range_ := range proxyRanges {
		if strings.HasPrefix(ip, range_) {
			return true
		}
	}

	return false
}
