// Package db provides database service.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	// Importing postgres driver for database/sql
	_ "github.com/lib/pq"

	"go.uber.org/zap"
	"l0_wb/internal/config"
	"l0_wb/internal/util"
)

// InitDB инициализирует подключение к базе данных, используя переданную конфигурацию.
//
//	Возвращает:
//	- *sql.DB: экземпляр подключения к базе данных.
//	- error: ошибка, если не удалось установить подключение.
func InitDB(cfg *config.Config) (*sql.DB, error) {
	logger := util.GetLogger()

	// Формируем строку подключения на основе конфигурации.
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	// Открываем подключение к БД.
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error("Failed to open DB connection", zap.Error(err))
		return nil, fmt.Errorf("failed to open DB connection: %w", err)
	}

	// Проверяем подключение.
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping DB", zap.Error(err))
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	// Выполняем миграции.
	if err := runMigrations(db); err != nil {
		logger.Error("Failed to run migrations", zap.Error(err))
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database initialized successfully")
	return db, nil
}

// runMigrations выполняет SQL-миграции.
//
//	Данный метод опционален и зависит от потребностей проекта.
//	Возвращает:
//	- error: ошибка, если не удалось применить миграции.
func runMigrations(db *sql.DB) error {
	logger := util.GetLogger()
	migrationsDir := "internal/db/migrations"

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		logger.Error("Failed to find migration files", zap.Error(err))
		return fmt.Errorf("failed to find migration files: %w", err)
	}

	// Получаем абсолютный путь к папке миграций.
	absDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		logger.Error("Failed to resolve absolute path for migrations directory", zap.String("dir", migrationsDir), zap.Error(err))
		return fmt.Errorf("failed to resolve absolute path for migrations directory %s: %w", migrationsDir, err)
	}

	for _, file := range files {
		// Очистка пути для предотвращения path traversal атак
		cleanFile := filepath.Clean(file)
		absFile, err := filepath.Abs(cleanFile)
		if err != nil {
			logger.Error("Failed to resolve absolute path", zap.String("file", cleanFile), zap.Error(err))
			return fmt.Errorf("failed to resolve absolute path for %s: %w", cleanFile, err)
		}

		// Проверяем, что файл действительно находится в папке миграций
		relPath, err := filepath.Rel(absDir, absFile)
		if err != nil || relPath == "" || relPath[0] == '.' {
			logger.Error("File is outside the migrations directory", zap.String("file", absFile))
			return fmt.Errorf("file %s is outside the migrations directory", absFile)
		}

		logger.Info("Applying migration", zap.String("file", absFile))

		// Безопасное чтение файла
		//nolint:gosec // Безопасность гарантирована проверкой пути выше
		content, err := os.ReadFile(absFile)
		if err != nil {
			logger.Error("Failed to read migration file", zap.String("file", absFile), zap.Error(err))
			return fmt.Errorf("failed to read migration file %s: %w", absFile, err)
		}

		// Выполняем SQL-запрос
		if _, err := db.Exec(string(content)); err != nil {
			logger.Error("Failed to execute migration", zap.String("file", absFile), zap.Error(err))
			return fmt.Errorf("failed to execute migration %s: %w", absFile, err)
		}
	}

	logger.Info("All migrations applied successfully")
	return nil
}
