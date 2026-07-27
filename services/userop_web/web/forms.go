package web

// UserFavForm 收藏/取消收藏：以 goodsId 标识（(userId,goodsId) 唯一）。
type UserFavForm struct {
	GoodsId int32 `form:"goods_id" json:"goods_id" binding:"required"`
}

// AddressForm 新建/更新收货地址（更新时 id 由 URL 路径参数提供）。
type AddressForm struct {
	Province     string `form:"province" json:"province" binding:"required"`
	City         string `form:"city" json:"city" binding:"required"`
	District     string `form:"district" json:"district" binding:"required"`
	Address      string `form:"address" json:"address" binding:"required"`
	SignerName   string `form:"signer_name" json:"signer_name" binding:"required"`
	SignerMobile string `form:"signer_mobile" json:"signer_mobile" binding:"required,mobile"`
}

// CreateMessageForm 新建留言：type 1留言 2投诉 3询问 4售后 5求购。
type CreateMessageForm struct {
	Subject string `form:"subject" json:"subject" binding:"required"`
	Message string `form:"message" json:"message" binding:"required"`
	Type    int32  `form:"type" json:"type" binding:"required,oneof=1 2 3 4 5"`
	File    string `form:"file" json:"file"`
}