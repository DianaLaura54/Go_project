.PHONY: all build test clean run-server run-pomodoro

BINARY_DIR=bin

all: build

build:
	@mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_DIR)/pomodoro ./cmd/pomodoro
	go build -o $(BINARY_DIR)/server    ./cmd/server
	@echo " Built: bin/pomodoro, bin/server"

test:
	go test ./... -v

vet:
	go vet ./...

clean:
	rm -rf $(BINARY_DIR)

run-server:
	go run ./cmd/server --port 8080

run-pomodoro:
	go run ./cmd/pomodoro --work 25 --short 5 --long 15 --sessions 4
