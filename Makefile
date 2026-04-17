.PHONY: help install-hooks test fmt vet build clean

help:
	@echo "Litmus - Alertmanager Validator"
	@echo ""
	@echo "Available targets:"
	@echo "  install-hooks   Install git pre-commit hooks"
	@echo "  test            Run tests"
	@echo "  fmt             Format code with go fmt"
	@echo "  vet             Run go vet"
	@echo "  build           Build litmus binary"
	@echo "  clean           Clean build artifacts"
	@echo "  help            Show this help message"

install-hooks:
	@echo "Installing git hooks..."
	@mkdir -p .git/hooks
	@cp .git-hooks/pre-commit .git/hooks/pre-commit
	@cp .git-hooks/commit-msg .git/hooks/commit-msg
	@chmod +x .git/hooks/pre-commit
	@chmod +x .git/hooks/commit-msg
	@echo "✓ Git hooks installed"

test:
	@echo "Running tests..."
	@go test ./... -v

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

build:
	@echo "Building litmus..."
	@go build -o litmus ./cmd/litmus

clean:
	@echo "Cleaning up..."
	@rm -f litmus
	@go clean
