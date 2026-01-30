package tests

import (
	"os"
	"testing"

	"shop/good_srv/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var goodClient proto.GoodClient
var conn *grpc.ClientConn

// TestMain 初始化 gRPC 连接，所有测试共用，结束后关闭
func TestMain(m *testing.M) {
	var err error
	conn, err = grpc.NewClient(":8082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	goodClient = proto.NewGoodClient(conn)

	code := m.Run()

	_ = conn.Close()
	os.Exit(code)
}
