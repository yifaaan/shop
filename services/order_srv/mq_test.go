package main

import (
	"context"
	"errors"
	"testing"

	rmq "github.com/apache/rocketmq-clients/golang/v5"
	"go.uber.org/zap"
)

type stubOrderTimeoutProcessor struct {
	err   error
	calls []string
}

func (s *stubOrderTimeoutProcessor) processOrderTimeout(_ context.Context, orderSn string) error {
	s.calls = append(s.calls, orderSn)
	return s.err
}

func TestOrderTimeoutHandlerConsume(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		processErr error
		want       rmq.ConsumerResultType
		wantCalls  int
	}{
		{
			name:      "valid timeout",
			body:      `{"order_sn":"20260730001"}`,
			want:      rmq.ConsumerResultTypeSuccess,
			wantCalls: 1,
		},
		{
			name:       "processing error requests retry",
			body:       `{"order_sn":"20260730001"}`,
			processErr: errors.New("inventory unavailable"),
			want:       rmq.ConsumerResultTypeFailure,
			wantCalls:  1,
		},
		{
			name: "missing order number is discarded",
			body: `{"order_sn":""}`,
			want: rmq.ConsumerResultTypeSuccess,
		},
		{
			name: "invalid json is discarded",
			body: `{`,
			want: rmq.ConsumerResultTypeSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &stubOrderTimeoutProcessor{err: tt.processErr}
			handler := &orderTimeoutHandler{service: service, log: zap.NewNop().Sugar()}

			got := handler.consume([]byte(tt.body))
			if got.Type != tt.want {
				t.Fatalf("consume() result = %v, want %v", got.Type, tt.want)
			}
			if len(service.calls) != tt.wantCalls {
				t.Fatalf("processOrderTimeout() calls = %d, want %d", len(service.calls), tt.wantCalls)
			}
			if tt.wantCalls == 1 && service.calls[0] != "20260730001" {
				t.Fatalf("processOrderTimeout() order_sn = %q", service.calls[0])
			}
		})
	}
}
