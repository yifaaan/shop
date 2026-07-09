package main

import (
	"fmt"
	"shop/services/user_web/global"
	"shop/services/user_web/initialize"

	"go.uber.org/zap"
)

func main() {
	initialize.InitLogger()
	initialize.InitConfig()

	if err := initialize.InitTranslator("zh"); err != nil {
		zap.S().Panic("failed to init translator: ", err.Error())
	}

	router := initialize.Routers()

	zap.S().Infof("starting server, port: %d", global.ServerConfig.Port)
	if err := router.Run(fmt.Sprintf(":%d", global.ServerConfig.Port)); err != nil {
		zap.S().Panic("failed to start server: ", err.Error())
	}
}
