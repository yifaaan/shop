package config

type ServerConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	IP      string `mapstructure:"ip"`
	Port    int    `mapstructure:"port"`
}
