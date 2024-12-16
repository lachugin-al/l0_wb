package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"l0_wb/internal/model"
)

func main() {
	// Параметры командной строки
	seedFilePath := flag.String("seed-file", "internal/db/migrations/seed.sql", "Path for the seed file")
	seedRecordCount := flag.Int("seed-count", 10, "Number of seed records to generate")
	flag.Parse()

	// Проверяем, что файл не пустой
	if *seedFilePath == "" {
		log.Fatal("File path is required")
	}

	// Генерируем seed-данные
	log.Printf("Generating %d records into %s", *seedRecordCount, *seedFilePath)
	if err := GenerateSeedData(*seedFilePath, *seedRecordCount); err != nil {
		log.Fatalf("Failed to generate seed data: %v", err)
	}

	log.Println("Seed data generation completed successfully")
}

// GenerateSeedData генерирует случайные данные для всех таблиц и добавляет их в seed.sql.
func GenerateSeedData(filePath string, recordCount int) error {
	var file *os.File

	// Проверяем, существует ли файл
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err = os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create seed file: %w", err)
		}
		log.Printf("Created new seed file: %s", filePath)
	}
	defer file.Close()

	// Генерируем данные и записываем их в файл
	for i := 0; i < recordCount; i++ {
		// Генерация данных для orders
		order := model.Order{
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
		delivery := model.Delivery{
			Name:    gofakeit.Name(),
			Phone:   gofakeit.Phone(),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.Street(),
			Region:  gofakeit.State(),
			Email:   gofakeit.Email(),
		}

		// Генерация данных для payments
		payment := model.Payment{
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
		item := model.Item{
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
		}

		// SQL для orders
		orderSQL := fmt.Sprintf(
			`INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) 
			VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', %d, '%s', '%s');`,
			order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID,
			order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated.Format("2006-01-02 15:04:05"), order.OofShard,
		)

		// SQL для deliveries
		deliverySQL := fmt.Sprintf(
			`INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email) 
			VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s');`,
			order.OrderUID, delivery.Name, delivery.Phone, delivery.Zip, delivery.City, delivery.Address, delivery.Region, delivery.Email,
		)

		// SQL для payments
		paymentSQL := fmt.Sprintf(
			`INSERT INTO payments (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) 
			VALUES ('%s', '%s', '%s', '%s', '%s', %d, %d, '%s', %d, %d, %d);`,
			order.OrderUID, payment.Transaction, payment.RequestID, payment.Currency, payment.Provider, payment.Amount, payment.PaymentDt,
			payment.Bank, payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee,
		)

		// SQL для items
		itemSQL := fmt.Sprintf(
			`INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) 
			VALUES ('%s', %d, '%s', %d, '%s', '%s', %d, '%s', %d, %d, '%s', %d);`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size,
			item.TotalPrice, item.NmID, item.Brand, item.Status,
		)

		// Записываем SQL-запросы в файл
		if _, err := file.WriteString(orderSQL + "\n"); err != nil {
			return fmt.Errorf("failed to write to seed file: %w", err)
		}
		if _, err := file.WriteString(deliverySQL + "\n"); err != nil {
			return fmt.Errorf("failed to write to seed file: %w", err)
		}
		if _, err := file.WriteString(paymentSQL + "\n"); err != nil {
			return fmt.Errorf("failed to write to seed file: %w", err)
		}
		if _, err := file.WriteString(itemSQL + "\n"); err != nil {
			return fmt.Errorf("failed to write to seed file: %w", err)
		}
	}

	log.Printf("Appended %d records to %s", recordCount, filePath)
	return nil
}
