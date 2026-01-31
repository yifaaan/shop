package middleware

import (
	"errors"
	"net/http"
	"shop/good_web/global"
	"shop/good_web/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	TokenExpired     = errors.New("token已过期")
	TokenNotValidYet = errors.New("token尚未生效")
	TokenMalformed   = errors.New("token格式错误")
	TokenInvalid     = errors.New("token无效")
)

func JWTAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("x-token")
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "请先登录"})
			ctx.Abort()
			return
		}
		j := NewJWT()
		// 解析JWT的信息
		claims, err := j.ParseToken(token)
		if err != nil {
			switch err {
			case TokenExpired:
				ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "token已过期"})
				ctx.Abort()
				return
			case TokenNotValidYet:
				ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "token尚未生效"})
				ctx.Abort()
				return
			case TokenMalformed:
				ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "token格式错误"})
				ctx.Abort()
				return
			case TokenInvalid:
				ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "token无效"})
				ctx.Abort()
				return
			default:
				ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "token解析失败"})
				ctx.Abort()
				return
			}
		}
		ctx.Set("claims", claims)
		ctx.Set("userID", claims.ID)
		ctx.Next()
	}
}

type JWT struct {
	SigningKey []byte
	ExpiresAt  int64
	Issuer     string
}

func NewJWT() *JWT {
	return &JWT{
		SigningKey: []byte(global.ServerConfig.JWTConfig.SigningKey),
		ExpiresAt:  global.ServerConfig.JWTConfig.ExpiresAt,
		Issuer:     global.ServerConfig.JWTConfig.Issuer,
	}
}

// CreateToken 创建JWT
func (j *JWT) CreateToken(claims model.CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// ParseToken 解析JWT
// 解析成功返回CustomClaims，解析失败返回error
func (j *JWT) ParseToken(tokenString string) (*model.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		switch err {
		case jwt.ErrTokenMalformed:
			return nil, TokenMalformed
		case jwt.ErrTokenExpired:
			return nil, TokenExpired
		case jwt.ErrTokenNotValidYet:
			return nil, TokenNotValidYet
		default:
			return nil, TokenInvalid
		}
	}
	if token.Valid {
		return token.Claims.(*model.CustomClaims), nil
	}
	return nil, TokenInvalid
}
