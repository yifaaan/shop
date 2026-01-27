package config

type MysqlConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
}

type ServerConfig struct {
	Host        string      `mapstructure:"host"`
	Port        int         `mapstructure:"port"`
	MysqlConfig MysqlConfig `mapstructure:"mysql"`
}
