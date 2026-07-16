package config

type Config struct {
	Name      string          `mapstructure:"name"`
	Port      int             `mapstructure:"port"`
	UserSrv   UserSrvConfig   `mapstructure:"user-srv"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	AliyunSMS AliyunSMSConfig `mapstructure:"aliyun-sms"`
	Redis     RedisConfig     `mapstructure:"redis"`
}

type UserSrvConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type JWTConfig struct {
	SigningKey string `mapstructure:"signing_key"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AliyunSMSConfig struct {
	RegionID        string `mapstructure:"region_id"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	SignName        string `mapstructure:"sign_name"`
	TemplateCode    string `mapstructure:"template_code"`
	Expire          int    `mapstructure:"expire"`
}
