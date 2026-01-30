package tests

import (
	"context"
	"shop/good_srv/proto"
	"testing"
)

func TestGoodList(t *testing.T) {
	resp, err := goodClient.GoodList(context.Background(), &proto.GoodFilterRequest{
		PriceMin:    1,
		PriceMax:    10000,
		Pages:       1,
		PagePerNums: 10,
		TopCategory: 130358,
	})
	if err != nil {
		t.Fatalf("GoodList err: %v", err)
	}
	t.Log(resp)
}
