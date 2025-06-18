// Package server provides http server.
package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"go.uber.org/zap"
	"l0_wb/internal/cache"
	"l0_wb/internal/kafka"
	"l0_wb/internal/metrics"
	"l0_wb/internal/util"
)

// Server представляет HTTP-сервер для работы с заказами.
type Server struct {
	httpServer *http.Server
	cache      *cache.OrderCache
	staticDir  string
	logger     *zap.Logger
}

// NewServer создаёт новый экземпляр Server.
//
//	Параметры:
//	- port: порт, на котором будет работать сервер.
//	- orderCache: кэш для доступа к заказам.
//	- staticDir: директория для статических файлов (например, index.html).
//	Возвращает:
//	- *Server: экземпляр HTTP-сервера.
func NewServer(port string, orderCache *cache.OrderCache, staticDir string) *Server {
	logger := util.GetLogger()

	s := &Server{
		cache:     orderCache,
		staticDir: staticDir,
		logger:    logger,
	}

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.httpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	logger.Info("HTTP server initialized", zap.String("port", port))
	return s
}

// metricsMiddleware оборачивает HTTP-обработчик для сбора метрик.
//
//	Параметры:
//	- next: следующий обработчик в цепочке.
//	- endpoint: имя эндпоинта для метрик.
//	Возвращает:
//	- http.HandlerFunc: обработчик с метриками.
func (s *Server) metricsMiddleware(next http.HandlerFunc, endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Создаем ResponseWriter, который отслеживает статус ответа
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // По умолчанию 200 OK
		}

		// Вызываем следующий обработчик
		next(rw, r)

		// Записываем метрики
		duration := time.Since(startTime)
		metrics.RecordHTTPRequest(r.Method, endpoint, rw.statusCode, duration)

		// Если произошла ошибка (статус >= 400), записываем ее
		if rw.statusCode >= 400 {
			metrics.RecordError("http", endpoint)
		}

		// Записываем размер ответа как исходящий трафик
		metrics.RecordNetworkTraffic("out", rw.bytesWritten)

		// Оцениваем размер запроса как входящий трафик
		contentLength := r.ContentLength
		if contentLength > 0 {
			metrics.RecordNetworkTraffic("in", int(contentLength))
		}
	}
}

// responseWriter оборачивает http.ResponseWriter для отслеживания статуса ответа и размера.
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

// WriteHeader переопределяет метод для отслеживания статуса ответа.
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write переопределяет метод для отслеживания размера ответа.
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// registerRoutes регистрирует маршруты HTTP для обработки запросов.
//
//	Параметры:
//	- mux: HTTP маршрутизатор (ServeMux).
func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Маршрут для получения заказа по ID
	mux.HandleFunc("/order/", s.metricsMiddleware(s.handleGetOrderByID, "/order/{id}"))
	mux.HandleFunc("/api/orders", s.metricsMiddleware(s.handleGetOrders, "/api/orders"))
	mux.HandleFunc("/api/send-test-order", s.metricsMiddleware(s.handleSendTestOrder, "/api/send-test-order"))

	// Health check endpoint
	mux.HandleFunc("/health", s.metricsMiddleware(s.handleHealth, "/health"))
	s.logger.Info("Health check endpoint registered")

	// Статический контент (index.html)
	if s.staticDir != "" {
		mux.HandleFunc("/", s.metricsMiddleware(s.handleStatic, "/static"))
		s.logger.Info("Static content route registered", zap.String("staticDir", s.staticDir))
	}
}

// handleGetOrderByID обрабатывает запросы вида: GET /order/{id}.
//
//	Возвращает заказ с указанным ID, если он есть в кэше.
//	Если ID отсутствует или не найден, возвращается ошибка 404 или 400.
//	Параметры:
//	- w: HTTP-ответ.
//	- r: HTTP-запрос.
func (s *Server) handleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	// Удаляем префикс "/order/" чтобы получить {id}
	orderID := strings.TrimPrefix(r.URL.Path, "/order/")
	s.logger.Info("Received order request", zap.String("orderID", orderID))

	if orderID == "" {
		http.Error(w, "order id is required", http.StatusBadRequest)
		s.logger.Warn("Order ID is missing in request")
		return
	}

	order := s.cache.Get(orderID)
	if order == nil {
		http.Error(w, "order not found", http.StatusNotFound)
		s.logger.Warn("Order not found", zap.String("orderID", orderID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		s.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// handleGetOrders возвращает список всех заказов из кэша.
func (s *Server) handleGetOrders(w http.ResponseWriter, _ *http.Request) {
	s.logger.Info("Received request to fetch all orders")

	orders := s.cache.GetAll()
	if len(orders) == 0 {
		http.Error(w, "no orders available", http.StatusNotFound)
		s.logger.Warn("No orders found in cache")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		s.logger.Error("Failed to encode orders response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// handleSendTestOrder отправляет тестовый заказ в Kafka.
func (s *Server) handleSendTestOrder(w http.ResponseWriter, _ *http.Request) {
	s.logger.Info("Received request to send test order")

	orderUID, err := kafka.ProduceTestMessage()
	if err != nil {
		s.logger.Error("Failed to send test order", zap.Error(err))
		http.Error(w, "failed to send test order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Test order sent successfully! Order UID: " + orderUID)); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

// handleHealth обрабатывает запросы к эндпоинту /health.
//
//	Возвращает 200 OK, если сервер работает нормально.
//	Параметры:
//	- w: HTTP-ответ.
//	- r: HTTP-запрос.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	s.logger.Info("Health check requested")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		s.logger.Error("Failed to write health check response", zap.Error(err))
	}
}

// handleStatic раздаёт статические файлы из s.staticDir.
//
//	Если запрашивается "/", возвращается "index.html".
//	Параметры:
//	- w: HTTP-ответ.
//	- r: HTTP-запрос.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path
	if filePath == "/" {
		filePath = "/index.html"
	}
	fp := path.Join(s.staticDir, filePath)

	s.logger.Info("Serving static file", zap.String("filePath", fp))
	http.ServeFile(w, r, fp)
}

// Start запускает сервер и блокируется до завершения работы.
//
//	Сервер завершает работу при отмене переданного контекста или при возникновении ошибки.
//	Параметры:
//	- ctx: контекст выполнения.
//	Возвращает:
//	- error: ошибку, если сервер не удалось запустить или корректно завершить.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("HTTP server is starting")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Failed to shut down server", zap.Error(err))
			return err
		}
		s.logger.Info("HTTP server shut down gracefully")
		return nil
	case err := <-errCh:
		s.logger.Error("HTTP server encountered an error", zap.Error(err))
		return err
	}
}
