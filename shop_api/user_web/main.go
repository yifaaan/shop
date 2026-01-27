package main

import (
	"fmt"
	"os"
	"shop/shop_api/user_web/global"
	"shop/shop_api/user_web/initialize"

	"go.uber.org/zap"
)

func main() {
	initialize.InitLogger()
	initialize.InitTrans("zh")
	initialize.InitConfig()
	r := initialize.Routers()

	zap.S().Debugf("server run at port %s:%d", global.ServerConfig.IP, global.ServerConfig.Port)
	err := r.Run(fmt.Sprintf("%s:%d", global.ServerConfig.IP, global.ServerConfig.Port))
	if err != nil {
		zap.S().Errorf("server run failed: %v", err)
		os.Exit(1)
	}
}
