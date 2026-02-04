package global

import (
	"shop/order_srv/config"
	"shop/order_srv/proto"

	"github.com/go-redsync/redsync/v4"
	"gorm.io/gorm"
)

var (
	DB                 *gorm.DB
	ServerConfig       *config.ServerConfig
	NacosConfig        *config.NacosConfig
	RedisSync          *redsync.Redsync
	GoodSrvClient      proto.GoodClient
	InventorySrvClient proto.InventoryClient
)
