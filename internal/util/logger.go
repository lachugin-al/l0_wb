package util

import (
	"go.uber.org/zap"
)

var logger *zap.Logger

// InitLogger инициализирует глобальный логгер.
//
//	Возвращает:
//	- error: если не удалось создать логгер.
func InitLogger() error {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		return err
	}
	return nil
}

// GetLogger возвращает глобальный логгер.
//
// Возвращает:
//   - *zap.Logger: указатель на глобальный логгер.
func GetLogger() *zap.Logger {
	if logger == nil {
		panic("Logger is not initialized. Call InitLogger first.")
	}
	return logger
}

// SyncLogger завершает работу логгера, очищая буфер.
//
//	Примечание: SyncLogger должен вызываться в main.go через defer после инициализации логгера.
func SyncLogger() {
	if logger != nil {
		_ = logger.Sync()
	}
}
