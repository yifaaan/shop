package main

import (
	"context"
	"testing"

	"shop/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func sampleAddrReq(userId, id int32) *proto.AddressRequest {
	return &proto.AddressRequest{
		Id: id, UserId: userId,
		Province: "广东省", City: "深圳市", District: "南山区",
		Address: "科技园1号", SignerName: "张三", SignerMobile: "13800000000",
	}
}

func TestAddress_CRUD(t *testing.T) {
	db := setupTestDB(t)
	srv := newTestServer(t, db)
	ctx := context.Background()

	// 新建（id=0）
	created, err := srv.CreateAddress(ctx, sampleAddrReq(1, 0))
	must(t, err)
	if created.Id == 0 || created.UserId != 1 || created.Province != "广东省" {
		t.Fatalf("CreateAddress 异常: %+v", created)
	}

	// 列表
	list, err := srv.GetAddressList(ctx, &proto.AddressListRequest{UserId: 1})
	must(t, err)
	if list.Total != 1 || len(list.Data) != 1 {
		t.Fatalf("列表应 1 条, got total=%d len=%d", list.Total, len(list.Data))
	}

	// 更新（本人，按 id+userId）
	upd := sampleAddrReq(1, created.Id)
	upd.Address = "科技园2号"
	must(t, errOf(srv.UpdateAddress(ctx, upd)))
	got, err := srv.GetAddressList(ctx, &proto.AddressListRequest{UserId: 1})
	must(t, err)
	if got.Data[0].Address != "科技园2号" {
		t.Fatalf("更新未生效, got %s", got.Data[0].Address)
	}

	// 越权更新他人地址 -> NotFound
	other := sampleAddrReq(2, created.Id)
	other.Address = "越权"
	if _, err := srv.UpdateAddress(ctx, other); status.Code(err) != codes.NotFound {
		t.Fatalf("越权更新期望 NotFound, got %v", err)
	}
	// 原地址未被改
	got, _ = srv.GetAddressList(ctx, &proto.AddressListRequest{UserId: 1})
	if got.Data[0].Address != "科技园2号" {
		t.Fatalf("越权更新不应改动原地址, got %s", got.Data[0].Address)
	}

	// 删除（本人）
	must(t, errOf(srv.DeleteAddress(ctx, &proto.DeleteAddressRequest{UserId: 1, Id: created.Id})))
	if c := count(db, &Address{}); c != 0 {
		t.Fatalf("删除后应 0 条, got %d", c)
	}
	// 越权删除 -> NotFound
	must(t, errOf(srv.CreateAddress(ctx, sampleAddrReq(1, 0))))
	list, _ = srv.GetAddressList(ctx, &proto.AddressListRequest{UserId: 1})
	otherId := list.Data[0].Id
	if _, err := srv.DeleteAddress(ctx, &proto.DeleteAddressRequest{UserId: 2, Id: otherId}); status.Code(err) != codes.NotFound {
		t.Fatalf("越权删除期望 NotFound, got %v", err)
	}
	if c := count(db, &Address{}); c != 1 {
		t.Fatalf("越权删除不应影响数据, got %d", c)
	}
}
