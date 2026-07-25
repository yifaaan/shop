package web

// AddCartItemForm 添加商品到购物车
type AddCartItemForm struct {
	GoodsId int32 `form:"goods_id" json:"goods_id" binding:"required"`
	Num     int32 `form:"num" json:"num" binding:"required,min=1"`
	Checked bool  `form:"checked" json:"checked"`
}

// UpdateCartItemForm 更新购物车商品（数量 / 选中状态）
type UpdateCartItemForm struct {
	Num     int32 `form:"num" json:"num" binding:"required,min=1"`
	Checked bool  `form:"checked" json:"checked"`
}

// CreateOrderForm 创建订单（商品来自购物车已选项，故只填收货信息）
type CreateOrderForm struct {
	Address string  `form:"address" json:"address" binding:"required"`
	Name    string  `form:"name" json:"name" binding:"required"`
	Mobile  string  `form:"mobile" json:"mobile" binding:"required,mobile"`
	Post    string  `form:"post" json:"post"`
	PayType int32   `form:"pay_type" json:"pay_type" binding:"required,oneof=1 2"`
	PostFee float32 `form:"post_fee" json:"post_fee"`
}

// UpdateOrderStatusForm 更新订单状态（按 orderSn）
type UpdateOrderStatusForm struct {
	OrderSn string `form:"order_sn" json:"order_sn" binding:"required"`
	Status  int32  `form:"status" json:"status" binding:"required,oneof=1 2 3 4 5"`
}
