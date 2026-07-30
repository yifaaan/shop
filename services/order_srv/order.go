package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"

	rmq "github.com/apache/rocketmq-clients/golang/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// 订单状态
const (
	StatusPending       = 1 // 待支付
	StatusPaid          = 2 // 已支付
	StatusCancel        = 3 // 已取消/超时关闭
	StatusTradeCreated  = 4 // 交易创建
	StatusTradeFinished = 5 // 交易结束
)

// ShoppingCart 购物车表
type ShoppingCart struct {
	basemodel.BaseModel
	UserID  int32 `gorm:"column:user_id;type:int;not null;index"`
	GoodsID int32 `gorm:"column:goods_id;type:int;not null;index"`
	Num     int32 `gorm:"type:int;not null"`
	Checked bool  // 在购物车中是否选中
}

// OrderInfo 订单主表
type OrderInfo struct {
	basemodel.BaseModel
	OrderSn string `gorm:"column:order_sn;type:varchar(40);not null;uniqueIndex"` // 订单号
	UserID  int32  `gorm:"column:user_id;type:int;not null;index"`                // 下单用户ID

	Status  int32      `gorm:"type:int;not null;default:1;comment '1-待支付 2-已支付 3-已取消/超时关闭 4-交易创建 5-交易结束'"` // 1-待支付 2-已支付 3-已取消/超时关闭 4-交易创建 5-交易结束
	PayType int32      `gorm:"column:pay_type;type:int;not null;default:1; comment '1-微信 2-支付宝'"`          // 1-微信 2-支付宝
	TradeNo string     `gorm:"column:trade_no;type:varchar(100); comment '交易号/支付宝(微信)订单号'"`
	PayTime *time.Time `gorm:"column:pay_time;comment '支付时间，未支付为 NULL'"`
	Total   float32    `gorm:"type:float;not null"`                           // 商品总金额（不含运费）
	PostFee float32    `gorm:"column:post_fee;type:float;not null;default:0"` // 运费

	Address string `gorm:"type:varchar(100);not null"`            // 收货地址
	Name    string `gorm:"type:varchar(30);not null"`             // 收货人
	Mobile  string `gorm:"type:varchar(20);not null"`             // 联系电话
	Post    string `gorm:"type:varchar(100);not null;default:''"` // 留言
}

// OrderGoods 订单商品快照（下单时刻的商品信息）
type OrderGoods struct {
	basemodel.BaseModel
	OrderID    int32   `gorm:"column:order_id;type:int;not null;index"`
	GoodsID    int32   `gorm:"column:goods_id;type:int;not null;index"`
	GoodsName  string  `gorm:"column:goods_name;type:varchar(100);not null"`
	GoodsImage string  `gorm:"column:goods_image;type:varchar(200);not null"`
	GoodsPrice float32 `gorm:"column:goods_price;type:float;not null"`
	Num        int32   `gorm:"type:int;not null"`
}

// OrderRebackEvent 是"归还库存"事务消息的载体。
// 消费者（inventory_srv）收到消息后据此调 RebackDetail 归还库存。
type OrderRebackEvent struct {
	OrderSn    string       `json:"order_sn"`
	OrderGoods []RebackItem `json:"order_goods"`
}

type RebackItem struct {
	GoodsId int32 `json:"goods_id"`
	Num     int32 `json:"num"`
}

type transactionProducer interface {
	BeginTransaction() rmq.Transaction
	SendWithTransaction(ctx context.Context, msg *rmq.Message, tx rmq.Transaction) ([]*rmq.SendReceipt, error)
}

// OrderServer implements proto.OrderServer over an injected *gorm.DB,
// orchestrating goods_srv（查价/快照）与 inventory_srv（扣减/归还库存）。
type OrderServer struct {
	proto.UnimplementedOrderServer
	db         *gorm.DB
	goodsSrv   proto.GoodsClient
	invSrv     proto.InventoryClient
	log        *zap.SugaredLogger
	txProducer transactionProducer
}

// NewOrderServer wires an OrderServer to its data store and downstream services.
func NewOrderServer(db *gorm.DB, goodsSrv proto.GoodsClient, invSrv proto.InventoryClient, log *zap.SugaredLogger, txProducer transactionProducer) *OrderServer {
	return &OrderServer{db: db, goodsSrv: goodsSrv, invSrv: invSrv, log: log, txProducer: txProducer}
}

// CreateOrder 创建订单（编排跨服务下单流程）：
//  1. 从购物车取出该用户已选中(checked=true)的商品
//  2. 调 goods_srv.BatchGetGoods 拉取商品名/图/本店价，构建订单商品快照并计算总价
//  3. 发送"归还库存"事务半消息（补偿消息，先于扣减库存发出）
//  4. 调 inventory_srv.StockSellDetail 扣减库存
//  5. 本地事务：写订单主表 + 商品快照，并删除购物车中已购买的商品
//  6. 本地事务成功 → RollBack 半消息（订单已建，保持扣减）；失败 → Commit 半消息（消费者异步归还库存）
func (s *OrderServer) CreateOrder(ctx context.Context, req *proto.OrderInfoRequest) (*proto.OrderInfoResponse, error) {
	// 1. 取出购物车中已选中的商品
	var carts []ShoppingCart
	if err := s.db.Where("user_id = ? AND checked = ?", req.UserId, true).Find(&carts).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "查询购物车失败: %v", err)
	}
	if len(carts) == 0 {
		return nil, status.Error(codes.FailedPrecondition, "购物车中没有选中的商品")
	}

	// 2. 批量查询商品信息
	goodsIDs := make([]int32, 0, len(carts))
	for i := range carts {
		goodsIDs = append(goodsIDs, carts[i].GoodsID)
	}
	goodsRsp, err := s.goodsSrv.BatchGetGoods(ctx, &proto.BatchGoodsIdInfo{Id: goodsIDs})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询商品信息失败: %v", err)
	}
	goodsMap := make(map[int32]*proto.GoodsInfoResponse, len(goodsRsp.Data))
	for _, g := range goodsRsp.Data {
		goodsMap[g.Id] = g
	}

	// 3. 构建订单商品快照 + 总价 + 库存扣减明细
	var total float32
	items := make([]OrderGoods, 0, len(carts))
	sellDetails := make([]*proto.OrderGoodsDetail, 0, len(carts))
	rebackItems := make([]RebackItem, 0, len(carts))
	for i := range carts {
		g, ok := goodsMap[carts[i].GoodsID]
		if !ok {
			return nil, status.Errorf(codes.NotFound, "商品 %d 不存在或已下架", carts[i].GoodsID)
		}
		items = append(items, OrderGoods{
			GoodsID:    g.Id,
			GoodsName:  g.Name,
			GoodsImage: g.GoodsFrontImage,
			GoodsPrice: g.ShopPrice,
			Num:        carts[i].Num,
		})
		total += g.ShopPrice * float32(carts[i].Num)
		sellDetails = append(sellDetails, &proto.OrderGoodsDetail{
			GoodsId: g.Id,
			Num:     carts[i].Num,
		})
		rebackItems = append(rebackItems, RebackItem{
			GoodsId: g.Id,
			Num:     carts[i].Num,
		})
	}

	// 4. 构建订单（提前生成 orderSn 供半消息 keys 使用）
	order := OrderInfo{
		OrderSn: newOrderSn(req.UserId),
		UserID:  req.UserId,
		Address: req.Address,
		Name:    req.Name,
		Mobile:  req.Mobile,
		Post:    req.Post,
		Status:  StatusPending,
		PayType: req.PayType,
		Total:   total,
		PostFee: req.PostFee,
	}
	cartIDs := make([]int32, 0, len(carts))
	for i := range carts {
		cartIDs = append(cartIDs, carts[i].ID)
	}

	// 5. 先发送"归还库存"事务半消息（先于扣减库存）
	eventBody, _ := json.Marshal(OrderRebackEvent{
		OrderSn:    order.OrderSn,
		OrderGoods: rebackItems,
	})
	mqTx := s.txProducer.BeginTransaction()
	msg := &rmq.Message{
		Topic: orderRebackTopic,
		Body:  eventBody,
	}
	msg.SetKeys(order.OrderSn)
	if _, err := s.txProducer.SendWithTransaction(ctx, msg, mqTx); err != nil {
		return nil, status.Errorf(codes.Unavailable, "发送事务半消息失败: %v", err)
	}

	// 6. 扣减库存
	if _, err := s.invSrv.StockSellDetail(ctx, &proto.OrderStockDetail{
		OrderGoods: sellDetails,
	}); err != nil {
		// 扣减失败：回滚半消息
		if rerr := mqTx.RollBack(); rerr != nil {
			s.log.Errorf("扣减失败后回滚事务消息异常，等待 Broker 回查: %v", rerr)
		}
		return nil, status.Errorf(codes.Internal, "扣减库存失败: %v", err)
	}

	// 7. 本地事务：写订单 + 商品快照 + 删除购物车中已购买的商品
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].OrderID = order.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		if err := tx.Where("id IN ?", cartIDs).Delete(&ShoppingCart{}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		// 本地事务失败：提交半消息，消费者将异步归还已扣减的库存
		s.log.Errorf("创建订单失败，提交归还库存消息: %v", err)
		if rerr := mqTx.Commit(); rerr != nil {
			s.log.Errorf("提交事务消息失败，等待 Broker 回查: %v", rerr)
		}
		return nil, status.Errorf(codes.Internal, "创建订单失败: %v", err)
	}

	// 8. 本地事务成功：回滚半消息（订单已创建，库存保持扣减状态）
	if err := mqTx.RollBack(); err != nil {
		// 回查时 checker 查到订单已存在会返回 ROLLBACK，最终一致
		s.log.Errorf("回滚事务消息失败，等待 Broker 回查: %v", err)
	}
	return orderModelToResponse(&order, items), nil
}

// OrderList 订单列表
func (s *OrderServer) OrderList(ctx context.Context, req *proto.OrderFilterRequest) (*proto.OrderListResponse, error) {
	var orders []OrderInfo
	q := s.db.Model(&OrderInfo{})
	if req.UserId > 0 {
		q = q.Where("user_id = ?", req.UserId)
	}
	if req.Status > 0 {
		q = q.Where("status = ?", req.Status)
	}

	result := q.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&orders)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询订单列表失败: %v", result.Error)
	}

	rsp := &proto.OrderListResponse{
		Total: int32(result.RowsAffected),
		Data:  make([]*proto.OrderInfoResponse, 0, result.RowsAffected),
	}
	for i := range orders {
		rsp.Data = append(rsp.Data, orderModelToResponse(&orders[i], nil))
	}
	return rsp, nil
}

// GetOrderDetail 通过 id 查询订单（含商品明细）。传入 userId > 0 时校验订单归属。
func (s *OrderServer) GetOrderDetail(ctx context.Context, req *proto.OrderInfoRequest) (*proto.OrderInfoResponse, error) {
	var order OrderInfo
	q := s.db.Where("id = ?", req.Id)
	if req.UserId > 0 {
		q = q.Where("user_id = ?", req.UserId)
	}
	result := q.First(&order)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询订单失败: %v", result.Error)
	}
	var items []OrderGoods
	if err := s.db.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "查询订单商品失败: %v", err)
	}
	return orderModelToResponse(&order, items), nil
}

// UpdateOrderStatus 更新订单状态（支付 / 取消），按订单号定位。
func (s *OrderServer) UpdateOrderStatus(ctx context.Context, req *proto.UpdateOrderStatusInfo) (*emptypb.Empty, error) {
	if req.Status < StatusPending || req.Status > StatusTradeFinished {
		return nil, status.Errorf(codes.InvalidArgument, "非法的订单状态: %d", req.Status)
	}
	result := s.db.Model(&OrderInfo{}).Where("order_sn = ?", req.OrderSn).Update("status", req.Status)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "更新订单状态失败: %v", result.Error)
	}
	return &emptypb.Empty{}, nil
}

// DeleteOrder 软删除订单（BaseModel.DeletedAt）
func (s *OrderServer) DeleteOrder(ctx context.Context, req *proto.DeleteOrderInfo) (*emptypb.Empty, error) {
	result := s.db.Delete(&OrderInfo{}, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "删除订单失败: %v", result.Error)
	}
	return &emptypb.Empty{}, nil
}

// CartItemList 获取某用户的购物车列表
func (s *OrderServer) CartItemList(ctx context.Context, req *proto.CartItemListRequest) (*proto.CartItemListResponse, error) {
	var carts []ShoppingCart
	result := s.db.Where(&ShoppingCart{UserID: req.UserId}).Find(&carts)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询购物车失败: %v", result.Error)
	}
	rsp := &proto.CartItemListResponse{
		Total: int32(result.RowsAffected),
		Data:  make([]*proto.CartItemInfoResponse, 0, result.RowsAffected),
	}
	for _, c := range carts {
		rsp.Data = append(rsp.Data, &proto.CartItemInfoResponse{
			Id:      c.ID,
			UserId:  c.UserID,
			GoodsId: c.GoodsID,
			Num:     c.Num,
			Checked: c.Checked,
		})
	}
	return rsp, nil
}

// AddCartItem 添加商品到购物车：同一用户同一商品已存在则累加数量，否则新建一条记录。
func (s *OrderServer) AddCartItem(ctx context.Context, req *proto.AddCartItemRequest) (*emptypb.Empty, error) {
	var cart ShoppingCart
	result := s.db.Where(&ShoppingCart{UserID: req.UserId, GoodsID: req.GoodsId}).First(&cart)
	switch {
	case result.Error == nil:
		// 已存在，累加数量
		cart.Num += req.Num
		if err := s.db.Save(&cart).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "更新购物车数量失败: %v", err)
		}
	case result.Error == gorm.ErrRecordNotFound:
		// 不存在，新建（按入参决定是否选中）
		cart = ShoppingCart{
			UserID:  req.UserId,
			GoodsID: req.GoodsId,
			Num:     req.Num,
			Checked: req.Checked,
		}
		if err := s.db.Create(&cart).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "添加购物车失败: %v", err)
		}
	default:
		return nil, status.Errorf(codes.Internal, "查询购物车失败: %v", result.Error)
	}
	return &emptypb.Empty{}, nil
}

// UpdateCartItem 更新购物车商品的数量与选中状态（按 id + userId 定位，
// 确保用户只能改自己的购物车项，避免越权）。
func (s *OrderServer) UpdateCartItem(ctx context.Context, req *proto.UpdateCartItemRequest) (*emptypb.Empty, error) {
	result := s.db.Model(&ShoppingCart{}).Where("id = ? AND user_id = ?", req.Id, req.UserId).Updates(map[string]any{
		"num":     req.Num,
		"checked": req.Checked,
	})
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车商品不存在")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "更新购物车失败: %v", result.Error)
	}
	return &emptypb.Empty{}, nil
}

// DeleteCartItem 删除购物车商品
// DeleteCartItem 删除购物车商品（软删除）。按 id + userId 定位，
// 确保用户只能删除自己的购物车项，避免越权。
func (s *OrderServer) DeleteCartItem(ctx context.Context, req *proto.DeleteCartItemRequest) (*emptypb.Empty, error) {
	result := s.db.Where("user_id = ?", req.UserId).Delete(&ShoppingCart{}, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车商品不存在")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "删除购物车失败: %v", result.Error)
	}
	return &emptypb.Empty{}, nil
}

// Paginate 分页
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// newOrderSn 生成订单号：年月日时分秒 + 用户ID + 2位随机数。
// 例：20260724153045 + 1001 + 37 -> "20260724153045100137"。
func newOrderSn(userID int32) string {
	return time.Now().Format("20060102150405") +
		fmt.Sprintf("%d", userID) +
		fmt.Sprintf("%02d", rand.IntN(100))
}

func orderModelToResponse(o *OrderInfo, items []OrderGoods) *proto.OrderInfoResponse {
	rsp := &proto.OrderInfoResponse{
		Id:         o.ID,
		OrderSn:    o.OrderSn,
		UserId:     o.UserID,
		Address:    o.Address,
		Name:       o.Name,
		Mobile:     o.Mobile,
		Post:       o.Post,
		Status:     o.Status,
		PayType:    o.PayType,
		Total:      o.Total,
		PostFee:    o.PostFee,
		AddTime:    o.CreatedAt.Unix(),
		OrderGoods: make([]*proto.OrderItemInfoResponse, 0, len(items)),
	}
	for _, item := range items {
		rsp.OrderGoods = append(rsp.OrderGoods, &proto.OrderItemInfoResponse{
			Id:         item.ID,
			OrderId:    item.OrderID,
			GoodsId:    item.GoodsID,
			GoodsName:  item.GoodsName,
			GoodsImage: item.GoodsImage,
			GoodsPrice: item.GoodsPrice,
			Num:        item.Num,
		})
	}
	return rsp
}
