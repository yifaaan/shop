package auth

import (
	"errors"
	"net/http"

	"shop/pkg/model"
	"shop/services/user_web/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var (
	TokenExpired     = errors.New("Token is expired")
	TokenNotValidYet = errors.New("Token not active yet")
	TokenMalformed   = errors.New("That's not even a token")
	TokenInvalid     = errors.New("Couldn't handle this token:")
)

// JWT mints and parses the HS256 JWT used for user sessions.
type JWT struct {
	SigningKey []byte
}

// NewJWT builds a JWT from the configured signing key.
func NewJWT(cfg *config.Config) *JWT {
	return &JWT{SigningKey: []byte(cfg.JWT.SigningKey)}
}

func (j *JWT) CreateToken(claims *model.CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

func (j *JWT) ParseToken(tokenString string) (*model.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.CustomClaims{}, func(token *jwt.Token) (any, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			}
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, TokenExpired
			}
			if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			}
			return nil, TokenInvalid
		}
	}

	if claims, ok := token.Claims.(*model.CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, TokenInvalid
}

// JWTAuth is the gin middleware that authenticates a request via the x-token header.
func JWTAuth(j *JWT) gin.HandlerFunc {
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

// IsAdminAuth is the gin middleware that requires the authenticated user to be an admin.
func IsAdminAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, _ := ctx.Get("claims")
		customClaims, _ := claims.(*model.CustomClaims)
		if customClaims.AuthorityId != 2 {
			ctx.JSON(http.StatusForbidden, gin.H{
				"msg": "You do not have permission to access this resource",
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
