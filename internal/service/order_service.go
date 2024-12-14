package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"l0_wb/internal/model"
	"l0_wb/internal/repository"
)

// OrderService определяет бизнес-логику для работы с заказами.
type OrderService interface {
	SaveOrder(ctx context.Context, order *model.Order) error

	GetOrderByID(ctx context.Context, orderUID string) (*model.Order, error)
}

// orderService является конкретной реализацией интерфейса OrderService.
type orderService struct {
	db             *sql.DB
	ordersRepo     repository.OrdersRepository
	deliveriesRepo repository.DeliveriesRepository
	paymentsRepo   repository.PaymentsRepository
	itemsRepo      repository.ItemsRepository
}

// NewOrderService создает новый экземпляр orderService.
//
//	Параметры:
//	- db: подключение к базе данных.
//	- ordersRepo: репозиторий для работы с таблицей заказов.
//	- deliveriesRepo: репозиторий для работы с таблицей доставок.
//	- paymentsRepo: репозиторий для работы с таблицей оплат.
//	- itemsRepo: репозиторий для работы с таблицей товаров.
//	Возвращает:
//	- OrderService: экземпляр сервиса для работы с заказами.
func NewOrderService(
	db *sql.DB,
	ordersRepo repository.OrdersRepository,
	deliveriesRepo repository.DeliveriesRepository,
	paymentsRepo repository.PaymentsRepository,
	itemsRepo repository.ItemsRepository,
) OrderService {
	return &orderService{
		db:             db,
		ordersRepo:     ordersRepo,
		deliveriesRepo: deliveriesRepo,
		paymentsRepo:   paymentsRepo,
		itemsRepo:      itemsRepo,
	}
}

// SaveOrder сохраняет заказ в рамках одной транзакции базы данных.
//
//	Этапы:
//	1. Валидация структуры заказа (проверка order_uid, списка товаров и данных доставки).
//	2. Вставка данных в таблицы orders, deliveries, payments, items.
//	3. Завершение транзакции (commit) при успешной вставке всех данных.
//	Параметры:
//	- ctx: контекст выполнения.
//	- order: объект заказа.
//	Возвращает:
//	- error: ошибка, если произошел сбой на любом этапе.
func (s *orderService) SaveOrder(ctx context.Context, order *model.Order) error {
	if order == nil {
		return errors.New("order is nil")
	}
	if order.OrderUID == "" {
		return errors.New("order_uid is empty")
	}
	if len(order.Items) == 0 {
		return errors.New("order has no items")
	}
	if order.Delivery.Name == "" || order.Delivery.Phone == "" {
		return errors.New("invalid delivery data")
	}

	// Если дата создания заказа не указана, устанавливаем текущее время
	if order.DateCreated.IsZero() {
		order.DateCreated = time.Now().UTC()
	}

	// Начало транзакции
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}

	// Вставка данных в таблицы order
	if err := s.ordersRepoInsertTx(tx, order); err != nil {
		tx.Rollback()
		return err
	}

	// Вставка данных в таблицы delivery
	if err := s.deliveriesRepoInsertTx(tx, &order.Delivery, order.OrderUID); err != nil {
		tx.Rollback()
		return err
	}

	// Вставка данных в таблицы payment
	if err := s.paymentsRepoInsertTx(tx, &order.Payment, order.OrderUID); err != nil {
		tx.Rollback()
		return err
	}

	// Вставка данных в таблицы items
	if err := s.itemsRepoInsertTx(tx, order.Items, order.OrderUID); err != nil {
		tx.Rollback()
		return err
	}

	// Завершаем транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	return nil
}

// GetOrderByID получает заказ и сопутствующие данные из базы данных и возвращает заполненную структуру Order.
//
//	Параметры:
//	- ctx: контекст выполнения.
//	- orderUID: уникальный идентификатор заказа.
//	Возвращает:
//	- *model.Order: объект заказа.
//	- error: ошибка, если произошел сбой на любом этапе.
func (s *orderService) GetOrderByID(ctx context.Context, orderUID string) (*model.Order, error) {
	order, err := s.ordersRepo.GetByID(orderUID)
	if err != nil {
		return nil, err
	}

	delivery, err := s.deliveriesRepo.GetByOrderID(orderUID)
	if err != nil {
		return nil, err
	}

	payment, err := s.paymentsRepo.GetByOrderID(orderUID)
	if err != nil {
		return nil, err
	}

	items, err := s.itemsRepo.GetByOrderID(orderUID)
	if err != nil {
		return nil, err
	}

	order.Delivery = *delivery
	order.Payment = *payment
	order.Items = items

	return order, nil
}

// ordersRepoInsertTx вставляет заказ в таблицу orders с использованием транзакции (tx).
func (s *orderService) ordersRepoInsertTx(tx *sql.Tx, order *model.Order) error {
	query := `INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := tx.Exec(query,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("insert order failed: %w", err)
	}
	return nil
}

// deliveriesRepoInsertTx вставляет данные доставки в таблицу deliveries с использованием транзакции (tx).
func (s *orderService) deliveriesRepoInsertTx(tx *sql.Tx, delivery *model.Delivery, orderUID string) error {
	query := `INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := tx.Exec(query,
		orderUID,
		delivery.Name,
		delivery.Phone,
		delivery.Zip,
		delivery.City,
		delivery.Address,
		delivery.Region,
		delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("insert delivery failed: %w", err)
	}
	return nil
}

// paymentsRepoInsertTx вставляет данные оплаты в таблицу payments с использованием транзакции (tx).
func (s *orderService) paymentsRepoInsertTx(tx *sql.Tx, payment *model.Payment, orderUID string) error {
	query := `INSERT INTO payments (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := tx.Exec(query,
		orderUID,
		payment.Transaction,
		payment.RequestID,
		payment.Currency,
		payment.Provider,
		payment.Amount,
		payment.PaymentDt,
		payment.Bank,
		payment.DeliveryCost,
		payment.GoodsTotal,
		payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("insert payment failed: %w", err)
	}
	return nil
}

// itemsRepoInsertTx вставляет товары в таблицу items с использованием транзакции (tx).
func (s *orderService) itemsRepoInsertTx(tx *sql.Tx, items []model.Item, orderUID string) error {
	query := `INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	for _, it := range items {
		_, err := tx.Exec(query,
			orderUID,
			it.ChrtID,
			it.TrackNumber,
			it.Price,
			it.Rid,
			it.Name,
			it.Sale,
			it.Size,
			it.TotalPrice,
			it.NmID,
			it.Brand,
			it.Status,
		)
		if err != nil {
			return fmt.Errorf("insert item failed: %w", err)
		}
	}
	return nil
}
