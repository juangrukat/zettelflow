.PHONY: build run help

BINARY_NAME=zettelflow
CMD_PATH=./cmd/zettelflow
OUTPUT_DIR=./bin

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(OUTPUT_DIR)
	@go build -o $(OUTPUT_DIR)/$(BINARY_NAME) $(CMD_PATH)

run: build
	@echo "Running $(BINARY_NAME)..."
	@$(OUTPUT_DIR)/$(BINARY_NAME)

help:
	@echo "Available commands:"
	@echo "  make build    - Build the application"
	@echo "  make run      - Build and run the application"
	@echo "  make help     - Show this help message"

# Default target
all: build
