package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

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

	for _, file := range files {
		log.Printf("Applying migration: %s", file)
		content, err := os.ReadFile(file)
		if err != nil {
			logger.Error("Failed to read migration file", zap.String("file", file), zap.Error(err))
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			logger.Error("Failed to execute migration", zap.String("file", file), zap.Error(err))
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}

	logger.Info("All migrations applied successfully")
	return nil
}
