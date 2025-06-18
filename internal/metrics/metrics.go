// Package metrics provides metrics.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"runtime"
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

	// RPS (Requests Per Second) - счетчик запросов в секунду
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// TPS (Transactions Per Second) - счетчик транзакций в секунду
	TransactionsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "transactions_total",
			Help: "Total number of database transactions",
		},
	)

	// QPS (Queries Per Second) - счетчик запросов к БД в секунду
	QueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	// DBQueryDuration - время выполнения запросов к БД
	DBQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// ResponseTime - время ответа HTTP запросов
	HTTPResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_time_seconds",
			Help:    "HTTP response time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// ErrorRate - процент ошибок
	ErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "operation"},
	)

	// Traffic - сетевой трафик
	NetworkTrafficBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "network_traffic_bytes_total",
			Help: "Total network traffic in bytes",
		},
		[]string{"direction"}, // in, out
	)

	// CPU Usage
	CPUUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "cpu_usage_percent",
			Help: "Current CPU usage in percent",
		},
	)

	// Memory Usage
	MemoryUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_usage_bytes",
			Help: "Current memory usage in bytes",
		},
	)

	// Disk Usage
	DiskUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "disk_usage_bytes",
			Help: "Current disk usage in bytes",
		},
		[]string{"device", "mountpoint"},
	)

	// Uptime
	Uptime = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "uptime_seconds_total",
			Help: "Total uptime in seconds",
		},
	)

	// Queue Sizes
	QueueSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_size",
			Help: "Current size of the queue",
		},
		[]string{"queue_name"},
	)

	// Goroutines Count
	GoroutinesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "goroutines_count",
			Help: "Current number of goroutines",
		},
	)
)

// Init инициализирует метрики и регистрирует их в Prometheus.
func Init() {
	// Регистрация существующих метрик
	prometheus.MustRegister(OrdersProcessed)
	prometheus.MustRegister(OrderProcessingTime)
	prometheus.MustRegister(OrderProcessingErrors)

	// Регистрация новых метрик
	prometheus.MustRegister(RequestsTotal)
	prometheus.MustRegister(TransactionsTotal)
	prometheus.MustRegister(QueriesTotal)
	prometheus.MustRegister(DBQueryDuration)
	prometheus.MustRegister(HTTPResponseTime)
	prometheus.MustRegister(ErrorsTotal)
	prometheus.MustRegister(NetworkTrafficBytes)
	prometheus.MustRegister(CPUUsage)
	prometheus.MustRegister(MemoryUsage)
	prometheus.MustRegister(DiskUsage)
	prometheus.MustRegister(Uptime)
	prometheus.MustRegister(QueueSize)
	prometheus.MustRegister(GoroutinesCount)

	// Запуск обновления метрик времени работы
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Обновление Uptime
			Uptime.Add(1)

			// Обновление количества горутин
			GoroutinesCount.Set(float64(runtime.NumGoroutine()))

			// Обновление использования памяти
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			MemoryUsage.Set(float64(memStats.Alloc))

			// Здесь можно добавить обновление других метрик, требующих периодического обновления
		}
	}()
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

// RecordHTTPRequest записывает метрику HTTP запроса
func RecordHTTPRequest(method, endpoint string, status int, duration time.Duration) {
	RequestsTotal.WithLabelValues(method, endpoint, string(rune(status))).Inc()
	HTTPResponseTime.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordDBQuery записывает метрику запроса к базе данных
func RecordDBQuery(operation, table string, duration time.Duration) {
	QueriesTotal.WithLabelValues(operation, table).Inc()
	DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordTransaction записывает метрику транзакции
func RecordTransaction() {
	TransactionsTotal.Inc()
}

// RecordError записывает метрику ошибки
func RecordError(errorType, operation string) {
	ErrorsTotal.WithLabelValues(errorType, operation).Inc()
}

// RecordNetworkTraffic записывает метрику сетевого трафика
func RecordNetworkTraffic(direction string, bytes int) {
	NetworkTrafficBytes.WithLabelValues(direction).Add(float64(bytes))
}

// SetQueueSize устанавливает текущий размер очереди
func SetQueueSize(queueName string, size int) {
	QueueSize.WithLabelValues(queueName).Set(float64(size))
}
