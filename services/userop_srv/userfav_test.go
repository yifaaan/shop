package main

import (
	"context"
	"testing"

	"shop/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserFav_CRUD(t *testing.T) {
	db := setupTestDB(t)
	srv := newTestServer(t, db)
	ctx := context.Background()

	// 创建收藏
	fav, err := srv.CreateUserFav(ctx, &proto.UserFavRequest{UserId: 1, GoodsId: 101})
	must(t, err)
	if fav.UserId != 1 || fav.GoodsId != 101 || fav.Id == 0 {
		t.Fatalf("CreateUserFav 返回异常: %+v", fav)
	}

	// 查是否已收藏
	got, err := srv.GetUserFav(ctx, &proto.UserFavRequest{UserId: 1, GoodsId: 101})
	must(t, err)
	if got.Id != fav.Id {
		t.Fatalf("GetUserFav id = %d, 期望 %d", got.Id, fav.Id)
	}

	// 重复收藏 -> AlreadyExists
	if _, err := srv.CreateUserFav(ctx, &proto.UserFavRequest{UserId: 1, GoodsId: 101}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("重复收藏期望 AlreadyExists, got %v", err)
	}

	// 列表（含两条）
	must(t, errOf(srv.CreateUserFav(ctx, &proto.UserFavRequest{UserId: 1, GoodsId: 102})))
	list, err := srv.GetUserFavList(ctx, &proto.UserFavListRequest{UserId: 1})
	must(t, err)
	if list.Total != 2 || len(list.Data) != 2 {
		t.Fatalf("列表 total=%d len=%d, 期望 2/2", list.Total, len(list.Data))
	}

	// 删除（按 user+goods）
	must(t, errOf(srv.DeleteUserFav(ctx, &proto.UserFavRequest{UserId: 1, GoodsId: 101})))
	if c := count(db, &UserFav{}); c != 1 {
		t.Fatalf("删除后应剩 1 条, got %d", c)
	}

	// 删除不存在的 -> NotFound
	if _, err := srv.DeleteUserFav(ctx, &proto.UserFavRequest{UserId: 1, GoodsId: 999}); status.Code(err) != codes.NotFound {
		t.Fatalf("删除不存在期望 NotFound, got %v", err)
	}
}

func TestUserFav_GetNotFound(t *testing.T) {
	db := setupTestDB(t)
	srv := newTestServer(t, db)
	ctx := context.Background()

	if _, err := srv.GetUserFav(ctx, &proto.UserFavRequest{UserId: 1, GoodsId: 101}); status.Code(err) != codes.NotFound {
		t.Fatalf("未收藏期望 NotFound, got %v", err)
	}
	// 列表只看本人
	must(t, errOf(srv.CreateUserFav(ctx, &proto.UserFavRequest{UserId: 1, GoodsId: 101})))
	if list, _ := srv.GetUserFavList(ctx, &proto.UserFavListRequest{UserId: 2}); list.Total != 0 || len(list.Data) != 0 {
		t.Fatalf("用户 2 列表应为空, got %+v", list)
	}
}

func TestUserFav_ReFavAfterUnfav(t *testing.T) {
	db := setupTestDB(t)
	srv := newTestServer(t, db)
	ctx := context.Background()

	// 收藏 -> 取消 -> 再收藏应成功（硬删释放唯一索引，不残留软删行）
	must(t, errOf(srv.CreateUserFav(ctx, &proto.UserFavRequest{UserId: 1, GoodsId: 101})))
	must(t, errOf(srv.DeleteUserFav(ctx, &proto.UserFavRequest{UserId: 1, GoodsId: 101})))
	if _, err := srv.CreateUserFav(ctx, &proto.UserFavRequest{UserId: 1, GoodsId: 101}); err != nil {
		t.Fatalf("取消后再收藏应成功, got %v", err)
	}
	// 活跃收藏 1 条
	if c := count(db, &UserFav{}); c != 1 {
		t.Fatalf("活跃收藏应 1 条, got %d", c)
	}
	// 含软删的总数也应为 1（验证硬删、无残留）
	var total int64
	db.Unscoped().Model(&UserFav{}).Count(&total)
	if total != 1 {
		t.Fatalf("Unscoped 总数应为 1（硬删无残留）, got %d", total)
	}
}