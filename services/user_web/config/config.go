package config

import (
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Name      string          `mapstructure:"name"`
	Port      int             `mapstructure:"port"`
	UserSrv   UserSrvConfig   `mapstructure:"user-srv"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	AliyunSMS AliyunSMSConfig `mapstructure:"aliyun-sms"`
	Redis     RedisConfig     `mapstructure:"redis"`
}

type UserSrvConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type JWTConfig struct {
	SigningKey string `mapstructure:"signing_key"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AliyunSMSConfig struct {
	RegionID        string `mapstructure:"region_id"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	SignName        string `mapstructure:"sign_name"`
	TemplateCode    string `mapstructure:"template_code"`
	Expire          int    `mapstructure:"expire"`
}

// Load reads the YAML config (debug or pro) plus SHOP_-prefixed env overrides
// and returns a populated Config. It watches the file for live reloads, mutating
// the returned *Config in place so callers see updates.
func Load() (*Config, error) {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")

	v := viper.New()
	v.SetEnvPrefix("SHOP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	cfgFileName := "services/user_web/config_pro.yaml"
	if v.GetBool("debug") {
		cfgFileName = "services/user_web/config_debug.yaml"
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
