# L0 WB

### Demonstration Service with a Simple Interface to Display Order Data:
- Accepts order messages from Kafka (topic: `orders`).
- Validates and saves order data to PostgreSQL.
- Caches received data in memory for quick access.
- Restores the cache from the database after a service restart.
- Provides an HTTP endpoint to fetch order data by `order_uid`.
- Offers a simple web interface to view order details.

### Project Structure
```
l0_wb/
├── cmd/
│   └── app/
│       └── main.go            # Application entry point
│
├── internal/
│   ├── config/                # Configuration
│   ├── db/                    # Database connection and migrations
│   ├── model/                 # Data models
│   ├── repository/            # Database repositories
│   ├── service/               # Business logic (SaveOrder, GetOrderByID)
│   ├── cache/                 # In-memory order caching
│   ├── kafka/                 # Kafka consumer and producer modules
│   ├── server/                # HTTP server and routes
│   ├── tools/                 # Utility scripts
│   └── util/                  # Utilities
│
├── web/                       # Static content (index.html)
├── docker-compose.yml
├── Makefile
├── go.mod
└── go.sum
```

#### Data Model:
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

### Running the Project
1. Start dependencies (PostgreSQL, Zookeeper, Kafka, Kafka UI):
```bash
  docker-compose up -d
```
2. Build and run the application:
```bash
  make run
```
3. The application is ready for demonstration:
    - Applies database migrations.
    - Loads the cache from the database.
    - Starts the Kafka consumer to read new orders.
    - Starts the HTTP server at `http://localhost:8081`.

### Verifying Application Functionality
- Open the browser and navigate to `http://localhost:8081`. You will see a simple page to input an `order_uid`.
- If test data exists in the database, it will be loaded into the cache. Enter the `order_uid` of a test order and click "Show" to view the data in JSON format.
- To test incoming data from Kafka, you can:
    - Use Kafka UI at `http://localhost:8080`:
        - Send a test message to the `orders` topic.
        - After processing the message, the service will save the order to the database and add it to the cache.
        - Retrieve the `order_uid` via the web interface or execute the following command:
      ```bash
        curl http://localhost:8081/order/<order_uid>
      ```
    - Use the internal tools:
        - Generate and send a test message to Kafka by running:
      ```bash
        go run internal/tools/kafka/producer.go
      ```
The static UI at `http://localhost:8081` provides the following features:
1. **Search for Orders**: Enter an `order_uid` and click the "Show" button to retrieve and display order details in JSON format.
2. **Send Test Order**: Click the "Send Test Order" button to generate and send a test order to the Kafka topic. The order is processed and displayed in the list.
3. **View All Orders**: Click the "Show Orders" button to display a list of all cached orders with their respective `order_uid`.

### Testing
- To run unit tests, execute:
```bash
  make test
```
- For stress testing, use the script:
```bash
  go run internal/tools/ht/stress_tester.go -url=http://localhost:8081/order/<order_uid> -rate=1000 -duration=10
```

### Shutting Down
To stop the services and the application:
```bash
  docker-compose down
```


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
│   ├── tools/                 # Скрипты
│   └── util/                  # Утилиты
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
2. Соберите и запустите приложение:
```bash
  make run
```
3. Приложение готово к демонстрации:
    - Применит миграции к БД.
    - Загрузит кэш из БД.
    - Запустит Kafka-консьюмер для чтения новых заказов.
    - Поднимет HTTP-сервер по адресу http://localhost:8081.

### Проверка работы приложения
- Перейти в браузере по адресу http://localhost:8081. Отобразится простая страница для ввода order_uid.
- При наличии тестовых данных в БД они будут загружены в кэш, введите order_uid тестового заказа и нажмите “Показать”. Должны отобразиться данные в JSON.
- Чтобы проверить поступление новых данных из Kafka можно следующими способами:
    - При помощи Kafka UI по адресу http://localhost:8080.
        - Отправить тестовое сообщение в топик orders.
        - После обработки сообщения сервис сохранит заказ в БД, добавит в кэш.
        - Повторно запросить order_uid через веб-интерфейс или выполнить в терминале команду:
      ```bash
        curl http://localhost:8081/order/<order_uid>
      ```
    - При помощи internal/tools
        - Для генерации и отправки тестового сообщения в kafka, выполните:
      ```bash
        go run internal/tools/kafka/producer.go
      ```

### Тестирование
- Для запуска unit тестов, выполните:
```bash
  make test
```
- Для проведения stress тестирования, можно воспользоваться скриптом:
```
  go run internal/tools/ht/stress_tester.go -url=http://localhost:8081/order/<order_uid> -rate=1000 -duration=10
```

### Завершение работы
Для остановки сервисов и остановки приложения:
```bash
  docker-compose down
```