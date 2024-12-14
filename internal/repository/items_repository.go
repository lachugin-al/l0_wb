package repository

import (
	"database/sql"

	"l0_wb/internal/model"
)

// ItemsRepository определяет методы для взаимодействия с таблицей 'items'.
type ItemsRepository interface {
	Insert(items []model.Item, orderUID string) error
	GetByOrderID(orderUID string) ([]model.Item, error)
}

type itemsRepository struct {
	db *sql.DB
}

// NewItemsRepository создает новый экземпляр ItemsRepository.
//
//	Параметры:
//	- db: подключение к базе данных (sql.DB).
//	Возвращает:
//	- ItemsRepository: экземпляр интерфейса для взаимодействия с таблицей 'items'.
func NewItemsRepository(db *sql.DB) ItemsRepository {
	return &itemsRepository{db: db}
}

// Insert добавляет несколько записей о товарах в таблицу 'items'.
//
//	Параметры:
//	- items: массив объектов model.Item, представляющих товары.
//	- orderUID: уникальный идентификатор заказа.
//	Возвращает:
//	- error: ошибка при выполнении запроса (если возникла).
func (r *itemsRepository) Insert(items []model.Item, orderUID string) error {
	query := `INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	for _, it := range items {
		_, err := r.db.Exec(query,
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
			return err
		}
	}
	return nil
}

// GetByOrderID получает все записи о товарах, связанных с указанным order_uid.
//
//	Параметры:
//	- orderUID: уникальный идентификатор заказа.
//	Возвращает:
//	- []model.Item: массив объектов товаров, если записи найдены.
//	- error: ошибка при выполнении запроса (если возникла).
func (r *itemsRepository) GetByOrderID(orderUID string) ([]model.Item, error) {
	query := `SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
              FROM items WHERE order_uid = $1`
	rows, err := r.db.Query(query, orderUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Item
	for rows.Next() {
		var it model.Item
		err := rows.Scan(
			&it.ChrtID,
			&it.TrackNumber,
			&it.Price,
			&it.Rid,
			&it.Name,
			&it.Sale,
			&it.Size,
			&it.TotalPrice,
			&it.NmID,
			&it.Brand,
			&it.Status,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}
