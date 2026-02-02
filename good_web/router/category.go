package router

import (
	"shop/good_web/api/category"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func InitCategoryRouter(router *gin.RouterGroup) {
	categoryRouter := router.Group("category")
	zap.S().Info("配置分类相关的路由...")
	{
		categoryRouter.GET("", category.List)
		categoryRouter.POST("", category.New)
		categoryRouter.GET("/:id", category.Detail)
		categoryRouter.DELETE("/:id", category.Delete)
		categoryRouter.PUT("/:id", category.Update)
	}
}
