package repository

import (
	"database/sql"

	"l0_wb/internal/model"
)

// PaymentsRepository определяет методы для взаимодействия с таблицей 'payments'.
type PaymentsRepository interface {
	Insert(payment *model.Payment, orderUID string) error
	GetByOrderID(orderUID string) (*model.Payment, error)
}

type paymentsRepository struct {
	db *sql.DB
}

// NewPaymentsRepository создает новый экземпляр PaymentsRepository.
//
//	Параметры:
//	- db: подключение к базе данных (sql.DB).
//	Возвращает:
//	- PaymentsRepository: экземпляр интерфейса для взаимодействия с таблицей 'payments'.
func NewPaymentsRepository(db *sql.DB) PaymentsRepository {
	return &paymentsRepository{db: db}
}

// Insert добавляет новую запись о платеже в таблицу 'payments'.
//
//	Параметры:
//	- payment: объект model.Payment, содержащий данные о платеже.
//	- orderUID: уникальный идентификатор заказа.
//	Возвращает:
//	- error: ошибка при выполнении запроса (если возникла).
func (r *paymentsRepository) Insert(payment *model.Payment, orderUID string) error {
	query := `INSERT INTO payments (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.Exec(query,
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
	return err
}

// GetByOrderID получает запись о платеже по order_uid.
//
//	Параметры:
//	- orderUID: уникальный идентификатор заказа.
//	Возвращает:
//	- *model.Payment: объект платежа, если запись найдена.
//	- error: ошибка при выполнении запроса (если возникла) или sql.ErrNoRows, если запись не найдена.
func (r *paymentsRepository) GetByOrderID(orderUID string) (*model.Payment, error) {
	query := `SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
              FROM payments WHERE order_uid = $1`
	row := r.db.QueryRow(query, orderUID)
	var p model.Payment
	err := row.Scan(
		&p.Transaction,
		&p.RequestID,
		&p.Currency,
		&p.Provider,
		&p.Amount,
		&p.PaymentDt,
		&p.Bank,
		&p.DeliveryCost,
		&p.GoodsTotal,
		&p.CustomFee,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
