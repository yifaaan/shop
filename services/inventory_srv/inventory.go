package main

import (
	"context"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	result := s.db.Where("goods_id = ?", req.GoodsId).First(&inv)
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

// // StockSellDetail 订单扣减库存
// func (s *InventoryServer) StockSellDetail(ctx context.Context, req *proto.OrderStockDetail) (*emptypb.Empty, error) {
// 	// 事务解决：部分扣减问题
// 	// 并发情况会出现超卖: 悲观锁 FOR UPDATE，查询条件是索引匹配时才会锁行，否则会锁表
// 	tx := s.db.Begin()

// 	for _, goods := range req.OrderGoods {
// 		var inv Inventory
// 		// GoodsID 设置了索引
// 		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&Inventory{GoodsID: goods.GoodsId}).First(&inv)
// 		if result.Error != nil {
// 			tx.Rollback()
// 			if result.Error == gorm.ErrRecordNotFound {
// 				return nil, status.Errorf(codes.NotFound, "库存记录不存在：%v", result.Error)
// 			} else {
// 				return nil, status.Errorf(codes.Internal, "查询库存记录失败：%v", result.Error)
// 			}
// 		}
// 		// 判断库存是否充足
// 		if inv.Stocks < goods.Num {
// 			tx.Rollback()
// 			return nil, status.Error(codes.ResourceExhausted, "库存不足")
// 		}
// 		inv.Stocks -= goods.Num
// 		tx.Save(&inv)
// 	}

// 	tx.Commit()

// 	return &emptypb.Empty{}, nil
// }

// StockSellDetail 订单扣减库存
func (s *InventoryServer) StockSellDetail(ctx context.Context, req *proto.OrderStockDetail) (*emptypb.Empty, error) {
	// 事务解决：部分扣减问题
	// 并发情况会出现超卖: 乐观锁
	tx := s.db.Begin()

	for _, goods := range req.OrderGoods {
		var inv Inventory

		for {
			result := tx.Where(&Inventory{GoodsID: goods.GoodsId}).First(&inv)
			if result.Error != nil {
				tx.Rollback()
				if result.Error == gorm.ErrRecordNotFound {
					return nil, status.Errorf(codes.NotFound, "库存记录不存在：%v", result.Error)
				} else {
					return nil, status.Errorf(codes.Internal, "查询库存记录失败：%v", result.Error)
				}
			}
			// 判断库存是否充足
			if inv.Stocks < goods.Num {
				tx.Rollback()
				return nil, status.Error(codes.ResourceExhausted, "库存不足")
			}
			inv.Stocks -= goods.Num
			result = tx.Model(&Inventory{}).Where(&Inventory{GoodsID: inv.GoodsID, Version: inv.Version}).Select("stocks", "version").Updates(&Inventory{Stocks: inv.Stocks, Version: inv.Version + 1})
			if result.RowsAffected == 0 {
				continue
			}
			break
		}
	}

	tx.Commit()

	return &emptypb.Empty{}, nil
}

// RebackDetail 订单归还库存
//
//	1.订单超时后需要归还
//	2.订单创建失败时，需要归还之前扣减的库存
//	3.用户手动归还
func (s *InventoryServer) RebackDetail(ctx context.Context, req *proto.OrderStockDetail) (*emptypb.Empty, error) {
	tx := s.db.Begin()

	for _, goods := range req.OrderGoods {
		var inv Inventory
		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&Inventory{GoodsID: goods.GoodsId}).First(&inv)
		if result.Error != nil {
			tx.Rollback()
			if result.Error == gorm.ErrRecordNotFound {
				return nil, status.Errorf(codes.NotFound, "库存记录不存在：%v", result.Error)
			} else {
				return nil, status.Errorf(codes.Internal, "查询库存记录失败：%v", result.Error)
			}
		}
		inv.Stocks += goods.Num
		tx.Save(&inv)
	}

	tx.Commit()

	return &emptypb.Empty{}, nil
}

// ShowInvDetail 后台分页查询库存
func (s *InventoryServer) ShowInvDetail(ctx context.Context, req *proto.ShowInvDetailRequest) (*proto.ShowInvDetailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ShowInvDetail not implemented")
}
