package main

import (
	"fmt"
	"os"
	"shop/good_web/global"
	"shop/good_web/initialize"
	"shop/good_web/utils"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	initialize.InitLogger()
	// 初始化翻译器
	initialize.InitTrans("zh")
	// 初始化配置
	initialize.InitConfig()
	// 初始化rpc连接
	initialize.InitSrvConn()

	viper.AutomaticEnv()
	// debug时，port固定
	debug := viper.GetBool("SHOP_DEBUG")
	fmt.Println("SHOP_DEBUG env value:", os.Getenv("SHOP_DEBUG"))
	fmt.Println("debug ", debug)
	if !debug {
		port, err := utils.GetFreePort()
		if err != nil {
			zap.S().Fatalf("get free port failed: %v", err)
		}
		global.ServerConfig.Port = port
	}
	// 初始化路由
	r := initialize.Routers()
	zap.S().Infof("server run at port %s:%d", global.ServerConfig.IP, global.ServerConfig.Port)
	err := r.Run(fmt.Sprintf("%s:%d", global.ServerConfig.IP, global.ServerConfig.Port))
	if err != nil {
		zap.S().Errorf("server run failed: %v", err)
		os.Exit(1)
	}
}
