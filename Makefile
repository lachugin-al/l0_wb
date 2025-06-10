APP_NAME = l0_wb
CMD_DIR = ./cmd/app
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: all build run test clean run-seed seed lint docker-build docker-run docker-compose-up docker-compose-down

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

docker-build:
	@echo ">>> Building Docker image..."
	docker build -t $(APP_NAME) .

docker-run: docker-build
	@echo ">>> Running application in Docker container..."
	docker run -p 8081:8081 -p 9100:9100 --name $(APP_NAME) $(APP_NAME)

docker-compose-up:
	@echo ">>> Starting all services with Docker Compose..."
	docker-compose up -d

docker-compose-rebuild:
	@echo ">>> Rebuilding and restarting all services with Docker Compose..."
	docker-compose up -d --build

docker-compose-down:
	@echo ">>> Stopping all services..."
	docker-compose down
