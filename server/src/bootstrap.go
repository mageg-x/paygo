package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"paygo/src/config"
	"paygo/src/router"
	"paygo/src/service"

	"github.com/gin-gonic/gin"
)

type runtimeOptions struct {
	DBPath  string
	Host    string
	Port    string
	Migrate bool
}

func initRuntime(opts runtimeOptions) (*gin.Engine, string, string, string) {
	config.LoadConfig(opts.DBPath, opts.Port)
	config.InitDB()

	service.InitSmsService()
	service.InitSystemCrons()

	if opts.Migrate {
		log.Println("[migrate] running database migrations...")
		if err := runMigrations(); err != nil {
			log.Fatalf("[migrate] database migration failed: %v", err)
		}
	}

	dbDir := filepath.Dir(config.AppConfig.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("[init] create database directory failed: %v", err)
	}

	host := strings.TrimSpace(opts.Host)
	if host == "" {
		host = "0.0.0.0"
	}
	port := strings.TrimSpace(opts.Port)
	if port == "" {
		port = "8080"
	}
	addr := net.JoinHostPort(host, port)
	openURL := buildOpenURL(host, port)

	r := router.SetupRouter()
	return r, addr, openURL, config.AppConfig.DBPath
}

func printStartupLogs(addr, openURL, dbPath string) {
	base := strings.TrimRight(openURL, "/")
	log.Printf("[server] payment system started, listen=%s, db=%s", addr, dbPath)
	log.Printf("[server] admin panel: %s/admin", base)
	log.Printf("[server] merchant panel: %s/user", base)
	log.Printf("[server] api endpoint: %s/api", base)
}

func buildOpenURL(host, port string) string {
	h := strings.TrimSpace(host)
	if h == "" || h == "0.0.0.0" || h == "::" {
		h = "localhost"
	}
	return fmt.Sprintf("http://%s:%s/", h, port)
}

func runMigrations() error {
	migrations := []struct {
		name string
		sql  string
	}{
		{
			name: "idx_order_uid_status_addtime",
			sql:  "CREATE INDEX IF NOT EXISTS idx_order_uid_status_addtime ON `order`(uid, status, addtime)",
		},
		{
			name: "idx_order_notify_status_time",
			sql:  "CREATE INDEX IF NOT EXISTS idx_order_notify_status_time ON `order`(notify, status, notifytime)",
		},
		{
			name: "idx_order_out_trade_no_uid",
			sql:  "CREATE INDEX IF NOT EXISTS idx_order_out_trade_no_uid ON `order`(out_trade_no, uid)",
		},
		{
			name: "idx_settle_uid_status_addtime",
			sql:  "CREATE INDEX IF NOT EXISTS idx_settle_uid_status_addtime ON settle(uid, status, addtime)",
		},
		{
			name: "idx_transfer_uid_status_paytime",
			sql:  "CREATE INDEX IF NOT EXISTS idx_transfer_uid_status_paytime ON transfer(uid, status, paytime)",
		},
		{
			name: "idx_record_uid_action_date",
			sql:  "CREATE INDEX IF NOT EXISTS idx_record_uid_action_date ON record(uid, action, date)",
		},
		{
			name: "idx_log_uid_date",
			sql:  "CREATE INDEX IF NOT EXISTS idx_log_uid_date ON log(uid, date)",
		},
		{
			name: "idx_regcode_scene_to_status_time",
			sql:  "CREATE INDEX IF NOT EXISTS idx_regcode_scene_to_status_time ON regcode(scene, `to`, status, time)",
		},
		{
			name: "idx_invitecode_code_unique",
			sql:  "CREATE UNIQUE INDEX IF NOT EXISTS idx_invitecode_code_unique ON invitecode(code)",
		},
	}

	for _, m := range migrations {
		key := "migration_" + m.name
		if config.Get(key) != "" {
			continue
		}
		if err := config.DB.Exec(m.sql).Error; err != nil {
			return fmt.Errorf("应用迁移 %s 失败: %w", m.name, err)
		}
		if err := config.Set(key, time.Now().Format(time.RFC3339)); err != nil {
			return fmt.Errorf("写入迁移标记 %s 失败: %w", m.name, err)
		}
		log.Printf("[migrate] applied: name=%s", m.name)
	}

	log.Println("[migrate] completed")
	return nil
}
