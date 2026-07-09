package router

import (
	"shop/services/user_web/api"

	"github.com/gin-gonic/gin"
)

func InitUserRouter(router *gin.RouterGroup) {
	router.Group("user")

	router.GET("/list", api.GetUserList)
}
