# L0 WB

### Демонстрационный сервис с простейшим интерфейсом, отображающий данные о заказе:
 - Принимает сообщения о заказах из Kafka (топик orders).
 - Валидирует и сохраняет данные о заказе в PostgreSQL.
 - Кэширует полученные данные в памяти для быстрого доступа.
 - При рестарте сервиса восстанавливает кэш из базы данных.
 - Предоставляет HTTP-эндпоинт для получения данных заказа по order_uid.
 - Предоставляет простой веб-интерфейс для просмотра заказа.

### Структура проекта
```
l0_wb/
├── cmd/
│   └── app/
│       └── main.go            # Точка входа в приложение
│
├── internal/
│   ├── config/                # Конфиг
│   ├── db/                    # Подключение к БД и миграции
│   ├── model/                 # Модели данных
│   ├── repository/            # Репозитории для работы с БД
│   ├── service/               # Бизнес-логика (SaveOrder, GetOrderByID)
│   ├── cache/                 # Кэширование заказов in-memory
│   ├── kafka/                 # Консьюмер для чтения сообщений из Kafka
│   ├── server/                # HTTP-сервер и роуты
│   └── tools/                 # Скрипты и консольные утилиты
│
├── web/                       # Статический контент (index.html)
├── docker-compose.yml
├── Makefile
├── go.mod
└── go.sum
```

#### Модель данных:
```json
{
   "order_uid": "b563feb7b2b84b6test",
   "track_number": "WBILMTESTTRACK",
   "entry": "WBIL",
   "delivery": {
      "name": "Test Testov",
      "phone": "+9720000000",
      "zip": "2639809",
      "city": "Kiryat Mozkin",
      "address": "Ploshad Mira 15",
      "region": "Kraiot",
      "email": "test@gmail.com"
   },
   "payment": {
      "transaction": "b563feb7b2b84b6test",
      "request_id": "",
      "currency": "USD",
      "provider": "wbpay",
      "amount": 1817,
      "payment_dt": 1637907727,
      "bank": "alpha",
      "delivery_cost": 1500,
      "goods_total": 317,
      "custom_fee": 0
   },
   "items": [
      {
         "chrt_id": 9934930,
         "track_number": "WBILMTESTTRACK",
         "price": 453,
         "rid": "ab4219087a764ae0btest",
         "name": "Mascaras",
         "sale": 30,
         "size": "0",
         "total_price": 317,
         "nm_id": 2389212,
         "brand": "Vivienne Sabo",
         "status": 202
      }
   ],
   "locale": "en",
   "internal_signature": "",
   "customer_id": "test",
   "delivery_service": "meest",
   "shardkey": "9",
   "sm_id": 99,
   "date_created": "2021-11-26T06:22:19Z",
   "oof_shard": "1"
}
```

### Запуск проекта
1. Запустите зависимости (PostgreSQL, Zookeeper, Kafka, Kafka UI):
```bash
docker-compose up -d
```
2. Если необходимо сгенерировать тестовые данные для миграции, выполните:
```bash
go run internal/tools/db/order-seeder.go
```
3. Соберите и запустите приложение:
```bash
make run
```
4. Приложение готово к демонстрации:
   - Применит миграции к БД.
   - Загрузит кэш из БД.
   - Запустит Kafka-консьюмер для чтения новых заказов.
   - Поднимет HTTP-сервер по адресу http://localhost:8081.

### Проверка работы приложения
- Перейти в браузере по адресу http://localhost:8081. Отобразится простая страница для ввода order_uid.
- При наличии тестовых данных в БД они будут загружены в кэш, введите order_uid тестового заказа и нажмите “Показать”. Должны отобразиться данные в JSON.
- Чтобы проверить поступление новых данных из Kafka:
  - Открыть Kafka UI по адресу http://localhost:8080.
  - Отправить тестовое сообщение в топик orders.
  - После обработки сообщения сервис сохранит заказ в БД, добавит в кэш.
  - Повторно запросить order_uid через веб-интерфейс или выполнить в терминале команду:
  ```bash
  curl http://localhost:8081/order/<order_uid>
  ```
  - Для генерации и отправки тестового сообщения в kafka, выполните:
  ```bash
  go run internal/tools/kafka/producer.go
  ```

### Тестирование
- Для запуска unit тестов, выполните:
```bash
make test
```
- Для проведения stress тестирования, выполните:
```bash
go run internal/tools/tester/stress_tester.go
```

### Завершение работы
Для остановки сервисов и остановки приложения:
```bash
docker-compose down
```