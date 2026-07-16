package api

import (
	"fmt"
	"net/http"
	"shop/services/user_web/forms"
	"shop/services/user_web/global"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dypnsapi"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func SendSms(ctx *gin.Context) {
	var form forms.SendSmsForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", global.ServerConfig.Redis.Host, global.ServerConfig.Redis.Port),
		Password: global.ServerConfig.Redis.Password,
		DB:       global.ServerConfig.Redis.DB,
	})
	defer rdb.Close()

	cooldownKey := "sms:cooldown:" + form.Mobile
	codeKey := "sms:code:" + form.Mobile
	requestCtx := ctx.Request.Context()

	ok, err := rdb.SetNX(requestCtx, cooldownKey, "1", time.Minute).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "短信服务暂时不可用"})
		return
	}
	if !ok {
		ctx.JSON(http.StatusTooManyRequests, gin.H{"msg": "请勿频繁发送验证码"})
		return
	}

	rollback := func() {
		_ = rdb.Del(requestCtx, cooldownKey, codeKey).Err()
	}

	cfg := global.ServerConfig.AliyunSMS

	client, err := dypnsapi.NewClientWithAccessKey(
		cfg.RegionID,
		cfg.AccessKeyID,
		cfg.AccessKeySecret,
	)
	if err != nil {
		rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "短信服务初始化失败"})
		return
	}

	client.SetConnectTimeout(3 * time.Second)
	client.SetReadTimeout(8 * time.Second)

	request := dypnsapi.CreateSendSmsVerifyCodeRequest()
	request.Scheme = "https"
	request.PhoneNumber = form.Mobile
	request.SignName = cfg.SignName
	request.TemplateCode = cfg.TemplateCode
	request.TemplateParam = `{"code":"##code##","min":"5"}`
	request.ReturnVerifyCode = requests.NewBoolean(true)

	response, err := client.SendSmsVerifyCode(request)
	if err != nil {
		rollback()
		ctx.JSON(http.StatusBadGateway, gin.H{"msg": "短信发送失败"})
		return
	}

	if response == nil || !response.Success || response.Code != "OK" {
		rollback()
		ctx.JSON(http.StatusBadGateway, gin.H{"msg": "短信发送失败"})
		return
	}

	if response.Model.VerifyCode == "" {
		rollback()
		ctx.JSON(http.StatusBadGateway, gin.H{"msg": "没有获取到验证码"})
		return
	}

	if err := rdb.Set(
		requestCtx,
		codeKey,
		response.Model.VerifyCode,
		time.Duration(cfg.Expire)*time.Second,
	).Err(); err != nil {
		rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "验证码保存失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "发送成功"})
}
