package global

import (
	"shop/oss_web/config"

	ut "github.com/go-playground/universal-translator"
)

var (
	Trans        ut.Translator
	ServerConfig *config.ServerConfig
	NacosConfig  *config.NacosConfig
)
