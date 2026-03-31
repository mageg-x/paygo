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
