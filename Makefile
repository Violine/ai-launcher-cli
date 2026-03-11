# AI Launcher CLI — build for multiple platforms
BINARY_NAME := ai-launcher
MAIN_PKG   := ./cmd/ai-launcher
OUT_DIR    := dist

# Version for release (optional; set via: make build-all VERSION=1.0.0)
VERSION ?= dev

.PHONY: build build-all clean build-darwin-arm64 build-darwin-amd64 build-linux-amd64 build-windows-amd64

# Build for current OS/arch only
build:
	go build -o $(OUT_DIR)/$(BINARY_NAME) $(MAIN_PKG)

# Cross-compile: macOS Apple Silicon
build-darwin-arm64:
	@mkdir -p $(OUT_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(OUT_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PKG)

# Cross-compile: macOS Intel
build-darwin-amd64:
	@mkdir -p $(OUT_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(OUT_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PKG)

# Cross-compile: Linux amd64
build-linux-amd64:
	@mkdir -p $(OUT_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(OUT_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PKG)

# Cross-compile: Windows amd64
build-windows-amd64:
	@mkdir -p $(OUT_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(OUT_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PKG)

# Build all platforms (macOS arm64 + amd64, Linux, Windows)
build-all:
	@mkdir -p $(OUT_DIR)
	$(MAKE) build-darwin-arm64 build-darwin-amd64 build-linux-amd64 build-windows-amd64
	@echo "Built in $(OUT_DIR)/:"
	@ls -la $(OUT_DIR)/

clean:
	rm -rf $(OUT_DIR)
