.PHONY: build clean test all

BINARY_NAME=globe-nmea-server
BUILD_DIR=build

all: clean build

build:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux cmd/server/main.go
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-mac cmd/server/main.go
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME).exe cmd/server/main.go

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)