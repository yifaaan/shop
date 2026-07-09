package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"go.uber.org/zap"
)

var store = base64Captcha.DefaultMemStore

func GetCaptcha(ctx *gin.Context) {
	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)
	base64Captcha := base64Captcha.NewCaptcha(driver, store)
	id, b64s, _, err := base64Captcha.Generate()
	if err != nil {
		zap.S().Error("failed to generate captcha: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "failed to generate captcha",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"captchaId": id,
		"picPath":   b64s,
	})
}
