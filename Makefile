.PHONY: help build run test clean lint docker-up docker-down setup

help:
	@echo "Available commands:"
	@echo "  make build       - Build the binary"
	@echo "  make run         - Run the application"
	@echo "  make docker-up   - Start Docker services (Qdrant)"
	@echo "  make docker-down - Stop Docker services"
	@echo "  make test        - Run tests"
	@echo "  make lint        - Run linter"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make setup       - Initial project setup"

build:
	go build -o bin/go-local-rag-email ./cmd/go-local-rag-email

run:
	go run ./cmd/go-local-rag-email

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/

docker-up:
	docker-compose up -d
	@echo "✓ Qdrant running on http://localhost:6333"

docker-down:
	docker-compose down

setup:
	@echo "Setting up Go-Local-RAG-Email..."
	@mkdir -p ~/.go-local-rag-email
	@if [ ! -f ~/.go-local-rag-email/config.yaml ]; then \
		cp configs/config.yaml.example ~/.go-local-rag-email/config.yaml; \
		echo "✓ Created config file at ~/.go-local-rag-email/config.yaml"; \
	else \
		echo "✓ Config file already exists"; \
	fi
	@echo "✓ Setup complete! Edit ~/.go-local-rag-email/config.yaml to add your API keys"
