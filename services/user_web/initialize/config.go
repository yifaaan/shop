package initialize

import (
	"fmt"
	"shop/services/user_web/global"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func InitConfig() {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")

	viper.AutomaticEnv()
	debug := viper.GetBool("SHOP_DEBUG")
	configFilePrefix := "config"
	cfgFileName := fmt.Sprintf("services/user_web/%s_pro.yaml", configFilePrefix)
	if debug {
		cfgFileName = fmt.Sprintf("services/user_web/%s_debug.yaml", configFilePrefix)
	}

	v := viper.New()
	v.SetEnvPrefix("SHOP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()
	v.SetConfigFile(cfgFileName)
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	if err := v.Unmarshal(global.ServerConfig); err != nil {
		panic(fmt.Errorf("fatal error unmarshaling config: %s", err))
	}

	zap.S().Infof("config file used: %s", v.ConfigFileUsed())

	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		_ = v.ReadInConfig()
		_ = v.Unmarshal(global.ServerConfig)
		zap.S().Infof("config file used: %s", v.ConfigFileUsed())
	})
}
