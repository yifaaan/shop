package handler

import (
	"context"
	"shop/good_srv/proto"
)

// 品牌分类
func (s *GoodServer) CategoryBrandList(ctx context.Context, in *proto.CategoryBrandFilterRequest) (*proto.CategoryBrandListResponse, error) {
	return nil, nil
}

// 通过category获取brands
func (s *GoodServer) GetCategoryBrandList(ctx context.Context, in *proto.CategoryInfoRequest) (*proto.BrandListResponse, error) {
	return nil, nil
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
