package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitBaseRouter(router *gin.RouterGroup) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"success": true,
		})
	})
}
