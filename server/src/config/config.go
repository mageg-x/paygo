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
		&model.Weixin{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 设置自增ID从10000开始
	setAutoIncrementStart(sqlDB)

	// 初始化默认配置（必须在 loadAdminConfig 之前）
	initDefaultConfig()

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
	tables := []string{"user", "user_group", "record", "log", "order", "refundorder", "settle", "batch", "transfer", "pay_type", "plugin", "channel", "roll", "sub_channel", "config", "cache", "anounce", "reg_code", "invite_code", "risk", "domain", "blacklist", "ps_receiver", "ps_receiver2", "ps_order", "weixin"}
	for _, table := range tables {
		db.Exec("INSERT INTO sqlite_sequence (name, seq) VALUES (?, 9999) ON CONFLICT(name) DO UPDATE SET seq = 9999", table)
	}
}

// 初始化默认配置
func initDefaultConfig() {
	defaultConfigs := []struct {
		k string
		v string
	}{
		{"reg_open", "1"},        // 注册开放
		{"reg_pay", "0"},         // 注册收费关闭
		{"reg_pay_price", "0"},   // 注册费用
		{"user_review", "0"},     // 注册无需审核
		{"test_open", "0"},        // 测试支付关闭
		{"test_pay_uid", "1000"}, // 测试支付收款商户
		{"sitename", "PayGo支付"},
		{"title", "PayGo支付"},
		{"localurl", "http://127.0.0.1:8080/"},
		{"apiurl", "http://127.0.0.1:8080/"},
		{"settle_money", "30"},
		{"settle_cycle", "1"},
		{"settle_alipay", "1"},
		{"settle_wxpay", "1"},
		{"transfer_min", "1"},
		{"transfer_max", "50000"},
		{"transfer_fee", "0"},
	}

	for _, cfg := range defaultConfigs {
		var count int64
		DB.Model(&model.Config{}).Where("k = ?", cfg.k).Count(&count)
		if count == 0 {
			DB.Create(&model.Config{K: cfg.k, V: cfg.v})
		}
	}
}
