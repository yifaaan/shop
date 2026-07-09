package initialize

import (
	"shop/services/user_web/middlewares"
	rt "shop/services/user_web/router"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Routers() *gin.Engine {
	router := gin.Default()
	router.Use(middlewares.Cors())

	apiGroup := router.Group("/u/v1")

	zap.S().Debug("registering user router")
	rt.InitUserRouter(apiGroup)

	return router
}
