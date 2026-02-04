package global

import (
	"shop/order_web/config"
	"shop/order_web/proto"

	ut "github.com/go-playground/universal-translator"
)

var (
	Trans          ut.Translator
	ServerConfig   *config.ServerConfig
	NacosConfig    *config.NacosConfig
	GoodSrvClient  proto.GoodClient
	OrderSrvClient proto.OrderClient
	InvSrvClient   proto.InventoryClient
)
