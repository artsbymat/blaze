APP_NAME := blaze
CMD_PATH := ./cmd/ssg
BINARY := ./bin/$(APP_NAME)

.PHONY: all build run serve build-site clean

all: build

build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(BINARY) $(CMD_PATH)/main.go
	@echo "Done."

run: build
	@$(BINARY)

serve: build
	@$(BINARY) serve

build-site: build
	@$(BINARY) build

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY)
