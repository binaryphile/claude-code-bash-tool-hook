.PHONY: all build test clean install uninstall linux darwin-amd64 darwin-arm64 windows coverage

BINARY_NAME=claude-code-bash-tool-hook
VERSION=1.0.0
BIN_DIR=bin

# Default target
all: build

# Build for current platform
build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) -ldflags="-s -w" .

# Run tests
test:
	go test -v ./...

# Run tests with coverage
coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	go tool cover -func=coverage.out | grep total

# Clean build artifacts
clean:
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html

# Install symlink to ~/.claude/hooks/
install: build
	@mkdir -p ~/.claude/hooks
	@if [ -L ~/.claude/hooks/$(BINARY_NAME) ]; then rm ~/.claude/hooks/$(BINARY_NAME); fi
	@if [ -f ~/.claude/hooks/$(BINARY_NAME) ]; then echo "Error: ~/.claude/hooks/$(BINARY_NAME) exists and is not a symlink"; exit 1; fi
	@ln -s $(CURDIR)/$(BIN_DIR)/$(BINARY_NAME) ~/.claude/hooks/$(BINARY_NAME)
	@echo "Symlinked $(BIN_DIR)/$(BINARY_NAME) -> ~/.claude/hooks/$(BINARY_NAME)"
	@echo ""
	@echo "To enable the hook, add this to ~/.claude/settings.json:"
	@echo '{'
	@echo '  "hooks": {'
	@echo '    "PreToolUse": [{'
	@echo '      "matcher": "Bash",'
	@echo '      "hooks": [{'
	@echo '        "type": "command",'
	@echo '        "command": "~/.claude/hooks/$(BINARY_NAME)",'
	@echo '        "timeout": 5'
	@echo '      }]'
	@echo '    }]'
	@echo '  }'
	@echo '}'

# Uninstall symlink
uninstall:
	@rm -f ~/.claude/hooks/$(BINARY_NAME)
	@echo "Removed ~/.claude/hooks/$(BINARY_NAME)"

# Cross-compilation targets

linux: linux-amd64

linux-amd64:
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 -ldflags="-s -w" .
	@echo "Built: $(BIN_DIR)/$(BINARY_NAME)-linux-amd64"

darwin: darwin-amd64 darwin-arm64

darwin-amd64:
	@mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 -ldflags="-s -w" .
	@echo "Built: $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64"

darwin-arm64:
	@mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 -ldflags="-s -w" .
	@echo "Built: $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64"

windows: windows-amd64

windows-amd64:
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe -ldflags="-s -w" .
	@echo "Built: $(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe"

# Build all platforms
all-platforms: linux-amd64 darwin-amd64 darwin-arm64 windows-amd64
	@echo ""
	@echo "All platform builds complete:"
	@ls -lh $(BIN_DIR)/$(BINARY_NAME)-*

# Run benchmarks
benchmark: build
	@echo "Running startup benchmark..."
	@./benchmark-hook.sh

# Help target
help:
	@echo "Available targets:"
	@echo "  make build          - Build for current platform (to bin/)"
	@echo "  make test           - Run tests"
	@echo "  make coverage       - Run tests with coverage report"
	@echo "  make install        - Symlink bin/binary to ~/.claude/hooks/"
	@echo "  make uninstall      - Remove symlink from ~/.claude/hooks/"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make linux          - Build for Linux (amd64)"
	@echo "  make darwin         - Build for macOS (amd64 + arm64)"
	@echo "  make darwin-amd64   - Build for macOS Intel"
	@echo "  make darwin-arm64   - Build for macOS Apple Silicon"
	@echo "  make windows        - Build for Windows (amd64)"
	@echo "  make all-platforms  - Build for all platforms"
	@echo "  make benchmark      - Run performance benchmarks"
	@echo "  make help           - Show this help"
