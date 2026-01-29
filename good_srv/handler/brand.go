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

// 品牌和轮播图

// BrandList 品牌列表
func (s *GoodServer) BrandList(ctx context.Context, in *proto.BrandFilterRequest) (*proto.BrandListResponse, error) {
	var brands []*model.Brand
	result := global.DB.WithContext(ctx).Scopes(paginate(int(in.Pages), int(in.PagePerNums))).Find(&brands)
	if result.Error != nil {
		return nil, result.Error
	}

	var total int64
	result = global.DB.WithContext(ctx).Model(&model.Brand{}).Count(&total)
	if result.Error != nil {
		return nil, result.Error
	}

	brandInfos := make([]*proto.BrandInfoResponse, 0, len(brands))
	for _, b := range brands {
		brandInfos = append(brandInfos, &proto.BrandInfoResponse{
			Id:   b.ID,
			Name: b.Name,
			Logo: b.Logo,
		})
	}
	return &proto.BrandListResponse{
		Total: int32(total),
		Data:  brandInfos,
	}, nil
}

// CreateBrand 创建品牌
func (s *GoodServer) CreateBrand(ctx context.Context, in *proto.BrandRequest) (*proto.BrandInfoResponse, error) {
	brand := model.Brand{}
	if result := global.DB.WithContext(ctx).Where("name = ?", in.Name).First(&brand); result.RowsAffected == 1 {
		return nil, status.Errorf(codes.AlreadyExists, "品牌名称已存在")
	}

	brand.Logo = in.Logo
	brand.Name = in.Name
	result := global.DB.WithContext(ctx).Create(&brand)
	if result.Error != nil {
		return nil, result.Error
	}

	return &proto.BrandInfoResponse{
		Id:   brand.ID,
		Name: brand.Name,
		Logo: brand.Logo,
	}, nil
}

// DeleteBrand 删除品牌
func (s *GoodServer) DeleteBrand(ctx context.Context, in *proto.BrandRequest) (*proto.Empty, error) {
	result := global.DB.WithContext(ctx).Delete(&model.Brand{}, in.Id)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "品牌不存在")
	}
	return &proto.Empty{}, nil
}

// UpdateBrand 更新品牌
func (s *GoodServer) UpdateBrand(ctx context.Context, in *proto.BrandRequest) (*proto.Empty, error) {
	var brand model.Brand
	result := global.DB.WithContext(ctx).First(&brand, in.Id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "品牌不存在")
		}
		return nil, result.Error
	}

	if in.Name != "" {
		brand.Name = in.Name
	}
	if in.Logo != "" {
		brand.Logo = in.Logo
	}
	result = global.DB.WithContext(ctx).Model(&brand).Updates(&brand)
	if result.Error != nil {
		return nil, result.Error
	}
	return &proto.Empty{}, nil
}
