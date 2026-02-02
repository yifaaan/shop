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
	global.DB.First(&inv, in.GoodId)

	inv.Good = in.GoodId
	inv.Stock = in.Nums
	global.DB.Save(&inv)
	return &proto.Empty{}, nil
}
func (s *InventoryServer) InvDetail(ctx context.Context, in *proto.GoodInvInfo) (*proto.GoodInvInfo, error) {
	var inv model.Inventory
	global.DB.First(&inv, in.GoodId)
	if inv.ID == 0 {
		return nil, status.Errorf(codes.NotFound, "库存信息不存在")
	}
	return &proto.GoodInvInfo{
		GoodId: inv.Good,
		Nums:   inv.Stock,
	}, nil
}
func (s *InventoryServer) Sell(ctx context.Context, in *proto.SellInfo) (*proto.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Sell not implemented")
}
func (s *InventoryServer) Reback(ctx context.Context, in *proto.SellInfo) (*proto.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Reback not implemented")
}
