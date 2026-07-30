package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"shop/pkg/proto"
	"shop/services/inventory_srv/config"

	rmq "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	orderRebackTopic         = "order_reback"
	orderRebackConsumerGroup = "inventory_srv_order_reback"
	orderRebackTimeout       = 10 * time.Second
)

type orderRebackEvent struct {
	OrderSn string `json:"order_sn"`
}

type inventoryRebacker interface {
	RebackDetail(context.Context, *proto.OrderStockDetail) (*emptypb.Empty, error)
}

type orderRebackHandler struct {
	service inventoryRebacker
	log     *zap.SugaredLogger
}

func (h *orderRebackHandler) consume(body []byte) rmq.ConsumerResult {
	orderSn, req, err := decodeOrderRebackEvent(body)
	if err != nil {
		// Retrying cannot repair an invalid payload, so acknowledge it after logging.
		h.log.Warnf("discard invalid inventory reback message: %v", err)
		return rmq.SUCCESS
	}

	ctx, cancel := context.WithTimeout(context.Background(), orderRebackTimeout)
	defer cancel()
	if _, err := h.service.RebackDetail(ctx, req); err != nil {
		h.log.Errorf("reback inventory for order %s failed, message will be retried: %v", orderSn, err)
		return rmq.FAILURE
	}

	h.log.Infof("inventory reback message consumed, order_sn=%s", orderSn)
	return rmq.SUCCESS
}

func decodeOrderRebackEvent(body []byte) (string, *proto.OrderStockDetail, error) {
	var event orderRebackEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return "", nil, fmt.Errorf("decode payload: %w", err)
	}
	if event.OrderSn == "" {
		return "", nil, fmt.Errorf("order_sn is required")
	}
	return event.OrderSn, &proto.OrderStockDetail{
		OrderSn: event.OrderSn,
	}, nil
}

type rocketMQConsumer struct {
	consumer rmq.PushConsumer
}

func newRocketMQConsumer(cfg config.RocketMQConfig, service inventoryRebacker, log *zap.SugaredLogger) (*rocketMQConsumer, error) {
	topic := cfg.Topic
	if topic == "" {
		topic = orderRebackTopic
	}
	consumerGroup := cfg.ConsumerGroup
	if consumerGroup == "" {
		consumerGroup = orderRebackConsumerGroup
	}

	handler := &orderRebackHandler{service: service, log: log}
	consumer, err := rmq.NewPushConsumer(
		&rmq.Config{
			Endpoint:      cfg.Endpoint,
			ConsumerGroup: consumerGroup,
			Credentials: &credentials.SessionCredentials{
				AccessKey:    cfg.AccessKey,
				AccessSecret: cfg.AccessSecret,
			},
		},
		rmq.WithPushSubscriptionExpressions(map[string]*rmq.FilterExpression{
			topic: rmq.SUB_ALL,
		}),
		rmq.WithPushMessageListener(&rmq.FuncMessageListener{
			Consume: func(message *rmq.MessageView) rmq.ConsumerResult {
				return handler.consume(message.GetBody())
			},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create rocketmq consumer: %w", err)
	}
	if err := consumer.Start(); err != nil {
		return nil, fmt.Errorf("start rocketmq consumer: %w", err)
	}

	log.Infof("listening for inventory reback messages, topic=%s group=%s", topic, consumerGroup)
	return &rocketMQConsumer{consumer: consumer}, nil
}

func (c *rocketMQConsumer) Close() error {
	return c.consumer.GracefulStop()
}
