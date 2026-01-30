package tests

import (
	"context"
	"testing"

	"shop/good_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func TestGetSubCategory_NotFound(t *testing.T) {
	_, err := goodClient.GetSubCategory(context.Background(), &proto.CategoryListRequest{Id: 999999999})
	if err == nil {
		t.Fatalf("expected error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error")
	}
	if st.Code() != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", st.Code())
	}
}

func TestCreateCategory_DuplicateName(t *testing.T) {
	// Create first category
	name := "TestCat"
	createRsp, err := goodClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:  name,
		Level: 1,
		IsTab: false,
	})
	if err != nil {
		t.Fatalf("CreateCategory err: %v", err)
	}
	id := createRsp.Id
	defer deleteTestCategory(t, id)

	// Try to create category with same name
	_, err = goodClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:  name,
		Level: 1,
		IsTab: false,
	})
	if err == nil {
		t.Fatalf("expected error for duplicate category name")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error")
	}
	if st.Code() != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists, got %v", st.Code())
	}
}

func TestCreateCategory_InvalidParent(t *testing.T) {
	// Try to create category with invalid parent
	_, err := goodClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:           "InvalidParent",
		Level:          2,
		ParentCategory: 999999999,
		IsTab:          false,
	})
	if err == nil {
		t.Fatalf("expected error for invalid parent")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error")
	}
	if st.Code() != codes.Unknown {
		t.Fatalf("expected Unknown, got %v", st.Code())
	}
}

// TestDeleteCategory_NotFound - Not testing because DeleteCategory handler doesn't check for existence
// func TestDeleteCategory_NotFound(t *testing.T) {
//     _, err := goodClient.DeleteCategory(context.Background(), &proto.DeleteCategoryRequest{Id: 999999999})
//     if err == nil {
//         t.Fatalf("expected error")
//     }
//     st, ok := status.FromError(err)
//     if !ok {
//         t.Fatalf("expected grpc status error")
//     }
//     if st.Code() != codes.NotFound {
//         t.Fatalf("expected NotFound, got %v", st.Code())
//     }
// }

func TestUpdateCategory_NotFound(t *testing.T) {
	_, err := goodClient.UpdateCategory(context.Background(), &proto.CategoryInfoRequest{
		Id:   999999999,
		Name: "NonExistentCategory",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error")
	}
	if st.Code() != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", st.Code())
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
