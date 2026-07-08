package main

import (
	"fmt"
	"log"
	"os"
	"shop/services/user_srv/model"
	"time"

	"crypto/sha512"

	"github.com/anaskhan96/go-password-encoder"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func main() {
	dsn := "shop_user:shop123456@tcp(127.0.0.1:3306)/shop_user_srv?charset=utf8mb4&parseTime=True&loc=Local"
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{

			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // Disable color
		},
	)

	// Globally mode
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 表名默认为单数
		},
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}

	_ = db.AutoMigrate(&model.User{})

}

func genPwd(code string) string {

	// Using custom options
	options := &password.Options{10, 20, 16, sha512.New}
	salt, encodedPwd := password.Encode(code, options)

	newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)
	// parts := strings.Split(newPassword, "$")

	// check := password.Verify(code, parts[2], parts[3], options)
	// fmt.Println(check) // true
	return newPassword
}
