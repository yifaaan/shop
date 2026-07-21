package main

import (
	"context"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GoodsCategoryBrand 商品分类和品牌的多对多关系表, 一个商品分类可以有多个品牌，一个品牌也可以对应多个商品分类
type GoodsCategoryBrand struct {
	basemodel.BaseModel
	CategoryID int32     `gorm:"column:category_id;type:int;not null;index:goodscategorybrand_category_id_brand_id,unique"`
	Category   *Category `gorm:"foreignKey:CategoryID;references:ID"`
	BrandID    int32     `gorm:"column:brands_id;type:int;not null;index:goodscategorybrand_category_id_brand_id,unique"`
	Brand      *Brands   `gorm:"foreignKey:BrandID;references:ID"`
}

func (GoodsCategoryBrand) TableName() string {
	return "goodscategorybrand"
}

// CategoryBrandList 分类-品牌关系列表（分页）
func (s *GoodsServer) CategoryBrandList(ctx context.Context, req *proto.CategoryBrandFilterRequest) (*proto.CategoryBrandListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CategoryBrandList not implemented")
}

// GetCategoryBrandList 通过 category 获取其关联的 brands
func (s *GoodsServer) GetCategoryBrandList(ctx context.Context, req *proto.CategoryInfoRequest) (*proto.BrandListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCategoryBrandList not implemented")
}

// CreateCategoryBrand 新建分类-品牌关系
func (s *GoodsServer) CreateCategoryBrand(ctx context.Context, req *proto.CategoryBrandRequest) (*proto.CategoryBrandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateCategoryBrand not implemented")
}

// DeleteCategoryBrand 删除分类-品牌关系
func (s *GoodsServer) DeleteCategoryBrand(ctx context.Context, req *proto.CategoryBrandRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteCategoryBrand not implemented")
}

// UpdateCategoryBrand 更新分类-品牌关系
func (s *GoodsServer) UpdateCategoryBrand(ctx context.Context, req *proto.CategoryBrandRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCategoryBrand not implemented")
}