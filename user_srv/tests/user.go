package main

import (
	"context"
	"fmt"
	"shop/user_srv/global"
	"shop/user_srv/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var userClient proto.UserClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.NewClient(fmt.Sprintf("%s:%d", global.ServerConfig.Host, global.ServerConfig.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	userClient = proto.NewUserClient(conn)
}

func TestGetUserList() {
	resp, err := userClient.GetUserList(context.Background(), &proto.PageInfoRequest{
		PageNumber: 1,
		PageSize:   4,
	})
	if err != nil {
		panic(err)
	}
	for _, u := range resp.Data {
		fmt.Println(u.Mobile, u.NickName, u.Gender, u.Birthday)
		resp, err := userClient.CheckPassword(context.Background(), &proto.CheckPasswordInfoRequest{
			Password:          "123456",
			EncryptedPassword: u.Password,
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(resp.Success)
	}
}

func main() {
	Init()
	TestGetUserList()
	defer conn.Close()

}
