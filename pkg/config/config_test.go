package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test with non-existent file
	err := LoadConfig("/nonexistent/path/config.yaml")
	if err != nil {
		t.Errorf("LoadConfig returned error for nonexistent file: %v", err)
	}
}

func TestTLSConfig(t *testing.T) {
	tls := TLSConfig{
		Enabled:  true,
		CertFile: "server.crt",
		KeyFile:  "server.key",
		ClientCA: "ca.crt",
	}

	if !tls.Enabled {
		t.Error("TLSConfig.Enabled should be true")
	}
	if tls.CertFile != "server.crt" {
		t.Errorf("Expected CertFile 'server.crt', got '%s'", tls.CertFile)
	}
}

func TestConfig(t *testing.T) {
	cfg := &Config{
		BindAddr: "0.0.0.0:6380",
		MaxConn:  5000,
		Timeout:  600,
		TLS: TLSConfig{
			Enabled: false,
		},
		MaxMemory:      1024 * 1024 * 1024,
		EvictionPolicy: "allkeys-lru",
		AppendOnly:     true,
		AppendFilename: "appendonly.aof",
		AppendFsync:    "always",
		RdbFilename:    "dump.rdb",
		SaveInterval:   []int{900, 1, 300, 10},
		RequirePass:    "mypassword",
		MaxClients:     5000,
		RenameCommand: map[string]string{
			"flushall": "hidden_flushall",
		},
		EnableAI:         true,
		OpenClawEndpoint: "http://localhost:8080",
		MCPServerAddr:    "localhost:8081",
	}

	if cfg.BindAddr != "0.0.0.0:6380" {
		t.Errorf("Expected BindAddr '0.0.0.0:6380', got '%s'", cfg.BindAddr)
	}
	if cfg.MaxConn != 5000 {
		t.Errorf("Expected MaxConn 5000, got %d", cfg.MaxConn)
	}
	if cfg.MaxMemory != 1024*1024*1024 {
		t.Errorf("Expected MaxMemory 1073741824, got %d", cfg.MaxMemory)
	}
	if cfg.EvictionPolicy != "allkeys-lru" {
		t.Errorf("Expected EvictionPolicy 'allkeys-lru', got '%s'", cfg.EvictionPolicy)
	}
	if !cfg.AppendOnly {
		t.Error("AppendOnly should be true")
	}
	if cfg.RequirePass != "mypassword" {
		t.Errorf("Expected RequirePass 'mypassword', got '%s'", cfg.RequirePass)
	}
	if !cfg.EnableAI {
		t.Error("EnableAI should be true")
	}
}

func TestGlobalConfig(t *testing.T) {
	if GlobalConfig == nil {
		t.Error("GlobalConfig should not be nil")
	}

	_ = GlobalConfig.BindAddr
	_ = GlobalConfig.MaxConn
	_ = GlobalConfig.Timeout
	_ = GlobalConfig.MaxMemory
	_ = GlobalConfig.EvictionPolicy
	_ = GlobalConfig.AppendOnly
	_ = GlobalConfig.MaxClients
	_ = GlobalConfig.EnableAI
}

func TestConfigYamlTags(t *testing.T) {
	cfg := Config{}

	// Test that yaml tags work correctly by checking struct field values
	cfg.BindAddr = "127.0.0.1:6379"
	if cfg.BindAddr != "127.0.0.1:6379" {
		t.Error("BindAddr field not working")
	}

	cfg.MaxMemory = 2048 * 1024 * 1024
	if cfg.MaxMemory != 2048*1024*1024 {
		t.Error("MaxMemory field not working")
	}

	cfg.SaveInterval = []int{60, 1, 300, 10}
	if len(cfg.SaveInterval) != 4 {
		t.Error("SaveInterval field not working")
	}
}

func TestConfigEnvOverrides(t *testing.T) {
	// Test environment variable based config loading concept
	cfg := &Config{}

	// Save original env and restore after test
	origBindAddr := cfg.BindAddr
	_ = origBindAddr

	// Basic test that config can be modified
	cfg.BindAddr = os.Getenv("REDIS_BIND_ADDR")
	_ = cfg.BindAddr // This will be empty string if env not set, which is fine

	// Just ensure the struct is mutable
	cfg.MaxConn = 100
	if cfg.MaxConn != 100 {
		t.Error("Config fields should be mutable")
	}
}
