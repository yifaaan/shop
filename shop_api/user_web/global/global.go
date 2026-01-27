package global

import (
	"shop/shop_api/user_web/config"

	ut "github.com/go-playground/universal-translator"
)

var (
	Trans        ut.Translator
	ServerConfig *config.ServerConfig
)
