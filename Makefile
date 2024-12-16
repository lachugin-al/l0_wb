APP_NAME = l0_wb
CMD_DIR = ./cmd/app
SEED_FILE = internal/db/migrations/seed.sql
SEED_COUNT = 10
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: all build run test clean run-seed seed

all: build

build:
	@echo ">>> Building the application..."
	go build -o $(APP_NAME) $(CMD_DIR)

# Запустить приложение
run: build
	@echo ">>> Running the application..."
	rm -f $(SEED_FILE)
	./$(APP_NAME)

# Запустить приложение с генерацией seed.sql и заполнением тестовыми данными
run--seed: build
	@echo ">>> Generating seed data and running the application..."
	rm -f $(SEED_FILE)
	./$(APP_NAME) --seed --seed-file=$(SEED_FILE) --seed-count=$(SEED_COUNT)

test:
	@echo ">>> Running tests..."
	go test -v ./...

clean:
	@echo ">>> Cleaning up..."
	rm -f $(APP_NAME)