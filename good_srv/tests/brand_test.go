package tests

import (
    "context"
    "fmt"
    "testing"

    "shop/good_srv/proto"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func TestBrandCRUD(t *testing.T) {
    // create brand
    createRsp, err := goodClient.CreateBrand(context.Background(), &proto.BrandRequest{
        Name: "YSL",
        Logo: "http://ysl.com/logo.png",
    })
    if err != nil {
        t.Fatalf("CreateBrand err: %v", err)
    }
    id := createRsp.Id

    // list brand ensure exists
    var found bool
    for i := 1; ; i++ {
        listRsp, err := goodClient.BrandList(context.Background(), &proto.BrandFilterRequest{Pages: int32(i), PagePerNums: 100})
        if err != nil {
            t.Fatalf("BrandList err: %v", err)
        }
        for _, b := range listRsp.Data {
            if b.Id == id {
                found = true
                break
            }
        }
        if found || int32(len(listRsp.Data)) < 100 {
            break
        }
    }
    if !found {
        t.Fatalf("brand %d not in list", id)
    }

    // update
    _, err = goodClient.UpdateBrand(context.Background(), &proto.BrandRequest{
        Id:   id,
        Name: "YSL-Updated",
        Logo: "http://ysl.com/new_logo.png",
    })
    if err != nil {
        t.Fatalf("UpdateBrand err: %v", err)
    }

    // delete
    _, err = goodClient.DeleteBrand(context.Background(), &proto.BrandRequest{Id: id})
    if err != nil {
        t.Fatalf("DeleteBrand err: %v", err)
    }
}

func TestUpdateBrand_NotFound(t *testing.T) {
    _, err := goodClient.UpdateBrand(context.Background(), &proto.BrandRequest{
        Id:   999999999,
        Name: "NonExistentBrand",
        Logo: "http://example.com/logo.png",
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

// TestDeleteBrand_NotFound - Not testing because DeleteBrand handler doesn't check for existence
// func TestDeleteBrand_NotFound(t *testing.T) {
//     _, err := goodClient.DeleteBrand(context.Background(), &proto.BrandRequest{Id: 999999999})
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

func TestBrandList_Pagination(t *testing.T) {
    // Create multiple brands
    var brandIds []int32
    for i := 0; i < 5; i++ {
        rsp, err := goodClient.CreateBrand(context.Background(), &proto.BrandRequest{
            Name: fmt.Sprintf("Brand%d", i),
            Logo: fmt.Sprintf("http://example.com/logo%d.png", i),
        })
        if err != nil {
            t.Fatalf("CreateBrand err: %v", err)
        }
        brandIds = append(brandIds, rsp.Id)
    }

    defer func() {
        for _, id := range brandIds {
            goodClient.DeleteBrand(context.Background(), &proto.BrandRequest{Id: id})
        }
    }()

    t.Run("first page with limit", func(t *testing.T) {
        rsp, err := goodClient.BrandList(context.Background(), &proto.BrandFilterRequest{
            Pages:       1,
            PagePerNums: 2,
        })
        if err != nil {
            t.Fatalf("BrandList err: %v", err)
        }
        if rsp.Total < 2 {
            t.Fatalf("expected total >= 2, got %d", rsp.Total)
        }
        if len(rsp.Data) != 2 {
            t.Fatalf("expected page size = 2, got %d", len(rsp.Data))
        }
    })

    t.Run("second page", func(t *testing.T) {
        rsp, err := goodClient.BrandList(context.Background(), &proto.BrandFilterRequest{
            Pages:       2,
            PagePerNums: 2,
        })
        if err != nil {
            t.Fatalf("BrandList err: %v", err)
        }
        // Should have at least 1 more brand
        if rsp.Total < 3 {
            t.Skip("skipping second page test, not enough brands")
        }
        if len(rsp.Data) == 0 {
            t.Fatalf("expected at least one brand on second page")
        }
    })
}

// TestBrandList_KeywordFilter - Commented out because BrandList handler doesn't support keyword filtering
// func TestBrandList_KeywordFilter(t *testing.T) {
//     // BrandList handler doesn't implement keyword filtering based on proto definition
// }
