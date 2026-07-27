package web

import "shop/pkg/proto"

// userFavResponse 收藏的 snake_case JSON。
// 商品名/图/价由 web 层调用 goods_srv.BatchGetGoods 补全（收藏表只存 goodsId）。
type userFavResponse struct {
	Id         int32   `json:"id"`
	UserId     int32   `json:"user_id"`
	GoodsId    int32   `json:"goods_id"`
	GoodsName  string  `json:"goods_name"`
	GoodsImage string  `json:"goods_image"`
	GoodsPrice float32 `json:"goods_price"`
}

// userFavListResponse 收藏列表响应。
type userFavListResponse struct {
	Total int32              `json:"total"`
	Data  []*userFavResponse `json:"data"`
}

// addressResponse 收货地址的 snake_case JSON。
type addressResponse struct {
	Id           int32  `json:"id"`
	UserId       int32  `json:"user_id"`
	Province     string `json:"province"`
	City         string `json:"city"`
	District     string `json:"district"`
	Address      string `json:"address"`
	SignerName   string `json:"signer_name"`
	SignerMobile string `json:"signer_mobile"`
}

// addressListResponse 地址列表响应。
type addressListResponse struct {
	Total int32              `json:"total"`
	Data  []*addressResponse `json:"data"`
}

// messageResponse 留言的 snake_case JSON。
type messageResponse struct {
	Id      int32  `json:"id"`
	UserId  int32  `json:"user_id"`
	Subject string `json:"subject"`
	Message string `json:"message"`
	Type    int32  `json:"type"`
	File    string `json:"file"`
	AddTime int64  `json:"add_time"`
}

// messageListResponse 留言列表响应。
type messageListResponse struct {
	Total int32              `json:"total"`
	Data  []*messageResponse `json:"data"`
}

func userFavToResponse(f *proto.UserFavInfoResponse, goods *proto.GoodsInfoResponse) *userFavResponse {
	if f == nil {
		return nil
	}
	rsp := &userFavResponse{
		Id:      f.Id,
		UserId:  f.UserId,
		GoodsId: f.GoodsId,
	}
	if goods != nil {
		rsp.GoodsName = goods.Name
		rsp.GoodsImage = goods.GoodsFrontImage
		rsp.GoodsPrice = goods.ShopPrice
	}
	return rsp
}

func addressToResponse(a *proto.AddressInfoResponse) *addressResponse {
	if a == nil {
		return nil
	}
	return &addressResponse{
		Id:           a.Id,
		UserId:       a.UserId,
		Province:     a.Province,
		City:         a.City,
		District:     a.District,
		Address:      a.Address,
		SignerName:   a.SignerName,
		SignerMobile: a.SignerMobile,
	}
}

func messageToResponse(m *proto.MessageInfoResponse) *messageResponse {
	if m == nil {
		return nil
	}
	return &messageResponse{
		Id:      m.Id,
		UserId:  m.UserId,
		Subject: m.Subject,
		Message: m.Message,
		Type:    m.Type,
		File:    m.File,
		AddTime: m.AddTime,
	}
}