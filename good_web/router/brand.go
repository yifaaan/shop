package router

import (
	"shop/good_web/api/brand"
	"shop/good_web/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func InitBrandRouter(router *gin.RouterGroup) {
	brandRouter := router.Group("brand")
	zap.S().Info("配置品牌相关的路由...")
	{
		brandRouter.GET("", middleware.JWTAuth(), middleware.AdminAuth(), brand.List)
		brandRouter.POST("", middleware.JWTAuth(), middleware.AdminAuth(), brand.New)
		brandRouter.DELETE("/:id", middleware.JWTAuth(), middleware.AdminAuth(), brand.Delete)
		brandRouter.PUT("/:id", middleware.JWTAuth(), middleware.AdminAuth(), brand.Update)
	}
}
