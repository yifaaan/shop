package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"shop/pkg/proto"
	"shop/services/user_srv/config"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func main() {
	log := newLogger(false)
	defer func() { _ = log.Sync() }()

	cfg, err := config.Load()
	if err != nil {
		log.Panic("failed to load config: ", err)
	}
	log = newLogger(cfg.Debug)

	db, err := openDB(cfg.MySQL, log)
	if err != nil {
		log.Panic("failed to open db: ", err)
	}
	log.Info("DB init done")

	server := grpc.NewServer()
	proto.RegisterUserServer(server, NewUserServer(db))

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	if err != nil {
		log.Panic("failed to listen: ", err)
	}
	log.Infof("starting grpc server on %s:%d", cfg.Host, cfg.Port)
	if err := server.Serve(lis); err != nil {
		log.Panic("failed to start grpc: ", err)
	}
}

// openDB opens the MySQL connection and auto-migrates the User table.
// gorm log output is routed through the injected zap logger via zapWriter.
func openDB(cfg config.MySQLConfig, log *zap.SugaredLogger) (*gorm.DB, error) {
	gormLogger := logger.New(
		zapWriter{log: log},
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      false,
		},
	)

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, err
	}
	return db, nil
}

// newLogger returns a dev (console, debug-level) or prod (json, info-level) sugared logger.
func newLogger(debug bool) *zap.SugaredLogger {
	var (
		l   *zap.Logger
		err error
	)
	if debug {
		l, err = zap.NewDevelopment()
	} else {
		l, err = zap.NewProduction()
	}
	if err != nil {
		panic("failed to init logger: " + err.Error())
	}
	return l.Sugar()
}

// zapWriter adapts a *zap.SugaredLogger to gorm's logger.Writer interface.
type zapWriter struct {
	log *zap.SugaredLogger
}

func (w zapWriter) Printf(format string, args ...interface{}) {
	w.log.Info(strings.TrimRight(fmt.Sprintf(format, args...), "\n"))
}