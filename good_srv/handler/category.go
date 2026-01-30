package handler

import (
	"context"
	"encoding/json"
	"shop/good_srv/global"
	"shop/good_srv/model"
	"shop/good_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
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
	var category model.Category
	result := global.DB.WithContext(ctx).Preload("SubCategories").First(&category, in.Id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "商品分类不存在")
		}

		return nil, status.Errorf(codes.Internal, "查询分类失败: %v", result.Error)
	}

	resp := &proto.SubCategoryListResponse{}
	resp.Info = &proto.CategoryInfoResponse{
		Name:           category.Name,
		Id:             category.ID,
		ParentCategory: category.ParentCategoryID,
		Level:          category.Level,
		IsTab:          category.IsTab,
	}
	if category.Level == 1 || category.Level == 2 {
		subCategorys := make([]*proto.CategoryInfoResponse, 0, len(category.SubCategories))
		for _, sub := range category.SubCategories {
			subCategorys = append(subCategorys, &proto.CategoryInfoResponse{
				Id:             sub.ID,
				Name:           sub.Name,
				ParentCategory: sub.ParentCategoryID,
				Level:          sub.Level,
				IsTab:          sub.IsTab,
			})
		}
		resp.SubCategorys = subCategorys
		resp.Total = int32(len(subCategorys))
	} else {
		resp.SubCategorys = []*proto.CategoryInfoResponse{}
		resp.Total = 0
	}
	return resp, nil
}
func (s *GoodServer) CreateCategory(ctx context.Context, in *proto.CategoryInfoRequest) (*proto.CategoryInfoResponse, error) {
	category := model.Category{
		Name:  in.Name,
		Level: in.Level,
	}
	if in.Level != 1 {
		category.ParentCategoryID = in.ParentCategory
	}

	// Check if category with same name already exists
	var existingCategory model.Category
	if result := global.DB.WithContext(ctx).Where("name = ?", in.Name).First(&existingCategory); result.RowsAffected == 1 {
		return nil, status.Errorf(codes.AlreadyExists, "分类已存在")
	}

	category.IsTab = in.IsTab
	result := global.DB.WithContext(ctx).Create(&category)
	if result.Error != nil {
		return nil, result.Error
	}

	return &proto.CategoryInfoResponse{
		Id:             category.ID,
		Name:           category.Name,
		ParentCategory: category.ParentCategoryID,
		Level:          category.Level,
		IsTab:          category.IsTab,
	}, nil
}
func (s *GoodServer) DeleteCategory(ctx context.Context, in *proto.DeleteCategoryRequest) (*proto.Empty, error) {
	result := global.DB.WithContext(ctx).Delete(&model.Category{}, in.Id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &proto.Empty{}, nil
}
func (s *GoodServer) UpdateCategory(ctx context.Context, in *proto.CategoryInfoRequest) (*proto.Empty, error) {
	var category model.Category
	result := global.DB.WithContext(ctx).First(&category, in.Id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "分类不存在")
		}
		return nil, result.Error
	}

	if in.Name != "" {
		category.Name = in.Name
	}
	if in.ParentCategory > 0 {
		category.ParentCategoryID = in.ParentCategory
	}
	if in.Level > 0 {
		category.Level = in.Level
	}
	category.IsTab = in.IsTab
	result = global.DB.WithContext(ctx).Model(&category).Updates(&category)
	if result.Error != nil {
		return nil, result.Error
	}
	return &proto.Empty{}, nil
}
