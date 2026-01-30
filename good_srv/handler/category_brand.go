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

// 品牌分类
func (s *GoodServer) CategoryBrandList(ctx context.Context, in *proto.CategoryBrandFilterRequest) (*proto.CategoryBrandListResponse, error) {
	var total int64
	global.DB.Model(&model.GoodCategoryBrand{}).Count(&total)
	var categoryBrands []model.GoodCategoryBrand
	global.DB.WithContext(ctx).Scopes(paginate(int(in.Pages), int(in.PagePerNums))).Find(&categoryBrands)

	var categoryInfos []*proto.CategoryBrandResponse
	for _, cb := range categoryBrands {
		categoryInfos = append(categoryInfos, &proto.CategoryBrandResponse{
			Id: cb.ID,
			Brand: &proto.BrandInfoResponse{
				Id:   cb.Brand.ID,
				Name: cb.Brand.Name,
				Logo: cb.Brand.Logo,
			},
			Category: &proto.CategoryInfoResponse{
				Id:             cb.Category.ID,
				Name:           cb.Category.Name,
				ParentCategory: cb.Category.ParentCategoryID,
				Level:          cb.Category.Level,
				IsTab:          cb.Category.IsTab,
			},
		})
	}
	return &proto.CategoryBrandListResponse{
		Total: int32(total),
		Data:  categoryInfos,
	}, nil
}

// 通过category获取brands
func (s *GoodServer) GetCategoryBrandList(ctx context.Context, in *proto.CategoryInfoRequest) (*proto.BrandListResponse, error) {
	// 先查询分类是否存在
	var category model.Category
	result := global.DB.WithContext(ctx).First(&category, in.Id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "商品分类不存在")
		}
		return nil, status.Errorf(codes.Internal, "查询分类失败: %v", result.Error)
	}

	resp := &proto.BrandListResponse{}
	// 查询分类对应的品牌
	var categoryBrands []model.GoodCategoryBrand
	result = global.DB.WithContext(ctx).Where("category_id = ?", in.Id).Find(&categoryBrands)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.Internal, "查询分类品牌失败: %v", result.Error)
	}
	resp.Total = int32(result.RowsAffected)

	var brandInfos []*proto.BrandInfoResponse
	for _, cb := range categoryBrands {
		brandInfos = append(brandInfos, &proto.BrandInfoResponse{
			Id:   cb.BrandID,
			Name: cb.Brand.Name,
			Logo: cb.Brand.Logo,
		})
	}
	resp.Data = brandInfos
	return resp, nil
}
func (s *GoodServer) CreateCategoryBrand(ctx context.Context, in *proto.CategoryBrandRequest) (*proto.CategoryBrandResponse, error) {
	return nil, nil
}
func (s *GoodServer) DeleteCategoryBrand(ctx context.Context, in *proto.CategoryBrandRequest) (*proto.Empty, error) {
	return nil, nil
}
func (s *GoodServer) UpdateCategoryBrand(ctx context.Context, in *proto.CategoryBrandRequest) (*proto.Empty, error) {
	return nil, nil
}
