package initialize

import (
	"shop/oss_web/middleware"
	"shop/oss_web/router"

	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	r := gin.Default()

	// CORS 应该尽量挂在全局：
	// - 预检 OPTIONS 可能不会命中具体路由（404/405），导致浏览器报“没有 Access-Control-Allow-Origin”
	// - 放在 group 里只能覆盖该 group 下命中的路由
	r.Use(middleware.Cors())

	apiGroup := r.Group("oss/v1")
	router.InitBaseRouter(apiGroup)
	router.InitOssRouter(apiGroup)
	return r
}
