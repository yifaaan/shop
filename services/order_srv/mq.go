package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"shop/services/order_srv/config"

	rmq "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	orderRebackTopic          = "order_reback"
	orderTimeoutTopic         = "order_timeout"
	orderTimeoutConsumerGroup = "order_srv_order_timeout"
	orderTimeoutDelay         = 15 * time.Minute
	orderTimeoutHandleTimeout = 10 * time.Second
)

type rocketMQProducer struct {
	prod rmq.Producer
}

// newRocketMQProducer 创建并启动一个事务消息生产者。
//
// 事务回查在 Broker 未收到 Commit/Rollback（例如进程崩溃）时被调用：
// 按 OrderSn 查库——订单存在说明本地事务成功（库存应保持扣减）→ ROLLBACK；
// 订单不存在说明失败 → COMMIT（归还库存）。
func newRocketMQProducer(cfg config.RocketMQConfig, db *gorm.DB, log *zap.SugaredLogger) (*rocketMQProducer, error) {
	cred := &credentials.SessionCredentials{
		AccessKey:    cfg.AccessKey,
		AccessSecret: cfg.AccessSecret,
	}
	prod, err := rmq.NewProducer(
		&rmq.Config{
			Endpoint:    cfg.Endpoint,
			Credentials: cred,
		},
		rmq.WithTransactionChecker(&rmq.TransactionChecker{
			Check: func(msg *rmq.MessageView) rmq.TransactionResolution {
				var event OrderRebackEvent
				if err := json.Unmarshal(msg.GetBody(), &event); err != nil {
					log.Errorf("事务回查反序列化失败, rollback: %v", err)
					return rmq.ROLLBACK
				}
				var count int64
				if err := db.Model(&OrderInfo{}).Where("order_sn = ?", event.OrderSn).Count(&count).Error; err != nil {
					log.Errorf("事务回查查询订单 %s 失败, 返回 UNKNOWN: %v", event.OrderSn, err)
					return rmq.UNKNOWN
				}
				if count > 0 {
					log.Infof("事务回查: 订单 %s 已创建, ROLLBACK（保持扣减）", event.OrderSn)
					return rmq.ROLLBACK
				}
				log.Infof("事务回查: 订单 %s 不存在, COMMIT（归还库存）", event.OrderSn)
				return rmq.COMMIT
			},
		}),
		rmq.WithTopics(orderRebackTopic, orderTimeoutTopic),
	)
	if err != nil {
		return nil, err
	}
	if err := prod.Start(); err != nil {
		return nil, err
	}
	return &rocketMQProducer{prod: prod}, nil
}

func (p *rocketMQProducer) BeginTransaction() rmq.Transaction {
	return p.prod.BeginTransaction()
}

func (p *rocketMQProducer) Send(ctx context.Context, msg *rmq.Message) ([]*rmq.SendReceipt, error) {
	return p.prod.Send(ctx, msg)
}

func (p *rocketMQProducer) SendWithTransaction(ctx context.Context, msg *rmq.Message, tx rmq.Transaction) ([]*rmq.SendReceipt, error) {
	return p.prod.SendWithTransaction(ctx, msg, tx)
}

func (p *rocketMQProducer) Close() error {
	return p.prod.GracefulStop()
}

type orderTimeoutProcessor interface {
	processOrderTimeout(context.Context, string) error
}

type orderTimeoutHandler struct {
	service orderTimeoutProcessor
	log     *zap.SugaredLogger
}

func (h *orderTimeoutHandler) consume(body []byte) rmq.ConsumerResult {
	orderSn, err := decodeOrderTimeoutEvent(body)
	if err != nil {
		// Retrying cannot repair malformed payloads.
		h.log.Warnf("discard invalid order timeout message: %v", err)
		return rmq.SUCCESS
	}

	ctx, cancel := context.WithTimeout(context.Background(), orderTimeoutHandleTimeout)
	defer cancel()
	if err := h.service.processOrderTimeout(ctx, orderSn); err != nil {
		h.log.Errorf("process timeout for order %s failed, message will be retried: %v", orderSn, err)
		return rmq.FAILURE
	}

	h.log.Infof("order timeout message consumed, order_sn=%s", orderSn)
	return rmq.SUCCESS
}

func decodeOrderTimeoutEvent(body []byte) (string, error) {
	var event OrderTimeoutEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return "", fmt.Errorf("decode payload: %w", err)
	}
	orderSn := strings.TrimSpace(event.OrderSn)
	if orderSn == "" {
		return "", fmt.Errorf("order_sn is required")
	}
	return orderSn, nil
}

type orderTimeoutConsumer struct {
	consumer rmq.PushConsumer
}

func newOrderTimeoutConsumer(cfg config.RocketMQConfig, service orderTimeoutProcessor, log *zap.SugaredLogger) (*orderTimeoutConsumer, error) {
	handler := &orderTimeoutHandler{service: service, log: log}
	consumer, err := rmq.NewPushConsumer(
		&rmq.Config{
			Endpoint:      cfg.Endpoint,
			ConsumerGroup: orderTimeoutConsumerGroup,
			Credentials: &credentials.SessionCredentials{
				AccessKey:    cfg.AccessKey,
				AccessSecret: cfg.AccessSecret,
			},
		},
		rmq.WithPushSubscriptionExpressions(map[string]*rmq.FilterExpression{
			orderTimeoutTopic: rmq.SUB_ALL,
		}),
		rmq.WithPushMessageListener(&rmq.FuncMessageListener{
			Consume: func(message *rmq.MessageView) rmq.ConsumerResult {
				return handler.consume(message.GetBody())
			},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create order timeout consumer: %w", err)
	}
	if err := consumer.Start(); err != nil {
		return nil, fmt.Errorf("start order timeout consumer: %w", err)
	}

	log.Infof("listening for order timeout messages, topic=%s group=%s", orderTimeoutTopic, orderTimeoutConsumerGroup)
	return &orderTimeoutConsumer{consumer: consumer}, nil
}

func (c *orderTimeoutConsumer) Close() error {
	return c.consumer.GracefulStop()
}
