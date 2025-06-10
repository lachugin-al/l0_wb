APP_NAME = l0_wb
CMD_DIR = ./cmd/app
COMPOSE_BUILD_FLAG =

.PHONY: all build run test clean lint docker-build docker-run docker-compose docker-compose-rebuild docker-compose-down

all: build

build:
	@echo ">>> Building the application..."
	go build -o $(APP_NAME) $(CMD_DIR)

run: build
	@echo ">>> Running the application..."
	./$(APP_NAME)

test:
	@echo ">>> Running tests..."
	go test -v ./...

lint:
	@echo ">>> Running linters..."
	golangci-lint run --timeout=5m

clean:
	@echo ">>> Cleaning up..."
	rm -f $(APP_NAME)

docker-compose:
	@echo ">>> Starting all services with Docker Compose..."
	docker-compose up -d $(COMPOSE_BUILD_FLAG)

docker-compose-rebuild: COMPOSE_BUILD_FLAG = --build
docker-compose-rebuild: docker-compose

docker-compose-down:
	@echo ">>> Stopping all services..."
	docker-compose down
