package router

import (
	"shop/user_web/api"

	"github.com/gin-gonic/gin"
)

func InitBaseRouter(router *gin.RouterGroup) {
	baseRouter := router.Group("base")
	{
		// 获取验证码
		baseRouter.GET("captcha", api.GetCaptcha)
		// 发送验证码
		baseRouter.POST("send_sms", api.SendSMS)
	}
}
