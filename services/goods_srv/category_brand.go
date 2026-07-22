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
	var cbs []GoodsCategoryBrand
	result := s.db.Preload("Category").Preload("Brand").
		Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).
		Find(&cbs)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询分类品牌列表失败: %v", result.Error)
	}
	rsp := &proto.CategoryBrandListResponse{
		Total: int32(result.RowsAffected),
		Data:  make([]*proto.CategoryBrandResponse, 0, result.RowsAffected),
	}
	for i := range cbs {
		rsp.Data = append(rsp.Data, CategoryBrandModelToResponse(&cbs[i]))
	}
	return rsp, nil
}

// GetCategoryBrandList 通过 category 获取其关联的 brands
func (s *GoodsServer) GetCategoryBrandList(ctx context.Context, req *proto.CategoryInfoRequest) (*proto.BrandListResponse, error) {
	var cbs []GoodsCategoryBrand
	result := s.db.Preload("Brand").
		Where("category_id = ?", req.Id).
		Find(&cbs)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询分类下品牌失败: %v", result.Error)
	}
	rsp := &proto.BrandListResponse{
		Total: int32(result.RowsAffected),
		Data:  make([]*proto.BrandInfoResponse, 0, result.RowsAffected),
	}
	for i := range cbs {
		if cbs[i].Brand != nil {
			rsp.Data = append(rsp.Data, BrandModelToResponse(cbs[i].Brand))
		}
	}
	return rsp, nil
}

// CreateCategoryBrand 新建分类-品牌关系
func (s *GoodsServer) CreateCategoryBrand(ctx context.Context, req *proto.CategoryBrandRequest) (*proto.CategoryBrandResponse, error) {
	cb := GoodsCategoryBrand{
		CategoryID: req.CategoryId,
		BrandID:    req.BrandId,
	}
	if err := s.db.Create(&cb).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "创建分类品牌关系失败: %v", err)
	}
	// 回填关联对象，便于 Response 直接带出分类和品牌信息
	var category Category
	var brand Brands
	s.db.First(&category, cb.CategoryID)
	s.db.First(&brand, cb.BrandID)
	cb.Category = &category
	cb.Brand = &brand
	return CategoryBrandModelToResponse(&cb), nil
}

// DeleteCategoryBrand 删除分类-品牌关系
func (s *GoodsServer) DeleteCategoryBrand(ctx context.Context, req *proto.CategoryBrandRequest) (*emptypb.Empty, error) {
	result := s.db.Delete(&GoodsCategoryBrand{}, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "分类品牌关系不存在")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "删除分类品牌关系失败: %v", result.Error)
	}
	return &emptypb.Empty{}, nil
}

// UpdateCategoryBrand 更新分类-品牌关系
func (s *GoodsServer) UpdateCategoryBrand(ctx context.Context, req *proto.CategoryBrandRequest) (*emptypb.Empty, error) {
	result := s.db.Model(&GoodsCategoryBrand{}).Where("id = ?", req.Id).Updates(map[string]any{
		"category_id": req.CategoryId,
		"brands_id":   req.BrandId,
	})
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "分类品牌关系不存在")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "更新分类品牌关系失败: %v", result.Error)
	}
	return &emptypb.Empty{}, nil
}

func CategoryBrandModelToResponse(cb *GoodsCategoryBrand) *proto.CategoryBrandResponse {
	rsp := &proto.CategoryBrandResponse{
		Id: cb.ID,
	}
	if cb.Category != nil {
		rsp.Category = CategoryModelToResponse(cb.Category)
	}
	if cb.Brand != nil {
		rsp.Brand = BrandModelToResponse(cb.Brand)
	}
	return rsp
}