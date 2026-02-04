package tests

import (
	"context"
	"testing"
	"time"

	"shop/order_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newUserID() int32 {
	return int32(time.Now().UnixNano()%1_000_000_000) + 1000
}

func newGoodsID() int32 {
	return int32(time.Now().UnixNano()%1_000_000_000) + 2000
}

func assertStatusCode(t *testing.T, err error, code codes.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error")
	}
	if st.Code() != code {
		t.Fatalf("expected %v, got %v", code, st.Code())
	}
}

func TestCartItemCRUD(t *testing.T) {
	userID := newUserID()
	goodsID := newGoodsID()

	createResp, err := orderClient.CreateCartItem(context.Background(), &proto.CartItemRequest{
		UserId:  userID,
		GoodsId: goodsID,
		Nums:    2,
		Checked: true,
	})
	if err != nil {
		t.Fatalf("CreateCartItem err: %v", err)
	}
	if createResp.Id == 0 {
		t.Fatalf("CreateCartItem returned id=0")
	}

	listResp, err := orderClient.CartItemList(context.Background(), &proto.UserInfo{Id: userID})
	if err != nil {
		t.Fatalf("CartItemList err: %v", err)
	}
	if listResp.Total == 0 || len(listResp.Data) == 0 {
		t.Fatalf("expected cart list data")
	}

	found := false
	for _, item := range listResp.Data {
		if item.Id == createResp.Id {
			found = true
			if item.Nums != 2 {
				t.Fatalf("expected nums=2, got %d", item.Nums)
			}
			if !item.Checked {
				t.Fatalf("expected checked=true")
			}
		}
	}
	if !found {
		t.Fatalf("created cart item not found in list")
	}

	_, err = orderClient.UpdateCartItem(context.Background(), &proto.CartItemRequest{
		Id:      createResp.Id,
		Nums:    5,
		Checked: true,
	})
	if err != nil {
		t.Fatalf("UpdateCartItem err: %v", err)
	}

	listResp, err = orderClient.CartItemList(context.Background(), &proto.UserInfo{Id: userID})
	if err != nil {
		t.Fatalf("CartItemList after update err: %v", err)
	}
	found = false
	for _, item := range listResp.Data {
		if item.Id == createResp.Id {
			found = true
			if item.Nums != 5 {
				t.Fatalf("expected nums=5, got %d", item.Nums)
			}
		}
	}
	if !found {
		t.Fatalf("updated cart item not found in list")
	}

	_, err = orderClient.DeleteCartItem(context.Background(), &proto.CartItemRequest{Id: createResp.Id})
	if err != nil {
		t.Fatalf("DeleteCartItem err: %v", err)
	}

	_, err = orderClient.CartItemList(context.Background(), &proto.UserInfo{Id: userID})
	assertStatusCode(t, err, codes.NotFound)
}

func TestCartItemList_NotFound(t *testing.T) {
	userID := newUserID()
	_, err := orderClient.CartItemList(context.Background(), &proto.UserInfo{Id: userID})
	assertStatusCode(t, err, codes.NotFound)
}

func TestUpdateCartItem_NotFound(t *testing.T) {
	_, err := orderClient.UpdateCartItem(context.Background(), &proto.CartItemRequest{
		Id:   -1,
		Nums: 1,
	})
	assertStatusCode(t, err, codes.NotFound)
}

func TestDeleteCartItem_NotFound(t *testing.T) {
	_, err := orderClient.DeleteCartItem(context.Background(), &proto.CartItemRequest{Id: -1})
	assertStatusCode(t, err, codes.NotFound)
}

func TestCreateOrder_NoChecked(t *testing.T) {
	userID := newUserID()

	_, err := orderClient.CreateOrder(context.Background(), &proto.OrderRequest{
		UserId:  userID,
		Address: "addr",
		Name:    "name",
		Mobile:  "18800000000",
		Post:    "note",
	})
	assertStatusCode(t, err, codes.NotFound)
}

func TestOrderList_NotFound(t *testing.T) {
	userID := newUserID()
	_, err := orderClient.OrderList(context.Background(), &proto.OrderFilterRequest{
		UserId:      userID,
		Pages:       1,
		PagePerNums: 10,
	})
	assertStatusCode(t, err, codes.NotFound)
}

func TestOrderDetail_NotFound(t *testing.T) {
	_, err := orderClient.OrderDetail(context.Background(), &proto.OrderRequest{Id: -1})
	assertStatusCode(t, err, codes.NotFound)
}

func TestUpdateOrderStatus_NotFound(t *testing.T) {
	_, err := orderClient.UpdateOrderStatus(context.Background(), &proto.OrderStatus{
		OrderSn: "not-exists",
		Status:  "paid",
	})
	assertStatusCode(t, err, codes.Internal)
}
