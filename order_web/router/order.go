package router

import (
	"shop/order_web/api/order"
	"shop/order_web/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func InitOrderRouter(router *gin.RouterGroup) {
	orderRouter := router.Group("order")
	zap.S().Info("配置订单相关的路由...")
	{
		orderRouter.GET("", middleware.JWTAuth(), middleware.AdminAuth(), order.List)
		orderRouter.POST("", middleware.JWTAuth(), order.New)
		orderRouter.GET("/:id", middleware.JWTAuth(), order.Detail)
	}
}
