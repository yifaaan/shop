package main

import (
	"fmt"
	"log"

	"shop/services/user_srv/global"
	"shop/services/user_srv/model"

	"crypto/sha512"

	"github.com/anaskhan96/go-password-encoder"
)

func main() {
	log.Println("DB init done")
	_ = global.DB.AutoMigrate(&model.User{})
}

func genPwd(code string) string {
	options := &password.Options{10, 20, 16, sha512.New}
	salt, encodedPwd := password.Encode(code, options)

	newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)
	return newPassword
}