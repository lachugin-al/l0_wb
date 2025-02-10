// Package metrics provides metrics.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"sync"
	"time"
)

// Метрики сервиса
var (
	// OrdersProcessed считает общее количество обработанных заказов.
	OrdersProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_processed_total",
			Help: "Total number of processed orders",
		},
	)

	// OrderProcessingTime измеряет время обработки заказа (гистограмма).
	OrderProcessingTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "order_processing_duration_seconds",
			Help:    "Histogram of order processing times",
			Buckets: prometheus.LinearBuckets(0.01, 0.05, 10),
		},
	)

	// OrderProcessingErrors считает общее количество ошибок при обработке заказов.
	OrderProcessingErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "order_processing_errors_total",
			Help: "Total number of order processing errors",
		},
	)
)

// Init инициализирует метрики и регистрирует их в Prometheus.
func Init() {
	prometheus.MustRegister(OrdersProcessed)
	prometheus.MustRegister(OrderProcessingTime)
	prometheus.MustRegister(OrderProcessingErrors)
}

// StartMetricsServer запускает HTTP-сервер для экспорта метрик Prometheus.
func StartMetricsServer(port string, wg *sync.WaitGroup) {
	http.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic("Failed to start metrics server: " + err.Error())
		}
	}()
}
