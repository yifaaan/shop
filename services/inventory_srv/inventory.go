package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"

	"github.com/go-redsync/redsync/v4"
	"go.uber.org/zap"
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

const (
	stockDeductionStatusDeducted int32 = 1
	stockDeductionStatusReturned int32 = 2
)

// StockDeduction records one order-level inventory deduction.
type StockDeduction struct {
	basemodel.BaseModel
	OrderSn string `gorm:"column:order_sn;type:varchar(40);not null;uniqueIndex:uniq_stock_sell_detail_order_sn"`
	Status  int32  `gorm:"column:status;type:tinyint;not null;default:1;check:chk_stock_sell_detail_status,status IN (1,2)"`
}

func (StockDeduction) TableName() string {
	return "stock_sell_detail"
}

// StockDeductionItem records the quantity deducted for one product.
type StockDeductionItem struct {
	basemodel.BaseModel
	StockSellDetailID int32 `gorm:"column:stock_sell_detail_id;type:int;not null;uniqueIndex:uniq_stock_sell_detail_goods"`
	GoodsID           int32 `gorm:"column:goods_id;type:int;not null;uniqueIndex:uniq_stock_sell_detail_goods"`
	Num               int32 `gorm:"column:num;type:int;not null"`
}

func (StockDeductionItem) TableName() string {
	return "stock_sell_detail_item"
}

type stockDeductionGoods struct {
	GoodsID int32
	Num     int32
}

// InventoryServer implements proto.InventoryServer over an injected *gorm.DB.
type InventoryServer struct {
	proto.UnimplementedInventoryServer
	db  *gorm.DB
	rs  *redsync.Redsync
	log *zap.SugaredLogger
}

// NewInventoryServer wires a InventoryServer to its data store.
func NewInventoryServer(db *gorm.DB, rs *redsync.Redsync, log *zap.SugaredLogger) *InventoryServer {
	return &InventoryServer{db: db, rs: rs, log: log}
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

// // StockSellDetail 订单扣减库存
// func (s *InventoryServer) StockSellDetail(ctx context.Context, req *proto.OrderStockDetail) (*emptypb.Empty, error) {
// 	// 事务解决：部分扣减问题
// 	// 并发情况会出现超卖: 乐观锁
// 	tx := s.db.Begin()

// 	for _, goods := range req.OrderGoods {
// 		var inv Inventory

// 		for {
// 			result := tx.Where(&Inventory{GoodsID: goods.GoodsId}).First(&inv)
// 			if result.Error != nil {
// 				tx.Rollback()
// 				if result.Error == gorm.ErrRecordNotFound {
// 					return nil, status.Errorf(codes.NotFound, "库存记录不存在：%v", result.Error)
// 				} else {
// 					return nil, status.Errorf(codes.Internal, "查询库存记录失败：%v", result.Error)
// 				}
// 			}
// 			// 判断库存是否充足
// 			if inv.Stocks < goods.Num {
// 				tx.Rollback()
// 				return nil, status.Error(codes.ResourceExhausted, "库存不足")
// 			}
// 			inv.Stocks -= goods.Num
// 			result = tx.Model(&Inventory{}).Where(&Inventory{GoodsID: inv.GoodsID, Version: inv.Version}).Select("stocks", "version").Updates(&Inventory{Stocks: inv.Stocks, Version: inv.Version + 1})
// 			if result.RowsAffected == 0 {
// 				continue
// 			}
// 			break
// 		}
// 	}

// 	tx.Commit()

// 	return &emptypb.Empty{}, nil
// }

// StockSellDetail 订单扣减库存
func (s *InventoryServer) StockSellDetail(ctx context.Context, req *proto.OrderStockDetail) (*emptypb.Empty, error) {
	goods, err := normalizeOrderGoods(req.GetOrderGoods())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	locks, err := s.lockInventory(ctx, goods)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "获取库存锁失败: %v", err)
	}
	defer s.unlockInventory(locks)

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, status.Errorf(codes.Internal, "开始库存扣减事务失败: %v", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var existing StockDeduction
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_sn = ?", req.GetOrderSn()).First(&existing).Error
	if err == nil {
		matches, matchErr := deductionMatches(tx, existing.ID, goods)
		tx.Rollback()
		if matchErr != nil {
			return nil, status.Errorf(codes.Internal, "查询库存扣减明细失败: %v", matchErr)
		}
		if !matches {
			return nil, status.Error(codes.AlreadyExists, "同一订单号的库存扣减明细不一致")
		}
		switch existing.Status {
		case stockDeductionStatusDeducted:
			return &emptypb.Empty{}, nil
		case stockDeductionStatusReturned:
			return nil, status.Error(codes.FailedPrecondition, "该订单库存已经归还")
		default:
			return nil, status.Errorf(codes.Internal, "未知库存扣减状态: %d", existing.Status)
		}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "查询库存扣减记录失败: %v", err)
	}

	for _, item := range goods {
		var inv Inventory
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&Inventory{GoodsID: item.GoodsID}).First(&inv).Error
		if err != nil {
			tx.Rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, status.Errorf(codes.NotFound, "库存记录不存在")
			}
			return nil, status.Errorf(codes.Internal, "查询库存失败: %v", err)
		}

		if inv.Stocks < item.Num {
			tx.Rollback()
			return nil, status.Errorf(codes.ResourceExhausted, "商品 %d 库存不足", item.GoodsID)
		}

		inv.Stocks -= item.Num
		if err := tx.Save(&inv).Error; err != nil {
			tx.Rollback()
			return nil, status.Errorf(codes.Internal, "更新库存失败: %v", err)
		}
	}

	deduction := StockDeduction{
		OrderSn: req.GetOrderSn(),
		Status:  stockDeductionStatusDeducted,
	}
	if err := tx.Create(&deduction).Error; err != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "创建库存扣减记录失败: %v", err)
	}
	details := make([]StockDeductionItem, 0, len(goods))
	for _, item := range goods {
		details = append(details, StockDeductionItem{
			StockSellDetailID: deduction.ID,
			GoodsID:           item.GoodsID,
			Num:               item.Num,
		})
	}
	if err := tx.Create(&details).Error; err != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "创建库存扣减明细失败: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, status.Errorf(codes.Internal, "提交库存扣减事务失败: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// RebackDetail 订单归还库存
//
//	1.订单超时后需要归还
//	2.订单创建失败时，需要归还之前扣减的库存
//	3.用户手动归还
func (s *InventoryServer) RebackDetail(ctx context.Context, req *proto.OrderStockDetail) (*emptypb.Empty, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, status.Errorf(codes.Internal, "开始库存归还事务失败: %v", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var deduction StockDeduction
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_sn = ?", req.GetOrderSn()).First(&deduction).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "库存扣减记录不存在")
		}
		return nil, status.Errorf(codes.Internal, "查询库存扣减记录失败: %v", err)
	}
	if deduction.Status == stockDeductionStatusReturned {
		tx.Rollback()
		return &emptypb.Empty{}, nil
	}
	if deduction.Status != stockDeductionStatusDeducted {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "未知库存扣减状态: %d", deduction.Status)
	}

	var details []StockDeductionItem
	if err := tx.Where("stock_sell_detail_id = ?", deduction.ID).Order("goods_id ASC").Find(&details).Error; err != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "查询库存扣减明细失败: %v", err)
	}
	if len(details) == 0 {
		tx.Rollback()
		return nil, status.Error(codes.FailedPrecondition, "库存扣减明细为空")
	}

	for _, detail := range details {
		var inv Inventory
		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&Inventory{GoodsID: detail.GoodsID}).First(&inv)
		if result.Error != nil {
			tx.Rollback()
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil, status.Errorf(codes.NotFound, "库存记录不存在：%v", result.Error)
			}
			return nil, status.Errorf(codes.Internal, "查询库存记录失败：%v", result.Error)
		}
		inv.Stocks += detail.Num
		if err := tx.Save(&inv).Error; err != nil {
			tx.Rollback()
			return nil, status.Errorf(codes.Internal, "更新归还库存失败: %v", err)
		}
	}

	result := tx.Model(&deduction).
		Where("status = ?", stockDeductionStatusDeducted).
		Update("status", stockDeductionStatusReturned)
	if result.Error != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "更新库存扣减状态失败: %v", result.Error)
	}
	if result.RowsAffected != 1 {
		tx.Rollback()
		return nil, status.Error(codes.Aborted, "库存扣减状态已发生变化")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, status.Errorf(codes.Internal, "提交库存归还事务失败: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func normalizeOrderGoods(orderGoods []*proto.OrderGoodsDetail) ([]stockDeductionGoods, error) {
	totals := make(map[int32]int64, len(orderGoods))
	for _, item := range orderGoods {
		totals[item.GetGoodsId()] += int64(item.GetNum())
		if totals[item.GetGoodsId()] > math.MaxInt32 {
			return nil, fmt.Errorf("商品 %d 数量超出范围", item.GetGoodsId())
		}
	}

	goodsIDs := make([]int32, 0, len(totals))
	for goodsID := range totals {
		goodsIDs = append(goodsIDs, goodsID)
	}
	slices.Sort(goodsIDs)

	goods := make([]stockDeductionGoods, 0, len(goodsIDs))
	for _, goodsID := range goodsIDs {
		goods = append(goods, stockDeductionGoods{GoodsID: goodsID, Num: int32(totals[goodsID])})
	}
	return goods, nil
}

func deductionMatches(tx *gorm.DB, deductionID int32, goods []stockDeductionGoods) (bool, error) {
	var details []StockDeductionItem
	if err := tx.Where("stock_sell_detail_id = ?", deductionID).Order("goods_id ASC").Find(&details).Error; err != nil {
		return false, err
	}
	if len(details) != len(goods) {
		return false, nil
	}
	for i := range details {
		if details[i].GoodsID != goods[i].GoodsID || details[i].Num != goods[i].Num {
			return false, nil
		}
	}
	return true, nil
}

func (s *InventoryServer) lockInventory(ctx context.Context, goods []stockDeductionGoods) ([]*redsync.Mutex, error) {
	locks := make([]*redsync.Mutex, 0, len(goods))
	for _, item := range goods {
		mutex := s.rs.NewMutex(
			fmt.Sprintf("lock:inventory:%d", item.GoodsID),
			redsync.WithExpiry(30*time.Second),
			redsync.WithTries(20),
			redsync.WithRetryDelay(100*time.Millisecond),
		)
		if err := mutex.LockContext(ctx); err != nil {
			s.unlockInventory(locks)
			return nil, err
		}
		locks = append(locks, mutex)
	}
	return locks, nil
}

func (s *InventoryServer) unlockInventory(locks []*redsync.Mutex) {
	for i := len(locks) - 1; i >= 0; i-- {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		ok, err := locks[i].UnlockContext(ctx)
		cancel()
		if err != nil || !ok {
			s.log.Warnf("释放库存锁失败: ok=%t err=%v", ok, err)
		}
	}
}

// ShowInvDetail 后台分页查询库存
func (s *InventoryServer) ShowInvDetail(ctx context.Context, req *proto.ShowInvDetailRequest) (*proto.ShowInvDetailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ShowInvDetail not implemented")
}
