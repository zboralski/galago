# Galago Makefile
# ARM64 emulation trace analyzer
#
# Prerequisites:
#   brew install unicorn go
#
# Quick start:
#   make setup    # Install deps and build
#   make build    # Build CLI binary
#   make test     # Run all tests

# CGO flags for Unicorn (installed via Homebrew)
export CGO_LDFLAGS := -L/opt/homebrew/lib
export CGO_CFLAGS := -I/opt/homebrew/include

.PHONY: all build test clean setup check-deps demo

all: build

# Full setup from scratch
setup: check-deps deps build
	@echo "Setup complete. Run './galago demo/libcocos2dlua.so' to verify."

# Check system dependencies
check-deps:
	@echo "Checking dependencies..."
	@command -v go >/dev/null 2>&1 || { echo "ERROR: go not found. Install with: brew install go"; exit 1; }
	@test -f /opt/homebrew/lib/libunicorn.dylib || test -f /usr/local/lib/libunicorn.dylib || \
		{ echo "ERROR: libunicorn not found. Install with: brew install unicorn"; exit 1; }
	@echo "All dependencies found."

# Build the main CLI binary
build:
	go build -o galago ./cmd/galago

# Run all tests (requires libunicorn)
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f galago
	go clean

# Install Go dependencies
deps:
	go mod tidy

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# Quick development cycle: test emulator package
dev:
	go test -v ./internal/emulator/...

# Generate demo GIF and MP4 using VHS
demo: build
	@command -v vhs >/dev/null 2>&1 || { echo "ERROR: vhs not found. Install with: brew install vhs"; exit 1; }
	vhs demo/demo.tape

# Show help
help:
	@echo "Galago - ARM64 emulation trace analyzer"
	@echo ""
	@echo "Setup (run once):"
	@echo "  make setup       - Install deps and build"
	@echo "  make check-deps  - Verify system dependencies"
	@echo ""
	@echo "Development:"
	@echo "  make build       - Build galago binary"
	@echo "  make test        - Run all tests (requires libunicorn)"
	@echo "  make fmt         - Format Go code"
	@echo ""
	@echo "Demo:"
	@echo "  make demo        - Generate demo.gif using VHS"
	@echo ""
	@echo "Usage:"
	@echo "  ./galago <binary.so>       - Extract keys with colorized trace"
	@echo "  ./galago <binary.so> -q    - Quiet mode (keys + stats only)"
	@echo "  ./galago info <binary.so>  - Show binary info"
	@echo ""
	@echo "Prerequisites:"
	@echo "  brew install unicorn go"
