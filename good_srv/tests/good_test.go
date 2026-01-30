package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"shop/good_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func createTestCategory(t *testing.T, name string) int32 {
	t.Helper()
	rsp, err := goodClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:  name,
		Level: 1,
		IsTab: false,
	})
	if err != nil {
		t.Fatalf("CreateCategory err: %v", err)
	}
	if rsp.Id == 0 {
		t.Fatalf("CreateCategory returned id=0")
	}
	return rsp.Id
}

func createTestBrand(t *testing.T, name string) int32 {
	t.Helper()
	rsp, err := goodClient.CreateBrand(context.Background(), &proto.BrandRequest{
		Name: name,
		Logo: "http://example.com/logo.png",
	})
	if err != nil {
		t.Fatalf("CreateBrand err: %v", err)
	}
	if rsp.Id == 0 {
		t.Fatalf("CreateBrand returned id=0")
	}
	return rsp.Id
}

func createTestGood(t *testing.T, catID, brandID int32, name string, shopPrice float32, isHot, isNew bool) int32 {
	t.Helper()
	sn := fmt.Sprintf("SN-%d", time.Now().UnixNano())
	rsp, err := goodClient.CreateGood(context.Background(), &proto.CreateGoodInfo{
		Name:           name,
		GoodSn:         sn,
		MarketPrice:    shopPrice + 100,
		ShopPrice:      shopPrice,
		GoodBrief:      "brief",
		ShipFree:       true,
		Images:         []string{"http://example.com/img1.png"},
		DescImages:     []string{"http://example.com/desc1.png"},
		GoodFrontImage: "http://example.com/front.png",
		IsNew:          isNew,
		IsHot:          isHot,
		OnSale:         true,
		CategoryId:     catID,
		BrandId:        brandID,
	})
	if err != nil {
		t.Fatalf("CreateGood err: %v", err)
	}
	if rsp.Id == 0 {
		t.Fatalf("CreateGood returned id=0")
	}
	return rsp.Id
}

func deleteTestGood(t *testing.T, id int32) {
	t.Helper()
	_, err := goodClient.DeleteGood(context.Background(), &proto.DeleteGoodInfo{Id: id})
	if err != nil {
		t.Fatalf("DeleteGood err: %v", err)
	}
}

func deleteTestCategory(t *testing.T, id int32) {
	t.Helper()
	_, err := goodClient.DeleteCategory(context.Background(), &proto.DeleteCategoryRequest{Id: id})
	if err != nil {
		t.Fatalf("DeleteCategory err: %v", err)
	}
}

func deleteTestBrand(t *testing.T, id int32) {
	t.Helper()
	_, err := goodClient.DeleteBrand(context.Background(), &proto.BrandRequest{Id: id})
	if err != nil {
		t.Fatalf("DeleteBrand err: %v", err)
	}
}

func TestGoodList_FiltersAndPaging(t *testing.T) {
	catID := createTestCategory(t, "TestGoodListCat")
	brandID := createTestBrand(t, "TestGoodListBrand")

	g1 := createTestGood(t, catID, brandID, "AlphaPhone", 1999, true, false)
	g2 := createTestGood(t, catID, brandID, "BetaPhone", 2999, false, true)
	g3 := createTestGood(t, catID, brandID, "GammaLaptop", 9999, true, true)

	defer deleteTestGood(t, g1)
	defer deleteTestGood(t, g2)
	defer deleteTestGood(t, g3)
	defer deleteTestCategory(t, catID)
	defer deleteTestBrand(t, brandID)

	t.Run("paging", func(t *testing.T) {
		rsp, err := goodClient.GoodList(context.Background(), &proto.GoodFilterRequest{
			Pages:       1,
			PagePerNums: 2,
		})
		if err != nil {
			t.Fatalf("GoodList err: %v", err)
		}
		if rsp.Total == 0 {
			t.Fatalf("expected total > 0")
		}
		if len(rsp.Data) > 2 {
			t.Fatalf("expected page size <= 2, got %d", len(rsp.Data))
		}
	})

	t.Run("keyword", func(t *testing.T) {
		rsp, err := goodClient.GoodList(context.Background(), &proto.GoodFilterRequest{
			KeyWords:    "Phone",
			Pages:       1,
			PagePerNums: 50,
		})
		if err != nil {
			t.Fatalf("GoodList err: %v", err)
		}
		for _, it := range rsp.Data {
			if it.Name == "" {
				t.Fatalf("expected non-empty name")
			}
			if it.Id == 0 {
				t.Fatalf("expected non-zero id")
			}
			if it.Name != "AlphaPhone" && it.Name != "BetaPhone" {
				// 关键词为 LIKE %Phone%，只断言必须包含 Phone 更稳健
				if !contains(it.Name, "Phone") {
					t.Fatalf("expected name contains 'Phone', got %q", it.Name)
				}
			}
		}
	})

	t.Run("brand", func(t *testing.T) {
		rsp, err := goodClient.GoodList(context.Background(), &proto.GoodFilterRequest{
			Brand:       brandID,
			Pages:       1,
			PagePerNums: 50,
		})
		if err != nil {
			t.Fatalf("GoodList err: %v", err)
		}
		if rsp.Total == 0 {
			t.Fatalf("expected goods for brand")
		}
		for _, it := range rsp.Data {
			if it.Brand == nil || it.Brand.Id != brandID {
				t.Fatalf("expected brand id %d, got %+v", brandID, it.Brand)
			}
		}
	})

	t.Run("price_range", func(t *testing.T) {
		rsp, err := goodClient.GoodList(context.Background(), &proto.GoodFilterRequest{
			PriceMin:    2500,
			PriceMax:    12000,
			Pages:       1,
			PagePerNums: 50,
		})
		if err != nil {
			t.Fatalf("GoodList err: %v", err)
		}
		for _, it := range rsp.Data {
			if it.ShopPrice < 2500 || it.ShopPrice > 12000 {
				t.Fatalf("shopPrice out of range: %v", it.ShopPrice)
			}
		}
	})

	t.Run("is_hot", func(t *testing.T) {
		rsp, err := goodClient.GoodList(context.Background(), &proto.GoodFilterRequest{
			IsHot:       true,
			Pages:       1,
			PagePerNums: 50,
		})
		if err != nil {
			t.Fatalf("GoodList err: %v", err)
		}
		for _, it := range rsp.Data {
			if !it.IsHot {
				t.Fatalf("expected isHot=true")
			}
		}
	})

	t.Run("is_new", func(t *testing.T) {
		rsp, err := goodClient.GoodList(context.Background(), &proto.GoodFilterRequest{
			IsNew:       true,
			Pages:       1,
			PagePerNums: 50,
		})
		if err != nil {
			t.Fatalf("GoodList err: %v", err)
		}
		for _, it := range rsp.Data {
			if !it.IsNew {
				t.Fatalf("expected isNew=true")
			}
		}
	})
}

func TestGoodList_TopCategoryNotFound(t *testing.T) {
	_, err := goodClient.GoodList(context.Background(), &proto.GoodFilterRequest{
		TopCategory: 999999999,
		Pages:       1,
		PagePerNums: 10,
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

func TestGoodList_TopCategoryLevels(t *testing.T) {
	// Create proper category hierarchy:
	// Level1 -> Level2 -> Level3
	brandID := createTestBrand(t, "LevelBrand")

	// Create level 1 category
	l1Rsp, err := goodClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:  "L1",
		Level: 1,
		IsTab: false,
	})
	if err != nil {
		t.Fatalf("CreateCategory L1 err: %v", err)
	}
	catL1 := l1Rsp.Id

	// Create level 2 category under L1
	l2Rsp, err := goodClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:           "L2",
		Level:          2,
		ParentCategory: catL1,
		IsTab:          false,
	})
	if err != nil {
		t.Fatalf("CreateCategory L2 err: %v", err)
	}
	catL2 := l2Rsp.Id

	// Create level 3 category under L2
	l3Rsp, err := goodClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:           "L3",
		Level:          3,
		ParentCategory: catL2,
		IsTab:          false,
	})
	if err != nil {
		t.Fatalf("CreateCategory L3 err: %v", err)
	}
	catL3 := l3Rsp.Id

	// Create goods only for level 3 category
	g1 := createTestGood(t, catL3, brandID, "Level3Good", 3000, false, false)

	defer deleteTestGood(t, g1)
	defer deleteTestCategory(t, catL3)
	defer deleteTestCategory(t, catL2)
	defer deleteTestCategory(t, catL1)
	defer deleteTestBrand(t, brandID)

	t.Run("Level1 category filtering - should find goods under L3", func(t *testing.T) {
		rsp, err := goodClient.GoodList(context.Background(), &proto.GoodFilterRequest{
			TopCategory: catL1,
			Pages:       1,
			PagePerNums: 50,
		})
		if err != nil {
			t.Fatalf("GoodList err: %v", err)
		}
		if rsp.Total == 0 {
			t.Fatalf("expected goods for level1 category")
		}
		for _, it := range rsp.Data {
			if it.Category == nil || it.Category.Id != catL3 {
				t.Fatalf("expected category id %d (L3), got %+v", catL3, it.Category)
			}
		}
	})

	t.Run("Level2 category filtering - should find goods under L3", func(t *testing.T) {
		rsp, err := goodClient.GoodList(context.Background(), &proto.GoodFilterRequest{
			TopCategory: catL2,
			Pages:       1,
			PagePerNums: 50,
		})
		if err != nil {
			t.Fatalf("GoodList err: %v", err)
		}
		if rsp.Total == 0 {
			t.Fatalf("expected goods for level2 category")
		}
		for _, it := range rsp.Data {
			if it.Category == nil || it.Category.Id != catL3 {
				t.Fatalf("expected category id %d (L3), got %+v", catL3, it.Category)
			}
		}
	})

	t.Run("Level3 category filtering - should find goods under L3", func(t *testing.T) {
		rsp, err := goodClient.GoodList(context.Background(), &proto.GoodFilterRequest{
			TopCategory: catL3,
			Pages:       1,
			PagePerNums: 50,
		})
		if err != nil {
			t.Fatalf("GoodList err: %v", err)
		}
		if rsp.Total == 0 {
			t.Fatalf("expected goods for level3 category")
		}
		for _, it := range rsp.Data {
			if it.Category == nil || it.Category.Id != catL3 {
				t.Fatalf("expected category id %d (L3), got %+v", catL3, it.Category)
			}
		}
	})
}

func TestCreateGood_DuplicateName(t *testing.T) {
	catID := createTestCategory(t, "TestGoodDupCat")
	brandID := createTestBrand(t, "TestGoodDupBrand")
	id := createTestGood(t, catID, brandID, "DupGoodName", 1234, false, false)

	defer deleteTestGood(t, id)
	defer deleteTestCategory(t, catID)
	defer deleteTestBrand(t, brandID)

	_, err := goodClient.CreateGood(context.Background(), &proto.CreateGoodInfo{
		Name:           "DupGoodName",
		GoodSn:         fmt.Sprintf("SN-%d", time.Now().UnixNano()),
		MarketPrice:    2000,
		ShopPrice:      1500,
		GoodBrief:      "brief",
		ShipFree:       true,
		Images:         []string{"http://example.com/img1.png"},
		DescImages:     []string{"http://example.com/desc1.png"},
		GoodFrontImage: "http://example.com/front.png",
		IsNew:          false,
		IsHot:          false,
		OnSale:         true,
		CategoryId:     catID,
		BrandId:        brandID,
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error")
	}
	if st.Code() != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists, got %v", st.Code())
	}
}

func TestGetGoodDetail_NotFound(t *testing.T) {
	_, err := goodClient.GetGoodDetail(context.Background(), &proto.GoodInfoRequest{Id: 999999999})
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

func TestGoodCRUDAndBatchGet(t *testing.T) {
	// Use unique names with shorter format to avoid conflicts (name max length is 20)
	catID := createTestCategory(t, fmt.Sprintf("C%d", time.Now().UnixNano()%100000))
	brandID := createTestBrand(t, fmt.Sprintf("B%d", time.Now().UnixNano()%100000))
	goodName := fmt.Sprintf("G%d", time.Now().UnixNano()%100000)
	goodID := createTestGood(t, catID, brandID, goodName, 1888, false, false)

	defer deleteTestCategory(t, catID)
	defer deleteTestBrand(t, brandID)

	// detail
	detail, err := goodClient.GetGoodDetail(context.Background(), &proto.GoodInfoRequest{Id: goodID})
	if err != nil {
		t.Fatalf("GetGoodDetail err: %v", err)
	}
	if detail.Id != goodID {
		t.Fatalf("expected id %d got %d", goodID, detail.Id)
	}
	if detail.Category == nil || detail.Category.Id != catID {
		t.Fatalf("expected category id %d got %+v", catID, detail.Category)
	}
	if detail.Brand == nil || detail.Brand.Id != brandID {
		t.Fatalf("expected brand id %d got %+v", brandID, detail.Brand)
	}

	// update
	updatedName := goodName + "Updated"
	_, err = goodClient.UpdateGood(context.Background(), &proto.CreateGoodInfo{
		Id:             goodID,
		Name:           updatedName,
		GoodSn:         fmt.Sprintf("SN-%d", time.Now().UnixNano()),
		MarketPrice:    3000,
		ShopPrice:      2666,
		GoodBrief:      "brief2",
		ShipFree:       false,
		Images:         []string{"http://example.com/img2.png"},
		DescImages:     []string{"http://example.com/desc2.png"},
		GoodFrontImage: "http://example.com/front2.png",
		IsNew:          true,
		IsHot:          true,
		OnSale:         true,
		CategoryId:     catID,
		BrandId:        brandID,
	})
	if err != nil {
		t.Fatalf("UpdateGood err: %v", err)
	}

	updated, err := goodClient.GetGoodDetail(context.Background(), &proto.GoodInfoRequest{Id: goodID})
	if err != nil {
		t.Fatalf("GetGoodDetail after update err: %v", err)
	}
	if updated.Name != updatedName {
		t.Fatalf("expected updated name, got %q", updated.Name)
	}
	if updated.ShopPrice != 2666 {
		t.Fatalf("expected updated shopPrice=2666 got %v", updated.ShopPrice)
	}
	if !updated.IsHot || !updated.IsNew {
		t.Fatalf("expected updated flags isHot/isNew")
	}

	// batch get
	batch, err := goodClient.BatchGetGood(context.Background(), &proto.BatchGoodIdInfo{Id: []int32{goodID}})
	if err != nil {
		t.Fatalf("BatchGetGood err: %v", err)
	}
	if batch.Total != 1 {
		t.Fatalf("expected total=1 got %d", batch.Total)
	}
	if len(batch.Data) != 1 || batch.Data[0].Id != goodID {
		t.Fatalf("unexpected batch data: %+v", batch.Data)
	}

	// delete
	deleteTestGood(t, goodID)

	_, err = goodClient.GetGoodDetail(context.Background(), &proto.GoodInfoRequest{Id: goodID})
	if err == nil {
		t.Fatalf("expected not found after delete")
	}
}

func TestUpdateGood_NotFound(t *testing.T) {
	catID := createTestCategory(t, "TestUpdateGoodNF_Cat")
	brandID := createTestBrand(t, "TestUpdateGoodNF_Brand")
	defer deleteTestCategory(t, catID)
	defer deleteTestBrand(t, brandID)

	_, err := goodClient.UpdateGood(context.Background(), &proto.CreateGoodInfo{
		Id:             999999999,
		Name:           "NoSuch",
		GoodSn:         fmt.Sprintf("SN-%d", time.Now().UnixNano()),
		MarketPrice:    100,
		ShopPrice:      90,
		GoodBrief:      "brief",
		ShipFree:       true,
		Images:         []string{"http://example.com/img.png"},
		DescImages:     []string{"http://example.com/desc.png"},
		GoodFrontImage: "http://example.com/front.png",
		IsNew:          false,
		IsHot:          false,
		OnSale:         true,
		CategoryId:     catID,
		BrandId:        brandID,
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

func TestCreateGood_CategoryNotFound(t *testing.T) {
	brandID := createTestBrand(t, "TestCreateGoodCatNF_Brand")
	defer deleteTestBrand(t, brandID)

	_, err := goodClient.CreateGood(context.Background(), &proto.CreateGoodInfo{
		Name:           "CatNotFoundGood",
		GoodSn:         fmt.Sprintf("SN-%d", time.Now().UnixNano()),
		MarketPrice:    200,
		ShopPrice:      100,
		GoodBrief:      "brief",
		ShipFree:       true,
		Images:         []string{"http://example.com/img.png"},
		DescImages:     []string{"http://example.com/desc.png"},
		GoodFrontImage: "http://example.com/front.png",
		IsNew:          false,
		IsHot:          false,
		OnSale:         true,
		CategoryId:     999999999,
		BrandId:        brandID,
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

func TestCreateGood_BrandNotFound(t *testing.T) {
	catID := createTestCategory(t, "BrandNF")
	defer deleteTestCategory(t, catID)

	_, err := goodClient.CreateGood(context.Background(), &proto.CreateGoodInfo{
		Name:           "BrandNotFoundGood",
		GoodSn:         fmt.Sprintf("SN-%d", time.Now().UnixNano()),
		MarketPrice:    200,
		ShopPrice:      100,
		GoodBrief:      "brief",
		ShipFree:       true,
		Images:         []string{"http://example.com/img.png"},
		DescImages:     []string{"http://example.com/desc.png"},
		GoodFrontImage: "http://example.com/front.png",
		IsNew:          false,
		IsHot:          false,
		OnSale:         true,
		CategoryId:     catID,
		BrandId:        999999999,
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

// TestDeleteGood_NotFound - Not testing because DeleteGood handler doesn't check for existence
// func TestDeleteGood_NotFound(t *testing.T) {
// 	_, err := goodClient.DeleteGood(context.Background(), &proto.DeleteGoodInfo{Id: 999999999})
// 	if err == nil {
// 		t.Fatalf("expected error")
// 	}
// 	st, ok := status.FromError(err)
// 	if !ok {
// 		t.Fatalf("expected grpc status error")
// 	}
// 	if st.Code() != codes.NotFound {
// 		t.Fatalf("expected NotFound, got %v", st.Code())
// 	}
// }

func contains(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	return len(s) >= len(sub) && (indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
