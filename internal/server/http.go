package server

import (
	"context"
	"encoding/json"
	"net/http"
	"path"
	"strings"
	"time"

	"l0_wb/internal/cache"
)

// Server представляет HTTP-сервер для работы с заказами.
type Server struct {
	httpServer *http.Server
	cache      *cache.OrderCache
	staticDir  string
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
	s := &Server{
		cache:     orderCache,
		staticDir: staticDir,
	}
	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return s
}

// registerRoutes регистрирует маршруты HTTP для обработки запросов.
//
//	Параметры:
//	- mux: HTTP маршрутизатор (ServeMux).
func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Маршрут для получения заказа по ID
	mux.HandleFunc("/order/", s.handleGetOrderByID)

	// Статический контент (index.html)
	if s.staticDir != "" {
		mux.HandleFunc("/", s.handleStatic)
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
	if orderID == "" {
		http.Error(w, "order id is required", http.StatusBadRequest)
		return
	}

	order := s.cache.Get(orderID)
	if order == nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
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
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Остановка сервера по сигналу из контекста
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		return err // Ошибка запуска сервера
	}
}