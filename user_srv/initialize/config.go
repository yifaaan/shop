package initialize

import (
	"fmt"
	"shop/user_srv/global"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func getEnvInfo(env string) bool {
	viper.AutomaticEnv()
	return viper.GetBool(env)
}

func InitConfig() {
	if getEnvInfo("SHOP_DEBUG") {
		viper.SetConfigFile("./user_srv/config-debug.yaml")
	} else {
		viper.SetConfigFile("./user_srv/config-pro.yaml")
	}
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = viper.Unmarshal(&global.ServerConfig)
	if err != nil {
		panic(err)
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		viper.Unmarshal(&global.ServerConfig)
	})
}
