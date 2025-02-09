// Package service provides order service.
package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"l0_wb/internal/model"
	"l0_wb/internal/repository"
	"l0_wb/internal/util"
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
	logger         *zap.Logger
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
	logger := util.GetLogger()
	return &orderService{
		db:             db,
		ordersRepo:     ordersRepo,
		deliveriesRepo: deliveriesRepo,
		paymentsRepo:   paymentsRepo,
		itemsRepo:      itemsRepo,
		logger:         logger,
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
	// Валидация заказа перед сохранением
	if err := s.validateOrder(order); err != nil {
		return err
	}

	// Устанавливаем дату создания заказа, если она не указана
	if order.DateCreated.IsZero() {
		order.DateCreated = time.Now().UTC()
	}

	// Открываем транзакцию
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("SaveOrder: begin transaction failed", zap.Error(err))
		return fmt.Errorf("begin transaction failed: %w", err)
	}

	// Отложенный откат транзакции при ошибке
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Вставка данных заказа
	if err = s.insertOrderData(tx, order); err != nil {
		return err
	}

	// Подтверждаем транзакцию
	if err = tx.Commit(); err != nil {
		s.logger.Error("SaveOrder: commit transaction failed", zap.String("orderUID", order.OrderUID), zap.Error(err))
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	s.logger.Info("SaveOrder: order saved successfully", zap.String("orderUID", order.OrderUID))
	return nil
}

// validateOrder выполняет базовую валидацию заказа.
//
// Параметры:
// - order: объект заказа.
//
// Возвращает:
// - error: если заказ некорректен.
func (s *orderService) validateOrder(order *model.Order) error {
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
	return nil
}

// insertOrderData выполняет вставку данных заказа в базу данных в рамках транзакции.
//
// Параметры:
// - tx: активная транзакция базы данных.
// - order: объект заказа.
//
// Возвращает:
// - error: если произошла ошибка при вставке.
func (s *orderService) insertOrderData(tx *sql.Tx, order *model.Order) error {
	if err := s.ordersRepoInsertTx(tx, order); err != nil {
		return err
	}

	if err := s.deliveriesRepoInsertTx(tx, &order.Delivery, order.OrderUID); err != nil {
		return err
	}

	if err := s.paymentsRepoInsertTx(tx, &order.Payment, order.OrderUID); err != nil {
		return err
	}

	if err := s.itemsRepoInsertTx(tx, order.Items, order.OrderUID); err != nil {
		return err
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
func (s *orderService) GetOrderByID(_ context.Context, orderUID string) (*model.Order, error) {
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
