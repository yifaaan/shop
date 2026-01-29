package handler

import (
	"context"
	"shop/good_srv/proto"
)

// 商品分类
func (s *GoodServer) GetAllCategorysList(ctx context.Context, in *proto.Empty) (*proto.CategoryListResponse, error) {
	return nil, nil

}

// 获取⼦分类
func (s *GoodServer) GetSubCategory(ctx context.Context, in *proto.CategoryListRequest) (*proto.SubCategoryListResponse, error) {
	return nil, nil

}
func (s *GoodServer) CreateCategory(ctx context.Context, in *proto.CategoryInfoRequest) (*proto.CategoryInfoResponse, error) {
	return nil, nil

}
func (s *GoodServer) DeleteCategory(ctx context.Context, in *proto.DeleteCategoryRequest) (*proto.Empty, error) {
	return nil, nil

}
func (s *GoodServer) UpdateCategory(ctx context.Context, in *proto.CategoryInfoRequest) (*proto.Empty, error) {
	return nil, nil

}
