package main

import (
	"context"
	"encoding/json"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Category struct {
	basemodel.BaseModel
	Name             string     `gorm:"type:varchar(20);not null"`
	ParentCategoryID *int32     `gorm:"column:parent_category_id;type:int"`               // 父类目ID，nil 代表一级类目（dump 中顶级分类为 NULL）
	ParentCategory   *Category  `gorm:"foreignKey:ParentCategoryID;references:ID"`        // 关联父类目
	SubCategory      []Category `gorm:"foreignKey:ParentCategoryID;references:ID"`        // 子分类列表
	Level            int32      `gorm:"type:int;default:1;not null"`                      // 分类级别：1-一级分类，2-二级分类，3-三级分类
	IsTab            bool       `gorm:"column:is_tab;type:tinyint(1);default:0;not null"` // 是否显示在导航栏：0-不显示，1-显示
}

// categoryNode 是分类树的 JSON 序列化节点
type categoryNode struct {
	Id           int32          `json:"id"`
	Name         string         `json:"name"`
	Level        int32          `json:"level"`
	IsTab        bool           `json:"is_tab"`
	SubCategorys []categoryNode `json:"subCategorys"`
}

// CategoryModelToNode 把 Category 模型递归转换成 JSON 节点。
func CategoryModelToNode(c *Category) categoryNode {
	node := categoryNode{
		Id:    c.ID,
		Name:  c.Name,
		Level: c.Level,
		IsTab: c.IsTab,
	}
	if len(c.SubCategory) > 0 {
		node.SubCategorys = make([]categoryNode, 0, len(c.SubCategory))
		for i := range c.SubCategory {
			node.SubCategorys = append(node.SubCategorys, CategoryModelToNode(&c.SubCategory[i]))
		}
	}
	return node
}

// GetAllCategorysList 获取全部分类列表（含完整树形 JSON）
func (s *GoodsServer) GetAllCategorysList(ctx context.Context, req *emptypb.Empty) (*proto.CategoryListResponse, error) {
	var topCategories []Category
	// 一级分类 parent_category_id IS NULL，递归 Preload 到第三层
	result := s.db.Where("parent_category_id IS NULL").
		Preload("SubCategory.SubCategory").
		Find(&topCategories)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询分类列表失败: %v", result.Error)
	}

	rsp := &proto.CategoryListResponse{
		Total: int32(result.RowsAffected),
		Data:  make([]*proto.CategoryInfoResponse, 0, result.RowsAffected),
	}
	for i := range topCategories {
		rsp.Data = append(rsp.Data, CategoryModelToResponse(&topCategories[i]))
	}

	// 整棵树转 JSON 填 JsonData
	nodes := make([]categoryNode, 0, len(topCategories))
	for i := range topCategories {
		nodes = append(nodes, CategoryModelToNode(&topCategories[i]))
	}
	jsonBytes, err := json.Marshal(nodes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "序列化分类树失败: %v", err)
	}
	rsp.JsonData = string(jsonBytes)
	return rsp, nil
}

// GetSubCategory 获取子分类（递归加载多层）
func (s *GoodsServer) GetSubCategory(ctx context.Context, req *proto.CategoryListRequest) (*proto.SubCategoryListResponse, error) {
	var parent Category
	// Preload 两层子分类：直接子分类 + 孙分类，共三层
	result := s.db.Preload("SubCategory.SubCategory").First(&parent, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "父分类不存在")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询分类失败: %v", result.Error)
	}
	rsp := &proto.SubCategoryListResponse{
		Total:        int32(len(parent.SubCategory)),
		Info:         CategoryModelToResponse(&parent),
		SubCategorys: make([]*proto.CategoryInfoResponse, 0, len(parent.SubCategory)),
	}

	for i := range parent.SubCategory {
		rsp.SubCategorys = append(rsp.SubCategorys, CategoryModelToResponse(&parent.SubCategory[i]))
	}
	return rsp, nil
}

// CreateCategory 新建分类
func (s *GoodsServer) CreateCategory(ctx context.Context, req *proto.CategoryInfoRequest) (*proto.CategoryInfoResponse, error) {
	category := Category{
		Name:  req.Name,
		Level: req.Level,
		IsTab: req.IsTab,
	}
	if req.ParentCategory != 0 {
		category.ParentCategoryID = &req.ParentCategory
	}
	if err := s.db.Create(&category).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "创建分类失败: %v", err)
	}
	return CategoryModelToResponse(&category), nil
}

// DeleteCategory 删除分类
func (s *GoodsServer) DeleteCategory(ctx context.Context, req *proto.DeleteCategoryRequest) (*emptypb.Empty, error) {
	result := s.db.Delete(&Category{}, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "分类不存在")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "删除分类失败: %v", result.Error)
	}
	return &emptypb.Empty{}, nil
}

// UpdateCategory 更新分类
func (s *GoodsServer) UpdateCategory(ctx context.Context, req *proto.CategoryInfoRequest) (*emptypb.Empty, error) {
	updates := map[string]any{
		"name":   req.Name,
		"level":  req.Level,
		"is_tab": req.IsTab,
	}
	if req.ParentCategory != 0 {
		updates["parent_category_id"] = req.ParentCategory
	}
	result := s.db.Model(&Category{}).Where("id = ?", req.Id).Updates(updates)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "分类不存在")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "更新分类失败: %v", result.Error)
	}
	return &emptypb.Empty{}, nil
}

func CategoryModelToResponse(c *Category) *proto.CategoryInfoResponse {
	rsp := &proto.CategoryInfoResponse{
		Id:    c.ID,
		Name:  c.Name,
		Level: c.Level,
		IsTab: c.IsTab,
	}
	if c.ParentCategoryID != nil {
		rsp.ParentCategory = *c.ParentCategoryID
	}
	return rsp
}
