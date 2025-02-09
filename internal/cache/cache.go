// Package cache provides in memory cache service.
package cache

import (
	"context"
	"database/sql"
	"log"
	"sync"

	"go.uber.org/zap"
	"l0_wb/internal/model"
	"l0_wb/internal/repository"
	"l0_wb/internal/util"
)

// OrderCache представляет собой кэш для хранения заказов в памяти.
type OrderCache struct {
	mu     sync.RWMutex            // Мьютекс для синхронизации доступа к кэшу
	cache  map[string]*model.Order // Словарь, где ключ — order_uid, значение — объект заказа
	logger *zap.Logger
}

// NewOrderCache создает новый пустой кэш заказов.
//
//	Возвращает:
//	- *OrderCache: экземпляр кэша.
func NewOrderCache() *OrderCache {
	return &OrderCache{
		cache:  make(map[string]*model.Order),
		logger: util.GetLogger(),
	}
}

// LoadFromDB загружает все заказы из базы данных в кэш.
//
//	Этот метод рекомендуется вызывать при старте приложения после инициализации БД.
//	Параметры:
//	- ctx: контекст выполнения.
//	- ordersRepo: репозиторий для работы с таблицей orders.
//	- deliveriesRepo: репозиторий для работы с таблицей deliveries.
//	- paymentsRepo: репозиторий для работы с таблицей payments.
//	- itemsRepo: репозиторий для работы с таблицей items.
//	- db: подключение к базе данных.
//	Возвращает:
//	- error: ошибку, если произошел сбой при загрузке данных из БД.
func (c *OrderCache) LoadFromDB(
	ctx context.Context,
	ordersRepo repository.OrdersRepository,
	deliveriesRepo repository.DeliveriesRepository,
	paymentsRepo repository.PaymentsRepository,
	itemsRepo repository.ItemsRepository,
	db *sql.DB,
) error {
	c.logger.Info("Starting to load orders into cache")
	// TODO если нет возможности получить все order_uid из БД, реализовать метод GetAllOrderIDs() из ordersRepo

	// Получаем список всех order_uid из БД
	orderUIDs, err := getAllOrderUIDs(db)
	if err != nil {
		c.logger.Error("Failed to fetch order UIDs from database", zap.Error(err))
		return err
	}
	c.logger.Info("Fetched order UIDs", zap.Int("count", len(orderUIDs)))

	// Загружаем полный заказ для каждого order_uid и сохраняем в кэш
	for _, uid := range orderUIDs {
		o, err := loadFullOrder(ctx, uid, ordersRepo, deliveriesRepo, paymentsRepo, itemsRepo)
		if err != nil {
			c.logger.Warn("Failed to load order", zap.String("order_uid", uid), zap.Error(err))
			continue
		}
		c.mu.Lock()
		c.cache[uid] = o
		c.mu.Unlock()
	}

	c.logger.Info("Finished loading orders into cache", zap.Int("cached_orders", len(c.cache)))
	return nil
}

// Get возвращает заказ из кэша по его order_uid.
//
//	Параметры:
//	- orderUID: уникальный идентификатор заказа.
//	Возвращает:
//	- *model.Order: объект заказа (nil, если не найден).
func (c *OrderCache) Get(orderUID string) *model.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order := c.cache[orderUID]
	if order == nil {
		c.logger.Warn("Order not found in cache", zap.String("order_uid", orderUID))
	}
	return order
}

// Set добавляет или обновляет заказ в кэше.
//
//	Параметры:
//	- order: объект заказа.
func (c *OrderCache) Set(order *model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[order.OrderUID] = order
	c.logger.Info("Order added to cache", zap.String("order_uid", order.OrderUID))
}

// loadFullOrder загружает полный заказ из базы данных, включая связанные данные (доставка, оплата, товары).
//
//	Параметры:
//	- ctx: контекст выполнения.
//	- orderUID: уникальный идентификатор заказа.
//	- ordersRepo: репозиторий для работы с таблицей orders.
//	- deliveriesRepo: репозиторий для работы с таблицей deliveries.
//	- paymentsRepo: репозиторий для работы с таблицей payments.
//	- itemsRepo: репозиторий для работы с таблицей items.
//	Возвращает:
//	- *model.Order: заполненный объект заказа.
//	- error: ошибку, если не удалось загрузить данные.
func loadFullOrder(
	_ context.Context,
	orderUID string,
	ordersRepo repository.OrdersRepository,
	deliveriesRepo repository.DeliveriesRepository,
	paymentsRepo repository.PaymentsRepository,
	itemsRepo repository.ItemsRepository,
) (*model.Order, error) {
	o, err := ordersRepo.GetByID(orderUID)
	if err != nil {
		return nil, err
	}

	d, err := deliveriesRepo.GetByOrderID(orderUID)
	if err != nil {
		return nil, err
	}
	p, err := paymentsRepo.GetByOrderID(orderUID)
	if err != nil {
		return nil, err
	}
	it, err := itemsRepo.GetByOrderID(orderUID)
	if err != nil {
		return nil, err
	}

	o.Delivery = *d
	o.Payment = *p
	o.Items = it
	return o, nil
}

// GetAll возвращает список всех заказов, хранящихся в кэше.
//
//	Возвращает:
//	- []model.Order: список всех заказов.
func (c *OrderCache) GetAll() []*model.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()

	orders := make([]*model.Order, 0, len(c.cache))
	for _, order := range c.cache {
		orders = append(orders, order)
	}

	c.logger.Info("Fetched all orders from cache", zap.Int("count", len(orders)))
	return orders
}

// getAllOrderUIDs возвращает список всех order_uid из таблицы orders.
//
//	Параметры:
//	- db: подключение к базе данных.
//	Возвращает:
//	- []string: список order_uid.
//	- error: ошибку, если не удалось выполнить запрос.
func getAllOrderUIDs(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT order_uid FROM orders`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	var uids []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		uids = append(uids, uid)
	}
	return uids, rows.Err()
}
