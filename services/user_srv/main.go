package main

import (
	"log"
	"net"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"shop/pkg/proto"

	"google.golang.org/grpc"
)

func main() {
	db, err := openDB()
	if err != nil {
		panic("failed to open db: " + err.Error())
	}
	log.Println("DB init done")

	server := grpc.NewServer()
	proto.RegisterUserServer(server, NewUserServer(db))

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		panic("failed to listen:" + err.Error())
	}
	if err := server.Serve(lis); err != nil {
		panic("failed to start grpc:" + err.Error())
	}
}

// openDB opens the MySQL connection and auto-migrates the User table.
func openDB() (*gorm.DB, error) {
	dsn := "shop_user:shop123456@tcp(127.0.0.1:3306)/shop_user_srv?charset=utf8mb4&parseTime=True&loc=Local"
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, err
	}
	return db, nil
}
