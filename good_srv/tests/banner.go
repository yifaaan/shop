package main

import (
	"context"
	"fmt"
	"shop/good_srv/proto"
)

func TestBannerList() {
	resp, err := goodClient.BannerList(context.Background(), &proto.Empty{})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Total)
	for _, banner := range resp.Data {
		fmt.Println(banner.Id, banner.Image, banner.Url)
	}
}

func TestCreateBanner() {
	rsp, err := goodClient.CreateBanner(context.Background(), &proto.BannerRequest{
		Image: "http://example.com/banner1.png",
		Url:   "http://example.com/link1",
		Index: 1,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Id)
}

func TestUpdateBanner(id int32) {
	_, err := goodClient.UpdateBanner(context.Background(), &proto.BannerRequest{
		Id:    id,
		Image: "http://example.com/banner1_updated.png",
		Url:   "http://example.com/link1_updated",
		Index: 2,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Update banner success")
}

func TestDeleteBanner(id int32) {
	_, err := goodClient.DeleteBanner(context.Background(), &proto.BannerRequest{
		Id: id,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Delete banner success")
}
