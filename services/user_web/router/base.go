package router

import (
	"shop/services/user_web/api"

	"github.com/gin-gonic/gin"
)

func InitBaseRouter(router *gin.RouterGroup) {
	baseRouter := router.Group("base")
	baseRouter.GET("captcha", api.GetCaptcha)
	baseRouter.POST("send_sms", api.SendSms)
}
