package initialize

import (
	rt "shop/services/user_web/router"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Routers() *gin.Engine {
	router := gin.Default()
	apiGroup := router.Group("/v1")

	zap.S().Debug("registering user router")
	rt.InitUserRouter(apiGroup)

	return router
}
