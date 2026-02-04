package config

type MysqlConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
	DBName   string `mapstructure:"db_name" json:"db_name"`
}

type RedisConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type GoodSrvConfig struct {
	Name string `mapstructure:"name" json:"name"`
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type InventorySrvConfig struct {
	Name string `mapstructure:"name" json:"name"`
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type ServerConfig struct {
	Name            string             `mapstructure:"name" json:"name"`
	Host            string             `mapstructure:"host" json:"host"`
	Port            int                `mapstructure:"port" json:"port"`
	MysqlConfig     MysqlConfig        `mapstructure:"mysql" json:"mysql"`
	RedisConfig     RedisConfig        `mapstructure:"redis" json:"redis"`
	ConsulConfig    ConsulConfig       `mapstructure:"consul" json:"consul"`
	GoodSrvCfg      GoodSrvConfig      `mapstructure:"good_srv" json:"good_srv"`
	InventorySrvCfg InventorySrvConfig `mapstructure:"inventory_srv" json:"inventory_srv"`
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
