package web

import (
	"net/http"

	"shop/pkg/proto"
	"shop/services/user_web/auth"
	"shop/services/user_web/config"
	"shop/services/user_web/sms"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server holds the web layer's dependencies, injected from main.
type Server struct {
	cfg     *config.Config
	log     *zap.SugaredLogger
	j       *auth.JWT
	userSrv proto.UserClient
	sms     *sms.Service
}

// New wires a Server with its dependencies.
func New(cfg *config.Config, log *zap.SugaredLogger, j *auth.JWT, userSrv proto.UserClient, smsSvc *sms.Service) *Server {
	return &Server{
		cfg:     cfg,
		log:     log,
		j:       j,
		userSrv: userSrv,
		sms:     smsSvc,
	}
}

// Routers builds the gin engine with all routes registered.
func (s *Server) Routers() *gin.Engine {
	router := gin.Default()
	router.Use(Cors())

	router.GET("/health", s.Health) // Consul 健康检查

	apiGroup := router.Group("/u/v1")
	s.log.Debug("registering user router")
	s.registerUserRoutes(apiGroup)
	s.registerBaseRoutes(apiGroup)

	return router
}

func (s *Server) registerUserRoutes(g *gin.RouterGroup) {
	ug := g.Group("user")
	ug.GET("/list", auth.JWTAuth(s.j), auth.IsAdminAuth(), s.GetUserList)
	ug.POST("/pwd_login", s.PasswordLogin)
	ug.POST("/register", s.Register)
}

func (s *Server) registerBaseRoutes(g *gin.RouterGroup) {
	bg := g.Group("base")
	bg.GET("captcha", s.GetCaptcha)
	bg.POST("send_sms", s.SendSms)
}

// Cors is the permissive CORS middleware.
func Cors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		method := ctx.Request.Method

		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, x-token")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, 	UPDATE, PATCH")
		if method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
	}
}
