APP_NAME = okane
BUILD_DIR = build

.PHONY: all build build-linux clean

all: build

build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/...

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/...

clean:
	rm -rf $(BUILD_DIR)
