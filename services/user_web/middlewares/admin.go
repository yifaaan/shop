package middlewares

import (
	"net/http"
	"shop/pkg/model"

	"github.com/gin-gonic/gin"
)

func IsAdminAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get the user role from the context
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
