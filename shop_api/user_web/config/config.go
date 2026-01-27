package config

// UserSrvConfig 用户rpc服务配置
type UserSrvConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Name       string        `mapstructure:"name"`
	Version    string        `mapstructure:"version"`
	IP         string        `mapstructure:"ip"`
	Port       int           `mapstructure:"port"`
	UserSrvCfg UserSrvConfig `mapstructure:"user_srv"`
	JWTConfig  JWTConfig     `mapstructure:"jwt"`
}

//
type JWTConfig struct {
	SigningKey string `mapstructure:"key"`
	ExpiresAt  int64  `mapstructure:"exp"`
}
