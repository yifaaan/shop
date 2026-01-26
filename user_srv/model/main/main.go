package main

import (
	"crypto/md5"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/anaskhan96/go-password-encoder"
)

func main() {
	// dsn := "shop_user:shop123456@tcp(127.0.0.1:3306)/shop_user_srv?charset=utf8mb4&parseTime=True&loc=Local"
	// newLogger := logger.New(
	// 	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	// 	logger.Config{

	// 		SlowThreshold: time.Second, // Slow SQL threshold
	// 		LogLevel:      logger.Info, // Log level
	// 		Colorful:      true,        // Disable color
	// 	},
	// )

	// // Globally mode
	// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
	// 	NamingStrategy: schema.NamingStrategy{
	// 		SingularTable: true, // 表名默认为单数
	// 	},
	// 	Logger: newLogger,
	// })
	// if err != nil {
	// 	panic(err)
	// }

	// _ = db.AutoMigrate(&model.User{})

	options := &password.Options{SaltLen: 16, Iterations: 100, KeyLen: 32, HashFunction: sha512.New}
	salt, encodedPwd := password.Encode("generic password", options)
	ps := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)
	passwordInfo := strings.Split(ps, "$")
	check := password.Verify("generic password", passwordInfo[2], passwordInfo[3], options)
	fmt.Println(check) // true
}

func genMd5(code string) string {
	hash := md5.New()
	_, _ = io.WriteString(hash, code)
	return hex.EncodeToString(hash.Sum(nil))
}
