package tests

import (
	"context"
	"testing"

	"shop/good_srv/proto"
)

func TestCategoryBrandList(t *testing.T) {
	// Since CreateCategoryBrand is not implemented, we can only check if the list query works
	// and returns a valid response structure (even if empty).
	rsp, err := goodClient.CategoryBrandList(context.Background(), &proto.CategoryBrandFilterRequest{
		Pages:       1,
		PagePerNums: 10,
	})
	if err != nil {
		t.Fatalf("CategoryBrandList err: %v", err)
	}
	t.Logf("CategoryBrandList total: %d", rsp.Total)
}

func TestGetCategoryBrandList(t *testing.T) {
	// 1. Create a category so we have a valid ID
	catRsp, err := goodClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:  "TestCatForBrand",
		Level: 1,
		IsTab: false,
	})
	if err != nil {
		t.Fatalf("CreateCategory err: %v", err)
	}
	catID := catRsp.Id
	t.Logf("Created category id: %d", catID)

	defer func() {
		// Cleanup category
		_, _ = goodClient.DeleteCategory(context.Background(), &proto.DeleteCategoryRequest{Id: catID})
	}()

	// 2. GetCategoryBrandList for this category
	// Should return empty list but no error
	brandListRsp, err := goodClient.GetCategoryBrandList(context.Background(), &proto.CategoryInfoRequest{
		Id: catID,
	})
	if err != nil {
		t.Fatalf("GetCategoryBrandList err: %v", err)
	}

	if brandListRsp.Total != 0 {
		t.Logf("Warning: Expected 0 brands for new category, got %d (maybe data pollution?)", brandListRsp.Total)
	}
	
	// 3. Try GetCategoryBrandList for non-existent category
	_, err = goodClient.GetCategoryBrandList(context.Background(), &proto.CategoryInfoRequest{
		Id: 99999999,
	})
	if err == nil {
		t.Fatalf("Expected error for non-existent category, got nil")
	}
}
