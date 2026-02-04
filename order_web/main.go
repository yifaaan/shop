package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"shop/order_web/global"
	"shop/order_web/initialize"
	"shop/order_web/utils"
	"shop/order_web/utils/register/consul"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	initialize.InitLogger()
	// 初始化翻译器
	initialize.InitTrans("zh")
	// 初始化配置
	initialize.InitConfig()
	// 初始化rpc连接
	initialize.InitSrvConn()

	viper.AutomaticEnv()
	// debug时，port固定
	debug := viper.GetBool("SHOP_DEBUG")
	fmt.Println("SHOP_DEBUG env value:", os.Getenv("SHOP_DEBUG"))
	fmt.Println("debug ", debug)
	if !debug {
		port, err := utils.GetFreePort()
		if err != nil {
			zap.S().Fatalf("get free port failed: %v", err)
		}
		global.ServerConfig.Port = port
	}
	// 初始化路由
	r := initialize.Routers()
	// 定义 http server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", global.ServerConfig.IP, global.ServerConfig.Port),
		Handler: r,
	}

	// 向consul注册服务
	serviceId := uuid.NewString()
	c := consul.NewRegistry(global.ServerConfig.ConsulConfig.Host, global.ServerConfig.ConsulConfig.Port)
	c.Register(global.ServerConfig.IP, global.ServerConfig.Port, global.ServerConfig.Name, []string{"order-web"}, serviceId)

	zap.S().Infof("server run at port %s:%d", global.ServerConfig.IP, global.ServerConfig.Port)
	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.S().Fatalf("listen: %s\n", err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.S().Info("Shutdown Server ...")

	// 1. 从 consul 注销
	if err := c.DeRegister(serviceId); err != nil {
		zap.S().Info("注销失败:", err)
	} else {
		zap.S().Info("注销成功")
	}

	// 2. 优雅关闭 http 服务
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zap.S().Fatal("Server Shutdown:", err)
	}
	zap.S().Info("Server exiting")
}
