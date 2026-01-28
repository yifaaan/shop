package global

import (
	"shop/shop_api/user_web/config"
	"shop/shop_api/user_web/proto"

	ut "github.com/go-playground/universal-translator"
)

var (
	Trans         ut.Translator
	ServerConfig  *config.ServerConfig
	UserSrvClient proto.UserClient
)
