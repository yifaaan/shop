package handler

import (
	"context"
	"shop/inventory_srv/global"
	"shop/inventory_srv/model"
	"shop/inventory_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *InventoryServer) Sell(ctx context.Context, in *proto.SellInfo) (*proto.Empty, error) {
	tx := global.DB.Begin()
	for _, goodInfo := range in.GoodInfos {
		var inv model.Inventory
		if err := tx.Where("good = ?", goodInfo.GoodId).First(&inv); err.RowsAffected == 0 {
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
		if err := tx.Where("good = ?", goodInfo.GoodId).First(&inv); err.RowsAffected == 0 {
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
