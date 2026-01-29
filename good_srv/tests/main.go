package main

import (
	"shop/good_srv/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var goodClient proto.GoodClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.NewClient(":8082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	goodClient = proto.NewGoodClient(conn)
}

func main() {
	Init()
	defer conn.Close()
	// TestBrandList()
	// TestCreateBrand()
	// TestUpdateBrand(1113)
	// TestDeleteBrand(1113)

	// TestBannerList()
	// TestCreateBanner()
	// TestUpdateBanner(5)
	// TestDeleteBanner(5)

	TestGetAllCategoryList()
}
