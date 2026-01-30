package tests

import (
	"context"
	"testing"

	"shop/good_srv/proto"
)

func TestGetAllCategoryList(t *testing.T) {
	rsp, err := goodClient.GetAllCategorysList(context.Background(), &proto.Empty{})
	if err != nil {
		t.Fatalf("GetAllCategorysList err: %v", err)
	}
	if rsp.Total == 0 {
		t.Fatalf("expect categories, got 0")
	}
}

func TestGetSubCategory(t *testing.T) {
	// choose a known parent id; adjust if seed data changes
	const parentID int32 = 130364
	rsp, err := goodClient.GetSubCategory(context.Background(), &proto.CategoryListRequest{Id: parentID})
	if err != nil {
		t.Fatalf("GetSubCategory err: %v", err)
	}
	if rsp.Info.Id != parentID {
		t.Fatalf("expected parent id %d got %d", parentID, rsp.Info.Id)
	}
}

func TestCategoryCRUD(t *testing.T) {
	// 1. create root category
	createRsp, err := goodClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:  "TestCatRoot",
		Level: 1,
		IsTab: true,
	})
	if err != nil {
		t.Fatalf("CreateCategory err: %v", err)
	}
	rootID := createRsp.Id
	if rootID == 0 {
		t.Fatalf("CreateCategory returned ID 0")
	}
	t.Logf("created root category id=%d", rootID)

	// 2. create sub category level 2
	subRsp, err := goodClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:           "TestCatSub",
		Level:          2,
		ParentCategory: rootID,
		IsTab:          false,
	})
	if err != nil {
		t.Fatalf("Create subcategory err: %v", err)
	}
	subID := subRsp.Id
	if subID == 0 {
		t.Fatalf("Create subcategory returned ID 0")
	}
	t.Logf("created sub category id=%d", subID)

	// 3. list subcategories via GetSubCategory
	subListRsp, err := goodClient.GetSubCategory(context.Background(), &proto.CategoryListRequest{Id: rootID})
	if err != nil {
		t.Fatalf("GetSubCategory err: %v", err)
	}
	found := false
	for _, c := range subListRsp.SubCategorys {
		if c.Id == subID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("sub category %d not found under root %d", subID, rootID)
	}

	// 4. update sub category name
	_, err = goodClient.UpdateCategory(context.Background(), &proto.CategoryInfoRequest{
		Id:   subID,
		Name: "TestCatSubUpdated",
	})
	if err != nil {
		t.Fatalf("UpdateCategory err: %v", err)
	}

	// 5. delete sub then root
	_, err = goodClient.DeleteCategory(context.Background(), &proto.DeleteCategoryRequest{Id: subID})
	if err != nil {
		t.Fatalf("Delete sub category err: %v", err)
	}
	_, err = goodClient.DeleteCategory(context.Background(), &proto.DeleteCategoryRequest{Id: rootID})
	if err != nil {
		t.Fatalf("Delete root category err: %v", err)
	}
}
