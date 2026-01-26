package main

import (
	"os"
	"shop/shop_api/user_web/initialize"

	"go.uber.org/zap"
)

func main() {
	initialize.InitLogger()
	r := initialize.Routers()

	zap.S().Debugf("server run at port %s", ":8080")
	err := r.Run(":8080")
	if err != nil {
		zap.S().Errorf("server run failed: %v", err)
		os.Exit(1)
	}
}
