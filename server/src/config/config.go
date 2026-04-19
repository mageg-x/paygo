package config

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopay/src/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	DBPath     string
	DBProvided bool
	Port       string
	AdminUser  string
	AdminPwd   string
	SysKey     string
}

var AppConfig *Config
var DB *gorm.DB

func LoadConfig(dbPath, port string) {
	resolvedDBPath := strings.TrimSpace(dbPath)
	if resolvedDBPath == "" {
		resolvedDBPath = resolveAutoDBPath()
	} else if absPath, err := filepath.Abs(resolvedDBPath); err == nil {
		resolvedDBPath = absPath
	}

	AppConfig = &Config{
		DBPath:     resolvedDBPath,
		DBProvided: strings.TrimSpace(dbPath) != "",
		Port:       port,
		AdminUser:  "admin",
		AdminPwd:   "12345678",
		SysKey:     "paygosyskey2024",
	}
}

// resolveAutoDBPath 在未显式传入 -db 时优先兼容历史部署路径，
// 避免升级后切到新默认目录导致“看起来像数据丢失”。
func resolveAutoDBPath() string {
	if v := strings.TrimSpace(os.Getenv("GOPAY_DB_PATH")); v != "" {
		if abs, err := filepath.Abs(v); err == nil {
			return abs
		}
		return v
	}

	defaultPath := DefaultDBPath()
	candidates := make([]string, 0, 8)

	if wd, err := os.Getwd(); err == nil && strings.TrimSpace(wd) != "" {
		candidates = append(candidates,
			filepath.Join(wd, "data", "pay.db"),
			filepath.Join(wd, "pay.db"),
			filepath.Join(wd, "data", "gopay.db"),
		)
	}

	if exe, err := os.Executable(); err == nil && strings.TrimSpace(exe) != "" {
		base := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(base, "data", "pay.db"),
			filepath.Join(base, "pay.db"),
			filepath.Join(base, "data", "gopay.db"),
		)
	}

	for _, p := range candidates {
		fi, err := os.Stat(p)
		if err != nil || fi.IsDir() {
			continue
		}
		if abs, err := filepath.Abs(p); err == nil {
			log.Printf("[init] auto-detected legacy database path: %s", abs)
			return abs
		}
		log.Printf("[init] auto-detected legacy database path: %s", p)
		return p
	}

	return defaultPath
}

// DefaultDBPath 返回各平台默认数据库路径：
// Windows: %APPDATA%\gopay\gopay.db
// macOS: ~/Library/Application Support/gopay/gopay.db
// Linux: ~/.gopay/gopay.db
func DefaultDBPath() string {
	switch runtime.GOOS {
	case "windows":
		base := strings.TrimSpace(os.Getenv("APPDATA"))
		if base == "" {
			if cfgDir, err := os.UserConfigDir(); err == nil {
				base = cfgDir
			}
		}
		if base == "" {
			base = "."
		}
		return filepath.Join(base, "gopay", "gopay.db")
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil || strings.TrimSpace(home) == "" {
			return filepath.Join(".", "gopay.db")
		}
		return filepath.Join(home, "Library", "Application Support", "gopay", "gopay.db")
	default:
		home, err := os.UserHomeDir()
		if err != nil || strings.TrimSpace(home) == "" {
			return filepath.Join(".", "gopay.db")
		}
		return filepath.Join(home, ".gopay", "gopay.db")
	}
}

func InitDB() {
	var err error
	dbPath := AppConfig.DBPath
	if fi, statErr := os.Stat(dbPath); statErr == nil && !fi.IsDir() {
		log.Printf("[init] database file exists: path=%s", dbPath)
	} else if AppConfig.DBProvided {
		log.Printf("[init] WARNING: explicit -db file not found, will create new database: path=%s", dbPath)
	}

	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("[init] create database directory failed: %v", err)
	}

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("[init] database connection failed: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("[init] get underlying sql.DB failed: %v", err)
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
		log.Fatalf("[init] database migration failed: %v", err)
	}

	// 补齐历史表结构
	ensureSettleTransferNoColumn(sqlDB)

	// 设置自增ID从10000开始
	setAutoIncrementStart(sqlDB)

	// 初始化默认配置（必须在 loadAdminConfig 之前）
	initDefaultConfig()

	// 初始化默认用户组
	initDefaultGroup()

	// 从数据库加载管理员配置
	loadAdminConfig()

	log.Println("[init] database initialized")
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
	const minSeq = int64(9999) // 下一条自增ID将从10000开始

	rows, err := db.Query(`
		SELECT name
		FROM sqlite_master
		WHERE type = 'table'
		  AND name NOT LIKE 'sqlite_%'
		  AND sql LIKE '%AUTOINCREMENT%'
	`)
	if err != nil {
		log.Printf("[init] query autoincrement tables failed: %v", err)
		return
	}
	defer rows.Close()

	tables := make([]string, 0, 32)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			log.Printf("[init] scan autoincrement table failed: %v", err)
			continue
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[init] iterate autoincrement tables failed: %v", err)
		return
	}

	for _, table := range tables {
		if _, err := db.Exec(`UPDATE sqlite_sequence SET seq = CASE WHEN seq < ? THEN ? ELSE seq END WHERE name = ?`, minSeq, minSeq, table); err != nil {
			log.Printf("[init] update sqlite_sequence failed: table=%s, error=%v", table, err)
			continue
		}
		if _, err := db.Exec(`INSERT INTO sqlite_sequence(name, seq) SELECT ?, ? WHERE NOT EXISTS (SELECT 1 FROM sqlite_sequence WHERE name = ?)`, table, minSeq, table); err != nil {
			log.Printf("[init] insert sqlite_sequence failed: table=%s, error=%v", table, err)
			continue
		}
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
		{"sitename", "GoPay支付"},
		{"title", "GoPay支付"},
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
		{"transfer_show_name", "GoPay支付"},
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
		{"trusted_proxies", "127.0.0.1,::1"},
		{"cookie_secure", "0"},
		{"cookie_samesite", "lax"},
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
		{"sms_sign_name", "GoPay支付"},
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
		{"cron_db_backup", "1"},
		{"cron_db_backup_spec", "0 0 2 * * ?"},
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
		log.Printf("[init] default group created: gid=%d", group.GID)
	}
}

func ensureSettleTransferNoColumn(db *sql.DB) {
	rows, err := db.Query(`PRAGMA table_info(settle)`)
	if err != nil {
		log.Printf("[init] query settle table info failed: %v", err)
		return
	}
	defer rows.Close()

	hasTransferNo := false
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			log.Printf("[init] scan settle table info failed: %v", err)
			continue
		}
		if name == "transfer_no" {
			hasTransferNo = true
			break
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("[init] iterate settle table info failed: %v", err)
		return
	}
	if hasTransferNo {
		return
	}

	if _, err := db.Exec(`ALTER TABLE settle ADD COLUMN transfer_no text`); err != nil {
		log.Printf("[init] add settle.transfer_no failed: %v", err)
		return
	}
	log.Printf("[init] migration applied: add settle.transfer_no")
}
