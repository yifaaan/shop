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
		brandRouter.GET("", middleware.JWTAuth(), middleware.AdminAuth(), brand.BrandList)
		brandRouter.POST("", middleware.JWTAuth(), middleware.AdminAuth(), brand.NewBrand)
		brandRouter.DELETE("/:id", middleware.JWTAuth(), middleware.AdminAuth(), brand.DeleteBrand)
		brandRouter.PUT("/:id", middleware.JWTAuth(), middleware.AdminAuth(), brand.UpdateBrand)
	}
	categoryBrandRouter := router.Group("categorybrand")
	zap.S().Info("配置品牌分类关联相关的路由...")
	{
		categoryBrandRouter.GET("", brand.GetCategoryBrandList)
		categoryBrandRouter.GET("/:id", brand.GetCategoryBrand)
		categoryBrandRouter.POST("", brand.NewCategoryBrand)
		categoryBrandRouter.DELETE("/:id", brand.DeleteCategoryBrand)
		categoryBrandRouter.PUT("/:id", brand.UpdateCategoryBrand)
	}
}
