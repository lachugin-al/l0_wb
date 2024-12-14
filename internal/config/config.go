package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config содержит все необходимые параметры конфигурации приложения.
type Config struct {
	// Параметры подключения к базе данных
	DBHost     string // Хост базы данных
	DBPort     int    // Порт базы данных
	DBUser     string // Имя пользователя базы данных
	DBPassword string // Пароль пользователя базы данных
	DBName     string // Имя базы данных

	// Параметры Kafka
	KafkaBrokers []string // Адреса брокеров Kafka
	KafkaTopic   string   // Топик Kafka для обработки заказов
	KafkaGroupID string   // Группа потребителей Kafka

	// Параметры HTTP-сервера
	HTTPPort string // Порт, на котором работает HTTP-сервер

	ShutdownTimeout time.Duration // Таймаут на завершение работы приложения
}

// LoadConfig загружает конфигурацию из переменных окружения или использует значения по умолчанию.
//
//	Возвращает:
//	- *Config: указатель на объект конфигурации.
//	- error: ошибку, если какие-либо из параметров не удалось обработать.
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	// Параметры базы данных
	cfg.DBHost = getEnv("DB_HOST", "localhost")
	dbPortStr := getEnv("DB_PORT", "5432")
	port, err := strconv.Atoi(dbPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}
	cfg.DBPort = port
	cfg.DBUser = getEnv("DB_USER", "orders_user")
	cfg.DBPassword = getEnv("DB_PASSWORD", "securepassword")
	cfg.DBName = getEnv("DB_NAME", "orders_db")

	// Параметры Kafka
	kafkaBrokersStr := getEnv("KAFKA_BROKERS", "localhost:9092")
	cfg.KafkaBrokers = []string{kafkaBrokersStr}
	cfg.KafkaTopic = getEnv("KAFKA_TOPIC", "orders")
	cfg.KafkaGroupID = getEnv("KAFKA_GROUP_ID", "orders_group")

	// Параметры HTTP-сервера
	cfg.HTTPPort = getEnv("HTTP_PORT", "8081")

	// Таймаут завершения работы приложения
	shutdownTimeoutStr := getEnv("SHUTDOWN_TIMEOUT", "5s")
	shutdownTimeout, err := time.ParseDuration(shutdownTimeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SHUTDOWN_TIMEOUT: %w", err)
	}
	cfg.ShutdownTimeout = shutdownTimeout

	return cfg, nil
}

// getEnv возвращает значение переменной окружения или значение по умолчанию, если переменная не установлена.
//
//	Параметры:
//	- key: имя переменной окружения.
//	- defaultVal: значение по умолчанию.
//	Возвращает:
//	- string: значение переменной окружения или значение по умолчанию.
func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
