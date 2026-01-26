package main

import (
	"flag"
	"fmt"
	"net"
	"shop/user_srv/handler"
	"shop/user_srv/proto"

	"google.golang.org/grpc"
)

func main() {
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 8080, "端口号")
	flag.Parse()
	fmt.Println("ip: ", *IP, "port: ", *Port)

	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handler.UserServer{})
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))

	if err != nil {
		panic("failed to listen: " + err.Error())
	}
	err = server.Serve(listener)
	if err != nil {
		panic("failed to serve: " + err.Error())
	}
}
