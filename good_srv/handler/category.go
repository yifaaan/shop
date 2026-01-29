package handler

import (
	"context"
	"encoding/json"
	"shop/good_srv/global"
	"shop/good_srv/model"
	"shop/good_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 商品分类
func (s *GoodServer) GetAllCategorysList(ctx context.Context, in *proto.Empty) (*proto.CategoryListResponse, error) {
	var categories []*model.Category
	result := global.DB.WithContext(ctx).Where("level = ?", 1).Preload("SubCategories.SubCategories").Find(&categories)
	if result.Error != nil {
		return nil, result.Error
	}
	categoryInfos := make([]*proto.CategoryInfoResponse, 0, len(categories))
	for _, c := range categories {
		categoryInfos = append(categoryInfos, &proto.CategoryInfoResponse{
			Id:             c.ID,
			Name:           c.Name,
			ParentCategory: c.ParentCategoryID,
			Level:          c.Level,
			IsTab:          c.IsTab,
		})
	}
	jsonData, err := json.Marshal(categories)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "json marshal error: %v", err)
	}
	return &proto.CategoryListResponse{
		Total:    int32(result.RowsAffected),
		Data:     categoryInfos,
		JsonData: string(jsonData),
	}, nil
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
