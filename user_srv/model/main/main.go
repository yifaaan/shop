package main

import (
	"crypto/md5"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"shop/user_srv/global"
	"shop/user_srv/model"

	"github.com/anaskhan96/go-password-encoder"
)

func main() {
	options := &password.Options{SaltLen: 16, Iterations: 100, KeyLen: 32, HashFunction: sha512.New}
	salt, encodedPwd := password.Encode("123456", options)
	newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)
	for i := 0; i < 10; i++ {
		user := model.User{
			NickName: fmt.Sprintf("user%d", i),
			Mobile:   fmt.Sprintf("1380000000%d", i),
			Password: newPassword,
		}
		global.DB.Create(&user)
	}
}

func genMd5(code string) string {
	hash := md5.New()
	_, _ = io.WriteString(hash, code)
	return hex.EncodeToString(hash.Sum(nil))
}
