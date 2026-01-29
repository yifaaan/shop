package main

import (
	"context"
	"fmt"
	"shop/good_srv/proto"
)

func TestBrandList() {
	resp, err := goodClient.BrandList(context.Background(), &proto.BrandFilterRequest{
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
	rsp, err := goodClient.CreateBrand(context.Background(), &proto.BrandRequest{
		Name: "YSL",
		Logo: "http://ysl.com/logo.png",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Id)
}

func TestUpdateBrand(id int32) {
	_, err := goodClient.UpdateBrand(context.Background(), &proto.BrandRequest{
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
	_, err := goodClient.DeleteBrand(context.Background(), &proto.BrandRequest{
		Id: id,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Delete success")
}
