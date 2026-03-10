package main

import (
	"flag"
	"goflashdb/pkg/config"
	"goflashdb/pkg/net"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	configPath = flag.String("c", "", "config file path")
	version    = flag.Bool("v", false, "show version")
)

const Version = "0.1.0"

func main() {
	flag.Parse()

	if *version {
		log.Printf("GoSwiftKV version %s", Version)
		return
	}

	// 加载配置
	if *configPath != "" {
		if err := config.LoadConfig(*configPath); err != nil {
			log.Fatalf("Load config failed: %v", err)
		}
	}

	// 初始化服务器
	server, err := net.NewServer(config.GlobalConfig.BindAddr)
	if err != nil {
		log.Fatalf("Create server failed: %v", err)
	}

	// 启动信号监听
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		log.Printf("Receive signal %s, shutting down...", sig)
		server.Close()
		os.Exit(0)
	}()

	// 启动服务
	log.Printf("GoSwiftKV server starting on %s", config.GlobalConfig.BindAddr)
	if err := server.Start(); err != nil {
		log.Fatalf("Server start failed: %v", err)
	}
}
