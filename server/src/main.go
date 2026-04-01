package main

import (
	"flag"
	"log"
	"os"

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
		// TODO: 实现数据库迁移
		runMigrations()
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

func runMigrations() {
	// TODO: 实现数据库迁移
	// 读取 schema.sql 并执行
	log.Println("迁移完成")
}
