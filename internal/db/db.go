// Package db provides database service.
package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"l0_wb/internal/config"
	"l0_wb/internal/util"
)

// InitDB инициализирует подключение к базе данных, используя переданную конфигурацию.
//
//	Возвращает:
//	- *pgxpool.Pool: пул соединений к базе данных.
//	- error: ошибка, если не удалось установить подключение.
func InitDB(cfg *config.Config) (*pgxpool.Pool, error) {
	logger := util.GetLogger()

	// Формируем строку подключения на основе конфигурации.
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	// Настройки пула соединений.
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Error("Failed to parse DB config", zap.Error(err))
		return nil, fmt.Errorf("failed to parse DB config: %w", err)
	}

	poolConfig.MaxConns = 500                       // Максимальное количество соединений
	poolConfig.MinConns = 100                       // Минимальное количество соединений
	poolConfig.HealthCheckPeriod = 30 * time.Second // Проверка соединений раз в 30 сек
	poolConfig.MaxConnLifetime = 5 * time.Minute    // Соединения живут не более 5 минут
	poolConfig.MaxConnIdleTime = 1 * time.Minute    // Простой соединения не больше 1 минуты

	// Создаем пул соединений
	dbPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		logger.Error("Failed to create DB pool", zap.Error(err))
		return nil, fmt.Errorf("failed to create DB pool: %w", err)
	}

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := dbPool.Ping(ctx); err != nil {
		logger.Error("Failed to ping DB", zap.Error(err))
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	// Выполняем миграции.
	if err := runMigrations(dbPool); err != nil {
		logger.Error("Failed to run migrations", zap.Error(err))
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database initialized successfully")
	return dbPool, nil
}

// runMigrations выполняет SQL-миграции.
//
//	Данный метод опционален и зависит от потребностей проекта.
//	Возвращает:
//	- error: ошибка, если не удалось применить миграции.
func runMigrations(db *pgxpool.Pool) error {
	logger := util.GetLogger()
	migrationsDir := "internal/db/migrations"

	logger.Info("Starting database migrations", zap.String("migrationsDir", migrationsDir))

	// Check if migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		logger.Error("Migrations directory does not exist", zap.String("dir", migrationsDir))
		// Try alternative path
		migrationsDir = "./internal/db/migrations"
		logger.Info("Trying alternative migrations path", zap.String("migrationsDir", migrationsDir))
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			logger.Error("Alternative migrations directory does not exist", zap.String("dir", migrationsDir))
			return fmt.Errorf("migrations directory does not exist: %s", migrationsDir)
		}
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		logger.Error("Failed to find migration files", zap.Error(err), zap.String("pattern", filepath.Join(migrationsDir, "*.sql")))
		return fmt.Errorf("failed to find migration files: %w", err)
	}

	logger.Info("Found migration files", zap.Int("count", len(files)), zap.Strings("files", files))

	// Получаем абсолютный путь к папке миграций.
	absDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		logger.Error("Failed to resolve absolute path for migrations directory", zap.String("dir", migrationsDir), zap.Error(err))
		return fmt.Errorf("failed to resolve absolute path for migrations directory %s: %w", migrationsDir, err)
	}

	logger.Info("Resolved absolute path for migrations directory", zap.String("absDir", absDir))

	for _, file := range files {
		// Очистка пути для предотвращения path traversal атак
		cleanFile := filepath.Clean(file)
		logger.Info("Processing migration file", zap.String("file", cleanFile))

		absFile, err := filepath.Abs(cleanFile)
		if err != nil {
			logger.Error("Failed to resolve absolute path", zap.String("file", cleanFile), zap.Error(err))
			return fmt.Errorf("failed to resolve absolute path for %s: %w", cleanFile, err)
		}
		logger.Info("Resolved absolute path", zap.String("absFile", absFile))

		// Проверяем, что файл действительно находится в папке миграций
		relPath, err := filepath.Rel(absDir, absFile)
		if err != nil {
			logger.Error("Failed to get relative path", zap.String("absDir", absDir), zap.String("absFile", absFile), zap.Error(err))
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		if relPath == "" || relPath[0] == '.' {
			logger.Error("File is outside the migrations directory", zap.String("file", absFile), zap.String("relPath", relPath))
			return fmt.Errorf("file %s is outside the migrations directory", absFile)
		}
		logger.Info("File is inside migrations directory", zap.String("relPath", relPath))

		logger.Info("Applying migration", zap.String("file", absFile))

		// Check if file exists and is readable
		fileInfo, err := os.Stat(absFile)
		if err != nil {
			logger.Error("Failed to stat migration file", zap.String("file", absFile), zap.Error(err))
			return fmt.Errorf("failed to stat migration file %s: %w", absFile, err)
		}
		logger.Info("File stats", zap.String("file", absFile), zap.Int64("size", fileInfo.Size()), zap.String("mode", fileInfo.Mode().String()))

		// Безопасное чтение файла
		//nolint:gosec // Безопасность гарантирована проверкой пути выше
		content, err := os.ReadFile(absFile)
		if err != nil {
			logger.Error("Failed to read migration file", zap.String("file", absFile), zap.Error(err))
			return fmt.Errorf("failed to read migration file %s: %w", absFile, err)
		}
		logger.Info("Read migration file content", zap.String("file", absFile), zap.Int("contentLength", len(content)))

		// Выполняем SQL-запрос
		logger.Info("Executing SQL migration", zap.String("file", absFile))
		if _, err := db.Exec(context.Background(), string(content)); err != nil {
			logger.Error("Failed to execute migration", zap.String("file", absFile), zap.Error(err))
			return fmt.Errorf("failed to execute migration %s: %w", absFile, err)
		}
		logger.Info("Successfully executed migration", zap.String("file", absFile))
	}

	logger.Info("All migrations applied successfully")
	return nil
}
