// Package repository provides repository service.
package repository

import (
	"database/sql"

	"l0_wb/internal/model"
)

// DeliveriesRepository определяет методы для взаимодействия с таблицей 'deliveries'.
type DeliveriesRepository interface {
	Insert(delivery *model.Delivery, orderUID string) error
	GetByOrderID(orderUID string) (*model.Delivery, error)
}

type deliveriesRepository struct {
	db *sql.DB
}

// NewDeliveriesRepository создает новый экземпляр DeliveriesRepository.
//
//	Параметры:
//	- db: подключение к базе данных (sql.DB).
//	Возвращает:
//	- DeliveriesRepository: экземпляр интерфейса для взаимодействия с таблицей 'deliveries'.
func NewDeliveriesRepository(db *sql.DB) DeliveriesRepository {
	return &deliveriesRepository{db: db}
}

// Insert добавляет новую запись о доставке в таблицу 'deliveries'.
//
//	Параметры:
//	- delivery: объект доставки, содержащий данные о получателе.
//	- orderUID: уникальный идентификатор заказа.
//	Возвращает:
//	- error: ошибка при выполнении запроса (если возникла).
func (r *deliveriesRepository) Insert(delivery *model.Delivery, orderUID string) error {
	query := `INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(query,
		orderUID,
		delivery.Name,
		delivery.Phone,
		delivery.Zip,
		delivery.City,
		delivery.Address,
		delivery.Region,
		delivery.Email,
	)
	return err
}

// GetByOrderID получает запись о доставке по order_uid.
//
//	Параметры:
//	- orderUID: уникальный идентификатор заказа.
//	Возвращает:
//	- *model.Delivery: объект доставки, если запись найдена.
//	- error: ошибка при выполнении запроса (если возникла) или sql.ErrNoRows, если запись не найдена.
func (r *deliveriesRepository) GetByOrderID(orderUID string) (*model.Delivery, error) {
	query := `SELECT name, phone, zip, city, address, region, email
              FROM deliveries WHERE order_uid = $1`
	row := r.db.QueryRow(query, orderUID)
	var d model.Delivery
	err := row.Scan(&d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
