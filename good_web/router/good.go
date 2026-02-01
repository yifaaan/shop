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
		goodRouter.GET("/:id", good.Detail)
		goodRouter.DELETE("/:id", middleware.JWTAuth(), middleware.AdminAuth(), good.Delete)
		goodRouter.GET("/:id/stock", good.Stock)                                                  // 库存
		goodRouter.PUT("/:id", middleware.JWTAuth(), middleware.AdminAuth(), good.Update)         // 更新商品
		goodRouter.PATCH("/:id", middleware.JWTAuth(), middleware.AdminAuth(), good.UpdateStatus) // 更新商品部分状态信息
	}
}
