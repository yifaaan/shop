package web

// CreateGoodsForm 新建商品表单
type CreateGoodsForm struct {
	Name        string   `form:"name" json:"name" binding:"required,min=1,max=100"`
	GoodsSn     string   `form:"goods_sn" json:"goods_sn" binding:"required,min=1,max=50"`
	Stocks      int32    `form:"stocks" json:"stocks"`
	MarketPrice float32  `form:"market_price" json:"market_price" binding:"required"`
	ShopPrice   float32  `form:"shop_price" json:"shop_price" binding:"required"`
	GoodsBrief  string   `form:"goods_brief" json:"goods_brief" binding:"max=200"`
	ShipFree    bool     `form:"ship_free" json:"ship_free"`
	Images      []string `form:"images" json:"images"`
	DescImages  []string `form:"desc_images" json:"desc_images"`
	FrontImage  string   `form:"front_image" json:"front_image" binding:"required"`
	IsNew       bool     `form:"is_new" json:"is_new"`
	IsHot       bool     `form:"is_hot" json:"is_hot"`
	OnSale      bool     `form:"on_sale" json:"on_sale"`
	Category    int32    `form:"category" json:"category" binding:"required"`
	Brand       int32    `form:"brand" json:"brand" binding:"required"`
}

// UpdateGoodsForm 更新商品表单（与 Create 相同，但 Id 必填）
type UpdateGoodsForm struct {
	Id          int32    `form:"id" json:"id" binding:"required"`
	Name        string   `form:"name" json:"name"`
	GoodsSn     string   `form:"goods_sn" json:"goods_sn"`
	Stocks      int32    `form:"stocks" json:"stocks"`
	MarketPrice float32  `form:"market_price" json:"market_price"`
	ShopPrice   float32  `form:"shop_price" json:"shop_price"`
	GoodsBrief  string   `form:"goods_brief" json:"goods_brief"`
	ShipFree    bool     `form:"ship_free" json:"ship_free"`
	Images      []string `form:"images" json:"images"`
	DescImages  []string `form:"desc_images" json:"desc_images"`
	FrontImage  string   `form:"front_image" json:"front_image"`
	IsNew       bool     `form:"is_new" json:"is_new"`
	IsHot       bool     `form:"is_hot" json:"is_hot"`
	OnSale      bool     `form:"on_sale" json:"on_sale"`
	Category    int32    `form:"category" json:"category"`
	Brand       int32    `form:"brand" json:"brand"`
}

// CreateBrandForm 新建品牌表单
type CreateBrandForm struct {
	Name string `form:"name" json:"name" binding:"required,min=1,max=50"`
	Logo string `form:"logo" json:"logo"`
}

// UpdateBrandForm 更新品牌表单
type UpdateBrandForm struct {
	Id   int32  `form:"id" json:"id" binding:"required"`
	Name string `form:"name" json:"name"`
	Logo string `form:"logo" json:"logo"`
}

// CreateBannerForm 新建轮播图表单
type CreateBannerForm struct {
	Image string `form:"image" json:"image" binding:"required"`
	Url   string `form:"url" json:"url" binding:"required"`
	Index int32  `form:"index" json:"index"`
}

// UpdateBannerForm 更新轮播图表单
type UpdateBannerForm struct {
	Id    int32  `form:"id" json:"id" binding:"required"`
	Image string `form:"image" json:"image"`
	Url   string `form:"url" json:"url"`
	Index int32  `form:"index" json:"index"`
}

// CreateCategoryForm 新建分类表单
type CreateCategoryForm struct {
	Name           string `form:"name" json:"name" binding:"required,min=1,max=20"`
	ParentCategory int32  `form:"parent_category" json:"parent_category"`
	Level          int32  `form:"level" json:"level" binding:"required,oneof=1 2 3"`
	IsTab          bool   `form:"is_tab" json:"is_tab"`
}

// UpdateCategoryForm 更新分类表单
type UpdateCategoryForm struct {
	Id    int32  `form:"id" json:"id" binding:"required"`
	Name  string `form:"name" json:"name"`
	IsTab bool   `form:"is_tab" json:"is_tab"`
}

// CreateCategoryBrandForm 新建分类品牌关系表单
type CreateCategoryBrandForm struct {
	Category int32 `form:"category" json:"category" binding:"required"`
	Brand    int32 `form:"brand" json:"brand" binding:"required"`
}

// UpdateCategoryBrandForm 更新分类品牌关系表单
type UpdateCategoryBrandForm struct {
	Id       int32 `form:"id" json:"id" binding:"required"`
	Category int32 `form:"category" json:"category"`
	Brand    int32 `form:"brand" json:"brand"`
}
