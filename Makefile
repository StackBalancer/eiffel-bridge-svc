.PHONY: build test clean deploy health-check

# Build the application
build:
	go build -o bin/adapter ./cmd/adapter

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Deploy to local Minikube
deploy:
	./deploy.py

# Check service health
health-check:
	curl -f http://localhost:8080/health || echo "Service not available"

# Test webhook endpoints
test-webhook-push:
	curl -X POST http://localhost:8080/webhook \
		-H 'Content-Type: application/json' \
		-H 'X-Gitlab-Event: Push Hook' \
		--data-binary @test-webhooks/push.json

test-webhook-push-form:
	curl -X POST http://localhost:8080/webhook \
		-H 'Content-Type: application/x-www-form-urlencoded' \
		-H 'X-Gitlab-Event: Push Hook' \
		--data-urlencode "payload@test-webhooks/push.json"

test-webhook-pipeline:
	curl -X POST http://localhost:8080/webhook \
		-H 'Content-Type: application/json' \
		-H 'X-Gitlab-Event: Pipeline Hook' \
		--data-binary @test-webhooks/pipeline-success.json

# Run locally (requires RabbitMQ)
run-local:
	RABBIT_URL="amqp://guest:guest@localhost:5672/" go run cmd/adapter/main.go

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the application"
	@echo "  test            - Run tests"
	@echo "  clean           - Clean build artifacts"
	@echo "  deploy          - Deploy to local Minikube"
	@echo "  health-check    - Check service health"
	@echo "  test-webhook-*  - Test webhook endpoints"
	@echo "  run-local       - Run locally"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code"
	@echo "  help            - Show this help"