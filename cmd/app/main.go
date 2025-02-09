// main provides the entry point for the application.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"l0_wb/internal/cache"
	"l0_wb/internal/config"
	"l0_wb/internal/db"
	"l0_wb/internal/kafka"
	"l0_wb/internal/repository"
	"l0_wb/internal/server"
	"l0_wb/internal/service"
	"l0_wb/internal/util"
)

// main инициализирует приложение, настраивает зависимости, запускает Kafka-консьюмер и HTTP-сервер.
func main() {
	// Инициализация логгера
	if err := util.InitLogger(); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer util.SyncLogger()

	logger := util.GetLogger()

	// Запуск приложения в стандартном режиме
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Настраиваем перехват сигналов для корректного завершения работы
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		cancel()
	}()

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("failed to load config: %v", zap.Error(err))
	}

	// Инициализация БД
	database, err := db.InitDB(cfg)
	if err != nil {
		logger.Fatal("failed to initialize database: %v", zap.Error(err))
	}

	// Создание репозиториев
	ordersRepo := repository.NewOrdersRepository(database)
	deliveriesRepo := repository.NewDeliveriesRepository(database)
	paymentsRepo := repository.NewPaymentsRepository(database)
	itemsRepo := repository.NewItemsRepository(database)

	// Инициализация кэша и загрузка данных из БД
	orderCache := cache.NewOrderCache()
	if err := orderCache.LoadFromDB(ctx, ordersRepo, deliveriesRepo, paymentsRepo, itemsRepo, database); err != nil {
		logger.Warn("failed to load cache from DB: %v", zap.Error(err))
	}

	// Инициализация сервисов
	orderService := service.NewOrderService(database, ordersRepo, deliveriesRepo, paymentsRepo, itemsRepo)

	// Запуск Kafka-консьюмера для получения новых заказов
	consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID, orderService, orderCache)
	go func() {
		if err := consumer.Run(ctx); err != nil {
			logger.Fatal("kafka consumer stopped with error: %v", zap.Error(err))
			cancel()
		}
	}()

	// Запуск HTTP-сервера
	// Раздача статических файлов из директории "web".
	srv := server.NewServer(cfg.HTTPPort, orderCache, "web")

	if err := srv.Start(ctx); err != nil {
		logger.Fatal("http server stopped with error: %v", zap.Error(err))
	}
}
