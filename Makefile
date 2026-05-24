.PHONY: dev build test clean docker-build docker-up swagger-ui

APP_NAME=api-docs-portal
MAIN_PATH=./cmd/server

# Run in development mode
dev:
	go run $(MAIN_PATH)/main.go

# Build binary
build:
	go build -o bin/$(APP_NAME) $(MAIN_PATH)/main.go

# Run tests
test:
	go test ./... -v

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f portal.db

# Build Docker image
docker-build:
	docker build -t $(APP_NAME) .

# Run with Docker Compose
docker-up:
	docker compose up -d

# Stop Docker Compose
docker-down:
	docker compose down

# Download Swagger UI assets
swagger-ui:
	@echo "Downloading Swagger UI dist..."
	@mkdir -p web/static/swagger-ui
	@curl -sL https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js -o web/static/swagger-ui/swagger-ui-bundle.js
	@curl -sL https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js -o web/static/swagger-ui/swagger-ui-standalone-preset.js
	@curl -sL https://unpkg.com/swagger-ui-dist@5/swagger-ui.css -o web/static/swagger-ui/swagger-ui.css
	@echo "Swagger UI assets downloaded."
