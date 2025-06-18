# L0 WB

### Demonstration Service with a Simple Interface to Display Order Data:
- Accepts order messages from Kafka (topic: `orders`).
- Validates and saves order data to PostgreSQL.
- Caches received data in memory for quick access.
- Restores the cache from the database after a service restart.
- Provides an HTTP endpoint to fetch order data by `order_uid`.
- Offers a simple web interface to view order details.
- Collects and exposes Prometheus metrics for monitoring.
- Integrates with Grafana for visualization.

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
│   ├── metrics/               # Prometheus metrics collection
│   ├── tools/                 # Utility scripts
│   └── util/                  # Utilities
│
├── web/                       # Static content (index.html)
├── docker-compose.yml
├── prometheus.yml             # Prometheus configuration
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
1. Start all services including the application using Docker Compose:
```bash
  docker-compose up -d
```
2. The application is ready for demonstration:

> **Note:** If database migrations are not running properly in Docker, use `make docker-compose-rebuild` to rebuild the Docker image with the latest changes. This ensures that migration files are properly included in the image.
    - Applies database migrations.
    - Loads the cache from the database.
    - Starts the Kafka consumer to read new orders.
    - Starts the HTTP server at `http://localhost:8081`.
    - Exposes Prometheus metrics at http://localhost:9100/metrics.

Note: If you prefer to run the application outside of Docker for development purposes, you can:
1. Start only the dependencies:
```bash
  docker-compose up -d postgres zookeeper kafka kafka-ui prometheus grafana postgres-exporter kafka-exporter
```
2. Build and run the application locally:
```bash
  make run
```

### Monitoring with Prometheus and Grafana
#### Prometheus
- Available at: http://localhost:9090
- Automatically scrapes:
  - http://localhost:9100/metrics (Application metrics)
  - http://localhost:9187/metrics (PostgreSQL metrics)
  - http://localhost:9308/metrics (Kafka metrics)

#### Grafana
- Available at: http://localhost:3000
- Default login:
  - Username: admin
  - Password: admin (can be changed in .env)
- Preconfigured dashboards:
  - **Application Metrics**: Displays general application metrics including orders processed, HTTP requests, response times, CPU and memory usage, goroutines count, and uptime.
  - **Database Query Performance**: Shows database-related metrics such as query rates, transaction rates, queries by table, and PostgreSQL statistics.
  - **Kafka Consumer Lag**: Monitors Kafka consumer lag, message processing rates, and broker metrics to ensure efficient message processing.

To access the dashboards:
1. Open http://localhost:3000 in your browser
2. Log in with the credentials above
3. Click on the "Dashboards" icon in the left sidebar
4. Select one of the preconfigured dashboards from the list

The dashboards are automatically provisioned when Grafana starts, so no manual setup is required.

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

### Makefile Commands
The project includes several Makefile targets to simplify development and deployment:

#### Standard Commands
- `make build`: Builds the application locally
- `make run`: Builds and runs the application locally
- `make test`: Runs all tests
- `make lint`: Runs linters
- `make clean`: Removes the compiled binary

#### Docker Commands
- `make docker-build`: Builds the Docker image for the application
- `make docker-run`: Builds and runs the application in a Docker container
- `make docker-compose-up`: Starts all services using Docker Compose (equivalent to `docker-compose up -d`)
- `make docker-compose-rebuild`: Rebuilds and restarts all services (equivalent to `docker-compose up -d --build`)
- `make docker-compose-down`: Stops all services (equivalent to `docker-compose down`)

### Environment Variables
Variables are loaded from .env:
```
POSTGRES_USER=<*****>
POSTGRES_PASSWORD=<*****>
POSTGRES_DB=<*****>
```

# L0 WB

### Демонстрационный сервис с простейшим интерфейсом, отображающий данные о заказе:
- Принимает сообщения о заказах из Kafka (топик orders).
- Валидирует и сохраняет данные о заказе в PostgreSQL.
- Кэширует полученные данные в памяти для быстрого доступа.
- При рестарте сервиса восстанавливает кэш из базы данных.
- Предоставляет HTTP-эндпоинт для получения данных заказа по order_uid.
- Предоставляет простой веб-интерфейс для просмотра заказа.
- Собирает и отображает метрики Prometheus для мониторинга.
- Интегрируется с Grafana для визуализации.

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
│   ├── metrics/               # Prometheus метрики
│   ├── tools/                 # Скрипты
│   └── util/                  # Утилиты
│
├── web/                       # Статический контент (index.html)
├── docker-compose.yml
├── prometheus.yml             # Prometheus конфигурация
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
1. Запустите все сервисы, включая приложение, используя Docker Compose:
```bash
  docker-compose up -d
```
2. Приложение готово к демонстрации:

> **Примечание:** Если миграции базы данных не запускаются корректно в Docker, используйте `make docker-compose-rebuild` для пересборки Docker-образа с последними изменениями. Это гарантирует, что файлы миграций правильно включены в образ.
    - Применит миграции к БД.
    - Загрузит кэш из БД.
    - Запустит Kafka-консьюмер для чтения новых заказов.
    - Поднимет HTTP-сервер по адресу http://localhost:8081.
    - Предоставит метрики Prometheus по адресу http://localhost:9100/metrics.

Примечание: Если вы предпочитаете запускать приложение вне Docker для целей разработки, вы можете:
1. Запустить только зависимости:
```bash
  docker-compose up -d postgres zookeeper kafka kafka-ui prometheus grafana postgres-exporter kafka-exporter
```
2. Собрать и запустить приложение локально:
```bash
  make run
```

### Мониторинг с помощью Prometheus и Grafana
#### Prometheus
- Доступно по адресу: http://localhost:9090.
- Автоматически скрапирует:
    - http://localhost:9100/metrics (метрики приложений)
    - http://localhost:9187/metrics (метрики PostgreSQL)
    - http://localhost:9308/metrics (метрики Kafka)

#### Grafana
- Доступна по адресу: http://localhost:3000
- Логин по умолчанию:
    - Имя пользователя: admin
    - Пароль: admin (может быть изменен в .env)
- Предварительно настроенные панели мониторинга:
    - **Application Metrics**: Отображает общие метрики приложения, включая обработанные заказы, HTTP-запросы, время отклика, использование CPU и памяти, количество горутин и время работы.
    - **Database Query Performance**: Показывает метрики, связанные с базой данных, такие как скорость запросов, скорость транзакций, запросы по таблицам и статистику PostgreSQL.
    - **Kafka Consumer Lag**: Мониторит отставание потребителей Kafka, скорость обработки сообщений и метрики брокера для обеспечения эффективной обработки сообщений.

Для доступа к панелям мониторинга:
1. Откройте http://localhost:3000 в браузере
2. Войдите, используя учетные данные выше
3. Нажмите на значок "Dashboards" в левой боковой панели
4. Выберите одну из предварительно настроенных панелей мониторинга из списка

Панели мониторинга автоматически настраиваются при запуске Grafana, поэтому ручная настройка не требуется.

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

### Команды Makefile
Проект включает несколько целей Makefile для упрощения разработки и развертывания:

#### Стандартные команды
- `make build`: Собирает приложение локально
- `make run`: Собирает и запускает приложение локально
- `make test`: Запускает все тесты
- `make lint`: Запускает линтеры
- `make clean`: Удаляет скомпилированный бинарный файл

#### Docker команды
- `make docker-build`: Собирает Docker-образ для приложения
- `make docker-run`: Собирает и запускает приложение в Docker-контейнере
- `make docker-compose-up`: Запускает все сервисы с помощью Docker Compose (эквивалент `docker-compose up -d`)
- `make docker-compose-rebuild`: Пересобирает и перезапускает все сервисы (эквивалент `docker-compose up -d --build`)
- `make docker-compose-down`: Останавливает все сервисы (эквивалент `docker-compose down`)
