package main

import (
	"log"
	"net"

	"shop/pkg/proto"
	"shop/services/user_srv/handler"

	"google.golang.org/grpc"
)

func main() {
	log.Println("DB init done")

	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handler.UserServer{})
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		panic("failed to listen:" + err.Error())
	}
	err = server.Serve(lis)
	if err != nil {
		panic("failed to start grpc:" + err.Error())
	}
}
