package web

import (
	"shop/pkg/proto"
)

// goodsResponse 是商品列表/详情的 JSON 响应结构，字段按前端约定使用 snake_case。
type goodsResponse struct {
	Id          int32              `json:"id"`
	Name        string             `json:"name"`         // 商品名（对应 proto.Name，详情页需要原始字段）
	Title       string             `json:"title"`       // 商品名（对应 proto.Name，与 name 同源，列表页用）
	Description string             `json:"description"` // 商品简短描述（对应 proto.GoodsBrief）
	GoodsBrief  string             `json:"goods_brief"`
	Desc        string             `json:"desc"`        // 商品详细描述（对应 proto.GoodsDesc）
	ShopPrice   float32            `json:"shop_price"`  // 本店价
	ShipFree    bool               `json:"ship_free"`   // 是否包邮
	Images      []string           `json:"images"`
	DescImages  []string           `json:"desc_images"`
	FrontImage  string             `json:"front_image"` // 商品封面图（对应 proto.GoodsFrontImage）
	IsHot       bool               `json:"is_hot"`
	IsNew       bool               `json:"is_new"`
	IsTab       bool               `json:"is_tab"`      // proto.GoodsInfoResponse 无此字段，预留输出 false
	Brand       *brandResponse     `json:"brand"`
	Category    *categoryResponse  `json:"category"`
}

type brandResponse struct {
	Id          int32  `json:"id"`
	Title       string `json:"title"`        // 品牌名（对应 proto.Brand.Name）
	Description string `json:"description"` // proto 未提供，预留
	Logo        string `json:"logo"`
}

type categoryResponse struct {
	Id          int32  `json:"id"`
	Title       string `json:"title"`        // 分类名（对应 proto.Category.Name）
	Description string `json:"description"` // proto 未提供，预留
}

// categoryInfoResponse 是单分类的 JSON 响应结构（CreateCategory 等单查场景）。
type categoryInfoResponse struct {
	Id           int32  `json:"id"`
	Name         string `json:"name"`
	Parent       int32  `json:"parent"`
	Level        int32  `json:"level"`
	IsTab        bool   `json:"is_tab"`
}

// subCategoryListResponse 是 GetSubCategory 的 JSON 响应结构，字段用 snake_case。
type subCategoryListResponse struct {
	Total            int32                  `json:"total"`
	Info             *categoryInfoResponse  `json:"info"`
	SubCategories    []*categoryInfoResponse `json:"sub_categories"`
}

// CategoryModelToResponse 把 proto.CategoryInfoResponse 转成前端约定的 JSON 结构。
func CategoryModelToResponse(c *proto.CategoryInfoResponse) *categoryInfoResponse {
	if c == nil {
		return nil
	}
	return &categoryInfoResponse{
		Id:    c.Id,
		Name:  c.Name,
		Parent: c.ParentCategory,
		Level: c.Level,
		IsTab: c.IsTab,
	}
}

// SubCategoryListToResponse 把 proto.SubCategoryListResponse 转成前端约定的 JSON 结构。
func SubCategoryListToResponse(r *proto.SubCategoryListResponse) *subCategoryListResponse {
	if r == nil {
		return nil
	}
	rsp := &subCategoryListResponse{
		Total: r.Total,
		Info:  CategoryModelToResponse(r.Info),
		SubCategories: make([]*categoryInfoResponse, 0, len(r.SubCategorys)),
	}
	for i := range r.SubCategorys {
		rsp.SubCategories = append(rsp.SubCategories, CategoryModelToResponse(r.SubCategorys[i]))
	}
	return rsp
}

// goodsListResponse 是商品列表的分页响应包装。
type goodsListResponse struct {
	Total int32           `json:"total"`
	Data  []*goodsResponse `json:"data"`
}

// GoodsModelToResponse 把 proto.GoodsInfoResponse 转成前端约定的 JSON 结构。
func GoodsModelToResponse(g *proto.GoodsInfoResponse) *goodsResponse {
	if g == nil {
		return nil
	}
	rsp := &goodsResponse{
		Id:          g.Id,
		Name:        g.Name,
		Title:       g.Name,
		Description: g.GoodsBrief,
		GoodsBrief:  g.GoodsBrief,
		Desc:        g.GoodsDesc,
		ShopPrice:   g.ShopPrice,
		ShipFree:    g.ShipFree,
		Images:      g.Images,
		DescImages:  g.DescImages,
		FrontImage:  g.GoodsFrontImage,
		IsHot:       g.IsHot,
		IsNew:       g.IsNew,
	}
	if g.Brand != nil {
		rsp.Brand = &brandResponse{
			Id:          g.Brand.Id,
			Title:       g.Brand.Name,
			Description: "",
			Logo:        g.Brand.Logo,
		}
	} else {
		rsp.Brand = &brandResponse{}
	}
	if g.Category != nil {
		rsp.Category = &categoryResponse{
			Id:          g.Category.Id,
			Title:       g.Category.Name,
			Description: "",
		}
	} else {
		rsp.Category = &categoryResponse{}
	}
	return rsp
}