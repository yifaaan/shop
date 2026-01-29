package model

// 商品分类表
type Category struct {
	BaseModel
	Name             string    `gorm:"type:varchar(20);not null"`
	ParentCategoryID int32     `gorm:"type:int;null;default:null;comment:'父分类ID'"`
	ParentCateGory   *Category `gorm:"foreignKey:ParentCategoryID;references:ID"`         // 自关联
	Level            int32     `gorm:"type:int;not null;default:1;comment:'分类级别'"`        // 分类级别
	IsTab            bool      `gorm:"type:bool;not null;default:false;comment:'是否是导航栏'"` // 是否是导航栏
}

// 商品品牌表
type Brand struct {
	BaseModel
	Name string `gorm:"type:varchar(50);not null comment '品牌名称'"`
	Logo string `gorm:"type:varchar(200);not null;default:'' comment '品牌logo'"`
}

// 一个品牌对应多个商品分类, 一个商品分类对应多个商品品牌 多对多关系
type GoodCategoryBrand struct {
	BaseModel
	CategoryID int32    `gorm:"type:int;not null;index:idx_category_brand,unique"`
	Category   Category `gorm:"foreignKey:CategoryID;references:ID"`
	BrandID    int32    `gorm:"type:int;not null;index:idx_category_brand,unique"`
	Brand      Brand    `gorm:"foreignKey:BrandID;references:ID"`
}

// 轮播图表
type Banner struct {
	BaseModel
	Image string `gorm:"type:varchar(200);not null comment '轮播图图片'"`
	Url   string `gorm:"type:varchar(200);not null comment '轮播图跳转链接'"`
	Index int32  `gorm:"type:int;not null;default:0;comment:'轮播图顺序'"`
}

type Good struct {
	BaseModel
	Name   string `gorm:"type:varchar(100);not null comment '商品名称'"`
	GoodSn string `gorm:"type:varchar(50);not null comment '商品唯一货号, 商家自定义'"`

	CategoryID int32    `gorm:"type:int;not null"`
	Category   Category `gorm:"foreignKey:CategoryID;references:ID"`
	BrandID    int32    `gorm:"type:int;not null"`
	Brand      Brand    `gorm:"foreignKey:BrandID;references:ID"`

	OnSale   bool `gorm:"type:bool;not null;default:false;comment:'是否上架'"`
	ShipFree bool `gorm:"type:bool;not null;default:false;comment:'是否包邮'"`
	IsNew    bool `gorm:"type:bool;not null;default:false;comment:'是否新品'"`
	IsHot    bool `gorm:"type:bool;not null;default:false;comment:'是否热销'"`

	ClickNum    int32   `gorm:"type:int;not null;default:0;comment:'点击数'"`
	SoldNum     int32   `gorm:"type:int;not null;default:0;comment:'商品销售量'"`
	FavNum      int32   `gorm:"type:int;not null;default:0;comment:'商品收藏数'"`
	MarketPrice float32 `gorm:"type:float;not null comment '市场价'"`
	ShopPrice   float32 `gorm:"type:float;not null comment '本店价'"`

	GoodBrief string `gorm:"type:varchar(100);not null default:'' comment '商品简短描述'"`

	GoodFrontImage string   `gorm:"type:varchar(1024);not null comment '商品封面图'"` // 商品封面图
	Images         GormList `gorm:"type:varchar(4096);not null comment '商品轮播图'"` // 商品轮播图
	DescImages     GormList `gorm:"type:varchar(4096);not null comment '商品详情图'"` // 商品详情图
}
