package config

// Config 全局配置结构体
type Config struct {
	// 网络配置
	BindAddr string `yaml:"bind_addr" default:"0.0.0.0:6379"`
	MaxConn  int    `yaml:"max_conn" default:"10000"`
	Timeout  int    `yaml:"timeout" default:"300"` // 连接超时，单位秒

	// 内存配置
	MaxMemory      int64  `yaml:"max_memory" default:"0"` // 0表示不限制，单位字节
	EvictionPolicy string `yaml:"eviction_policy" default:"volatile-lru"`

	// 持久化配置
	AppendOnly     bool   `yaml:"append_only" default:"false"`
	AppendFilename string `yaml:"append_filename" default:"appendonly.aof"`
	AppendFsync    string `yaml:"append_fsync" default:"everysec"`
	RdbFilename    string `yaml:"rdb_filename" default:"dump.rdb"`
	SaveInterval   []int  `yaml:"save_interval" default:"[900, 1, 300, 10, 60, 10000]"`

	// 安全配置
	RequirePass   string            `yaml:"require_pass" default:""`
	MaxClients    int               `yaml:"max_clients" default:"10000"`
	RenameCommand map[string]string `yaml:"rename_command" default:"{}"`

	// AI扩展配置
	EnableAI         bool   `yaml:"enable_ai" default:"false"`
	OpenClawEndpoint string `yaml:"openclaw_endpoint" default:""`
	MCPServerAddr    string `yaml:"mcp_server_addr" default:""`
}

// GlobalConfig 全局配置实例
var GlobalConfig = &Config{}

// LoadConfig 加载配置
func LoadConfig(path string) error {
	// TODO: 实现配置加载逻辑
	return nil
}
