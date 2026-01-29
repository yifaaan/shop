package main

import (
	"context"
	"fmt"
	"shop/good_srv/proto"
)

func TestGetAllCategoryList() {
	resp, err := goodClient.GetAllCategorysList(context.Background(), &proto.Empty{})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Total)
	fmt.Println(resp.JsonData)
}

// func TestCreateCategory() {
// 	rsp, err := goodClient.CreateCategory(context.Background(), &proto.CategoryRequest{
// 		Name: "YSL",
// 		Logo: "http://ysl.com/logo.png",
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(rsp.Id)
// }

// func TestUpdateCategory(id int32) {
// 	_, err := goodClient.UpdateCategory(context.Background(), &proto.CategoryRequest{
// 		Id:   id,
// 		Name: "YSL-Updated",
// 		Logo: "http://ysl.com/new_logo.png",
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Update success")
// }

// func TestDeleteCategory(id int32) {
// 	_, err := goodClient.DeleteCategory(context.Background(), &proto.CategoryRequest{
// 		Id: id,
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Delete success")
// }
