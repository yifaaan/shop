package model

import "time"

type ShoppingCart struct {
	BaseModel
	User    int32 `gorm:"type:int;not null;index:idx_user,unique;comment:'用户ID'"`
	Good    int32 `gorm:"type:int;not null;index:idx_good,unique;comment:'商品ID'"`
	Nums    int32 `gorm:"type:int;not null;comment:'数量'"`
	Checked bool  `gorm:"type:bool;not null;default:true;comment:'是否选中'"`
}

func (ShoppingCart) TableName() string {
	return "shoppingcart"
}

type OrderInfo struct {
	BaseModel
	User         int32     `gorm:"type:int;not null;index:idx_user,unique;comment:'用户ID'"`
	OrderSn      string    `gorm:"type:varchar(30);not null;index:idx_order_sn,unique;comment:'订单编号'"`
	PayType      string    `gorm:"type:varchar(20);not null;comment:'支付方式'"`
	Status       string    `gorm:"type:varchar(20);not null;comment:'订单状态'"`
	TradeNo      string    `gorm:"type:varchar(100);not null;comment:'交易号'"`
	OrderMount   float32   `gorm:"type:float;not null;comment:'订单金额'"`
	PayTime      time.Time `gorm:"type:datetime;not null;comment:'支付时间'"`
	Address      string    `gorm:"type:varchar(100);not null;comment:'收货地址'"`
	SignerName   string    `gorm:"type:varchar(20);not null;comment:'收货人姓名'"`
	SignerMobile string    `gorm:"type:varchar(11);not null;comment:'收货人手机号'"`
	Post         string    `gorm:"type:varchar(100);not null;comment:'用户留言'"`
}

func (OrderInfo) TableName() string {
	return "orderinfo"
}

type OrderGood struct {
	BaseModel
	Order     int32   `gorm:"type:int;not null;index:idx_order,unique;comment:'订单ID'"`
	Good      int32   `gorm:"type:int;not null;index:idx_good,unique;comment:'商品ID'"`
	Nums      int32   `gorm:"type:int;not null;comment:'数量'"`
	GoodsName string  `gorm:"type:varchar(100);not null;comment:'商品名称'"`
	GoodPrice float32 `gorm:"type:float;not null;comment:'商品价格'"`
}
