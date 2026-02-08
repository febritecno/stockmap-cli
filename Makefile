.PHONY: all build install clean test run

APP_NAME := stockmap
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-X 'github.com/febritecno/stockmap/cmd.version=$(VERSION)'"

all: build

build:
	@echo "Building $(APP_NAME)..."
	go build $(LDFLAGS) -o $(APP_NAME) main.go

install:
	@echo "Installing $(APP_NAME)..."
	go install $(LDFLAGS) .

clean:
	@echo "Cleaning..."
	rm -f $(APP_NAME)
	rm -rf release/

test:
	go test -v ./...

run: build
	./$(APP_NAME)
