package handler

import (
	"context"
	"shop/order_srv/global"
	"shop/order_srv/model"
	"shop/order_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderServer struct {
	proto.UnimplementedOrderServer
}

var _ proto.OrderServer = (*OrderServer)(nil)

// CartItemList 获取购物车列表
func (s *OrderServer) CartItemList(ctx context.Context, in *proto.UserInfo) (*proto.CartItemListResponse, error) {
	var cartList []*model.ShoppingCart
	// 查询购物车列表
	res := global.DB.WithContext(ctx).Where("user = ?", in.Id).Find(&cartList)
	if res.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车列表不存在")
	}
	if res.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询购物车列表失败: %v", res.Error)
	}

	data := make([]*proto.ShopCartInfoResponse, 0, res.RowsAffected)
	for _, item := range cartList {
		data = append(data, &proto.ShopCartInfoResponse{
			Id:      item.ID,
			UserId:  item.User,
			GoodsId: item.Good,
			Nums:    item.Nums,
			Checked: item.Checked,
		})
	}
	return &proto.CartItemListResponse{
		Total: int32(res.RowsAffected),
		Data:  data,
	}, nil
}

// CreateCartItem 将商品添加到购物车
func (s *OrderServer) CreateCartItem(ctx context.Context, in *proto.CartItemRequest) (*proto.ShopCartInfoResponse, error) {
	var cart model.ShoppingCart
	res := global.DB.WithContext(ctx).Where("user = ? AND good = ?", in.UserId, in.GoodsId).First(&cart)

	if res.RowsAffected == 0 {
		// 如果商品不存在，则创建新购物车项
		cart = model.ShoppingCart{
			User:    in.UserId,
			Good:    in.GoodsId,
			Nums:    in.Nums,
			Checked: in.Checked,
		}
		res = global.DB.WithContext(ctx).Save(&cart)
		if res.Error != nil {
			return nil, status.Errorf(codes.Internal, "创建购物车项失败: %v", res.Error)
		}
	} else {
		// 如果商品已存在，则更新数量
		cart.Nums += in.Nums
		res = global.DB.WithContext(ctx).Save(&cart)
		if res.Error != nil {
			return nil, status.Errorf(codes.Internal, "更新购物车项失败: %v", res.Error)
		}
	}
	return &proto.ShopCartInfoResponse{
		Id:      cart.ID,
		UserId:  cart.User,
		GoodsId: cart.Good,
		Nums:    cart.Nums,
		Checked: cart.Checked,
	}, nil
}

// UpdateCartItem 更新购物车,包括商品数量和选中状态
func (s *OrderServer) UpdateCartItem(ctx context.Context, in *proto.CartItemRequest) (*proto.Empty, error) {
	var cart model.ShoppingCart
	res := global.DB.WithContext(ctx).First(&cart, in.Id)
	if res.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车商品不存在")
	}

	if in.Nums > 0 {
		cart.Nums = in.Nums
	}
	if in.Checked {
		cart.Checked = in.Checked
	}
	res = global.DB.WithContext(ctx).Save(&cart)
	if res.RowsAffected == 0 {
		return nil, status.Errorf(codes.Internal, "更新购物车商品失败: %v", res.Error)
	}
	return &proto.Empty{}, nil
}

// DeleteCartItem 删除购物车商品
func (s *OrderServer) DeleteCartItem(ctx context.Context, in *proto.CartItemRequest) (*proto.Empty, error) {
	res := global.DB.WithContext(ctx).Delete(&model.ShoppingCart{}, in.Id)
	if res.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车商品不存在")
	}
	return &proto.Empty{}, nil
}

// Create 创建订单
func (s *OrderServer) Create(ctx context.Context, in *proto.OrderRequest) (*proto.OrderInfoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Create not implemented")
}

// OrderList 获取订单列表
func (s *OrderServer) OrderList(ctx context.Context, in *proto.OrderFilterRequest) (*proto.OrderListResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method OrderList not implemented")
}

// OrderDetail 获取订单详情
func (s *OrderServer) OrderDetail(ctx context.Context, in *proto.OrderRequest) (*proto.OrderInfoDetailResponse, error) {
	var orderInfo model.OrderInfo
	res := global.DB.WithContext(ctx).First(&orderInfo, in.Id)
	if res.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}

	var resp proto.OrderInfoDetailResponse
	resp.OrderInfo = &proto.OrderInfoResponse{
		Id:      orderInfo.ID,
		UserId:  orderInfo.User,
		OrderSn: orderInfo.OrderSn,
		PayType: orderInfo.PayType,
		Status:  orderInfo.Status,
		Post:    orderInfo.Post,
		Total:   orderInfo.OrderMount,
		Address: orderInfo.Address,
		Name:    orderInfo.SignerName,
		Mobile:  orderInfo.SignerMobile,
	}

	// 查询订单的商品列表
	var goodList []*model.OrderGood
	global.DB.WithContext(ctx).Where("order = ?", orderInfo.ID).Find(&goodList)
	for _, g := range goodList {
		resp.Goods = append(resp.Goods, &proto.OrderItemResponse{
			Id:         g.ID,
			OrderId:    g.Order,
			GoodsId:    g.Good,
			Nums:       g.Nums,
			GoodsName:  g.GoodsName,
			GoodsPrice: g.GoodPrice,
		})
	}
	return &resp, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderServer) UpdateOrderStatus(ctx context.Context, in *proto.OrderStatus) (*proto.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateOrderStatus not implemented")
}
