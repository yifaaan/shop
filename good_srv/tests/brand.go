package main

import (
	"context"
	"fmt"
	"shop/good_srv/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var brandClient proto.GoodClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.NewClient(":8082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	brandClient = proto.NewGoodClient(conn)
}

func TestBrandList() {
	resp, err := brandClient.BrandList(context.Background(), &proto.BrandFilterRequest{
		Pages:       4,
		PagePerNums: 4,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Total)
	for _, brand := range resp.Data {
		fmt.Println(brand.Name, brand.Logo)
	}
}

func TestCreateBrand() {
	rsp, err := brandClient.CreateBrand(context.Background(), &proto.BrandRequest{
		Name: "YSL",
		Logo: "http://ysl.com/logo.png",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Id)
}

func TestUpdateBrand(id int32) {
	_, err := brandClient.UpdateBrand(context.Background(), &proto.BrandRequest{
		Id:   id,
		Name: "YSL-Updated",
		Logo: "http://ysl.com/new_logo.png",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Update success")
}

func TestDeleteBrand(id int32) {
	_, err := brandClient.DeleteBrand(context.Background(), &proto.BrandRequest{
		Id: id,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Delete success")
}

func main() {
	Init()
	defer conn.Close()
	// TestBrandList()
	// TestCreateBrand()
	// TestUpdateBrand(1113)
	// TestDeleteBrand(1113)
}
