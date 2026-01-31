package middleware

import (
	"net/http"
	"shop/user_web/model"

	"github.com/gin-gonic/gin"
)

func AdminAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, ok := ctx.Get("claims")
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "请先登录"})
			ctx.Abort()
			return
		}
		claims, _ := c.(*model.CustomClaims)
		if claims.AuthorityID != 2 {
			ctx.JSON(http.StatusForbidden, gin.H{"msg": "您没有权限访问该资源"})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
