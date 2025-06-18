package repository

import (
	"context"
	"time"

	"l0_wb/internal/metrics"
)

// MetricsWrapper предоставляет способ записи метрик для операций с базой данных.
// Может использоваться для записи метрик TPS и QPS.
type MetricsWrapper struct{}

// NewMetricsWrapper создает новый экземпляр MetricsWrapper.
func NewMetricsWrapper() *MetricsWrapper {
	return &MetricsWrapper{}
}

// RecordDBOperation записывает метрики для операции с базой данных.
//
//	Измеряет продолжительность операции и записывает ее как метрику QPS.
//	Если операция является транзакцией, также записывает ее как метрику TPS.
//
//	Параметры:
//	- ctx: контекст для операции
//	- operation: тип операции (например, "select", "insert", "update", "delete")
//	- table: таблица, с которой выполняется операция
//	- isTransaction: является ли эта операция частью транзакции
//	- fn: функция для выполнения и измерения
//	Возвращает:
//	- error: любая ошибка, возвращаемая функцией
func (mw *MetricsWrapper) RecordDBOperation(
	ctx context.Context,
	operation string,
	table string,
	isTransaction bool,
	fn func(ctx context.Context) error,
) error {
	startTime := time.Now()

	// Выполнить операцию
	err := fn(ctx)

	// Записать продолжительность
	duration := time.Since(startTime)

	// Записать метрику QPS
	metrics.RecordDBQuery(operation, table, duration)

	// Если это транзакция, записать метрику TPS
	if isTransaction {
		metrics.RecordTransaction()
	}

	// Если произошла ошибка, записать ее
	if err != nil {
		metrics.RecordError("database", operation+":"+table)
	}

	return err
}
