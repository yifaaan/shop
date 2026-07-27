package web

import (
	"net/http"

	"shop/pkg/proto"
	"shop/services/userop_web/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server holds the web layer's dependencies, injected from main.
type Server struct {
	cfg       *config.Config
	log       *zap.SugaredLogger
	j         *JWT
	useropSrv proto.UserOpClient
	goodsSrv  proto.GoodsClient
}

// New wires a Server with its dependencies.
func New(cfg *config.Config, log *zap.SugaredLogger, j *JWT, useropSrv proto.UserOpClient, goodsSrv proto.GoodsClient) *Server {
	return &Server{
		cfg:       cfg,
		log:       log,
		j:         j,
		useropSrv: useropSrv,
		goodsSrv:  goodsSrv,
	}
}

// Routers builds the gin engine with all routes registered.
func (s *Server) Routers() *gin.Engine {
	router := gin.Default()
	router.Use(Cors())

	router.GET("/health", s.Health) // Consul 健康检查

	apiGroup := router.Group("/uo/v1")
	s.log.Debug("registering userop router")
	s.registerFavRoutes(apiGroup)
	s.registerAddressRoutes(apiGroup)
	s.registerMessageRoutes(apiGroup)

	return router
}

// registerFavRoutes 收藏路由：全部需登录（用户私有），以 goodsId 标识收藏项。
func (s *Server) registerFavRoutes(g *gin.RouterGroup) {
	fg := g.Group("favs", UserAuth(s.j))
	fg.GET("/", s.UserFavList)
	fg.GET("/:goods_id", s.UserFavDetail)
	fg.POST("/", s.CreateUserFav)
	fg.DELETE("/:goods_id", s.DeleteUserFav)
}

// registerAddressRoutes 收货地址路由：全部需登录（用户私有）。
func (s *Server) registerAddressRoutes(g *gin.RouterGroup) {
	ag := g.Group("addresses", UserAuth(s.j))
	ag.GET("/", s.AddressList)
	ag.POST("/", s.CreateAddress)
	ag.PUT("/:id", s.UpdateAddress)
	ag.DELETE("/:id", s.DeleteAddress)
}

// registerMessageRoutes 留言路由：全部需登录（用户私有）。
func (s *Server) registerMessageRoutes(g *gin.RouterGroup) {
	mg := g.Group("messages", UserAuth(s.j))
	mg.GET("/", s.MessageList)
	mg.POST("/", s.CreateMessage)
	mg.DELETE("/:id", s.DeleteMessage)
}

// Health is the liveness/readiness endpoint used by Consul.
func (s *Server) Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Cors is the permissive CORS middleware.
func Cors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		method := ctx.Request.Method

		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, x-token")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, UPDATE, PATCH")
		if method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
	}
}