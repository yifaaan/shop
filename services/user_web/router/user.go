package router

import (
	"shop/services/user_web/api"
	"shop/services/user_web/middlewares"

	"github.com/gin-gonic/gin"
)

func InitUserRouter(router *gin.RouterGroup) {
	userGroup := router.Group("user")

	userGroup.GET("/list", middlewares.JWTAuth(), middlewares.IsAdminAuth(), api.GetUserList)
	userGroup.POST("/pwd_login", api.PasswordLogin)
}
