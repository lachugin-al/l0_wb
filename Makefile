APP_NAME = l0_wb
CMD_DIR = ./cmd/app
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: all build run test clean run-seed seed lint

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