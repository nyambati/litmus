.PHONY: help install-hooks test fmt vet build build-ui clean

help:
	@echo "Litmus - Alertmanager Validator"
	@echo ""
	@echo "Available targets:"
	@echo "  install-hooks   Install git pre-commit hooks"
	@echo "  test            Run tests"
	@echo "  fmt             Format code with go fmt"
	@echo "  vet             Run go vet"
	@echo "  build-ui        Build the React UI (outputs to ui/dist/)"
	@echo "  build           Build UI then litmus binary (embeds UI)"
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

build-ui:
	@echo "Building UI..."
	@cd ui && npm run build

build: build-ui
	@echo "Building litmus..."
	@go build -o bin/litmus .

lint:
	golangci-lint run

clean:
	@echo "Cleaning up..."
	@rm -f bin/litmus
	@rm -rf ui/dist
	@go clean

all: fmt test vet lint
dev:
	@nohup sh -c 'cd ui && npm run dev' > /dev/null 2>&1 &
	@air
serve:
	@go run main.go serve
