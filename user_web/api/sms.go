package api

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"shop/user_web/form"
	"shop/user_web/global"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/openapi-util/service"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// SendSMS 发送验证码
func SendSMS(ctx *gin.Context) {
	smsForm := form.SendSmsForm{}
	if err := ctx.ShouldBind(&smsForm); err != nil {
		zap.S().Errorw("[SendSMS] 参数绑定失败", "msg", err.Error())
		HandleValidatorError(ctx, err)
		return
	}
	smsForm.Mobile = "13186102265"
	if err := run(smsForm.Mobile); err != nil {
		zap.S().Errorw("[SendSMS] 发送验证码失败", "msg", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "验证码发送成功"})
}

// newClient 基于环境变量创建 Dypnsapi Client
func newClient() (*openapi.Client, error) {
	akID := os.Getenv("ALIYUN_AK_ID")
	akSecret := os.Getenv("ALIYUN_AK_SECRET")
	if akID == "" || akSecret == "" {
		return nil, fmt.Errorf("missing ALIYUN_AK_ID/ALIYUN_AK_SECRET env vars")
	}

	credCfg := credential.Config{
		Type:            tea.String("access_key"),
		AccessKeyId:     tea.String(akID),
		AccessKeySecret: tea.String(akSecret),
	}
	cred, err := credential.NewCredential(&credCfg)
	if err != nil {
		return nil, err
	}

	cfg := &openapi.Config{
		Credential: cred,
		Endpoint:   tea.String("dypnsapi.aliyuncs.com"),
	}
	return openapi.NewClient(cfg)
}

func genSmsCode(width int) string {
	nums := "0123456789"
	len := len(nums)
	s := strings.Builder{}
	for i := 0; i < width; i++ {
		s.WriteByte(nums[rand.Intn(len)])
	}
	return s.String()
}

// newAPIParams 返回 SendSmsVerifyCode API 的固定参数
func newAPIParams() *openapi.Params {
	return &openapi.Params{
		Action:      tea.String("SendSmsVerifyCode"),
		Version:     tea.String("2017-05-25"),
		Protocol:    tea.String("HTTPS"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("RPC"),
		Pathname:    tea.String("/"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
}

func run(mobile string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	code := genSmsCode(5)
	queries := map[string]interface{}{
		"SchemeName":       "测试方案",
		"CountryCode":      "86",
		"PhoneNumber":      mobile,
		"SignName":         "速通互联验证码",
		"TemplateCode":     "100001",
		"TemplateParam":    `{"code":"` + code + `","min":"5"}`,
		"CodeLength":       4,
		"ValidTime":        300,
		"DuplicatePolicy":  1,
		"Interval":         60,
		"CodeType":         1,
		"ReturnVerifyCode": true,
		"AutoRetry":        1,
	}

	request := &openapi.OpenApiRequest{Query: openapiutil.Query(queries)}
	runtime := &util.RuntimeOptions{}

	resp, err := client.CallApi(newAPIParams(), request, runtime)
	if err != nil {
		return err
	}

	zap.S().Debugf("sms api response: %v", resp)

	// 保存验证码
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", global.ServerConfig.RedisConfig.Host, global.ServerConfig.RedisConfig.Port),
	})
	if err := rdb.Set(context.Background(), "sms_code_"+mobile, code, 300*time.Second).Err(); err != nil {
		zap.S().Errorw("[SendSMS] 保存验证码到Redis失败", "msg", err.Error())
		return err
	}
	zap.S().Infow("[SendSMS] 验证码保存成功", "mobile", mobile, "code", code)
	return nil
}
