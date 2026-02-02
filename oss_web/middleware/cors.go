package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, x-token, token")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	}
}
