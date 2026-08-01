package ratelimit

import (
	"fmt"
	"net/http"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/gin-gonic/gin"
)

const GoodsListResource = "goods_web.goods.list"

type Config struct {
	GoodsListQPS float64
}

// Init initializes Sentinel and installs the goods list QPS rule.
func Init(cfg Config) error {
	if cfg.GoodsListQPS <= 0 {
		return fmt.Errorf("goods list QPS must be greater than zero")
	}

	if err := sentinel.InitDefault(); err != nil {
		return fmt.Errorf("initialize Sentinel: %w", err)
	}
	if _, err := flow.LoadRules([]*flow.Rule{
		{
			Resource:               GoodsListResource,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			Threshold:              cfg.GoodsListQPS,
			StatIntervalInMs:       1000,
		},
	}); err != nil {
		return fmt.Errorf("load Sentinel flow rules: %w", err)
	}
	return nil
}

// Middleware rejects requests that exceed the rule for resource.
func Middleware(resource string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		entry, blockErr := sentinel.Entry(resource, sentinel.WithTrafficType(base.Inbound))
		if blockErr != nil {
			ctx.Header("Retry-After", "1")
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"msg": "too many requests",
			})
			return
		}

		defer entry.Exit()
		ctx.Next()
	}
}
