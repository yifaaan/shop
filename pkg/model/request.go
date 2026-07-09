package model

import (
	"github.com/dgrijalva/jwt-go"
)

type CustomClaims struct {
	jwt.StandardClaims
	ID          uint
	NickName    string
	AuthorityId uint
}
