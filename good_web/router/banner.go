package router

import (
	"shop/good_web/api/banner"
	"shop/good_web/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func InitBannerRouter(router *gin.RouterGroup) {
	bannerRouter := router.Group("banner")
	zap.S().Info("配置轮播图相关的路由...")
	{
		bannerRouter.GET("", middleware.JWTAuth(), middleware.AdminAuth(), banner.List)
		bannerRouter.POST("", middleware.JWTAuth(), middleware.AdminAuth(), banner.New)
		bannerRouter.DELETE("/:id", middleware.JWTAuth(), middleware.AdminAuth(), banner.Delete)
		bannerRouter.PUT("/:id", middleware.JWTAuth(), middleware.AdminAuth(), banner.Update)
	}
}
