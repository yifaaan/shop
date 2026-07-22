package web

import (
	"net/http"

	"shop/pkg/model"
	"shop/services/goods_web/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// JWT mints and parses the HS256 JWT used for admin operations.
type JWT struct {
	SigningKey []byte
}

// NewJWT builds a JWT from the configured signing key.
func NewJWT(cfg *config.Config) *JWT {
	return &JWT{SigningKey: []byte(cfg.JWT.SigningKey)}
}

func (j *JWT) ParseToken(tokenString string) (*model.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.CustomClaims{}, func(token *jwt.Token) (any, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*model.CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrInvalidKey
}

// AdminAuth is the gin middleware that authenticates a request via the x-token header
// and requires the user to be an admin (AuthorityId == 2) for goods management.
func AdminAuth(j *JWT) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Request.Header.Get("x-token")
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "未登录"})
			ctx.Abort()
			return
		}
		claims, err := j.ParseToken(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "登录已过期"})
			ctx.Abort()
			return
		}
		if claims.AuthorityId != 2 {
			ctx.JSON(http.StatusForbidden, gin.H{"msg": "无管理权限"})
			ctx.Abort()
			return
		}
		ctx.Set("claims", claims)
		ctx.Set("userId", claims.ID)
		ctx.Next()
	}
}