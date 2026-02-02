package handler

import (
	"context"
	"shop/inventory_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InventoryServer struct {
	proto.UnimplementedInventoryServer
}

var _ proto.InventoryServer = (*InventoryServer)(nil)

func (s *InventoryServer) SetInv(ctx context.Context, in *proto.GoodInvInfo) (*proto.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetInv not implemented")
}
func (s *InventoryServer) InvDetail(ctx context.Context, in *proto.GoodInvInfo) (*proto.GoodInvInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InvDetail not implemented")
}
func (s *InventoryServer) Sell(ctx context.Context, in *proto.SellInfo) (*proto.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Sell not implemented")
}
func (s *InventoryServer) Reback(ctx context.Context, in *proto.SellInfo) (*proto.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Reback not implemented")
}
