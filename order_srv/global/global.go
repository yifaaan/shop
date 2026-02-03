package global

import (
	"shop/order_srv/config"

	"github.com/go-redsync/redsync/v4"
	"gorm.io/gorm"
)

var (
	DB           *gorm.DB
	ServerConfig *config.ServerConfig
	NacosConfig  *config.NacosConfig
	RedisSync    *redsync.Redsync
)
