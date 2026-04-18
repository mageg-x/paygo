//go:build !gui

package main

import (
	"flag"
	"log"
)

func main() {
	dbPath := flag.String("db", "", "数据库路径(默认按系统平台自动选择)")
	host := flag.String("host", "0.0.0.0", "监听IP地址")
	port := flag.String("port", "8080", "服务端口")
	migrate := flag.Bool("migrate", false, "执行数据库迁移")
	flag.Parse()

	r, addr, openURL, dbFile := initRuntime(runtimeOptions{
		DBPath:  *dbPath,
		Host:    *host,
		Port:    *port,
		Migrate: *migrate,
	})
	printStartupLogs(addr, openURL, dbFile)

	if err := r.Run(addr); err != nil {
		log.Fatalf("[server] failed to start: %v", err)
	}
}
