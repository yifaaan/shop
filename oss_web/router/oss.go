package router

import (
	"shop/oss_web/handler"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func InitOssRouter(router *gin.RouterGroup) {
	ossRouter := router.Group("oss")
	zap.S().Info("配置商品相关的路由...")
	{
		ossRouter.GET("sign", handler.Sign)
		// ossRouter.POST("/callback", handler.HandlerRequest)
	}
}
