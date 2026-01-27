package main

import (
	"os"
	"shop/shop_api/user_web/initialize"

	"go.uber.org/zap"
)

func main() {
	initialize.InitLogger()
	initialize.InitTrans("zh")
	initialize.InitConfig()
	r := initialize.Routers()

	zap.S().Debugf("server run at port %s", ":8081")
	err := r.Run(":8081")
	if err != nil {
		zap.S().Errorf("server run failed: %v", err)
		os.Exit(1)
	}
}
