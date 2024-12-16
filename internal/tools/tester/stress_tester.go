package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// main запускает стресс-тест с использованием Vegeta.
//
//	Пример запуска:
//	go run internal/tools/stress_tester.go -url=http://localhost:8081/order/test-0 -rate=100 -duration=30 -output=stress_test_results.json
func main() {
	// Параметры командной строки
	url := flag.String("url", "http://localhost:8081/order/test-0", "Target URL for stress testing")
	rate := flag.Int("rate", 1000, "Requests per second")
	duration := flag.Int("duration", 30, "Test duration in seconds")
	output := flag.String("output", "stress_test_results.json", "Output file for test results")
	flag.Parse()

	// Проверка параметров
	if *url == "" {
		log.Fatal("Target URL is required")
	}

	log.Printf("Starting stress test: %d RPS for %d seconds on %s", *rate, *duration, *url)

	// Запуск стресс-теста
	if err := RunStressTest(*url, *rate, *duration, *output); err != nil {
		log.Fatalf("Stress test failed: %v", err)
	}

	log.Println("Stress test completed successfully")
}

// RunStressTest запускает стресс-тест и сохраняет результаты в файл.
//
//	Параметры:
//	- url: Целевой URL для тестирования.
//	- rate: Частота запросов в секунду.
//	- duration: Длительность теста в секундах.
//	- output: Файл для сохранения результатов.
func RunStressTest(url string, rate, duration int, output string) error {
	// Настройка Vegeta
	rateLimiter := vegeta.Rate{Freq: rate, Per: time.Second}
	durationTime := time.Duration(duration) * time.Second
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    url,
	})
	attacker := vegeta.NewAttacker()

	// Сбор метрик
	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rateLimiter, durationTime, "Stress Test") {
		metrics.Add(res)
	}
	metrics.Close()

	// Запись метрик
	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Экспорт метрик в JSON
	if err := json.NewEncoder(file).Encode(metrics); err != nil {
		return fmt.Errorf("failed to encode metrics to JSON: %w", err)
	}

	log.Printf("Stress test results saved to %s", output)
	return nil
}
