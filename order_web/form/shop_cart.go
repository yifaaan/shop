package form

type ShopCartItemForm struct {
	GoodsId int32 `json:"goods_id" form:"goods_id" binding:"required"`
	Nums    int32 `json:"nums" form:"nums" binding:"required,min=1"`
}
