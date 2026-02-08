.PHONY: all build install uninstall clean test run

APP_NAME := stockmap
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-X 'github.com/febritecno/stockmap-cli/cmd.version=$(VERSION)'"

# Default Go build command
GOBUILD := go build

# If on Termux (linux/aarch64), disable CGO
ifeq ($(shell uname -s)-$(shell uname -m), Linux-aarch64)
	GOBUILD := CGO_ENABLED=0 go build
endif

all: build

build:
	@echo "Building $(APP_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(APP_NAME) main.go

install:
	@echo "Installing $(APP_NAME)..."
	go install $(LDFLAGS) .

uninstall:
	@echo "Uninstalling $(APP_NAME)..."
	rm -f $(GOPATH)/bin/$(APP_NAME)
	rm -f $(GOPATH)/bin/$(APP_NAME)-cli
	rm -f /usr/local/bin/$(APP_NAME)
	rm -f ~/.local/bin/$(APP_NAME)

clean:
	@echo "Cleaning..."
	rm -f $(APP_NAME)
	rm -rf release/

test:
	go test -v ./...

run: build
	./$(APP_NAME)
