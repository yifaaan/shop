package web

import (
	"net/http"

	"shop/pkg/model"
	"shop/services/order_web/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// JWT mints and parses the HS256 JWT used for user sessions.
type JWT struct {
	SigningKey []byte
}

// NewJWT builds a JWT from the configured signing key。
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

// UserAuth 通过 x-token 校验登录态，并将 userId（claims.ID）写入 gin.Context。
// 购物车与订单均为用户私有数据，所有接口都依赖此中间件。
func UserAuth(j *JWT) gin.HandlerFunc {
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
		ctx.Set("claims", claims)
		ctx.Set("userId", claims.ID)
		ctx.Next()
	}
}

// currentUserID 从 gin.Context 取出 UserAuth 写入的 userId。
func currentUserID(ctx *gin.Context) int32 {
	return int32(ctx.GetUint("userId"))
}

// isAdmin 判断当前登录用户是否为管理员（AuthorityId == 2）。
func isAdmin(ctx *gin.Context) bool {
	c, _ := ctx.Get("claims")
	claims, ok := c.(*model.CustomClaims)
	return ok && claims.AuthorityId == 2
}
