// main provides stress-tester cli util.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// main запускает стресс-тест с использованием Vegeta.
//
//	Пример запуска:
//	go run internal/tools/ht/stress_tester.go -url=http://localhost:8081/order/test-0 -rate=100 -duration=30 -output=stress_test_results.json
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
	attacker := vegeta.NewAttacker(vegeta.Connections(10000))

	// Сбор метрик
	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rateLimiter, durationTime, "Stress Test") {
		metrics.Add(res)
	}
	metrics.Close()

	// Разрешенная директория
	allowedDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Очистка пути
	cleanPath := filepath.Clean(output)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	// Проверяем, что путь находится внутри allowedDir
	relPath, err := filepath.Rel(allowedDir, absPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("attempt to write outside allowed directory: %s", absPath)
	}

	// Кодируем метрики в JSON
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to encode metrics to JSON: %w", err)
	}

	// Безопасная запись в файл
	err = os.WriteFile(absPath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write data to file: %w", err)
	}

	log.Printf("Stress test results saved to %s", absPath)
	return nil
}
