package tests

import (
	"context"
	"testing"
	"time"

	"shop/inventory_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newGoodID() int32 {
	return int32(time.Now().UnixNano()%1_000_000_000) + 1000
}

func setInv(t *testing.T, goodID, stock int32) {
	t.Helper()
	_, err := inventoryClient.SetInv(context.Background(), &proto.GoodInvInfo{
		GoodId: goodID,
		Nums:   stock,
	})
	if err != nil {
		t.Fatalf("SetInv err: %v", err)
	}
}

func getInv(t *testing.T, goodID int32) *proto.GoodInvInfo {
	t.Helper()
	resp, err := inventoryClient.InvDetail(context.Background(), &proto.GoodInvInfo{
		GoodId: goodID,
	})
	if err != nil {
		t.Fatalf("InvDetail err: %v", err)
	}
	return resp
}

func TestSetInvAndDetail(t *testing.T) {
	goodID := newGoodID()
	setInv(t, goodID, 100)

	detail := getInv(t, goodID)
	if detail.GoodId != goodID {
		t.Fatalf("expected GoodId %d, got %d", goodID, detail.GoodId)
	}
	if detail.Nums != 100 {
		t.Fatalf("expected stock 100, got %d", detail.Nums)
	}

	setInv(t, goodID, 200)
	detail = getInv(t, goodID)
	if detail.Nums != 200 {
		t.Fatalf("expected stock 200, got %d", detail.Nums)
	}
}

func TestInvDetail_NotFound(t *testing.T) {
	_, err := inventoryClient.InvDetail(context.Background(), &proto.GoodInvInfo{GoodId: -1})
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

func TestSellAndReback(t *testing.T) {
	goodID := newGoodID()
	setInv(t, goodID, 10)

	_, err := inventoryClient.Sell(context.Background(), &proto.SellInfo{
		GoodInfos: []*proto.GoodInvInfo{
			{GoodId: goodID, Nums: 3},
		},
	})
	if err != nil {
		t.Fatalf("Sell err: %v", err)
	}

	detail := getInv(t, goodID)
	if detail.Nums != 7 {
		t.Fatalf("expected stock 7 after sell, got %d", detail.Nums)
	}

	_, err = inventoryClient.Reback(context.Background(), &proto.SellInfo{
		GoodInfos: []*proto.GoodInvInfo{
			{GoodId: goodID, Nums: 3},
		},
	})
	if err != nil {
		t.Fatalf("Reback err: %v", err)
	}

	detail = getInv(t, goodID)
	if detail.Nums != 10 {
		t.Fatalf("expected stock 10 after reback, got %d", detail.Nums)
	}
}

func TestSell_NotFound(t *testing.T) {
	_, err := inventoryClient.Sell(context.Background(), &proto.SellInfo{
		GoodInfos: []*proto.GoodInvInfo{
			{GoodId: -1, Nums: 1},
		},
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

func TestSell_InsufficientStock(t *testing.T) {
	goodID := newGoodID()
	setInv(t, goodID, 2)

	_, err := inventoryClient.Sell(context.Background(), &proto.SellInfo{
		GoodInfos: []*proto.GoodInvInfo{
			{GoodId: goodID, Nums: 3},
		},
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error")
	}
	if st.Code() != codes.ResourceExhausted {
		t.Fatalf("expected ResourceExhausted, got %v", st.Code())
	}
}

func TestReback_NotFound(t *testing.T) {
	_, err := inventoryClient.Reback(context.Background(), &proto.SellInfo{
		GoodInfos: []*proto.GoodInvInfo{
			{GoodId: -1, Nums: 1},
		},
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
