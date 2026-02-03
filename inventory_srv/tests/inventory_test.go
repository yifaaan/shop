package tests

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"shop/inventory_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	testGoodID421 int32 = 421
	testGoodID422 int32 = 422
	testGoodID423 int32 = 423
)

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

// func TestSetInvAndDetail(t *testing.T) {
// 	goodID := newGoodID()
// 	setInv(t, goodID, 100)

// 	detail := getInv(t, goodID)
// 	if detail.GoodId != goodID {
// 		t.Fatalf("expected GoodId %d, got %d", goodID, detail.GoodId)
// 	}
// 	if detail.Nums != 100 {
// 		t.Fatalf("expected stock 100, got %d", detail.Nums)
// 	}

// 	setInv(t, goodID, 200)
// 	detail = getInv(t, goodID)
// 	if detail.Nums != 200 {
// 		t.Fatalf("expected stock 200, got %d", detail.Nums)
// 	}
// }

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
	goodID := testGoodID421
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
	goodID := testGoodID422
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

func TestSell_MultiGoodsSuccess(t *testing.T) {
	goodID1 := testGoodID421
	goodID2 := testGoodID422
	setInv(t, goodID1, 5)
	setInv(t, goodID2, 6)

	_, err := inventoryClient.Sell(context.Background(), &proto.SellInfo{
		GoodInfos: []*proto.GoodInvInfo{
			{GoodId: goodID1, Nums: 2},
			{GoodId: goodID2, Nums: 3},
		},
	})
	if err != nil {
		t.Fatalf("Sell err: %v", err)
	}

	detail1 := getInv(t, goodID1)
	if detail1.Nums != 3 {
		t.Fatalf("expected stock 3 for good %d, got %d", goodID1, detail1.Nums)
	}
	detail2 := getInv(t, goodID2)
	if detail2.Nums != 3 {
		t.Fatalf("expected stock 3 for good %d, got %d", goodID2, detail2.Nums)
	}
}

func TestSell_MultiGoodsRollbackOnInsufficient(t *testing.T) {
	goodID1 := testGoodID421
	goodID2 := testGoodID423
	setInv(t, goodID1, 5)
	setInv(t, goodID2, 1)

	_, err := inventoryClient.Sell(context.Background(), &proto.SellInfo{
		GoodInfos: []*proto.GoodInvInfo{
			{GoodId: goodID1, Nums: 2},
			{GoodId: goodID2, Nums: 2},
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

	detail1 := getInv(t, goodID1)
	if detail1.Nums != 5 {
		t.Fatalf("expected stock 5 for good %d after rollback, got %d", goodID1, detail1.Nums)
	}
	detail2 := getInv(t, goodID2)
	if detail2.Nums != 1 {
		t.Fatalf("expected stock 1 for good %d after rollback, got %d", goodID2, detail2.Nums)
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

func TestSell_ConcurrentGoroutines(t *testing.T) {
	goodID := testGoodID421
	initialStock := int32(5)
	totalCalls := 10

	setInv(t, goodID, initialStock)

	var success int32
	errCh := make(chan error, totalCalls)
	var wg sync.WaitGroup

	for i := 0; i < totalCalls; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := inventoryClient.Sell(context.Background(), &proto.SellInfo{
				GoodInfos: []*proto.GoodInvInfo{
					{GoodId: goodID, Nums: 1},
				},
			})
			if err != nil {
				errCh <- err
				return
			}
			atomic.AddInt32(&success, 1)
		}()
	}

	wg.Wait()
	close(errCh)

	var exhausted int
	var unexpected []error
	for err := range errCh {
		st, ok := status.FromError(err)
		if !ok {
			unexpected = append(unexpected, err)
			continue
		}
		if st.Code() == codes.ResourceExhausted {
			exhausted++
			continue
		}
		unexpected = append(unexpected, err)
	}

	expectedSuccess := int32(initialStock)
	if success != expectedSuccess {
		t.Fatalf("expected success %d, got %d", expectedSuccess, success)
	}
	if exhausted != totalCalls-int(expectedSuccess) {
		t.Fatalf("expected exhausted %d, got %d", totalCalls-int(expectedSuccess), exhausted)
	}
	if len(unexpected) > 0 {
		t.Fatalf("unexpected errors: %v", unexpected)
	}

	detail := getInv(t, goodID)
	if detail.Nums != 0 {
		t.Fatalf("expected stock 0 after concurrent sells, got %d", detail.Nums)
	}
}
