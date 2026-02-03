package tests

import (
	"os"
	"testing"

	"shop/inventory_srv/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var inventoryClient proto.InventoryClient
var conn *grpc.ClientConn

// TestMain 初始化 gRPC 连接，所有测试共用，结束后关闭
func TestMain(m *testing.M) {
	addr := os.Getenv("INVENTORY_SRV_ADDR")
	if addr == "" {
		addr = "127.0.0.1:40004"
	}

	var err error
	conn, err = grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	inventoryClient = proto.NewInventoryClient(conn)

	code := m.Run()

	_ = conn.Close()
	os.Exit(code)
}
