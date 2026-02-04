package form

type GoodForm struct {
	Name        string   `form:"name" json:"name" binding:"required,min=2,max=100"`
	GoodSn      string   `form:"goodSn" json:"goodSn" binding:"required,min=2,max=100"`
	Stocks      int32    `form:"stocks" json:"stocks" binding:"required,min=1"`
	CategoryId  int32    `form:"category" json:"category" binding:"required,min=1"`
	BrandId     int32    `form:"brand" json:"brand" binding:"required,min=1"`
	MarketPrice float32  `form:"marketrice" json:"marketrice" binding:"required,min=0"`
	ShopPrice   float32  `form:"shoprice" json:"shoprice" binding:"required,min=0"`
	GoodBrief   string   `form:"good_brief" json:"good_brief" binding:"required,min=2,max=100"`
	Images      []string `form:"images" json:"images" binding:"required"`
	DescImages  []string `form:"desc_images" json:"desc_images" binding:"required"`
	GoodDesc    string   `form:"desc" json:"desc" binding:"required,min=2"`
	ShipFree    bool     `form:"ship_free" json:"ship_free" binding:"required"`
	FrontImage  string   `form:"front_image" json:"front_image" binding:"required"`
}

type GoodStatusForm struct {
	IsNew  bool `form:"new" json:"new" binding:"required"`
	IsHot  bool `form:"hot" json:"hot" binding:"required"`
	OnSale bool `form:"sale" json:"sale" binding:"required"`
}
