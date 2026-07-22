package web

import (
	"net/http"

	"shop/pkg/proto"
	"shop/services/goods_web/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server holds the web layer's dependencies, injected from main.
type Server struct {
	cfg      *config.Config
	log      *zap.SugaredLogger
	j        *JWT
	goodsSrv proto.GoodsClient
}

// New wires a Server with its dependencies.
func New(cfg *config.Config, log *zap.SugaredLogger, j *JWT, goodsSrv proto.GoodsClient) *Server {
	return &Server{
		cfg:      cfg,
		log:      log,
		j:        j,
		goodsSrv: goodsSrv,
	}
}

// Routers builds the gin engine with all routes registered.
func (s *Server) Routers() *gin.Engine {
	router := gin.Default()
	router.Use(Cors())

	router.GET("/health", s.Health) // Consul 健康检查

	apiGroup := router.Group("/g/v1")
	s.log.Debug("registering goods router")
	s.registerGoodsRoutes(apiGroup)
	s.registerCategoryRoutes(apiGroup)
	s.registerBrandRoutes(apiGroup)
	s.registerBannerRoutes(apiGroup)
	s.registerCategoryBrandRoutes(apiGroup)

	return router
}

// registerGoodsRoutes 商品路由：浏览公开，管理需管理员权限
func (s *Server) registerGoodsRoutes(g *gin.RouterGroup) {
	gg := g.Group("goods")
	// 公开浏览
	gg.GET("/", s.GoodsList)
	gg.GET("/:id", s.GetGoodsDetail)
	// 管理员操作
	admin := gg.Group("", AdminAuth(s.j))
	admin.POST("/", s.CreateGoods)

	admin.PUT("/:id", s.UpdateGoods)
	admin.DELETE("/:id", s.DeleteGoods)
}

// registerCategoryRoutes 分类路由
func (s *Server) registerCategoryRoutes(g *gin.RouterGroup) {
	cg := g.Group("category")
	cg.GET("/", s.GetAllCategorysList)
	cg.GET("/:id", s.GetSubCategory)
	admin := cg.Group("", AdminAuth(s.j))
	admin.POST("/", s.CreateCategory)
	admin.PUT("/:id", s.UpdateCategory)
	admin.DELETE("/:id", s.DeleteCategory)
}

// registerBrandRoutes 品牌路由
func (s *Server) registerBrandRoutes(g *gin.RouterGroup) {
	bg := g.Group("brand")
	bg.GET("/", s.BrandList)
	bg.GET("/:id", s.GetBrandDetail)
	admin := bg.Group("", AdminAuth(s.j))
	admin.POST("/", s.CreateBrand)
	admin.PUT("/:id", s.UpdateBrand)
	admin.DELETE("/:id", s.DeleteBrand)
}

// registerBannerRoutes 轮播图路由
func (s *Server) registerBannerRoutes(g *gin.RouterGroup) {
	bg := g.Group("banner")
	bg.GET("/", s.BannerList)
	admin := bg.Group("", AdminAuth(s.j))
	admin.POST("/", s.CreateBanner)
	admin.PUT("/:id", s.UpdateBanner)
	admin.DELETE("/:id", s.DeleteBanner)
}

// registerCategoryBrandRoutes 分类品牌关系路由
func (s *Server) registerCategoryBrandRoutes(g *gin.RouterGroup) {
	cbg := g.Group("category_brand")
	cbg.GET("/", s.CategoryBrandList)
	cbg.GET("/:category_id", s.GetCategoryBrandList)
	admin := cbg.Group("", AdminAuth(s.j))
	admin.POST("/", s.CreateCategoryBrand)
	admin.PUT("/:id", s.UpdateCategoryBrand)
	admin.DELETE("/:id", s.DeleteCategoryBrand)
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
