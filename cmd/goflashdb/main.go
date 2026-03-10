package main

import (
	"flag"
	"goflashdb/pkg/config"
	"goflashdb/pkg/net"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	configPath  = flag.String("c", "", "config file path")
	showVersion = flag.Bool("v", false, "show version")
)

const Version = "0.2.0"

func main() {
	flag.Parse()

	if *showVersion {
		log.Printf("GoFlashDB version %s", Version)
		return
	}

	cfg := config.GlobalConfig
	if *configPath != "" {
		if err := config.LoadConfig(*configPath); err != nil {
			log.Fatalf("Load config failed: %v", err)
		}
		cfg = config.GlobalConfig
	}

	opts := []net.ServerOption{}

	if cfg.RequirePass != "" {
		opts = append(opts, net.WithAuth(cfg.RequirePass))
	}

	opts = append(opts, net.WithRateLimit(10000, time.Second))

	if len(cfg.RenameCommand) > 0 {
		opts = append(opts, net.WithFilter(cfg.RenameCommand))
	}

	if cfg.AppendOnly || cfg.RdbFilename != "" {
		saveInterval := 5 * time.Minute
		if len(cfg.SaveInterval) >= 2 {
			saveInterval = time.Duration(cfg.SaveInterval[0]) * time.Second
		}
		opts = append(opts, net.WithPersist(
			cfg.AppendFilename,
			cfg.RdbFilename,
			cfg.AppendOnly,
			saveInterval,
		))
	}

	server, err := net.NewServer(cfg.BindAddr, opts...)
	if err != nil {
		log.Fatalf("Create server failed: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		log.Printf("Receive signal %s, shutting down...", sig)
		server.Close()
		os.Exit(0)
	}()

	log.Printf("GoFlashDB server starting on %s", cfg.BindAddr)
	if err := server.Start(); err != nil {
		log.Fatalf("Server start failed: %v", err)
	}
}
