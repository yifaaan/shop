package handler

import (
	"context"
	"shop/order_srv/global"
	"shop/order_srv/model"
	"shop/order_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
func (s *OrderServer) UpdateCartItem(ctx context.Context, in *proto.CartItemRequest) (*emptypb.Empty, error) {
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
	return &emptypb.Empty{}, nil
}

// DeleteCartItem 删除购物车商品
func (s *OrderServer) DeleteCartItem(ctx context.Context, in *proto.CartItemRequest) (*emptypb.Empty, error) {
	res := global.DB.WithContext(ctx).Delete(&model.ShoppingCart{}, in.Id)
	if res.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车商品不存在")
	}
	return &emptypb.Empty{}, nil
}

// Create 创建订单
func (s *OrderServer) CreateOrder(ctx context.Context, in *proto.OrderRequest) (*proto.OrderInfoResponse, error) {
	// 从购物车获取选中的商品
	var cartList []*model.ShoppingCart
	res := global.DB.WithContext(ctx).Where("user = ? AND checked = ?", in.UserId, true).Find(&cartList)
	if res.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车商品不存在")
	}

	// TODO:分布式锁
	// good-srv:批量查询商品价格
	var goodIds = make([]int32, 0, res.RowsAffected)
	var goodNumsMap = make(map[int32]int32, res.RowsAffected) // 商品ID -> 商品数量
	for _, item := range cartList {
		goodIds = append(goodIds, item.Good)
		goodNumsMap[item.Good] = item.Nums
	}
	goodList, err := global.GoodSrvClient.BatchGetGood(ctx, &proto.BatchGoodIdInfo{
		Id: goodIds,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询商品价格失败: %v", err)
	}

	var orderPrice float32                                         // 订单总价格
	var orderGoods = make([]*model.OrderGood, 0, res.RowsAffected) // 订单商品列表
	for _, good := range goodList.Data {
		orderPrice += good.ShopPrice * float32(goodNumsMap[good.Id])
		orderGoods = append(orderGoods, &model.OrderGood{
			Good:      good.Id,
			GoodImage: good.GoodFrontImage,
			Nums:      goodNumsMap[good.Id],
			GoodsName: good.Name,
			GoodPrice: good.ShopPrice,
		})
	}

	// inventory-srv:库存扣减
	goodNums := make([]*proto.GoodInvInfo, 0, res.RowsAffected)
	for _, good := range orderGoods {
		goodNums = append(goodNums, &proto.GoodInvInfo{
			GoodId: good.Good,
			Nums:   good.Nums,
		})
	}
	if _, err = global.InventorySrvClient.Sell(ctx, &proto.SellInfo{GoodInfos: goodNums}); err != nil {
		return nil, status.Errorf(codes.Internal, "库存扣减失败: %v", err)
	}
	// 构造订单商品表
	// 下面是本地事务，如果其中一步失败，则回滚
	tx := global.DB.WithContext(ctx).Begin()
	order := model.OrderInfo{
		User:         in.UserId,
		OrderSn:      generateOrderSn(in.UserId),
		OrderMount:   orderPrice,
		Address:      in.Address,
		SignerName:   in.Name,
		SignerMobile: in.Mobile,
		Post:         in.Post,
	}
	if res = tx.Save(&order); res.RowsAffected == 0 {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "创建订单失败: %v", err)
	}
	for _, good := range orderGoods {
		good.Order = order.ID
	}
	res = tx.CreateInBatches(orderGoods, 100) // 批量创建订单商品
	if res.Error != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "创建订单商品失败: %v", res.Error)
	}
	// 从购物车删除已购买的商品
	if err = tx.Where("user = ? AND checked = ?", in.UserId, true).Delete(&model.ShoppingCart{}).Error; err != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "删除购物车商品失败: %v", err)
	}
	tx.Commit()
	return &proto.OrderInfoResponse{
		Id:      order.ID,
		UserId:  order.User,
		OrderSn: order.OrderSn,
		Total:   order.OrderMount,
	}, nil
}

// OrderList 获取订单列表
func (s *OrderServer) OrderList(ctx context.Context, in *proto.OrderFilterRequest) (*proto.OrderListResponse, error) {
	var orderList []*model.OrderInfo
	res := global.DB.WithContext(ctx).Where("user = ?", in.UserId).Scopes(paginate(int(in.Pages), int(in.PagePerNums))).Find(&orderList)
	if res.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单列表不存在")
	}
	var orderInfos = make([]*proto.OrderInfoResponse, 0, res.RowsAffected)
	for _, item := range orderList {
		orderInfos = append(orderInfos, &proto.OrderInfoResponse{
			Id:      item.ID,
			UserId:  item.User,
			OrderSn: item.OrderSn,
			Total:   item.OrderMount,
			Address: item.Address,
			Name:    item.SignerName,
			Mobile:  item.SignerMobile,
		})
	}
	return &proto.OrderListResponse{
		Total: int32(res.RowsAffected),
		Data:  orderInfos,
	}, nil
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
func (s *OrderServer) UpdateOrderStatus(ctx context.Context, in *proto.OrderStatus) (*emptypb.Empty, error) {
	res := global.DB.WithContext(ctx).Where("order_sn = ?", in.OrderSn).Updates(&model.OrderInfo{Status: in.Status})
	if res.RowsAffected == 0 {
		return nil, status.Errorf(codes.Internal, "更新订单状态失败: %v", res.Error)
	}
	return &emptypb.Empty{}, nil
}
