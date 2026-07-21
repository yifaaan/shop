package main

import (
	"context"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Banner struct {
	basemodel.BaseModel
	Image string `gorm:"type:varchar(200);not null"`
	Url   string `gorm:"type:varchar(200);not null"`
	Index int32  `gorm:"type:int;default:1;not null"` // 轮播图顺序
}

// BannerList 轮播图列表
func (s *GoodsServer) BannerList(ctx context.Context, req *emptypb.Empty) (*proto.BannerListResponse, error) {
	var banners []Banner
	result := s.db.Find(&banners)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询轮播图列表失败: %v", result.Error)
	}

	rsp := &proto.BannerListResponse{
		Total: int32(result.RowsAffected),
		Data:  make([]*proto.BannerResponse, 0, result.RowsAffected),
	}
	for i := range banners {
		rsp.Data = append(rsp.Data, BannerModelToResponse(&banners[i]))
	}
	return rsp, nil
}

// CreateBanner 新建轮播图
func (s *GoodsServer) CreateBanner(ctx context.Context, req *proto.BannerRequest) (*proto.BannerResponse, error) {
	banner := Banner{
		Image: req.Image,
		Url:   req.Url,
		Index: req.Index,
	}
	if err := s.db.Create(&banner).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "创建轮播图失败: %v", err)
	}
	return BannerModelToResponse(&banner), nil
}

// DeleteBanner 删除轮播图
func (s *GoodsServer) DeleteBanner(ctx context.Context, req *proto.BannerRequest) (*emptypb.Empty, error) {
	result := s.db.Delete(&Banner{}, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "轮播图不存在")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "删除轮播图失败: %v", result.Error)
	}
	return &emptypb.Empty{}, nil
}

// UpdateBanner 更新轮播图
func (s *GoodsServer) UpdateBanner(ctx context.Context, req *proto.BannerRequest) (*emptypb.Empty, error) {
	result := s.db.Model(&Banner{}).Where("id = ?", req.Id).Updates(map[string]any{
		"image": req.Image,
		"url":   req.Url,
		"index": req.Index,
	})
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "轮播图不存在")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "更新轮播图失败: %v", result.Error)
	}
	return &emptypb.Empty{}, nil
}

func BannerModelToResponse(b *Banner) *proto.BannerResponse {
	return &proto.BannerResponse{
		Id:    b.ID,
		Index: b.Index,
		Image: b.Image,
		Url:   b.Url,
	}
}
