package main

import (
	"fmt"
	"shop/services/user_web/initialize"

	"go.uber.org/zap"
)

func main() {
	port := 50000

	initialize.InitLogger()

	router := initialize.Routers()

	zap.S().Infof("starting server, port: %d", port)
	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
		zap.S().Panic("failed to start server: ", err.Error())
	}
}
