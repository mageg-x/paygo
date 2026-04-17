package config

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"paygo/src/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	DBPath    string
	Port      string
	AdminUser string
	AdminPwd  string
	SysKey    string
}

var AppConfig *Config
var DB *gorm.DB

func LoadConfig(dbPath, port string) {
	AppConfig = &Config{
		DBPath:    dbPath,
		Port:      port,
		AdminUser: "admin",
		AdminPwd:  "12345678",
		SysKey:    "paygosyskey2024",
	}
}

func InitDB() {
	var err error
	dbPath := AppConfig.DBPath

	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("创建数据库目录失败: %v", err)
	}

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("获取底层sql.DB失败: %v", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// 自动迁移
	err = DB.AutoMigrate(
		&model.User{},
		&model.Group{},
		&model.Record{},
		&model.Log{},
		&model.Order{},
		&model.RefundOrder{},
		&model.Settle{},
		&model.Batch{},
		&model.Transfer{},
		&model.PayType{},
		&model.Plugin{},
		&model.Channel{},
		&model.Roll{},
		&model.SubChannel{},
		&model.Config{},
		&model.Cache{},
		&model.Anounce{},
		&model.RegCode{},
		&model.InviteCode{},
		&model.Risk{},
		&model.Domain{},
		&model.Blacklist{},
		&model.PsReceiver{},
		&model.PsReceiver2{},
		&model.PsOrder{},
		&model.PsRecord{},
		&model.Agent{},
		&model.Kefu{},
		&model.MailQueue{},
		&model.UserGroupTransfer{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 设置自增ID从10000开始
	setAutoIncrementStart(sqlDB)

	// 初始化默认配置（必须在 loadAdminConfig 之前）
	initDefaultConfig()

	// 初始化默认用户组
	initDefaultGroup()

	// 从数据库加载管理员配置
	loadAdminConfig()

	log.Println("数据库初始化成功")
}

func loadAdminConfig() {
	var cfg model.Config

	// 加载管理员用户名
	cfg = model.Config{}
	DB.Where("k = ?", "admin_user").Limit(1).Find(&cfg)
	if cfg.V != "" {
		AppConfig.AdminUser = cfg.V
	}

	// 加载管理员密码
	cfg = model.Config{}
	DB.Where("k = ?", "admin_pwd").Limit(1).Find(&cfg)
	if cfg.V != "" {
		AppConfig.AdminPwd = cfg.V
	}

	// 加载系统密钥
	cfg = model.Config{}
	DB.Where("k = ?", "sys_key").Limit(1).Find(&cfg)
	if cfg.V != "" {
		AppConfig.SysKey = cfg.V
	}
}

// 设置自增ID从10000开始
func setAutoIncrementStart(db *sql.DB) {
	tables := []string{"user", "user_group", "record", "log", "order", "refundorder", "settle", "batch", "transfer", "pay_type", "plugin", "channel", "roll", "sub_channel", "config", "cache", "anounce", "reg_code", "invite_code", "risk", "domain", "blacklist", "ps_receiver", "ps_receiver2", "ps_order"}
	for _, table := range tables {
		db.Exec("INSERT INTO sqlite_sequence (name, seq) VALUES (?, 9999) ON CONFLICT(name) DO UPDATE SET seq = 9999", table)
	}
}

// Get 获取配置值
func Get(key string) string {
	var cfg model.Config
	DB.Where("k = ?", key).Limit(1).Find(&cfg)
	return cfg.V
}

// Set 设置配置值
func Set(key, value string) error {
	var cfg model.Config
	DB.Where("k = ?", key).Limit(1).Find(&cfg)
	if cfg.K != "" {
		// 更新
		cfg.V = value
		DB.Save(&cfg)
	} else {
		// 创建
		DB.Create(&model.Config{K: key, V: value})
	}
	return nil
}

// 初始化默认配置
func initDefaultConfig() {
	defaultConfigs := []struct {
		k string
		v string
	}{
		// 网站信息
		{"reg_open", "1"},        // 注册开放
		{"reg_pay", "0"},         // 注册收费关闭
		{"reg_pay_price", "0"},   // 注册费用
		{"user_review", "0"},     // 注册无需审核
		{"test_open", "0"},       // 测试支付关闭
		{"test_pay_uid", "1000"}, // 测试支付收款商户
		{"sitename", "PayGo支付"},
		{"title", "PayGo支付"},
		{"localurl", "http://127.0.0.1:8080/"},
		{"apiurl", "http://127.0.0.1:8080/"},
		{"site_keywords", ""},
		{"site_description", ""},
		{"cdn_url", ""},
		{"user_verification", "0"}, // 用户验证方式: 0=无, 1=邮箱, 2=手机
		{"kfqq", ""},
		{"email", ""},
		// 商户设置
		{"default_group", "1"}, // 默认用户组ID
		// 支付设置
		{"pay_min_money", "1"},
		{"pay_max_money", "100000"},
		{"pay_block_goods", ""},
		{"pay_fee_rate", "0"},
		{"invite_cashback", "0"},
		{"qrcode_enabled", "0"},
		{"pay_success_page", ""},
		{"pay_error_page", ""},
		// 结算设置
		{"settle_money", "30"},
		{"settle_cycle", "1"},
		{"settle_alipay", "1"},
		{"settle_wxpay", "1"},
		{"settle_auto_transfer", "0"},
		// 转账设置
		{"transfer_min", "1"},
		{"transfer_max", "50000"},
		{"transfer_fee", "0"},
		{"transfer_alipay", "0"},
		{"transfer_wxpay", "0"},
		{"transfer_show_name", "PayGo支付"},
		// 快捷登录
		{"login_alipay", "0"},
		{"login_qq", "0"},
		{"login_wx", "0"},
		// 通知设置
		{"notify_email", ""},
		{"email_notify", "0"},
		{"order_notify", "0"},
		// 实名认证
		{"certificate_required", "0"},
		{"certificate_types", "1,2,3"},
		// IP类型
		{"ip_type", "0"}, // 0=REMOTE_ADDR, 1=X_FORWARDED_FOR, 2=X_REAL_IP
		// 代理设置
		{"proxy_enabled", "0"},
		{"proxy_host", ""},
		{"proxy_port", ""},
		{"proxy_user", ""},
		{"proxy_pass", ""},
		// 邮件设置
		{"mail_smtp_host", ""},
		{"mail_smtp_port", "587"},
		{"mail_username", ""},
		{"mail_password", ""},
		{"mail_from", ""},
		// 短信设置 - 阿里云
		{"sms_enabled", "0"},
		{"sms_provider", "aliyun"},
		{"sms_access_key_id", ""},
		{"sms_access_key_secret", ""},
		{"sms_sign_name", "PayGo支付"},
		{"sms_template_code", ""},
		{"sms_order_template_code", ""},
		// 计划任务设置
		{"cron_auto_settle", "1"},
		{"cron_auto_settle_spec", "0 0 * * * ?"},
		{"cron_retry_notify", "1"},
		{"cron_retry_notify_spec", "0 */5 * * * ?"},
		{"cron_order_query", "1"},
		{"cron_order_query_spec", "0 */3 * * * ?"},
		{"cron_risk_check", "1"},
		{"cron_risk_check_spec", "0 */30 * * * ?"},
		{"cron_cleanup", "1"},
		{"cron_cleanup_spec", "0 0 0 * * ?"},
		// 公告
		{"gonggao_content", ""},
	}

	for _, cfg := range defaultConfigs {
		var count int64
		DB.Model(&model.Config{}).Where("k = ?", cfg.k).Count(&count)
		if count == 0 {
			DB.Create(&model.Config{K: cfg.k, V: cfg.v})
		}
	}
}

// 初始化默认用户组
func initDefaultGroup() {
	var count int64
	DB.Model(&model.Group{}).Count(&count)
	if count == 0 {
		// 创建默认用户组
		group := model.Group{
			Name:       "默认组",
			Info:       "系统默认用户组",
			Sort:       1,
			Isbuy:      0,
			Price:      0,
			Expire:     0,
			SettleOpen: 1,
			SettleType: 1,
			SettleRate: "0.5",
		}
		DB.Create(&group)
		log.Printf("[初始化] 创建默认用户组, gid=%d", group.GID)
	}
}
