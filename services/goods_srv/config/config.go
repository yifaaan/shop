package config

import (
	"fmt"
	"os"
	"strconv"

	"shop/pkg/nacosconf"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Name   string       `mapstructure:"name"`
	Debug  bool         `mapstructure:"debug"` // from SHOP_DEBUG; selects dev logger + debug group
	Host   string       `mapstructure:"host"`  // gRPC 监听地址
	Port   int          `mapstructure:"port"`  // gRPC 监听端口
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

// DSNWithoutDB 返回不带库名的 MySQL DSN，用于连接到服务端执行建库等
// 无需选择具体库的操作（与 DSN 对照，后者指向已配置的业务库）。
func (m *MySQLConfig) DSNWithoutDB() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=%t&loc=%s",
		m.User, m.Password, m.Host, m.Port, m.Charset, m.ParseTime, m.Loc)
}

// Load fetches the service config from Nacos (DataID "goods-srv", Group
// "debug" or "pro" per SHOP_DEBUG) in the goods namespace, applies SHOP_-
// prefixed env overrides, and returns a populated Config. Nacos connection
// params come from the SHOP_NACOS_* env vars.
//
// On Nacos-side config change, the returned *Config is re-unmarshalled in
// place so callers holding the pointer see updates.
func Load() (*Config, error) {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")

	debug := envBool("SHOP_DEBUG", false)
	opts := nacosconf.Options{
		Host:      envStr("SHOP_NACOS_HOST", "127.0.0.1"),
		Port:      envInt("SHOP_NACOS_PORT", 8848),
		Namespace: envStr("SHOP_NACOS_NAMESPACE_GOODS", ""), // goods 命名空间 ID（未设则 public）
		Username:  envStr("SHOP_NACOS_USERNAME", "nacos"),
		Password:  envStr("SHOP_NACOS_PASSWORD", "nacos"),
		DataID:    "goods-srv",
		Group:     groupFor(debug),
	}

	cfg := &Config{}
	v, err := nacosconf.Load(opts, func(nv *viper.Viper) {
		_ = nv.Unmarshal(cfg)
	})
	if err != nil {
		return nil, err
	}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	cfg.Debug = debug
	return cfg, nil
}

func groupFor(debug bool) string {
	if debug {
		return "debug"
	}
	return "pro"
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}