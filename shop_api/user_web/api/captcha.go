package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"go.uber.org/zap"
)

// 验证码存储
var store = base64Captcha.DefaultMemStore

// GetCaptcha 获取验证码
func GetCaptcha(ctx *gin.Context) {
	// 验证码驱动类型
	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)
	cp := base64Captcha.NewCaptcha(driver, store)
	// id, img, ans, err
	id, b64s, _, err := cp.Generate()
	if err != nil {
		zap.S().Errorw("[GetCaptcha] 生成验证码失败", "msg", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "生成验证码失败"})
		return
	}
	zap.S().Debugw("[GetCaptcha] 验证码生成成功", "captchaId", id, "captcha", b64s)
	ctx.JSON(http.StatusOK, gin.H{"captchaId": id, "picPath": b64s})
}
