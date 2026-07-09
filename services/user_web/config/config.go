package config

type Config struct {
	Name    string        `mapstructure:"name"`
	Port    int           `mapstructure:"port"`
	UserSrv UserSrvConfig `mapstructure:"user-srv"`
	JWT     JWTConfig     `mapstructure:"jwt"`
}

type UserSrvConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type JWTConfig struct {
	SigningKey string `mapstructure:"signing-key"`
}
