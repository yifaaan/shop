package handler

import (
	"context"
	"shop/good_srv/global"
	"shop/good_srv/model"
	"shop/good_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// 轮播图
func (s *GoodServer) BannerList(ctx context.Context, in *proto.Empty) (*proto.BannerListResponse, error) {
	var banners []model.Banner
	result := global.DB.Find(&banners)
	if result.Error != nil {
		return nil, result.Error
	}
	bannerInfos := make([]*proto.BannerResponse, 0, len(banners))
	for _, b := range banners {
		bannerInfos = append(bannerInfos, &proto.BannerResponse{
			Id:    b.ID,
			Image: b.Image,
			Url:   b.Url,
			Index: b.Index,
		})
	}
	return &proto.BannerListResponse{
		Total: int32(result.RowsAffected),
		Data:  bannerInfos,
	}, nil
}

func (s *GoodServer) CreateBanner(ctx context.Context, in *proto.BannerRequest) (*proto.BannerResponse, error) {
	banner := model.Banner{
		Image: in.Image,
		Url:   in.Url,
		Index: in.Index,
	}
	if result := global.DB.WithContext(ctx).Where(&banner).First(&banner); result.RowsAffected == 1 {
		return nil, status.Errorf(codes.AlreadyExists, "轮播图已存在")
	}

	result := global.DB.WithContext(ctx).Create(&banner)
	if result.Error != nil {
		return nil, result.Error
	}

	return &proto.BannerResponse{
		Id:    banner.ID,
		Image: banner.Image,
		Url:   banner.Url,
		Index: banner.Index,
	}, nil
}

func (s *GoodServer) DeleteBanner(ctx context.Context, in *proto.BannerRequest) (*proto.Empty, error) {
	result := global.DB.WithContext(ctx).Delete(&model.Banner{}, in.Id)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "轮播图不存在")
	}
	return &proto.Empty{}, nil
}
func (s *GoodServer) UpdateBanner(ctx context.Context, in *proto.BannerRequest) (*proto.Empty, error) {
	var banner model.Banner
	result := global.DB.WithContext(ctx).First(&banner, in.Id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "轮播图不存在")
		}
		return nil, result.Error
	}

	if in.Image != "" {
		banner.Image = in.Image
	}
	if in.Url != "" {
		banner.Url = in.Url
	}
	if in.Index != 0 {
		banner.Index = in.Index
	}
	result = global.DB.WithContext(ctx).Model(&banner).Updates(&banner)
	if result.Error != nil {
		return nil, result.Error
	}
	return &proto.Empty{}, nil
}
