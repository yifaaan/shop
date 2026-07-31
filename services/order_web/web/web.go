package web

import (
	"net/http"

	"shop/pkg/proto"
	"shop/services/order_web/config"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
)

// Server holds the web layer's dependencies, injected from main.
type Server struct {
	cfg      *config.Config
	log      *zap.SugaredLogger
	j        *JWT
	orderSrv proto.OrderClient
	goodsSrv proto.GoodsClient
	invSrv   proto.InventoryClient
}

// New wires a Server with its dependencies.
func New(cfg *config.Config, log *zap.SugaredLogger, j *JWT, orderSrv proto.OrderClient, goodsSrv proto.GoodsClient, invSrv proto.InventoryClient) *Server {
	return &Server{
		cfg:      cfg,
		log:      log,
		j:        j,
		orderSrv: orderSrv,
		goodsSrv: goodsSrv,
		invSrv:   invSrv,
	}
}

// Routers builds the gin engine with all routes registered.
func (s *Server) Routers() *gin.Engine {
	router := gin.Default()
	router.Use(otelgin.Middleware(s.cfg.Name, otelgin.WithFilter(func(req *http.Request) bool {
		return req.URL.Path != "/health"
	})))
	router.Use(Cors())

	router.GET("/health", s.Health) // Consul 健康检查

	apiGroup := router.Group("/o/v1")
	s.log.Debug("registering order router")
	s.registerCartRoutes(apiGroup)
	s.registerOrderRoutes(apiGroup)
	// 支付宝回调（notify 异步 / return 同步）为公开端点，不走用户鉴权，
	// 由支付宝服务器或用户浏览器直接调用，签名校验在 handler 内完成。
	s.registerAlipayRoutes(apiGroup)

	return router
}

// registerCartRoutes 购物车路由：全部需登录（用户私有）
func (s *Server) registerCartRoutes(g *gin.RouterGroup) {
	cg := g.Group("cart", UserAuth(s.j))
	cg.GET("/", s.CartItemList)
	cg.POST("/", s.AddCartItem)
	cg.PUT("/:id", s.UpdateCartItem)
	cg.DELETE("/:id", s.DeleteCartItem)
}

// registerOrderRoutes 订单路由（全部需登录）：
//   - 订单列表角色感知：管理员看全部、普通用户看本人
func (s *Server) registerOrderRoutes(g *gin.RouterGroup) {
	og := g.Group("orders", UserAuth(s.j))
	og.POST("/", s.CreateOrder)
	og.GET("/", s.OrderList)
	og.GET("/:id", s.GetOrderDetail)
	og.PUT("/status", s.UpdateOrderStatus)
	og.DELETE("/:id", s.DeleteOrder)
}

func (s *Server) registerAlipayRoutes(g *gin.RouterGroup) {
	ag := g.Group("pay")
	ag.POST("alipay/notify", s.AlipayNotify)
	ag.GET("alipay/return", s.AlipayReturn)
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
