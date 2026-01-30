package tests

import (
	"context"
	"testing"

	"shop/good_srv/proto"
)

func TestCategoryBrandList(t *testing.T) {
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

	// 2. Create a brand
	brandRsp, err := goodClient.CreateBrand(context.Background(), &proto.BrandRequest{
		Name: "TestBrandForCat",
		Logo: "http://logo.com",
	})
	if err != nil {
		t.Fatalf("CreateBrand err: %v", err)
	}
	brandID := brandRsp.Id
	t.Logf("Created brand id: %d", brandID)

	defer func() {
		// Cleanup
		_, _ = goodClient.DeleteCategory(context.Background(), &proto.DeleteCategoryRequest{Id: catID})
		_, _ = goodClient.DeleteBrand(context.Background(), &proto.BrandRequest{Id: brandID})
	}()

	// 3. Create relation
	_, err = goodClient.CreateCategoryBrand(context.Background(), &proto.CategoryBrandRequest{
		CategoryId: catID,
		BrandId:    brandID,
	})
	if err != nil {
		t.Fatalf("CreateCategoryBrand err: %v", err)
	}

	// 4. GetCategoryBrandList for this category
	brandListRsp, err := goodClient.GetCategoryBrandList(context.Background(), &proto.CategoryInfoRequest{
		Id: catID,
	})
	if err != nil {
		t.Fatalf("GetCategoryBrandList err: %v", err)
	}

	if brandListRsp.Total != 1 {
		t.Fatalf("Expected 1 brand, got %d", brandListRsp.Total)
	}
	if brandListRsp.Data[0].Id != brandID {
		t.Fatalf("Expected brand ID %d, got %d", brandID, brandListRsp.Data[0].Id)
	}

	// 5. Try GetCategoryBrandList for non-existent category
	_, err = goodClient.GetCategoryBrandList(context.Background(), &proto.CategoryInfoRequest{
		Id: 99999999,
	})
	if err == nil {
		t.Fatalf("Expected error for non-existent category, got nil")
	}
}

func TestCreateCategoryBrand(t *testing.T) {
	// 1. Setup: Create Category and Brand
	catRsp, err := goodClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:  "TestCatCreateRel",
		Level: 1,
	})
	if err != nil {
		t.Fatalf("Setup category err: %v", err)
	}
	brandRsp, err := goodClient.CreateBrand(context.Background(), &proto.BrandRequest{
		Name: "TestBrandCreateRel",
		Logo: "http://logo.com",
	})
	if err != nil {
		t.Fatalf("Setup brand err: %v", err)
	}
	
	catID := catRsp.Id
	brandID := brandRsp.Id

	defer func() {
		_, _ = goodClient.DeleteCategory(context.Background(), &proto.DeleteCategoryRequest{Id: catID})
		_, _ = goodClient.DeleteBrand(context.Background(), &proto.BrandRequest{Id: brandID})
	}()

	// 2. Create Relation
	rsp, err := goodClient.CreateCategoryBrand(context.Background(), &proto.CategoryBrandRequest{
		CategoryId: catID,
		BrandId:    brandID,
	})
	if err != nil {
		t.Fatalf("CreateCategoryBrand err: %v", err)
	}

	// 3. Verify Response
	if rsp.Category.Id != catID {
		t.Errorf("Response category ID mismatch: want %d, got %d", catID, rsp.Category.Id)
	}
	if rsp.Brand.Id != brandID {
		t.Errorf("Response brand ID mismatch: want %d, got %d", brandID, rsp.Brand.Id)
	}

	// 4. Verify Duplicate Creation Fails
	_, err = goodClient.CreateCategoryBrand(context.Background(), &proto.CategoryBrandRequest{
		CategoryId: catID,
		BrandId:    brandID,
	})
	if err == nil {
		t.Errorf("Expected error when creating duplicate relation, got nil")
	}

	// 5. Verify Invalid IDs
	_, err = goodClient.CreateCategoryBrand(context.Background(), &proto.CategoryBrandRequest{
		CategoryId: 999999,
		BrandId:    brandID,
	})
	if err == nil {
		t.Errorf("Expected error for invalid category ID")
	}
}

func TestUpdateCategoryBrand(t *testing.T) {
	// Setup: 1 Category, 2 Brands, 1 Relation
	catRsp, _ := goodClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{Name: "TestCatUpdate", Level: 1})
	brandRsp1, _ := goodClient.CreateBrand(context.Background(), &proto.BrandRequest{Name: "TestBrandUpdate1", Logo: "l1"})
	brandRsp2, _ := goodClient.CreateBrand(context.Background(), &proto.BrandRequest{Name: "TestBrandUpdate2", Logo: "l2"})
	
	catID := catRsp.Id
	b1ID := brandRsp1.Id
	b2ID := brandRsp2.Id

	defer func() {
		goodClient.DeleteCategory(context.Background(), &proto.DeleteCategoryRequest{Id: catID})
		goodClient.DeleteBrand(context.Background(), &proto.BrandRequest{Id: b1ID})
		goodClient.DeleteBrand(context.Background(), &proto.BrandRequest{Id: b2ID})
	}()

	createRsp, err := goodClient.CreateCategoryBrand(context.Background(), &proto.CategoryBrandRequest{
		CategoryId: catID,
		BrandId:    b1ID,
	})
	if err != nil {
		t.Fatalf("Setup create relation err: %v", err)
	}
	relationID := createRsp.Id

	// 1. Update relation: Change Brand from b1 to b2
	_, err = goodClient.UpdateCategoryBrand(context.Background(), &proto.CategoryBrandRequest{
		Id:      relationID,
		BrandId: b2ID,
	})
	if err != nil {
		t.Fatalf("UpdateCategoryBrand err: %v", err)
	}

	// Verify update
	listRsp, _ := goodClient.GetCategoryBrandList(context.Background(), &proto.CategoryInfoRequest{Id: catID})
	if listRsp.Total != 1 || listRsp.Data[0].Id != b2ID {
		t.Errorf("Update failed: expected brand %d, got %+v", b2ID, listRsp.Data)
	}

	// 2. Try update with invalid Brand ID
	_, err = goodClient.UpdateCategoryBrand(context.Background(), &proto.CategoryBrandRequest{
		Id:      relationID,
		BrandId: 999999,
	})
	if err == nil {
		t.Errorf("Expected error for invalid brand ID")
	}
	
	// Clean up relation manually since we'll reuse IDs in other tests potentially? 
	// (Defer handles base entities, relation is deleted by cascade or manually?) 
	// Relation usually cascades but let's delete explicitly to test delete next.
	goodClient.DeleteCategoryBrand(context.Background(), &proto.CategoryBrandRequest{Id: relationID})
}

func TestDeleteCategoryBrand(t *testing.T) {
	// Setup
	catRsp, _ := goodClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{Name: "TestCatDel", Level: 1})
	brandRsp, _ := goodClient.CreateBrand(context.Background(), &proto.BrandRequest{Name: "TestBrandDel", Logo: "l"})
	
	catID := catRsp.Id
	brandID := brandRsp.Id

	defer func() {
		goodClient.DeleteCategory(context.Background(), &proto.DeleteCategoryRequest{Id: catID})
		goodClient.DeleteBrand(context.Background(), &proto.BrandRequest{Id: brandID})
	}()

	createRsp, err := goodClient.CreateCategoryBrand(context.Background(), &proto.CategoryBrandRequest{
		CategoryId: catID,
		BrandId:    brandID,
	})
	if err != nil {
		t.Fatalf("Setup create relation err: %v", err)
	}
	relationID := createRsp.Id

	// 1. Delete Relation
	_, err = goodClient.DeleteCategoryBrand(context.Background(), &proto.CategoryBrandRequest{
		Id: relationID,
	})
	if err != nil {
		t.Fatalf("DeleteCategoryBrand err: %v", err)
	}

	// 2. Verify Deletion
	listRsp, _ := goodClient.GetCategoryBrandList(context.Background(), &proto.CategoryInfoRequest{Id: catID})
	if listRsp.Total != 0 {
		t.Errorf("Expected 0 brands after delete, got %d", listRsp.Total)
	}

	// 3. Try Delete Non-existent
	_, err = goodClient.DeleteCategoryBrand(context.Background(), &proto.CategoryBrandRequest{
		Id: relationID,
	})
	if err == nil {
		t.Errorf("Expected error for non-existent relation")
	}
}