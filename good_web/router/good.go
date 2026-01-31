package router

import (
	"shop/good_web/api/good"
	"shop/good_web/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func InitGoodRouter(router *gin.RouterGroup) {
	goodRouter := router.Group("good")
	zap.S().Info("配置商品相关的路由...")
	{
		goodRouter.GET("", good.List)
		goodRouter.POST("", middleware.JWTAuth(), middleware.AdminAuth(), good.List)
	}
}
