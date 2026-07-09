package global

import (
	"shop/services/user_web/config"

	ut "github.com/go-playground/universal-translator"
)

var (
	ServerConfig *config.Config = &config.Config{}
	Translator   ut.Translator
)