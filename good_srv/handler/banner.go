package handler

import (
	"context"
	"shop/good_srv/proto"
)

// 轮播图
func (s *GoodServer) BannerList(ctx context.Context, in *proto.Empty) (*proto.BannerListResponse, error) {
	return nil, nil
}
func (s *GoodServer) CreateBanner(ctx context.Context, in *proto.BannerRequest) (*proto.BannerResponse, error) {
	return nil, nil
}
func (s *GoodServer) DeleteBanner(ctx context.Context, in *proto.BannerRequest) (*proto.Empty, error) {
	return nil, nil
}
func (s *GoodServer) UpdateBanner(ctx context.Context, in *proto.BannerRequest) (*proto.Empty, error) {
	return nil, nil
}
