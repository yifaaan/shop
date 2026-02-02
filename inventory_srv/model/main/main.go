package main

import (
	"log"
	"os"
	"shop/inventory_srv/global"
	"shop/inventory_srv/model"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func main() {
	// dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
	// 	global.ServerConfig.MysqlConfig.User,
	// 	global.ServerConfig.MysqlConfig.Password,
	// 	global.ServerConfig.MysqlConfig.Host,
	// 	global.ServerConfig.MysqlConfig.Port,
	// 	global.ServerConfig.MysqlConfig.DBName,
	// )
	dsn := "root:root123456@tcp(127.0.0.1:3306)/shop_inventory_srv?charset=utf8mb4&parseTime=True&loc=Local"
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
	_ = global.DB.AutoMigrate(&model.Inventory{})

}
