package handler

import (
	"context"
	"fmt"
	"shop/inventory_srv/global"
	"shop/inventory_srv/model"
	"shop/inventory_srv/proto"
	"sort"
	"time"

	"github.com/go-redsync/redsync/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm/clause"
)

type InventoryServer struct {
	proto.UnimplementedInventoryServer
}

var _ proto.InventoryServer = (*InventoryServer)(nil)

func (s *InventoryServer) SetInv(ctx context.Context, in *proto.GoodInvInfo) (*proto.Empty, error) {
	var inv model.Inventory
	global.DB.Where("good = ?", in.GoodId).First(&inv)
	if inv.ID == 0 {
		inv.Good = in.GoodId
		inv.Stock = in.Nums
		global.DB.Create(&inv)
	} else {
		inv.Stock = in.Nums
		global.DB.Save(&inv)
	}
	return &proto.Empty{}, nil
}

func (s *InventoryServer) InvDetail(ctx context.Context, in *proto.GoodInvInfo) (*proto.GoodInvInfo, error) {
	var inv model.Inventory
	global.DB.Where("good = ?", in.GoodId).First(&inv)
	if inv.ID == 0 {
		return nil, status.Errorf(codes.NotFound, "库存信息不存在")
	}
	return &proto.GoodInvInfo{
		GoodId: inv.Good,
		Nums:   inv.Stock,
	}, nil
}

/*
	一次订单扣多种商品时，如果不同事务锁行顺序不一致，容易死锁；常见做法是按 goodId 排序后依次加锁，保证锁行顺序一致
	Tx1（订单1）按顺序锁：先 421，再 422
	Tx2（订单2）按顺序锁：先 422，再 421（顺序相反）
	时间线如下：
	Tx1：FOR UPDATE 锁住行 good=421（成功，持有锁）
	Tx2：FOR UPDATE 锁住行 good=422（成功，持有锁）
	Tx1：继续处理下一件商品，尝试锁 good=422
	但 good=422 已被 Tx2 锁住 → Tx1 等待 Tx2 释放锁
	Tx2：继续处理下一件商品，尝试锁 good=421
	但 good=421 已被 Tx1 锁住 → Tx2 等待 Tx1 释放锁
	此时形成闭环等待：
	Tx1 等 Tx2 释放 422
	Tx2 等 Tx1 释放 421
*/

func (s *InventoryServer) Sell(ctx context.Context, in *proto.SellInfo) (*proto.Empty, error) {
	tx := global.DB.Begin()
	if tx.Error != nil {
		return nil, status.Errorf(codes.Internal, "开启事务失败: %v", tx.Error)
	}

	goods := make([]*proto.GoodInvInfo, len(in.GoodInfos))
	copy(goods, in.GoodInfos)
	sort.Slice(goods, func(i, j int) bool {
		return goods[i].GoodId < goods[j].GoodId
	})

	for _, goodInfo := range goods {

		// var inv model.Inventory
		// 悲观锁, 锁住商品库存，good是索引，会使用行锁
		// 当条件中没有索引时，会使用表锁
		// if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("good = ?", goodInfo.GoodId).First(&inv); err.RowsAffected == 0 {
		//  tx.Rollback()
		//  return nil, status.Errorf(codes.NotFound, "库存信息不存在")
		// }
		// if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("good = ?", goodInfo.GoodId).First(&inv); err.RowsAffected == 0 {
		//  tx.Rollback()
		//  return nil, status.Errorf(codes.NotFound, "库存信息不存在")
		// }
		// inv.Stock -= goodInfo.Nums
		// tx.Save(&inv)

		// 乐观锁, 使用版本号来保证数据一致性,失败重试
		// for {
		//  if err := tx.Where("good = ?", goodInfo.GoodId).First(&inv); err.RowsAffected == 0 {
		//      tx.Rollback()
		//      return nil, status.Errorf(codes.NotFound, "库存信息不存在")
		//  }

		//  if inv.Stock < goodInfo.Nums {
		//      tx.Rollback()
		//      return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		//  }

		//  // UPDATE inventory set stock = stock - 1, version = version + 1 WHERE good = x AND version = 0

		//  // Updates(struct)：默认跳过零值（0、""、false 不更新）,下面第一种方式会跳过零值，有问题
		//  // r := tx.Model(&model.Inventory{}).Where("good = ? AND version = ?", goodInfo.GoodId, inv.Version).Updates(&model.Inventory{Stock: inv.Stock - goodInfo.Nums, Version: inv.Version + 1})
		//  // Select或Updates(map)：会更新零值
		//  r := tx.Model(&model.Inventory{}).Select("stock", "version").Where("good = ? AND version = ?", goodInfo.GoodId, inv.Version).Updates(&model.Inventory{Stock: inv.Stock - goodInfo.Nums, Version: inv.Version + 1})
		//  if r.RowsAffected == 0 {
		//      zap.S().Errorf("库存扣减失败，库存信息不存在")
		//      continue
		//  }
		//  break
		// }

		// 分布式锁，防止多个节点同时扣减库存
		mutex := global.RedisSync.NewMutex(
			fmt.Sprintf("inventory_lock_good_%d", goodInfo.GoodId),
			redsync.WithExpiry(5*time.Second),
			redsync.WithTries(10),
			redsync.WithRetryDelay(100*time.Millisecond))

		// 获取分布式锁
		if err := mutex.Lock(); err != nil {
			tx.Rollback()
			return nil, status.Errorf(codes.Internal, "获取分布式锁失败: %v", err)
		}
		// Lock 成功后 defer Unlock，确保所有返回路径都释放
		defer func(m *redsync.Mutex) {
			_, _ = m.Unlock()
		}(mutex)

		var inv model.Inventory
		if err := tx.Where("good = ?", goodInfo.GoodId).First(&inv); err.RowsAffected == 0 {
			tx.Rollback()
			return nil, status.Errorf(codes.NotFound, "库存信息不存在")
		}
		if inv.Stock < goodInfo.Nums {
			tx.Rollback()
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		inv.Stock -= goodInfo.Nums
		if err := tx.Save(&inv).Error; err != nil {
			tx.Rollback()
			return nil, status.Errorf(codes.Internal, "更新库存失败: %v", err)
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, status.Errorf(codes.Internal, "提交事务失败: %v", err)
	}
	return &proto.Empty{}, nil
}

// 库存归还
// 订单超时未支付，库存归还
// 订单创建失败，库存归还
// 取消订单，库存归还
func (s *InventoryServer) Reback(ctx context.Context, in *proto.SellInfo) (*proto.Empty, error) {
	tx := global.DB.Begin()
	for _, goodInfo := range in.GoodInfos {
		var inv model.Inventory
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("good = ?", goodInfo.GoodId).First(&inv); err.RowsAffected == 0 {
			tx.Rollback()
			return nil, status.Errorf(codes.NotFound, "库存信息不存在")
		}
		inv.Stock += goodInfo.Nums
		tx.Save(&inv)
	}
	tx.Commit()
	return &proto.Empty{}, nil
}
