package handler

import (
	"context"
	"shop/good_srv/proto"
)

type GoodServer struct {
	proto.UnimplementedGoodServer
}

var _ proto.GoodServer = (*GoodServer)(nil)

// 商品接⼝
func (s *GoodServer) GoodList(ctx context.Context, in *proto.GoodFilterRequest) (*proto.GoodListResponse, error) {
	return nil, nil
}

// 现在⽤户提交订单有多个商品，你得批量查询商品的信息吧
func (s *GoodServer) BatchGetGood(ctx context.Context, in *proto.BatchGoodIdInfo) (*proto.GoodListResponse, error) {
	return nil, nil
}
func (s *GoodServer) CreateGood(ctx context.Context, in *proto.CreateGoodInfo) (*proto.GoodInfoResponse, error) {
	return nil, nil

}
func (s *GoodServer) DeleteGood(ctx context.Context, in *proto.DeleteGoodInfo) (*proto.Empty, error) {
	return nil, nil

}
func (s *GoodServer) UpdateGood(ctx context.Context, in *proto.CreateGoodInfo) (*proto.Empty, error) {
	return nil, nil

}
func (s *GoodServer) GetGoodDetail(ctx context.Context, in *proto.GoodInfoRequest) (*proto.GoodInfoResponse, error) {
	return nil, nil

}
