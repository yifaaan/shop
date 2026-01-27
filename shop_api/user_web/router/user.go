package router

import (
	"shop/shop_api/user_web/api"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func InitUserRouter(router *gin.RouterGroup) {
	userRouter := router.Group("user")
	zap.S().Info("配置用户相关的路由...")
	{
		userRouter.GET("/list", api.GetUserList)
		userRouter.POST("/login", api.Login)
	}
}
