package handler

import (
	"context"
	"shop/inventory_srv/global"
	"shop/inventory_srv/model"
	"shop/inventory_srv/proto"

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
	for _, goodInfo := range in.GoodInfos {
		var inv model.Inventory
		// 悲观锁, 锁住商品库存，good是索引，会使用行锁
		// 当条件中没有索引时，会使用表锁
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("good = ?", goodInfo.GoodId).First(&inv); err.RowsAffected == 0 {
			tx.Rollback()
			return nil, status.Errorf(codes.NotFound, "库存信息不存在")
		}
		if inv.Stock < goodInfo.Nums {
			tx.Rollback()
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		// 扣减库存，并发出现超卖问题，需要使用分布式锁来保证数据一致性
		inv.Stock -= goodInfo.Nums
		tx.Save(&inv)
	}
	tx.Commit()
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
		// 归还库存，需要使用分布式锁来保证数据一致性
		inv.Stock += goodInfo.Nums
		tx.Save(&inv)
	}
	tx.Commit()
	return &proto.Empty{}, nil
}
