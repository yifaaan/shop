package main

import (
	"context"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// UserFav 收藏：(user_id, goods_id) 唯一，防重复收藏。
type UserFav struct {
	basemodel.BaseModel
	UserID  int32 `gorm:"index:idx_user_goods,unique;not null"`
	GoodsID int32 `gorm:"index:idx_user_goods,unique;not null"`
}

// GetUserFavList 某用户的收藏列表（仅 id/userId/goodsId，商品详情由 web 层补全）。
func (s *UserOpServer) GetUserFavList(ctx context.Context, req *proto.UserFavListRequest) (*proto.UserFavListResponse, error) {
	var favs []UserFav
	result := s.db.Where("user_id = ?", req.UserId).Find(&favs)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询收藏失败: %v", result.Error)
	}
	rsp := &proto.UserFavListResponse{
		Total: int32(result.RowsAffected),
		Data:  make([]*proto.UserFavInfoResponse, 0, result.RowsAffected),
	}
	for _, f := range favs {
		rsp.Data = append(rsp.Data, &proto.UserFavInfoResponse{Id: f.ID, UserId: f.UserID, GoodsId: f.GoodsID})
	}
	return rsp, nil
}

// GetUserFav 查是否已收藏（未收藏 NotFound）。
func (s *UserOpServer) GetUserFav(ctx context.Context, req *proto.UserFavRequest) (*proto.UserFavInfoResponse, error) {
	var fav UserFav
	result := s.db.Where("user_id = ? AND goods_id = ?", req.UserId, req.GoodsId).First(&fav)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "未收藏")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询收藏失败: %v", result.Error)
	}
	return &proto.UserFavInfoResponse{Id: fav.ID, UserId: fav.UserID, GoodsId: fav.GoodsID}, nil
}

// CreateUserFav 新建收藏；(user_id, goods_id) 唯一，重复 AlreadyExists。
func (s *UserOpServer) CreateUserFav(ctx context.Context, req *proto.UserFavRequest) (*proto.UserFavInfoResponse, error) {
	fav := UserFav{UserID: req.UserId, GoodsID: req.GoodsId}
	if err := s.db.Create(&fav).Error; err != nil {
		// 唯一索引冲突视为已存在
		if isDuplicateKeyErr(err) {
			return nil, status.Errorf(codes.AlreadyExists, "已收藏")
		}
		return nil, status.Errorf(codes.Internal, "创建收藏失败: %v", err)
	}
	return &proto.UserFavInfoResponse{Id: fav.ID, UserId: fav.UserID, GoodsId: fav.GoodsID}, nil
}

// DeleteUserFav 按 user_id+goods_id 删除（防越权），不存在 NotFound。
func (s *UserOpServer) DeleteUserFav(ctx context.Context, req *proto.UserFavRequest) (*emptypb.Empty, error) {
	result := s.db.Where("user_id = ? AND goods_id = ?", req.UserId, req.GoodsId).Delete(&UserFav{})
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "删除收藏失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "未收藏")
	}
	return &emptypb.Empty{}, nil
}