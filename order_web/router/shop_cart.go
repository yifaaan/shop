package router

import (
	shopcart "shop/order_web/api/shop_cart"
	"shop/order_web/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func InitShopCartRouter(router *gin.RouterGroup) {
	shopCartRouter := router.Group("shopcart").Use(middleware.JWTAuth())
	zap.S().Info("配置购物车相关的路由...")
	{
		shopCartRouter.GET("", shopcart.List)          // 获取购物车商品列表
		shopCartRouter.POST("", shopcart.New)          // 添加商品到购物车
		shopCartRouter.DELETE("/:id", shopcart.Delete) // 删除购物车商品
		shopCartRouter.PATCH("/:id", shopcart.Update)  // 更新购物车商品
	}
}
