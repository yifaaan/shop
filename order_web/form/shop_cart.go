package form

type ShopCartItemForm struct {
	GoodId int32 `json:"good" form:"good" binding:"required"`
	Nums   int32 `json:"nums" form:"nums" binding:"required,min=1"`
}
