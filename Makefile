# AI Launcher CLI — build for multiple platforms
BINARY_NAME := ai-launcher
MAIN_PKG   := ./cmd/ai-launcher
OUT_DIR    := dist
RELEASE_DIR := release
VERSION    ?= dev

# LDFLAGS injects version into binary (autoupdate.Version)
LDFLAGS := -ldflags "-X github.com/ai-launcher/cli/internal/modules/autoupdate.Version=$(VERSION)"

.PHONY: build build-all clean release build-darwin-arm64 build-darwin-amd64 build-linux-amd64 build-windows-amd64

# Build for current OS/arch only
build:
	go build $(LDFLAGS) -o $(OUT_DIR)/$(BINARY_NAME) $(MAIN_PKG)

# Cross-compile: macOS Apple Silicon
build-darwin-arm64:
	@mkdir -p $(OUT_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(OUT_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PKG)

# Cross-compile: macOS Intel
build-darwin-amd64:
	@mkdir -p $(OUT_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(OUT_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PKG)

# Cross-compile: Linux amd64
build-linux-amd64:
	@mkdir -p $(OUT_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(OUT_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PKG)

# Cross-compile: Windows amd64
build-windows-amd64:
	@mkdir -p $(OUT_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(OUT_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PKG)

# Build all platforms (macOS arm64 + amd64, Linux, Windows)
build-all:
	@mkdir -p $(OUT_DIR)
	$(MAKE) build-darwin-arm64 build-darwin-amd64 build-linux-amd64 build-windows-amd64
	@echo "Built in $(OUT_DIR)/:"
	@ls -la $(OUT_DIR)/

# Build for release: build-all then copy to release/ with names for GitHub (underscores).
# Use: make release VERSION=1.0.0
release: build-all
	@mkdir -p $(RELEASE_DIR)
	cp $(OUT_DIR)/$(BINARY_NAME)-darwin-arm64   $(RELEASE_DIR)/$(BINARY_NAME)_darwin_arm64
	cp $(OUT_DIR)/$(BINARY_NAME)-darwin-amd64   $(RELEASE_DIR)/$(BINARY_NAME)_darwin_amd64
	cp $(OUT_DIR)/$(BINARY_NAME)-linux-amd64   $(RELEASE_DIR)/$(BINARY_NAME)_linux_amd64
	cp $(OUT_DIR)/$(BINARY_NAME)-windows-amd64.exe $(RELEASE_DIR)/$(BINARY_NAME)_windows_amd64.exe
	@echo "Release artifacts in $(RELEASE_DIR)/:"
	@ls -la $(RELEASE_DIR)/

clean:
	rm -rf $(OUT_DIR) $(RELEASE_DIR)
