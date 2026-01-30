package tests

import (
    "context"
    "testing"

    "shop/good_srv/proto"
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
