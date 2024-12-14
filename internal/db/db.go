package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	"l0_wb/internal/config"
)

// InitDB инициализирует подключение к базе данных, используя переданную конфигурацию.
//
//	Возвращает:
//	- *sql.DB: экземпляр подключения к базе данных.
//	- error: ошибка, если не удалось установить подключение.
func InitDB(cfg *config.Config) (*sql.DB, error) {
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
		return nil, fmt.Errorf("failed to open DB connection: %w", err)
	}

	// Проверяем подключение.
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	// Выполняем миграции.
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations выполняет SQL-миграции.
//
//	Данный метод опционален и зависит от потребностей проекта.
//	Возвращает:
//	- error: ошибка, если не удалось применить миграции.
func runMigrations(db *sql.DB) error {
	migrationsDir := "internal/db/migrations"

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to find migration files: %w", err)
	}

	for _, file := range files {
		log.Printf("Applying migration: %s", file)
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}

	return nil
}
