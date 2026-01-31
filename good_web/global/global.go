package global

import (
	"shop/good_web/config"
	"shop/good_web/proto"

	ut "github.com/go-playground/universal-translator"
)

var (
	Trans         ut.Translator
	ServerConfig  *config.ServerConfig
	NacosConfig   *config.NacosConfig
	GoodSrvClient proto.GoodClient
)
