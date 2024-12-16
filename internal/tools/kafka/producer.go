package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"
	"l0_wb/internal/config"
	"l0_wb/internal/model"
)

// main скрипт для генерации и отправки тестового сообщения в kafka.
//
//	go run internal/tools/kafka/producer.go
func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаем Kafka writer
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.KafkaBrokers...),
		Topic:    cfg.KafkaTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	// Инициализация gofakeit
	gofakeit.Seed(0)

	// Генерируем и публикуем сообщение
	order := generateOrder()

	// Преобразуем сообщение в JSON
	data, err := json.Marshal(order)
	if err != nil {
		log.Fatalf("Failed to marshal order: %v", err)
	}

	// Публикуем сообщение в Kafka
	err = writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(order.OrderUID),
		Value: data,
	})
	if err != nil {
		log.Printf("Failed to write message to Kafka: %v", err)
	} else {
		log.Printf("Message published to topic '%s': %s", cfg.KafkaTopic, order.OrderUID)
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
