package main

import (
	"context"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"
	"shop/services/goods_srv/esearch"

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

// GoodsList 商品列表（分页 + 过滤）。
// 带关键词时走 Elasticsearch IK 检索（按相关性返回 ID，MySQL 按 ID 补全全字段）；
// 无关键词或 ES 不可用时回退 MySQL 过滤分页。
func (s *GoodsServer) GoodsList(ctx context.Context, req *proto.GoodsFilterRequest) (*proto.GoodsListResponse, error) {
	if req.KeyWords != "" && s.es != nil {
		rsp, err := s.goodsListFromES(ctx, req)
		if err == nil {
			return rsp, nil
		}
		// ES 失败则降级到 MySQL 路径，保证可用性
	}
	return s.goodsListFromMySQL(ctx, req)
}

// goodsListFromES 关键词路径：ES bool query（关键词 must + 结构化 filter + 分页）。
func (s *GoodsServer) goodsListFromES(ctx context.Context, req *proto.GoodsFilterRequest) (*proto.GoodsListResponse, error) {
	// 顶级分类：复用 MySQL 子分类查询解析允许的 category_id 集合
	var categoryIDs []int32
	if req.TopCategory > 0 {
		if err := s.db.Model(&Category{}).
			Where("id = ? OR parent_category_id = ?", req.TopCategory, req.TopCategory).
			Pluck("id", &categoryIDs).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "查询分类失败: %v", err)
		}
	}
	// 分页参数与 Paginate 同口径
	page := int(req.Pages)
	if page <= 0 {
		page = 1
	}
	size := int(req.PagePerNums)
	switch {
	case size > 100:
		size = 100
	case size <= 0:
		size = 10
	}
	from := (page - 1) * size

	ids, total, err := s.es.SearchIDs(ctx,
		req.KeyWords,
		float64(req.PriceMin), float64(req.PriceMax),
		req.IsHot, req.IsNew, req.Brand,
		categoryIDs, from, size,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "搜索商品失败: %v", err)
	}
	if len(ids) == 0 {
		return &proto.GoodsListResponse{Total: total, Data: []*proto.GoodsInfoResponse{}}, nil
	}

	var goods []Goods
	if err := s.db.Preload("Category").Preload("Brand").Where("id IN ?", ids).Find(&goods).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "查询商品失败: %v", err)
	}
	// 按 ES 相关性顺序（ids）重排
	byID := make(map[int32]*Goods, len(goods))
	for i := range goods {
		byID[goods[i].ID] = &goods[i]
	}
	data := make([]*proto.GoodsInfoResponse, 0, len(ids))
	for _, id := range ids {
		if g := byID[id]; g != nil {
			data = append(data, GoodsModelToResponse(g))
		}
	}
	return &proto.GoodsListResponse{Total: total, Data: data}, nil
}

// goodsListFromMySQL 无关键词路径：纯 MySQL 过滤分页（ES 不参与）。
// 注意：Total 取 result.RowsAffected，在 Limit/Offset 下实为当前页条数，
// 属既有行为，本次保持不变。
func (s *GoodsServer) goodsListFromMySQL(ctx context.Context, req *proto.GoodsFilterRequest) (*proto.GoodsListResponse, error) {
	var goods []Goods
	q := s.db.Model(&Goods{}).Preload("Category").Preload("Brand")
	// 价格区间过滤
	if req.PriceMin > 0 {
		q = q.Where("shop_price >= ?", req.PriceMin)
	}
	if req.PriceMax > 0 {
		q = q.Where("shop_price <= ?", req.PriceMax)
	}
	// 促销标记过滤
	if req.IsHot {
		q = q.Where("is_hot = ?", true)
	}
	if req.IsNew {
		q = q.Where("is_new = ?", true)
	}
	// 关键词模糊匹配商品名（降级路径的兜底）
	if req.KeyWords != "" {
		q = q.Where("name LIKE ?", "%"+req.KeyWords+"%")
	}
	// 品牌过滤
	if req.Brand > 0 {
		q = q.Where("brands_id = ?", req.Brand)
	}
	// 顶级分类过滤：通过子分类查所有商品
	if req.TopCategory > 0 {
		q = q.Where("category_id IN (?)",
			s.db.Model(&Category{}).Select("id").Where("id = ? OR parent_category_id = ?", req.TopCategory, req.TopCategory))
	}

	result := q.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&goods)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询商品列表失败: %v", result.Error)
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

// goodsToDoc 把 Goods 模型转为 ES 文档（仅检索所需字段）。
func goodsToDoc(g *Goods) esearch.GoodsDoc {
	return esearch.GoodsDoc{
		ID:          g.ID,
		Name:        g.Name,
		GoodsBrief:  g.GoodsBrief,
		ShopPrice:   float64(g.ShopPrice),
		MarketPrice: float64(g.MarketPrice),
		ShipFree:    g.ShipFree,
		IsHot:       g.IsHot,
		IsNew:       g.IsNew,
		OnSale:      g.OnSale,
		ClickNum:    g.ClickNum,
		SoldNum:     g.SoldNum,
		FavNum:      g.FavNum,
		BrandsID:    g.BrandID,
		CategoryID:  g.CategoryID,
	}
}

// BatchGetGoods 批量查询商品信息（用户下单时一次性拉取多个商品）
func (s *GoodsServer) BatchGetGoods(ctx context.Context, req *proto.BatchGoodsIdInfo) (*proto.GoodsListResponse, error) {
	if len(req.Id) == 0 {
		return &proto.GoodsListResponse{Data: []*proto.GoodsInfoResponse{}}, nil
	}
	var goods []Goods
	result := s.db.Preload("Category").Preload("Brand").Where("id IN ?", req.Id).Find(&goods)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "批量查询商品失败: %v", result.Error)
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

// CreateGoods 新建商品
func (s *GoodsServer) CreateGoods(ctx context.Context, req *proto.CreateGoodsInfo) (*proto.GoodsInfoResponse, error) {
	g := Goods{
		Name:            req.Name,
		GoodsSn:         req.GoodsSn,
		CategoryID:      req.CategoryId,
		BrandID:         req.BrandId,
		OnSale:          req.OnSale,
		ShipFree:        req.ShipFree,
		IsNew:           req.IsNew,
		IsHot:           req.IsHot,
		MarketPrice:     req.MarketPrice,
		ShopPrice:       req.ShopPrice,
		GoodsBrief:      req.GoodsBrief,
		Images:          basemodel.GormList(req.Images),
		DescImages:      basemodel.GormList(req.DescImages),
		GoodsFrontImage: req.GoodsFrontImage,
	}
	if err := s.db.Create(&g).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "创建商品失败: %v", err)
	}
	// 回填关联对象，让返回的 Response 带出 Category 和 Brand 信息
	var category Category
	var brand Brands
	s.db.First(&category, g.CategoryID)
	s.db.First(&brand, g.BrandID)
	g.Category = &category
	g.Brand = &brand
	// 同步写入 ES 索引（失败不阻断主流程，下次启动重建会修正）
	_ = s.es.IndexGoods(ctx, goodsToDoc(&g))
	return GoodsModelToResponse(&g), nil
}

// GetGoodsDetail 通过 id 查询商品（含分类和品牌）
func (s *GoodsServer) GetGoodsDetail(ctx context.Context, req *proto.GoodInfoRequest) (*proto.GoodsInfoResponse, error) {
	var g Goods
	result := s.db.Preload("Category").Preload("Brand").First(&g, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询商品失败: %v", result.Error)
	}
	return GoodsModelToResponse(&g), nil
}

// UpdateGoods 更新商品
func (s *GoodsServer) UpdateGoods(ctx context.Context, req *proto.CreateGoodsInfo) (*emptypb.Empty, error) {
	updates := map[string]any{
		"name":              req.Name,
		"goods_sn":          req.GoodsSn,
		"category_id":       req.CategoryId,
		"brands_id":         req.BrandId,
		"on_sale":           req.OnSale,
		"ship_free":         req.ShipFree,
		"is_new":            req.IsNew,
		"is_hot":            req.IsHot,
		"market_price":      req.MarketPrice,
		"shop_price":        req.ShopPrice,
		"goods_brief":       req.GoodsBrief,
		"goods_front_image": req.GoodsFrontImage,
		"images":            basemodel.GormList(req.Images),
		"desc_images":       basemodel.GormList(req.DescImages),
	}
	result := s.db.Model(&Goods{}).Where("id = ?", req.Id).Updates(updates)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "更新商品失败: %v", result.Error)
	}
	// 同步覆盖 ES 文档：req 不含 click_num/sold_num/fav_num 等服务端计数字段，
	// 故从 MySQL 重取整条，避免用 req 构造时把计数清零。
	var g Goods
	if err := s.db.First(&g, req.Id).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "重取商品失败: %v", err)
	}
	_ = s.es.IndexGoods(ctx, goodsToDoc(&g))
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
	// 从 ES 索引移除，避免搜到已软删商品
	_ = s.es.DeleteGoodsDoc(ctx, req.Id)
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
