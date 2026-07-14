package initialize

import (
	"fmt"
	"strings"

	"shop/services/user_web/global"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func InitConfig() {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")

	v := viper.New()
	v.SetEnvPrefix("SHOP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	debug := v.GetBool("debug")
	cfgFileName := "services/user_web/config_pro.yaml"
	if debug {
		cfgFileName = "services/user_web/config_debug.yaml"
	}

	v.SetConfigFile(cfgFileName)
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("read config file %q: %w", cfgFileName, err))
	}

	if err := v.Unmarshal(global.ServerConfig); err != nil {
		panic(fmt.Errorf("unmarshal config: %w", err))
	}

	zap.S().Infof("config file used: %s", v.ConfigFileUsed())

	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		_ = v.ReadInConfig()
		_ = v.Unmarshal(global.ServerConfig)
		zap.S().Infof("config file used: %s", v.ConfigFileUsed())
	})
}
