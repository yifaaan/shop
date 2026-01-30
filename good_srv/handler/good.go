package handler

import (
	"context"
	"fmt"
	"shop/good_srv/global"
	"shop/good_srv/model"
	"shop/good_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GoodServer struct {
	proto.UnimplementedGoodServer
}

var _ proto.GoodServer = (*GoodServer)(nil)

// 商品接⼝
func (s *GoodServer) GoodList(ctx context.Context, in *proto.GoodFilterRequest) (*proto.GoodListResponse, error) {
	goodModel := global.DB.WithContext(ctx).Model(model.Good{})
	// 构建查询条件
	if in.PriceMin > 0 {
		goodModel = goodModel.Where("shop_price >= ?", in.PriceMin)
	}
	if in.PriceMax > 0 {
		goodModel = goodModel.Where("shop_price <= ?", in.PriceMax)
	}
	if in.IsHot {
		goodModel = goodModel.Where(model.Good{IsHot: true})
	}
	if in.IsNew {
		goodModel = goodModel.Where(model.Good{IsNew: true})
	}

	if in.Brand > 0 {
		goodModel = goodModel.Where("brand_id = ?", in.Brand)
	}
	if in.KeyWords != "" {
		goodModel = goodModel.Where("name LIKE ?", "%"+in.KeyWords+"%")
	}

	var subQuery string
	if in.TopCategory > 0 {
		// 查询该分类
		var category model.Category
		if result := global.DB.WithContext(ctx).First(&category, in.TopCategory); result.RowsAffected == 0 {
			return nil, status.Errorf(codes.NotFound, "商品分类不存在")
		}

		switch category.Level {
		case 1:
			// 需要先找出二级分类，再找出三级分类
			subQuery = fmt.Sprintf(`
				SELECT id FROM category WHERE parent_category_id IN
				(SELECT id FROM category WHERE parent_category_id = %d)`, in.TopCategory)
		case 2:
			// 查询二级分类下的所有三级分类
			subQuery = fmt.Sprintf("SELECT id FROM category WHERE parent_category_id = %d)", in.TopCategory)
		case 3:
			subQuery = fmt.Sprintf("SELECT id FROM category WHERE id = %d", in.TopCategory)
		}
		// 根据三级分类筛选商品
		goodModel = goodModel.Where(fmt.Sprintf("category_id in (%s)", subQuery))
	}

	var total int64
	goodModel.Count(&total)

	// 查询商品
	var goods []model.Good
	result := goodModel.Preload("Category").Preload("Brand").Scopes(paginate(int(in.Pages), int(in.PagePerNums))).Find(&goods)
	if result.Error != nil {
		return nil, result.Error
	}

	var goodInfos []*proto.GoodInfoResponse
	for _, good := range goods {
		goodInfos = append(goodInfos, ModelToResponse(&good))
	}
	return &proto.GoodListResponse{
		Total: int32(total),
		Data:  goodInfos,
	}, nil
}

// 现在⽤户提交订单有多个商品，你得批量查询商品的信息吧
func (s *GoodServer) BatchGetGood(ctx context.Context, in *proto.BatchGoodIdInfo) (*proto.GoodListResponse, error) {
	return nil, nil
}
func (s *GoodServer) CreateGood(ctx context.Context, in *proto.CreateGoodInfo) (*proto.GoodInfoResponse, error) {
	return nil, nil

}
func (s *GoodServer) DeleteGood(ctx context.Context, in *proto.DeleteGoodInfo) (*proto.Empty, error) {
	return nil, nil

}
func (s *GoodServer) UpdateGood(ctx context.Context, in *proto.CreateGoodInfo) (*proto.Empty, error) {
	return nil, nil

}
func (s *GoodServer) GetGoodDetail(ctx context.Context, in *proto.GoodInfoRequest) (*proto.GoodInfoResponse, error) {
	return nil, nil

}

func ModelToResponse(good *model.Good) *proto.GoodInfoResponse {
	return &proto.GoodInfoResponse{
		Id:             good.ID,
		CategoryId:     good.CategoryID,
		Name:           good.Name,
		GoodSn:         good.GoodSn,
		ClickNum:       good.ClickNum,
		SoldNum:        good.SoldNum,
		FavNum:         good.FavNum,
		MarketPrice:    good.MarketPrice,
		ShopPrice:      good.ShopPrice,
		GoodBrief:      good.GoodBrief,
		ShipFree:       good.ShipFree,
		Images:         good.Images,
		DescImages:     good.DescImages,
		GoodFrontImage: good.GoodFrontImage,
		IsNew:          good.IsNew,
		IsHot:          good.IsHot,
		OnSale:         good.OnSale,
		AddTime:        good.CreatedAt.Unix(),
		Category: &proto.CategoryBriefInfoResponse{
			Id:   good.Category.ID,
			Name: good.Category.Name,
		},
		Brand: &proto.BrandInfoResponse{
			Id:   good.Brand.ID,
			Name: good.Brand.Name,
			Logo: good.Brand.Logo,
		},
	}
}
