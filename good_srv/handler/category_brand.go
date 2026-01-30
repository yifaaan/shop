package handler

import (
	"context"
	"shop/good_srv/global"
	"shop/good_srv/model"
	"shop/good_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// 品牌分类
func (s *GoodServer) CategoryBrandList(ctx context.Context, in *proto.CategoryBrandFilterRequest) (*proto.CategoryBrandListResponse, error) {
	var total int64
	global.DB.Model(&model.GoodCategoryBrand{}).Count(&total)
	var categoryBrands []model.GoodCategoryBrand
	global.DB.WithContext(ctx).Scopes(paginate(int(in.Pages), int(in.PagePerNums))).Find(&categoryBrands)

	var categoryInfos []*proto.CategoryBrandResponse
	for _, cb := range categoryBrands {
		categoryInfos = append(categoryInfos, &proto.CategoryBrandResponse{
			Id: cb.ID,
			Brand: &proto.BrandInfoResponse{
				Id:   cb.Brand.ID,
				Name: cb.Brand.Name,
				Logo: cb.Brand.Logo,
			},
			Category: &proto.CategoryInfoResponse{
				Id:             cb.Category.ID,
				Name:           cb.Category.Name,
				ParentCategory: cb.Category.ParentCategoryID,
				Level:          cb.Category.Level,
				IsTab:          cb.Category.IsTab,
			},
		})
	}
	return &proto.CategoryBrandListResponse{
		Total: int32(total),
		Data:  categoryInfos,
	}, nil
}

// 通过category获取brands
func (s *GoodServer) GetCategoryBrandList(ctx context.Context, in *proto.CategoryInfoRequest) (*proto.BrandListResponse, error) {
	// 先查询分类是否存在
	var category model.Category
	result := global.DB.WithContext(ctx).First(&category, in.Id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "商品分类不存在")
		}
		return nil, status.Errorf(codes.Internal, "查询分类失败: %v", result.Error)
	}

	resp := &proto.BrandListResponse{}
	// 查询分类对应的品牌
	var categoryBrands []model.GoodCategoryBrand
	result = global.DB.WithContext(ctx).Where("category_id = ?", in.Id).Find(&categoryBrands)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.Internal, "查询分类品牌失败: %v", result.Error)
	}
	resp.Total = int32(result.RowsAffected)

	var brandInfos []*proto.BrandInfoResponse
	for _, cb := range categoryBrands {
		brandInfos = append(brandInfos, &proto.BrandInfoResponse{
			Id:   cb.BrandID,
			Name: cb.Brand.Name,
			Logo: cb.Brand.Logo,
		})
	}
	resp.Data = brandInfos
	return resp, nil
}
func (s *GoodServer) CreateCategoryBrand(ctx context.Context, in *proto.CategoryBrandRequest) (*proto.CategoryBrandResponse, error) {
	// 1. 验证分类是否存在
	var category model.Category
	if result := global.DB.WithContext(ctx).First(&category, in.CategoryId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}

	// 2. 验证品牌是否存在
	var brand model.Brand
	if result := global.DB.WithContext(ctx).First(&brand, in.BrandId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "品牌不存在")
	}

	// 3. 判断分类-品牌关系是否存在
	var categoryBrand model.GoodCategoryBrand
	if result := global.DB.WithContext(ctx).Where("category_id = ? AND brand_id = ?", in.CategoryId, in.BrandId).First(&categoryBrand); result.RowsAffected == 1 {
		return nil, status.Errorf(codes.AlreadyExists, "分类-品牌关系已存在")
	}

	// 4. 插入数据
	categoryBrand = model.GoodCategoryBrand{
		CategoryID: in.CategoryId,
		BrandID:    in.BrandId,
	}
	result := global.DB.WithContext(ctx).Create(&categoryBrand)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "创建分类-品牌失败: %v", result.Error)
	}

	// 5. 构造返回值 (需要完整的Category和Brand信息，已经在验证步骤查询到了，或者重新Preload)
	return &proto.CategoryBrandResponse{
		Id: categoryBrand.ID,
		Brand: &proto.BrandInfoResponse{
			Id:   brand.ID,
			Name: brand.Name,
			Logo: brand.Logo,
		},
		Category: &proto.CategoryInfoResponse{
			Id:             category.ID,
			Name:           category.Name,
			ParentCategory: category.ParentCategoryID,
			Level:          category.Level,
			IsTab:          category.IsTab,
		},
	}, nil
}
func (s *GoodServer) DeleteCategoryBrand(ctx context.Context, in *proto.CategoryBrandRequest) (*proto.Empty, error) {
	if in.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "参数错误")
	}

	// First check if record exists (GORM ignores soft-deleted records by default)
	var categoryBrand model.GoodCategoryBrand
	if result := global.DB.WithContext(ctx).First(&categoryBrand, in.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "记录不存在")
	}

	// Delete
	result := global.DB.WithContext(ctx).Delete(&categoryBrand)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "删除失败: %v", result.Error)
	}

	return &proto.Empty{}, nil
}

func (s *GoodServer) UpdateCategoryBrand(ctx context.Context, in *proto.CategoryBrandRequest) (*proto.Empty, error) {
	var categoryBrand model.GoodCategoryBrand
	if result := global.DB.WithContext(ctx).First(&categoryBrand, in.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "记录不存在")
	}

	// Validate new Category if changed
	if in.CategoryId != 0 {
		var category model.Category
		if res := global.DB.WithContext(ctx).First(&category, in.CategoryId); res.RowsAffected == 0 {
			return nil, status.Errorf(codes.NotFound, "分类不存在")
		}
		categoryBrand.CategoryID = in.CategoryId
	}

	// Validate new Brand if changed
	if in.BrandId != 0 {
		var brand model.Brand
		if res := global.DB.WithContext(ctx).First(&brand, in.BrandId); res.RowsAffected == 0 {
			return nil, status.Errorf(codes.NotFound, "品牌不存在")
		}
		categoryBrand.BrandID = in.BrandId
	}

	// Check for duplicates
	var checkDuplicate model.GoodCategoryBrand
	if result := global.DB.WithContext(ctx).Where("category_id = ? AND brand_id = ? AND id != ?", categoryBrand.CategoryID, categoryBrand.BrandID, in.Id).First(&checkDuplicate); result.RowsAffected > 0 {
		return nil, status.Errorf(codes.AlreadyExists, "分类-品牌关系已存在")
	}

	if err := global.DB.WithContext(ctx).Save(&categoryBrand).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "更新失败: %v", err)
	}

	return &proto.Empty{}, nil
}
