package model

import (
	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	ID          uint // user id
	NickName    string
	AuthorityID uint // 权限id
}
