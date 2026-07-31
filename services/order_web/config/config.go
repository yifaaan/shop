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
	Name         string             `mapstructure:"name"`
	Debug        bool               `mapstructure:"debug"` // from SHOP_DEBUG; selects dev logger + debug group
	Port         int                `mapstructure:"port"`
	OrderSrv     OrderSrvConfig     `mapstructure:"order-srv"`
	GoodsSrv     GoodsSrvConfig     `mapstructure:"goods-srv"`
	InventorySrv InventorySrvConfig `mapstructure:"inventory-srv"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	Consul       ConsulConfig       `mapstructure:"consul"`
	AliPay       AliPayConfig       `mapstructure:"alipay"`
	Trace        TraceConfig        `mapstructure:"trace"`
}

// OrderSrvConfig 描述 order_srv 在 Consul 注册的服务名，用于服务发现。
type OrderSrvConfig struct {
	Name string `mapstructure:"name"`
}

// GoodsSrvConfig 描述 goods_srv 在 Consul 注册的服务名，购物车列表需拉取商品名/图/价。
type GoodsSrvConfig struct {
	Name string `mapstructure:"name"`
}

// InventorySrvConfig 描述 inventory_srv 在 Consul 注册的服务名，加购/下单需校验库存。
type InventorySrvConfig struct {
	Name string `mapstructure:"name"`
}

type JWTConfig struct {
	SigningKey string `mapstructure:"signing_key"`
}

type ConsulConfig struct {
	Host    string `mapstructure:"host"`    // Consul agent 地址
	Port    int    `mapstructure:"port"`    // Consul agent 端口
	Address string `mapstructure:"address"` // 本服务对外宣告地址（健康检查 & 被发现用）
}

type AliPayConfig struct {
	AppId        string `mapstructure:"app_id"`
	PrivateKey   string `mapstructure:"private_key"`
	PublicKey    string `mapstructure:"public_key"`
	IsProduction bool   `mapstructure:"is_production"`
	NotifyUrl    string `mapstructure:"notify_url"`
	ReturnUrl    string `mapstructure:"return_url"`
}

type TraceConfig struct {
	Enabled     bool    `mapstructure:"enabled"`
	Endpoint    string  `mapstructure:"endpoint"`
	Insecure    bool    `mapstructure:"insecure"`
	SampleRatio float64 `mapstructure:"sample-ratio"`
}

// Load fetches the service config from Nacos (DataID "order-web", Group
// "debug" or "pro" per SHOP_DEBUG) in the order namespace, applies SHOP_-
// prefixed env overrides, and returns a populated Config.
func Load() (*Config, error) {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")

	debug := envBool("SHOP_DEBUG", false)
	opts := nacosconf.Options{
		Host:      envStr("SHOP_NACOS_HOST", "127.0.0.1"),
		Port:      envInt("SHOP_NACOS_PORT", 8848),
		Namespace: envStr("SHOP_NACOS_NAMESPACE_ORDER", ""),
		Username:  envStr("SHOP_NACOS_USERNAME", "nacos"),
		Password:  envStr("SHOP_NACOS_PASSWORD", "nacos"),
		DataID:    "order-web",
		Group:     groupFor(debug),
	}

	cfg := &Config{}
	v, err := nacosconf.Load(opts, func(nv *viper.Viper) {
		_ = unmarshalConfig(nv, cfg)
	})
	if err != nil {
		return nil, err
	}
	if err := unmarshalConfig(v, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	cfg.Debug = debug
	return cfg, nil
}

func unmarshalConfig(v *viper.Viper, cfg *Config) error {
	v.SetDefault("trace.enabled", true)
	v.SetDefault("trace.endpoint", "127.0.0.1:4317")
	v.SetDefault("trace.insecure", true)
	v.SetDefault("trace.sample-ratio", 1.0)
	return v.Unmarshal(cfg)
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
