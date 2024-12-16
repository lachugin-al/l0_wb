package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"

	"l0_wb/internal/cache"
	"l0_wb/internal/model"
	"l0_wb/internal/service"
)

// Consumer представляет собой Kafka-консумер, который слушает топик с заказами.
type Consumer struct {
	reader       *kafka.Reader
	orderService service.OrderService
	orderCache   *cache.OrderCache
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
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		StartOffset: kafka.FirstOffset, // Начинаем чтение с первого сообщения.
		MinBytes:    10e3,              // Минимальный размер данных 10KB
		MaxBytes:    10e6,              // Максимальный размер данных 10MB
	})
	return &Consumer{
		reader:       r,
		orderService: orderService,
		orderCache:   orderCache,
	}
}

// Run запускает процесс чтения сообщений из Kafka-топика до отмены контекста.
//
//	Параметры:
//	- ctx: контекст выполнения для управления остановкой консумера.
//	Возвращает:
//	- error: ошибку, если произошел сбой при чтении сообщений.
func (c *Consumer) Run(ctx context.Context) error {
	for {
		// Чтение следующего сообщения из топика
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("failed to read message: %w", err)
		}

		var order model.Order
		// Декодируем JSON-сообщение в структуру заказа
		if err := json.Unmarshal(m.Value, &order); err != nil {
			fmt.Printf("failed to unmarshal order: %v\n", err)
			continue
		}

		// Сохраняем заказ в базу данных через OrderService
		err = c.orderService.SaveOrder(ctx, &order)
		if err != nil {
			fmt.Printf("failed to save order: %v\n", err)
			continue
		}

		// Если заказ успешно сохранен, добавляем его в кэш
		c.orderCache.Set(&order)
	}
}

// Close закрывает Kafka reader.
//
//	Возвращает:
//	- error: ошибку, если не удалось закрыть соединение.
func (c *Consumer) Close() error {
	return c.reader.Close()
}
