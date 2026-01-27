package main

import (
	"fmt"
	"net"
	"shop/user_srv/global"
	"shop/user_srv/handler"
	"shop/user_srv/initialize"
	"shop/user_srv/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()

	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handler.UserServer{})
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", global.ServerConfig.Host, global.ServerConfig.Port))
	zap.S().Infof("server run at port %s:%d", global.ServerConfig.Host, global.ServerConfig.Port)

	if err != nil {
		panic("failed to listen: " + err.Error())
	}
	err = server.Serve(listener)
	if err != nil {
		panic("failed to serve: " + err.Error())
	}
}
