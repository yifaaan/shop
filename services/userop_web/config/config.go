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
	Name      string       `mapstructure:"name"`
	Debug     bool         `mapstructure:"debug"` // from SHOP_DEBUG; selects dev logger + debug group
	Port      int          `mapstructure:"port"`  // HTTP 监听端口
	UserOpSrv UserOpSrvCfg `mapstructure:"userop-srv"`
	GoodsSrv  GoodsSrvCfg  `mapstructure:"goods-srv"`
	JWT       JWTConfig    `mapstructure:"jwt"`
	Consul    ConsulConfig `mapstructure:"consul"`
}

// UserOpSrvCfg 描述 userop_srv 在 Consul 注册的服务名，用于服务发现。
type UserOpSrvCfg struct {
	Name string `mapstructure:"name"`
}

// GoodsSrvCfg 描述 goods_srv 在 Consul 注册的服务名，收藏列表需拉取商品名/图/价。
type GoodsSrvCfg struct {
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

// Load 从 Nacos（DataID "userop-web"）取配置，叠加 SHOP_* env 覆盖。
func Load() (*Config, error) {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")

	debug := envBool("SHOP_DEBUG", false)
	opts := nacosconf.Options{
		Host:      envStr("SHOP_NACOS_HOST", "127.0.0.1"),
		Port:      envInt("SHOP_NACOS_PORT", 8848),
		Namespace: envStr("SHOP_NACOS_NAMESPACE_USEROP", ""),
		Username:  envStr("SHOP_NACOS_USERNAME", "nacos"),
		Password:  envStr("SHOP_NACOS_PASSWORD", "nacos"),
		DataID:    "userop-web",
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