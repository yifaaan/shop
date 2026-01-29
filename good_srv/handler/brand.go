package handler

import (
	"context"
	"shop/good_srv/proto"
)

// 品牌和轮播图
func (s *GoodServer) BrandList(ctx context.Context, in *proto.BrandFilterRequest) (*proto.BrandListResponse, error) {
	return nil, nil
}
func (s *GoodServer) CreateBrand(ctx context.Context, in *proto.BrandRequest) (*proto.BrandInfoResponse, error) {
	return nil, nil
}
func (s *GoodServer) DeleteBrand(ctx context.Context, in *proto.BrandRequest) (*proto.Empty, error) {
	return nil, nil
}
func (s *GoodServer) UpdateBrand(ctx context.Context, in *proto.BrandRequest) (*proto.Empty, error) {
	return nil, nil
}
