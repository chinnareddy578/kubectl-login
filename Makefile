.PHONY: build install test clean help

BINARY_NAME=kubectl-login
GO_FILES=$(shell find . -name '*.go' -not -path './vendor/*')

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

install: build
	@echo "Installing $(BINARY_NAME)..."
	@mkdir -p ~/bin
	@cp $(BINARY_NAME) ~/bin/
	@chmod +x ~/bin/$(BINARY_NAME)
	@echo "Installed to ~/bin/$(BINARY_NAME)"
	@echo "Make sure ~/bin is in your PATH"

install-system: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

test:
	@echo "Running tests..."
	@go test -v ./...

test-short:
	@echo "Running short tests (skipping integration)..."
	@go test -v -short ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@echo ""
	@echo "Coverage report saved to coverage.out"
	@echo "View HTML report: go tool cover -html=coverage.out"

test-integration:
	@echo "Running integration tests..."
	@go test -v ./test -run Integration

test-mock:
	@echo "Running tests with mock OIDC provider..."
	@go test -v ./pkg/auth -run Mock

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@echo "Clean complete"

help:
	@echo "Available targets:"
	@echo "  build            - Build the binary"
	@echo "  install          - Build and install to ~/bin"
	@echo "  install-system   - Build and install to /usr/local/bin"
	@echo "  test             - Run all tests"
	@echo "  test-short       - Run tests (skip integration)"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-mock        - Run tests with mock OIDC provider"
	@echo "  clean            - Remove built binary"
	@echo "  help             - Show this help message"

