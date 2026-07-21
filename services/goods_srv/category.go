package main

import (
	"context"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Category struct {
	basemodel.BaseModel
	Name             string    `gorm:"type:varchar(20);not null"`
	ParentCategoryID *int32    `gorm:"column:parent_category_id;type:int"`               // 父类目ID，nil 代表一级类目（dump 中顶级分类为 NULL）
	ParentCategory   *Category `gorm:"foreignKey:ParentCategoryID;references:ID"`        // 关联父类目
	Level            int32     `gorm:"type:int;default:1;not null"`                      // 分类级别：1-一级分类，2-二级分类，3-三级分类
	IsTab            bool      `gorm:"column:is_tab;type:tinyint(1);default:0;not null"` // 是否显示在导航栏：0-不显示，1-显示
}

// GetAllCategorysList 获取全部分类列表
func (s *GoodsServer) GetAllCategorysList(ctx context.Context, req *emptypb.Empty) (*proto.CategoryListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllCategorysList not implemented")
}

// GetSubCategory 获取子分类
func (s *GoodsServer) GetSubCategory(ctx context.Context, req *proto.CategoryListRequest) (*proto.SubCategoryListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSubCategory not implemented")
}

// CreateCategory 新建分类
func (s *GoodsServer) CreateCategory(ctx context.Context, req *proto.CategoryInfoRequest) (*proto.CategoryInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateCategory not implemented")
}

// DeleteCategory 删除分类
func (s *GoodsServer) DeleteCategory(ctx context.Context, req *proto.DeleteCategoryRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteCategory not implemented")
}

// UpdateCategory 更新分类
func (s *GoodsServer) UpdateCategory(ctx context.Context, req *proto.CategoryInfoRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCategory not implemented")
}