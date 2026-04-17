package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"paygo/src/config"
	"paygo/src/router"
	"paygo/src/service"
)

func main() {
	// 命令行参数
	dbPath := flag.String("db", "../data/pay.db", "数据库路径")
	port := flag.String("port", "8080", "服务端口")
	migrate := flag.Bool("migrate", false, "执行数据库迁移")
	flag.Parse()

	// 初始化配置
	config.LoadConfig(*dbPath, *port)

	// 初始化数据库
	config.InitDB()

	// 初始化短信服务
	service.InitSmsService()

	// 初始化计划任务
	service.InitSystemCrons()

	// 如果指定了migrate，执行迁移
	if *migrate {
		log.Println("执行数据库迁移...")
		if err := runMigrations(); err != nil {
			log.Fatalf("数据库迁移失败: %v", err)
		}
	}

	// 确保数据目录存在
	dir := "../data"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}

	// 设置路由
	r := router.SetupRouter()

	// 启动服务
	addr := ":" + *port
	log.Printf("支付系统启动成功，监听端口: %s", *port)
	log.Printf("管理后台: http://localhost:%s/admin", *port)
	log.Printf("商户后台: http://localhost:%s/user", *port)
	log.Printf("API接口: http://localhost:%s/api", *port)

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
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
		log.Printf("迁移已应用: %s", m.name)
	}

	log.Println("迁移完成")
	return nil
}
