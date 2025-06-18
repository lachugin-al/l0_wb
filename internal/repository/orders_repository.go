package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"l0_wb/internal/model"
)

// OrdersRepository определяет методы для взаимодействия с таблицей 'orders'.
type OrdersRepository interface {
	Insert(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, orderUID string) (*model.Order, error)
}

type ordersRepository struct {
	db      *pgxpool.Pool
	metrics *MetricsWrapper
}

// NewOrdersRepository создает новый экземпляр OrdersRepository.
//
//	Параметры:
//	- db: подключение к базе данных (sql.DB).
//	Возвращает:
//	- OrdersRepository: экземпляр интерфейса для взаимодействия с таблицей 'orders'.
func NewOrdersRepository(db *pgxpool.Pool) OrdersRepository {
	return &ordersRepository{
		db:      db,
		metrics: NewMetricsWrapper(),
	}
}

// Insert добавляет новую запись о заказе в таблицу 'orders'.
//
//	Параметры:
//	- order: объект model.Order, представляющий данные заказа.
//	Возвращает:
//	- error: ошибка при выполнении запроса (если возникла).
func (r *ordersRepository) Insert(ctx context.Context, order *model.Order) error {
	return r.metrics.RecordDBOperation(ctx, "insert", "orders", true, func(ctx context.Context) error {
		query := `INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

		_, err := r.db.Exec(ctx, query,
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
		return err
	})
}

// GetByID получает запись о заказе по его order_uid.
//
//	Параметры:
//	- orderUID: уникальный идентификатор заказа.
//	Возвращает:
//	- *model.Order: объект заказа, если запись найдена.
//	- error: ошибка при выполнении запроса (если возникла) или sql.ErrNoRows, если запись не найдена.
func (r *ordersRepository) GetByID(ctx context.Context, orderUID string) (*model.Order, error) {
	var order *model.Order

	err := r.metrics.RecordDBOperation(ctx, "select", "orders", false, func(ctx context.Context) error {
		query := `SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
              FROM orders WHERE order_uid = $1`

		row := r.db.QueryRow(ctx, query, orderUID)
		var o model.Order
		var dateCreated time.Time

		err := row.Scan(
			&o.OrderUID,
			&o.TrackNumber,
			&o.Entry,
			&o.Locale,
			&o.InternalSignature,
			&o.CustomerID,
			&o.DeliveryService,
			&o.Shardkey,
			&o.SmID,
			&dateCreated,
			&o.OofShard,
		)
		if err != nil {
			return err
		}

		o.DateCreated = dateCreated
		order = &o
		return nil
	})

	return order, err
}
