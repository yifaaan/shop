package initialize

import (
	"fmt"
	"log"
	"os"
	"shop/user_srv/global"
	"shop/user_srv/model"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func InitDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		global.ServerConfig.MysqlConfig.User,
		global.ServerConfig.MysqlConfig.Password,
		global.ServerConfig.MysqlConfig.Host,
		global.ServerConfig.MysqlConfig.Port,
		global.ServerConfig.MysqlConfig.DBName,
	)
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{

			SlowThreshold: time.Second,   // Slow SQL threshold
			LogLevel:      logger.Silent, // Log level
			Colorful:      true,          // Disable color
		},
	)

	// Globally mode
	var err error
	global.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 表名默认为单数
		},
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}

	_ = global.DB.AutoMigrate(&model.User{})
}
