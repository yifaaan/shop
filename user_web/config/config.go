package config

// UserSrvConfig 用户rpc服务配置
type UserSrvConfig struct {
	Name string `mapstructure:"name" json:"name"`
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Name         string        `mapstructure:"name" json:"name"`
	Version      string        `mapstructure:"version" json:"version"`
	IP           string        `mapstructure:"ip" json:"ip"`
	Port         int           `mapstructure:"port" json:"port"`
	UserSrvCfg   UserSrvConfig `mapstructure:"user_srv" json:"user_srv"`
	JWTConfig    JWTConfig     `mapstructure:"jwt" json:"jwt"`
	RedisConfig  RedisConfig   `mapstructure:"redis" json:"redis"`
	ConsulConfig ConsulConfig  `mapstructure:"consul" json:"consul"`
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
