package main

import (
	"context"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Goods struct {
	basemodel.BaseModel
	Name    string `gorm:"type:varchar(100);not null"`
	GoodsSn string `gorm:"column:goods_sn;type:varchar(50);not null"` // 商品货号

	CategoryID int32 `gorm:"column:category_id;type:int;not null"` // 商品所属分类id
	Category   *Category
	BrandID    int32   `gorm:"column:brands_id;type:int;not null"` // 商品所属品牌id（与 mxshop 原版 brands 表对齐，列名为 brands_id）
	Brand      *Brands // 关联品牌表

	OnSale   bool `gorm:"column:on_sale;default:false;not null"`   // 是否上架销售
	ShipFree bool `gorm:"column:ship_free;default:false;not null"` // 是否包邮
	IsNew    bool `gorm:"column:is_new;default:false;not null"`    // 是否新品首发
	IsHot    bool `gorm:"column:is_hot;default:false;not null"`    // 是否热销推荐

	ClickNum        int32              `gorm:"column:click_num;default:0;not null"`                 // 商品点击数
	SoldNum         int32              `gorm:"column:sold_num;default:0;not null"`                  // 商品销售量
	FavNum          int32              `gorm:"column:fav_num;default:0;not null"`                   // 商品收藏数
	MarketPrice     float32            `gorm:"column:market_price;not null"`                        // 商品市场价
	ShopPrice       float32            `gorm:"column:shop_price;not null"`                          // 商品本店价
	GoodsBrief      string             `gorm:"column:goods_brief;type:varchar(200);not null"`       // 商品简短描述
	Images          basemodel.GormList `gorm:"column:images;type:json;not null"`                    // 商品图片，采用 JSON 数组格式
	DescImages      basemodel.GormList `gorm:"column:desc_images;type:json;not null"`               // 商品详情图片，采用 JSON 数组格式
	GoodsFrontImage string             `gorm:"column:goods_front_image;type:varchar(200);not null"` // 商品封面图
}

// GoodsList 商品列表（分页）
func (s *GoodsServer) GoodsList(ctx context.Context, req *proto.GoodsFilterRequest) (*proto.GoodsListResponse, error) {
	var goods []Goods
	result := s.db.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&goods)
	if result.Error != nil {
		return nil, result.Error
	}

	rsp := &proto.GoodsListResponse{
		Total: int32(result.RowsAffected),
		Data:  make([]*proto.GoodsInfoResponse, 0, result.RowsAffected),
	}
	for i := range goods {
		rsp.Data = append(rsp.Data, GoodsModelToResponse(&goods[i]))
	}
	return rsp, nil
}

// BatchGetGoods 批量查询商品信息（用户下单时一次性拉取多个商品）
func (s *GoodsServer) BatchGetGoods(ctx context.Context, req *proto.BatchGoodsIdInfo) (*proto.GoodsListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchGetGoods not implemented")
}

// CreateGoods 新建商品
func (s *GoodsServer) CreateGoods(ctx context.Context, req *proto.CreateGoodsInfo) (*proto.GoodsInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateGoods not implemented")
}

// GetGoodsDetail 通过 id 查询商品
func (s *GoodsServer) GetGoodsDetail(ctx context.Context, req *proto.GoodInfoRequest) (*proto.GoodsInfoResponse, error) {
	var g Goods
	result := s.db.First(&g, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return GoodsModelToResponse(&g), nil
}

// UpdateGoods 更新商品（name / market_price / stock）
func (s *GoodsServer) UpdateGoods(ctx context.Context, req *proto.CreateGoodsInfo) (*emptypb.Empty, error) {
	result := s.db.Model(&Goods{}).Where("id = ?", req.Id).Updates(map[string]any{
		"name":         req.Name,
		"market_price": req.MarketPrice,
		"stocks":       req.Stocks,
	})
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	if result.Error != nil {
		return nil, status.Error(codes.Internal, result.Error.Error())
	}
	return &emptypb.Empty{}, nil
}

// DeleteGoods 软删除商品（BaseModel.DeletedAt）
func (s *GoodsServer) DeleteGoods(ctx context.Context, req *proto.DeleteGoodsInfo) (*emptypb.Empty, error) {
	result := s.db.Delete(&Goods{}, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &emptypb.Empty{}, nil
}

func GoodsModelToResponse(g *Goods) *proto.GoodsInfoResponse {
	rsp := &proto.GoodsInfoResponse{
		Id:              g.ID,
		CategoryId:      g.CategoryID,
		Name:            g.Name,
		GoodsSn:         g.GoodsSn,
		ClickNum:        g.ClickNum,
		SoldNum:         g.SoldNum,
		FavNum:          g.FavNum,
		MarketPrice:     g.MarketPrice,
		ShopPrice:       g.ShopPrice,
		GoodsBrief:      g.GoodsBrief,
		ShipFree:        g.ShipFree,
		Images:          []string(g.Images),
		DescImages:      []string(g.DescImages),
		GoodsFrontImage: g.GoodsFrontImage,
		IsNew:           g.IsNew,
		IsHot:           g.IsHot,
		OnSale:          g.OnSale,
		AddTime:         g.CreatedAt.Unix(),
	}
	if g.Category != nil {
		rsp.Category = &proto.CategoryBriefInfoResponse{
			Id:   g.Category.ID,
			Name: g.Category.Name,
		}
	}
	if g.Brand != nil {
		rsp.Brand = &proto.BrandInfoResponse{
			Id:   g.Brand.ID,
			Name: g.Brand.Name,
			Logo: g.Brand.Logo,
		}
	}
	return rsp
}
