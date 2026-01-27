package global

import (
	"shop/user_srv/config"

	"gorm.io/gorm"
)

var (
	DB           *gorm.DB
	ServerConfig *config.ServerConfig
)
