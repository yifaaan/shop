package sms

import (
	"context"
	"errors"
	"fmt"
	"time"

	"shop/services/user_web/config"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dypnsapi"
	"github.com/redis/go-redis/v9"
)

var (
	ErrServiceUnavailable = errors.New("短信服务暂时不可用")
	ErrTooFrequent        = errors.New("请勿频繁发送验证码")
	ErrInitFailed         = errors.New("短信服务初始化失败")
	ErrSendFailed         = errors.New("短信发送失败")
	ErrNoCode             = errors.New("没有获取到验证码")
	ErrStoreFailed        = errors.New("验证码保存失败")
	ErrCodeExpired        = errors.New("验证码已过期")
)

const (
	cooldownKeyPrefix = "sms:cooldown:"
	codeKeyPrefix     = "sms:code:"
)

// Service sends SMS verification codes via Aliyun and stores them in Redis
// with a one-minute cooldown per mobile number.
type Service struct {
	cfg *config.Config
	rdb *redis.Client
}

// New builds a Service with its own Redis client derived from cfg.
func New(cfg *config.Config) *Service {
	return &Service{
		cfg: cfg,
		rdb: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		}),
	}
}

// Close releases the underlying Redis connection.
func (s *Service) Close() error { return s.rdb.Close() }

// SendCode sends a verification code to mobile and stores it for cfg.AliyunSMS.Expire seconds.
func (s *Service) SendCode(ctx context.Context, mobile string) error {
	cooldownKey := cooldownKeyPrefix + mobile
	codeKey := codeKeyPrefix + mobile

	ok, err := s.rdb.SetNX(ctx, cooldownKey, "1", time.Minute).Result()
	if err != nil {
		return ErrServiceUnavailable
	}
	if !ok {
		return ErrTooFrequent
	}

	rollback := func() { _ = s.rdb.Del(ctx, cooldownKey, codeKey).Err() }

	cfg := s.cfg.AliyunSMS
	client, err := dypnsapi.NewClientWithAccessKey(cfg.RegionID, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		rollback()
		return ErrInitFailed
	}
	client.SetConnectTimeout(3 * time.Second)
	client.SetReadTimeout(8 * time.Second)

	request := dypnsapi.CreateSendSmsVerifyCodeRequest()
	request.Scheme = "https"
	request.PhoneNumber = mobile
	request.SignName = cfg.SignName
	request.TemplateCode = cfg.TemplateCode
	request.TemplateParam = `{"code":"##code##","min":"5"}`
	request.ReturnVerifyCode = requests.NewBoolean(true)

	response, err := client.SendSmsVerifyCode(request)
	if err != nil {
		rollback()
		return ErrSendFailed
	}
	if response == nil || !response.Success || response.Code != "OK" {
		rollback()
		return ErrSendFailed
	}
	if response.Model.VerifyCode == "" {
		rollback()
		return ErrNoCode
	}

	if err := s.rdb.Set(ctx, codeKey, response.Model.VerifyCode, time.Duration(cfg.Expire)*time.Second).Err(); err != nil {
		rollback()
		return ErrStoreFailed
	}
	return nil
}

// VerifyCode reports whether code matches the stored code for mobile.
// It returns ErrCodeExpired when no code is stored.
func (s *Service) VerifyCode(ctx context.Context, mobile, code string) (bool, error) {
	stored, err := s.rdb.Get(ctx, codeKeyPrefix+mobile).Result()
	if err == redis.Nil {
		return false, ErrCodeExpired
	}
	if err != nil {
		return false, err
	}
	return stored == code, nil
}
