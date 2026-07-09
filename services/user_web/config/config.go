package config

type UserSrvConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type Config struct {
	Name    string        `mapstructure:"name"`
	Port    int           `mapstructure:"port"`
	UserSrv UserSrvConfig `mapstructure:"user-srv"`
}
