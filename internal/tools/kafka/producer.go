// main provides producer cli util.
package main

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"
	"l0_wb/internal/config"
	"l0_wb/internal/model"
	"l0_wb/internal/util"
)

// main скрипт для генерации и отправки тестового сообщения в kafka.
//
//	go run internal/tools/kafka/producer.go
func main() {
	// Инициализируем логгер, если он еще не был инициализирован
	if err := util.InitLogger(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	logger := util.GetLogger()
	defer util.SyncLogger()

	logger.Info("Starting Kafka producer")

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Создаем Kafka writer
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.KafkaBrokers...),
		Topic:    cfg.KafkaTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer func() {
		if err := writer.Close(); err != nil {
			logger.Warn("Failed to close Kafka writer", zap.Error(err))
		}
	}()

	logger.Info("Kafka writer initialized", zap.String("topic", cfg.KafkaTopic))

	// Инициализация gofakeit
	gofakeit.Seed(0)

	// Генерируем и публикуем сообщение
	order := generateOrder()

	// Преобразуем сообщение в JSON
	data, err := json.Marshal(order)
	if err != nil {
		logger.Fatal("Failed to marshal order", zap.Error(err))
	}

	// Публикуем сообщение в Kafka
	err = writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(order.OrderUID),
		Value: data,
	})
	if err != nil {
		logger.Error("Failed to write message to Kafka", zap.Error(err))
	} else {
		logger.Info("Message published successfully", zap.String("orderUID", order.OrderUID))
	}

}

// generateOrder генерирует случайный заказ со всеми связанными данными.
func generateOrder() *model.Order {
	// Генерация данных для orders
	order := &model.Order{
		OrderUID:          gofakeit.UUID(),
		TrackNumber:       gofakeit.Word(),
		Entry:             gofakeit.Word(),
		Locale:            gofakeit.LanguageAbbreviation(),
		InternalSignature: gofakeit.UUID(),
		CustomerID:        gofakeit.UUID(),
		DeliveryService:   gofakeit.Company(),
		Shardkey:          gofakeit.Word(),
		SmID:              gofakeit.Number(1, 100),
		DateCreated:       time.Now(),
		OofShard:          gofakeit.Word(),
	}

	// Генерация данных для deliveries
	order.Delivery = model.Delivery{
		Name:    gofakeit.Name(),
		Phone:   gofakeit.Phone(),
		Zip:     gofakeit.Zip(),
		City:    gofakeit.City(),
		Address: gofakeit.Street(),
		Region:  gofakeit.State(),
		Email:   gofakeit.Email(),
	}

	// Генерация данных для payments
	order.Payment = model.Payment{
		Transaction:  gofakeit.UUID(),
		RequestID:    gofakeit.UUID(),
		Currency:     gofakeit.CurrencyShort(),
		Provider:     gofakeit.Company(),
		Amount:       gofakeit.Number(100, 10000),
		PaymentDt:    time.Now().Unix(),
		Bank:         gofakeit.Company(),
		DeliveryCost: gofakeit.Number(10, 500),
		GoodsTotal:   gofakeit.Number(50, 5000),
		CustomFee:    gofakeit.Number(0, 100),
	}

	// Генерация данных для items
	for i := 0; i < gofakeit.Number(1, 5); i++ { // Случайное количество товаров в заказе
		order.Items = append(order.Items, model.Item{
			ChrtID:      gofakeit.Number(1000, 9999),
			TrackNumber: gofakeit.Word(),
			Price:       gofakeit.Number(100, 1000),
			Rid:         gofakeit.UUID(),
			Name:        gofakeit.Word(),
			Sale:        gofakeit.Number(0, 50),
			Size:        gofakeit.Letter(),
			TotalPrice:  gofakeit.Number(100, 2000),
			NmID:        gofakeit.Number(100000, 999999),
			Brand:       gofakeit.Company(),
			Status:      gofakeit.Number(1, 3),
		})
	}

	return order
}
