package config

import (
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Name   string       `mapstructure:"name"`
	Debug  bool         `mapstructure:"debug"` // from SHOP_DEBUG; selects dev logger + config_debug.yaml
	Host   string       `mapstructure:"host"` // gRPC 监听地址
	Port   int          `mapstructure:"port"` // gRPC 监听端口
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Consul ConsulConfig `mapstructure:"consul"`
}

type MySQLConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	User      string `mapstructure:"user"`
	Password  string `mapstructure:"password"`
	DBName    string `mapstructure:"dbname"`
	Charset   string `mapstructure:"charset"`
	ParseTime bool   `mapstructure:"parse_time"`
	Loc       string `mapstructure:"loc"`
}

type ConsulConfig struct {
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	Address string `mapstructure:"address"`
}

func (m *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		m.User, m.Password, m.Host, m.Port, m.DBName, m.Charset, m.ParseTime, m.Loc)
}

// Load reads the YAML config (debug or pro) plus SHOP_-prefixed env overrides
// and returns a populated Config. It watches the file for live reloads,
// mutating the returned *Config in place.
func Load() (*Config, error) {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")

	v := viper.New()
	v.SetEnvPrefix("SHOP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	cfgFileName := "services/user_srv/config_pro.yaml"
	if v.GetBool("debug") {
		cfgFileName = "services/user_srv/config_debug.yaml"
	}
	v.SetConfigFile(cfgFileName)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config file %q: %w", cfgFileName, err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		_ = v.ReadInConfig()
		_ = v.Unmarshal(cfg)
	})

	return cfg, nil
}
