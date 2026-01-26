package initialize

import (
	"shop/shop_api/user_web/router"

	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	r := gin.Default()
	apiGroup := r.Group("v1")
	router.InitUserRouter(apiGroup)
	return r
}
