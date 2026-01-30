package tests

import (
	"context"
	"testing"

	"shop/good_srv/proto"
)

func TestBannerCRUD(t *testing.T) {
	// create
	createRsp, err := goodClient.CreateBanner(context.Background(), &proto.BannerRequest{
		Image: "http://example.com/banner1.png",
		Url:   "http://example.com/link1",
		Index: 1,
	})
	if err != nil {
		t.Fatalf("CreateBanner err: %v", err)
	}

	id := createRsp.Id

	// list must include created banner
	listRsp, err := goodClient.BannerList(context.Background(), &proto.Empty{})
	if err != nil {
		t.Fatalf("BannerList err: %v", err)
	}
	found := false
	for _, b := range listRsp.Data {
		if b.Id == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created banner %d not found in list", id)
	}

	// update
	_, err = goodClient.UpdateBanner(context.Background(), &proto.BannerRequest{
		Id:    id,
		Image: "http://example.com/banner1_updated.png",
		Url:   "http://example.com/link1_updated",
		Index: 2,
	})
	if err != nil {
		t.Fatalf("UpdateBanner err: %v", err)
	}

	// delete
	_, err = goodClient.DeleteBanner(context.Background(), &proto.BannerRequest{Id: id})
	if err != nil {
		t.Fatalf("DeleteBanner err: %v", err)
	}
}
