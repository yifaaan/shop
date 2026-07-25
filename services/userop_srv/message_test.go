package main

import (
	"context"
	"testing"

	"shop/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMessage_CreateListDelete(t *testing.T) {
	db := setupTestDB(t)
	srv := newTestServer(t, db)
	ctx := context.Background()

	// 合法 type=1
	m1, err := srv.CreateMessage(ctx, &proto.MessageRequest{
		UserId: 1, Subject: "问询", Message: "在吗", Type: 3, File: "a.png",
	})
	must(t, err)
	if m1.Id == 0 || m1.Type != 3 || m1.File != "a.png" || m1.AddTime == 0 {
		t.Fatalf("CreateMessage 异常: %+v", m1)
	}

	// 列表
	must(t, errOf(srv.CreateMessage(ctx, &proto.MessageRequest{UserId: 1, Subject: "s2", Message: "m2", Type: 1})))
	list, err := srv.GetMessageList(ctx, &proto.MessageListRequest{UserId: 1})
	must(t, err)
	if list.Total != 2 || len(list.Data) != 2 {
		t.Fatalf("列表应 2 条, got total=%d len=%d", list.Total, len(list.Data))
	}

	// 删除（本人）
	must(t, errOf(srv.DeleteMessage(ctx, &proto.DeleteMessageRequest{UserId: 1, Id: m1.Id})))
	if c := count(db, &Message{}); c != 1 {
		t.Fatalf("删除后应剩 1 条, got %d", c)
	}
	// 越权删除 -> NotFound
	if _, err := srv.DeleteMessage(ctx, &proto.DeleteMessageRequest{UserId: 2, Id: m1.Id}); status.Code(err) != codes.NotFound {
		t.Fatalf("越权删除期望 NotFound, got %v", err)
	}
}

func TestMessage_InvalidType(t *testing.T) {
	db := setupTestDB(t)
	srv := newTestServer(t, db)
	ctx := context.Background()

	for _, bad := range []int32{0, 6, -1, 99} {
		if _, err := srv.CreateMessage(ctx, &proto.MessageRequest{
			UserId: 1, Subject: "s", Message: "m", Type: bad,
		}); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("非法 type=%d 期望 InvalidArgument, got %v", bad, err)
		}
	}
	if c := count(db, &Message{}); c != 0 {
		t.Fatalf("非法 type 不应落库, got %d", c)
	}
}