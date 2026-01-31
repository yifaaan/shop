package router

import (
	"shop/user_web/api"
	"shop/user_web/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func InitUserRouter(router *gin.RouterGroup) {
	userRouter := router.Group("user")
	zap.S().Info("配置用户相关的路由...")
	{
		userRouter.GET("/list", middleware.JWTAuth(), middleware.AdminAuth(), api.GetUserList)
		userRouter.POST("/pwd_login", api.LoginPassword)
		userRouter.POST("/register", api.Register)
	}
}
