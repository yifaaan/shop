package config

// ServerConfig 服务配置
type ServerConfig struct {
	Name         string       `mapstructure:"name" json:"name"`
	Version      string       `mapstructure:"version" json:"version"`
	IP           string       `mapstructure:"ip" json:"ip"`
	Port         int          `mapstructure:"port" json:"port"`
	JWTConfig    JWTConfig    `mapstructure:"jwt" json:"jwt"`
	RedisConfig  RedisConfig  `mapstructure:"redis" json:"redis"`
	ConsulConfig ConsulConfig `mapstructure:"consul" json:"consul"`
	OSSConfig    OSSConfig    `mapstructure:"oss" json:"oss"`
}

type JWTConfig struct {
	SigningKey string `mapstructure:"key" json:"key"`
	ExpiresAt  int64  `mapstructure:"exp" json:"exp"`
	Issuer     string `mapstructure:"issuer" json:"issuer"`
}

type RedisConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type ConsulConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type NacosConfig struct {
	Host      string `mapstructure:"host"`
	Port      uint64 `mapstructure:"port"`
	GrpcPort  uint64 `mapstructure:"grpc_port"`
	Namespace string `mapstructure:"namespace"`
	DataID    string `mapstructure:"data_id"`
	Group     string `mapstructure:"group"`
}

type OSSConfig struct {
	Endpoint        string `mapstructure:"endpoint" json:"endpoint"`
	Bucket          string `mapstructure:"bucket" json:"bucket"`
	AccessKeyID     string `mapstructure:"key" json:"key"`
	AccessKeySecret string `mapstructure:"secret" json:"secret"`
	UploadDir       string `mapstructure:"upload_dir" json:"upload_dir"`
	// CallBackURL     string `mapstructure:"callback_url" json:"callback_url"`
}
