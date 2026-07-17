package main

import (
	"context"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type Goods struct {
	basemodel.BaseModel
	Name    string `gorm:"type:varchar(50);not null"`
	GoodsSn string `gorm:"column:goods_sn;type:varchar(60);not null"` // 商品货号

	CategoryID int32 `gorm:"column:category_id;type:int;not null"` // 商品所属分类id
	Category   *Category
	BrandID    int32   `gorm:"column:brand_id;type:int;not null"` // 商品所属品牌id
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
	GoodsBrief      string             `gorm:"column:goods_brief;type:varchar(100);not null"`       // 商品简短描述
	Images          basemodel.GormList `gorm:"column:images;type:varchar(1000);not null"`           // 商品图片，采用 JSON 数组格式
	DescImages      basemodel.GormList `gorm:"column:desc_images;type:varchar(1000);not null"`      // 商品详情图片，采用 JSON 数组格式
	GoodsFrontImage string             `gorm:"column:goods_front_image;type:varchar(200);not null"` // 商品封面图
}

type Category struct {
	basemodel.BaseModel
	Name             string    `gorm:"type:varchar(20);not null"`
	ParentCategoryID int32     `gorm:"column:parent_category_id;type:int;default:0"`     // 父类目ID=0时，代表的是一级类目
	ParentCategory   *Category `gorm:"foreignKey:ParentCategoryID;references:ID"`        // 关联父类目
	Level            int32     `gorm:"type:int;default:1;not null"`                      // 分类级别：1-一级分类，2-二级分类，3-三级分类
	IsTab            bool      `gorm:"column:is_tab;type:tinyint(1);default:0;not null"` // 是否显示在导航栏：0-不显示，1-显示
}

type Brands struct {
	basemodel.BaseModel
	Name string `gorm:"type:varchar(20);not null"`
	Logo string `gorm:"type:varchar(200);default:'';not null"`
}

// GoodsCategoryBrand 商品分类和品牌的多对多关系表, 一个商品分类可以有多个品牌，一个品牌也可以对应多个商品分类
type GoodsCategoryBrand struct {
	basemodel.BaseModel
	CategoryID int32     `gorm:"column:category_id;type:int;not null;index:idx_category_brand,unique"`
	Category   *Category `gorm:"foreignKey:CategoryID;references:ID"`
	BrandID    int32     `gorm:"column:brand_id;type:int;not null;index:idx_category_brand,unique"`
	Brand      *Brands   `gorm:"foreignKey:BrandID;references:ID"`
}

func (GoodsCategoryBrand) TableName() string {
	return "goodscategorybrand"
}

type Banner struct {
	basemodel.BaseModel
	Image string `gorm:"type:varchar(200);not null"`
	Url   string `gorm:"type:varchar(200);not null"`
	Index int32  `gorm:"type:int;default:1;not null"` // 轮播图顺序
}

// GoodsServer implements proto.GoodsServer over an injected *gorm.DB.
type GoodsServer struct {
	proto.UnimplementedGoodsServer
	db *gorm.DB
}

// NewGoodsServer wires a GoodsServer to its data store.
func NewGoodsServer(db *gorm.DB) *GoodsServer {
	return &GoodsServer{db: db}
}

// GoodsList 商品列表（分页）
func (s *GoodsServer) GoodsList(ctx context.Context, req *proto.PageInfo) (*proto.GoodsListResponse, error) {
	var goods []Goods
	result := s.db.Scopes(Paginate(int(req.Pn), int(req.PSize))).Find(&goods)
	if result.Error != nil {
		return nil, result.Error
	}

	rsp := &proto.GoodsListResponse{
		Total: int32(result.RowsAffected),
		Data:  make([]*proto.GoodsInfoResponse, 0, result.RowsAffected),
	}
	for i := range goods {
		rsp.Data = append(rsp.Data, ModelToResponse(&goods[i]))
	}
	return rsp, nil
}

// GetGoodsDetail 通过 id 查询商品
func (s *GoodsServer) GetGoodsDetail(ctx context.Context, req *proto.IdRequest) (*proto.GoodsInfoResponse, error) {
	var g Goods
	result := s.db.First(&g, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return ModelToResponse(&g), nil
}

// CreateGoods 新建商品

// UpdateGoods 更新商品（name / market_price / stock）
func (s *GoodsServer) UpdateGoods(ctx context.Context, req *proto.UpdateGoodsInfo) (*emptypb.Empty, error) {
	result := s.db.Model(&Goods{}).Where("id = ?", req.Id).Updates(map[string]any{
		"name":         req.Name,
		"market_price": req.MarketPrice,
		"stock":        req.Stock,
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

// Paginate 分页
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func ModelToResponse(g *Goods) *proto.GoodsInfoResponse {
	return &proto.GoodsInfoResponse{}
}
