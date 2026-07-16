package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shop/pkg/proto"
	"shop/services/user_web/auth"
	"shop/services/user_web/config"
	"shop/services/user_web/sms"
	"shop/services/user_web/web"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log := newLogger()
	defer log.Sync()

	cfg, err := config.Load()
	if err != nil {
		log.Panic("failed to load config: ", err)
	}

	if err := web.ConfigureTranslator("zh"); err != nil {
		log.Panic("failed to init translator: ", err)
	}

	smsSvc := sms.New(cfg)
	defer smsSvc.Close()

	userConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.UserSrv.Host, cfg.UserSrv.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Panic("failed to connect to user service: ", err)
	}
	defer userConn.Close()

	srv := web.New(cfg, log, auth.NewJWT(cfg), proto.NewUserClient(userConn), smsSvc)
	httpSrv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           srv.Routers(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Infof("starting server, port: %d", cfg.Port)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Panic("failed to start server: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Error("server shutdown error: ", err)
	}
}

func newLogger() *zap.SugaredLogger {
	l, _ := zap.NewDevelopment()
	return l.Sugar()
}
