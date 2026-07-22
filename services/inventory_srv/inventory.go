package main

import (
	"context"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// Inventory 库存记录
type Inventory struct {
	basemodel.BaseModel
	GoodsID int32 `gorm:"column:goods_id;type:int;not null;uniqueIndex"` // 商品ID
	Stocks  int32 `gorm:"column:stocks;type:int;not null;default:0"`     // 库存数量
	Version int32 `gorm:"column:version;type:int;not null;default:0"`    // 分布式乐观锁版本号
}

// // InventoryHistory 扣减库存历史记录，用来回滚库存
// type InventoryHistory struct {
// 	userID  int32  `gorm:"column:user_id;type:int;not null"`          // 用户ID
// 	goodsID int32  `gorm:"column:goods_id;type:int;not null"`         // 商品ID
// 	orderSn string `gorm:"column:order_sn;type:varchar(50);not null"` // 订单号
// 	stocks  int32  `gorm:"column:stocks;type:int;not null"`           // 扣减库存数量
// 	status  int32  `gorm:"column:status;type:int;not null"`           // 状态：1-预扣减，2-已扣减（支付成功）
// }

// InventoryServer implements proto.InventoryServer over an injected *gorm.DB.
type InventoryServer struct {
	proto.UnimplementedInventoryServer
	db *gorm.DB
}

// NewInventoryServer wires a InventoryServer to its data store.
func NewInventoryServer(db *gorm.DB) *InventoryServer {
	return &InventoryServer{db: db}
}

// SetInv 设置库存
func (s *InventoryServer) SetInv(ctx context.Context, req *proto.GoodsInvInfo) (*emptypb.Empty, error) {
	var inv Inventory
	result := s.db.Where("goods_id = ?", req.GoodsId).First(&inv)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 如果库存记录不存在，则创建一条新的记录
			inv = Inventory{
				GoodsID: req.GoodsId,
				Stocks:  req.Stocks,
				Version: 0,
			}
			if err := s.db.Create(&inv).Error; err != nil {
				return nil, status.Errorf(codes.Internal, "创建库存记录失败: %v", err)
			}
		} else {
			return nil, status.Errorf(codes.Internal, "查询库存记录失败: %v", result.Error)
		}
	} else {
		// 如果库存记录存在，则更新库存数量
		inv.Stocks = req.Stocks
		if err := s.db.Save(&inv).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "更新库存记录失败: %v", err)
		}
	}
	return &emptypb.Empty{}, nil
}

// GetInvDetail 查询库存详情
func (s *InventoryServer) GetInvDetail(ctx context.Context, req *proto.GoodsInvInfo) (*proto.GoodsInvInfo, error) {
	var inv Inventory
	result := s.db.Where("goods_id = ?", req.GoodsId)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "库存记录不存在：%v", result.Error)
		} else {
			return nil, status.Errorf(codes.Internal, "查询库存记录失败：%v", result.Error)
		}
	}
	return &proto.GoodsInvInfo{
		GoodsId: inv.GoodsID,
		Stocks:  inv.Stocks,
	}, nil
}

// StockSellDetail 订单扣减库存
func (s *InventoryServer) StockSellDetail(ctx context.Context, req *proto.OrderStockDetail) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StockSellDetail not implemented")
}

// RebackDetail 订单归还库存
func (s *InventoryServer) RebackDetail(ctx context.Context, req *proto.OrderStockDetail) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RebackDetail not implemented")
}

// ShowInvDetail 后台分页查询库存
func (s *InventoryServer) ShowInvDetail(ctx context.Context, req *proto.ShowInvDetailRequest) (*proto.ShowInvDetailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ShowInvDetail not implemented")
}
