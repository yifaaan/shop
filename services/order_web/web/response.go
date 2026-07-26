package web

import "shop/pkg/proto"

// cartItemResponse 购物车项的 JSON 响应，字段按前端约定使用 snake_case。
// 商品名/图/价由 web 层调用 goods_srv.BatchGetGoods 补全（购物车表只存 goodsId）。
type cartItemResponse struct {
	Id          int32   `json:"id"`
	GoodsId     int32   `json:"goods_id"`
	GoodsName   string  `json:"goods_name"`
	GoodsImage  string  `json:"goods_image"`
	GoodsPrice  float32 `json:"goods_price"`
	Num         int32   `json:"num"`
	Checked     bool    `json:"checked"`
}

// cartItemListResponse 购物车列表响应。
type cartItemListResponse struct {
	Total int32               `json:"total"`
	Data  []*cartItemResponse `json:"data"`
}

// cartItemToResponse 合并购物车项与商品信息（可为 nil）生成前端响应。
func cartItemToResponse(c *proto.CartItemInfoResponse, goods *proto.GoodsInfoResponse) *cartItemResponse {
	if c == nil {
		return nil
	}
	rsp := &cartItemResponse{
		Id:      c.Id,
		GoodsId: c.GoodsId,
		Num:     c.Num,
		Checked: c.Checked,
	}
	if goods != nil {
		rsp.GoodsName = goods.Name
		rsp.GoodsImage = goods.GoodsFrontImage
		rsp.GoodsPrice = goods.ShopPrice
	}
	return rsp
}

// orderItemResponse 订单商品快照的 snake_case JSON。
type orderItemResponse struct {
	Id         int32   `json:"id"`
	OrderId    int32   `json:"order_id"`
	GoodsId    int32   `json:"goods_id"`
	GoodsName  string  `json:"goods_name"`
	GoodsImage string  `json:"goods_image"`
	GoodsPrice float32 `json:"goods_price"`
	Num        int32   `json:"num"`
}

// orderResponse 订单的 snake_case JSON。
type orderResponse struct {
	Id         int32                `json:"id"`
	OrderSn    string               `json:"order_sn"`
	UserId     int32                `json:"user_id"`
	Address    string               `json:"address"`
	Name       string               `json:"name"`
	Mobile     string               `json:"mobile"`
	Post       string               `json:"post"`
	Status     int32                `json:"status"`
	PayType    int32                `json:"pay_type"`
	Total      float32              `json:"total"`
	PostFee    float32              `json:"post_fee"`
	AddTime    int64                `json:"add_time"`
	OrderGoods []*orderItemResponse `json:"order_goods"`
	// AlipayUrl 仅在订单详情接口生成支付宝支付链接时填充，列表接口留空省略。
	AlipayUrl  string               `json:"alipay_url,omitempty"`
}

// orderListResponse 订单列表的 snake_case JSON。
type orderListResponse struct {
	Total int32            `json:"total"`
	Data  []*orderResponse `json:"data"`
}

func orderItemToResponse(item *proto.OrderItemInfoResponse) *orderItemResponse {
	if item == nil {
		return nil
	}
	return &orderItemResponse{
		Id:         item.Id,
		OrderId:    item.OrderId,
		GoodsId:    item.GoodsId,
		GoodsName:  item.GoodsName,
		GoodsImage: item.GoodsImage,
		GoodsPrice: item.GoodsPrice,
		Num:        item.Num,
	}
}

func orderToResponse(o *proto.OrderInfoResponse) *orderResponse {
	if o == nil {
		return nil
	}
	rsp := &orderResponse{
		Id:         o.Id,
		OrderSn:    o.OrderSn,
		UserId:     o.UserId,
		Address:    o.Address,
		Name:       o.Name,
		Mobile:     o.Mobile,
		Post:       o.Post,
		Status:     o.Status,
		PayType:    o.PayType,
		Total:      o.Total,
		PostFee:    o.PostFee,
		AddTime:    o.AddTime,
		OrderGoods: make([]*orderItemResponse, 0, len(o.OrderGoods)),
	}
	for _, g := range o.OrderGoods {
		rsp.OrderGoods = append(rsp.OrderGoods, orderItemToResponse(g))
	}
	return rsp
}
