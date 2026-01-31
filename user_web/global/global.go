package global

import (
	"shop/user_web/config"
	"shop/user_web/proto"

	ut "github.com/go-playground/universal-translator"
)

var (
	Trans         ut.Translator
	ServerConfig  *config.ServerConfig
	NacosConfig   *config.NacosConfig
	UserSrvClient proto.UserClient
)
