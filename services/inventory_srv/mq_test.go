package main

import (
	"context"
	"errors"
	"testing"

	"shop/pkg/proto"

	rmq "github.com/apache/rocketmq-clients/golang/v5"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type stubInventoryRebacker struct {
	err   error
	calls []*proto.OrderStockDetail
}

func (s *stubInventoryRebacker) RebackDetail(_ context.Context, req *proto.OrderStockDetail) (*emptypb.Empty, error) {
	s.calls = append(s.calls, req)
	if s.err != nil {
		return nil, s.err
	}
	return &emptypb.Empty{}, nil
}

func TestOrderRebackHandlerConsume(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		rebackErr error
		want      rmq.ConsumerResultType
		wantCalls int
	}{
		{
			name:      "valid message",
			body:      `{"order_sn":"20260730001","order_goods":[{"goods_id":11,"num":2}]}`,
			want:      rmq.ConsumerResultTypeSuccess,
			wantCalls: 1,
		},
		{
			name:      "inventory error requests retry",
			body:      `{"order_sn":"20260730001","order_goods":[{"goods_id":11,"num":2}]}`,
			rebackErr: errors.New("database unavailable"),
			want:      rmq.ConsumerResultTypeFailure,
			wantCalls: 1,
		},
		{
			name:      "invalid payload is discarded",
			body:      `{"order_sn":"","order_goods":[]}`,
			want:      rmq.ConsumerResultTypeSuccess,
			wantCalls: 0,
		},
		{
			name:      "invalid json is discarded",
			body:      `{`,
			want:      rmq.ConsumerResultTypeSuccess,
			wantCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &stubInventoryRebacker{err: tt.rebackErr}
			handler := &orderRebackHandler{service: service, log: zap.NewNop().Sugar()}

			got := handler.consume([]byte(tt.body))
			if got.Type != tt.want {
				t.Fatalf("consume() result = %v, want %v", got.Type, tt.want)
			}
			if len(service.calls) != tt.wantCalls {
				t.Fatalf("RebackDetail() calls = %d, want %d", len(service.calls), tt.wantCalls)
			}
			if tt.wantCalls == 1 {
				if service.calls[0].GetOrderSn() != "20260730001" {
					t.Fatalf("RebackDetail() order_sn = %q", service.calls[0].GetOrderSn())
				}
				if len(service.calls[0].GetOrderGoods()) != 0 {
					t.Fatalf("RebackDetail() should load goods from persisted details")
				}
			}
		})
	}
}
