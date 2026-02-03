package handler

import (
	"shop/order_srv/proto"
)

type OrderServer struct {
	proto.UnimplementedOrderServer
}

var _ proto.OrderServer = (*OrderServer)(nil)
