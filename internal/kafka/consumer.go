package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"l0_wb/internal/cache"
	"l0_wb/internal/model"
	"l0_wb/internal/service"
	"l0_wb/internal/util"
)

// Consumer представляет собой Kafka-консумер, который слушает топик с заказами.
type Consumer struct {
	reader       *kafka.Reader
	orderService service.OrderService
	orderCache   *cache.OrderCache
	logger       *zap.Logger
}

// NewConsumer создает новый экземпляр Consumer.
//
//	Параметры:
//	- brokers: список адресов Kafka-брокеров.
//	- topic: название топика Kafka для чтения сообщений.
//	- groupID: идентификатор группы потребителей Kafka.
//	- orderService: сервис для работы с заказами.
//	- orderCache: кэш для хранения заказов.
//	Возвращает:
//	- *Consumer: экземпляр Kafka-консумера.
func NewConsumer(brokers []string, topic, groupID string, orderService service.OrderService, orderCache *cache.OrderCache) *Consumer {
	logger := util.GetLogger()
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		StartOffset: kafka.FirstOffset, // Начинаем чтение с первого сообщения.
		MinBytes:    10e3,              // Минимальный размер данных 10KB
		MaxBytes:    10e6,              // Максимальный размер данных 10MB
	})

	logger.Info("Kafka consumer created",
		zap.String("topic", topic),
		zap.String("group_id", groupID),
	)

	return &Consumer{
		reader:       r,
		orderService: orderService,
		orderCache:   orderCache,
		logger:       logger,
	}
}

// Run запускает процесс чтения сообщений из Kafka-топика до отмены контекста.
//
//	Параметры:
//	- ctx: контекст выполнения для управления остановкой консумера.
//	Возвращает:
//	- error: ошибку, если произошел сбой при чтении сообщений.
func (c *Consumer) Run(ctx context.Context) error {
	c.logger.Info("Kafka consumer started")
	for {
		// Чтение следующего сообщения из топика
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			c.logger.Error("Failed to read message", zap.Error(err))
			return fmt.Errorf("failed to read message: %w", err)
		}

		var order model.Order
		// Декодируем JSON-сообщение в структуру заказа
		if err := json.Unmarshal(m.Value, &order); err != nil {
			c.logger.Warn("Failed to unmarshal order",
				zap.ByteString("message", m.Value),
				zap.Error(err),
			)
			continue
		}

		// Сохраняем заказ в базу данных через OrderService
		err = c.orderService.SaveOrder(ctx, &order)
		if err != nil {
			c.logger.Error("Failed to save order",
				zap.String("order_uid", order.OrderUID),
				zap.Error(err),
			)
			continue
		}

		// Если заказ успешно сохранен, добавляем его в кэш
		c.orderCache.Set(&order)
		c.logger.Info("Order processed successfully",
			zap.String("order_uid", order.OrderUID),
		)
	}
}

// Close закрывает Kafka reader.
//
//	Возвращает:
//	- error: ошибку, если не удалось закрыть соединение.
func (c *Consumer) Close() error {
	c.logger.Info("Closing Kafka consumer")
	return c.reader.Close()
}
