.PHONY: help build run run-boltdb test clean docker-build docker-run docker-stop k8s-deploy k8s-delete k8s-logs

# Variables
BINARY_NAME=todo-service
DOCKER_IMAGE=todo-service:latest
GO_FILES=$(shell find . -name '*.go' -type f)

# Default target
help:
	@echo "Available targets:"
	@echo "  build          - Build the Go binary"
	@echo "  run            - Run locally with in-memory storage"
	@echo "  run-boltdb     - Run locally with BoltDB storage"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container (in-memory)"
	@echo "  docker-stop    - Stop Docker container"
	@echo "  k8s-deploy     - Deploy to Kubernetes (in-memory)"
	@echo "  k8s-deploy-pvc - Deploy to Kubernetes (BoltDB with PVC)"
	@echo "  k8s-delete     - Delete Kubernetes resources"
	@echo "  k8s-logs       - View Kubernetes logs"
	@echo "  k8s-port-forward - Port forward to Kubernetes service"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) ./cmd/server/main.go
	@echo "Build complete: $(BINARY_NAME)"

# Run locally with in-memory storage
run:
	@echo "Running with in-memory storage..."
	go run cmd/server/main.go

# Run locally with BoltDB
run-boltdb:
	@echo "Running with BoltDB storage..."
	@mkdir -p data
	STORAGE_BACKEND=boltdb BOLTDB_PATH=./data/todos.db go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./internal/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf data/*.db
	@echo "Clean complete"

# Docker targets
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker build complete"
	@docker images $(DOCKER_IMAGE)

docker-run:
	@echo "Running Docker container..."
	docker run -d --name todo-api -p 8080:8080 \
		-e STORAGE_BACKEND=memory \
		$(DOCKER_IMAGE)
	@echo "Container started. Test with: curl http://localhost:8080/health/live"

docker-run-boltdb:
	@echo "Running Docker container with BoltDB..."
	@mkdir -p data
	docker run -d --name todo-api -p 8080:8080 \
		-e STORAGE_BACKEND=boltdb \
		-e BOLTDB_PATH=/data/todos.db \
		-v $$(pwd)/data:/data \
		$(DOCKER_IMAGE)
	@echo "Container started with persistent storage"

docker-stop:
	@echo "Stopping Docker container..."
	docker stop todo-api || true
	docker rm todo-api || true
	@echo "Container stopped"

docker-logs:
	docker logs -f todo-api

# Kubernetes targets (minikube)
k8s-load-image:
	@echo "Loading image to minikube..."
	minikube image load $(DOCKER_IMAGE)

k8s-load-image-kind:
	@echo "Loading image to kind..."
	kind load docker-image $(DOCKER_IMAGE)

k8s-deploy:
	@echo "Deploying to Kubernetes (in-memory)..."
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/service.yaml
	@echo "Waiting for deployment..."
	kubectl wait --for=condition=available deployment/todo-service -n todo-app --timeout=60s
	@echo "Deployment complete!"
	@echo "Run 'make k8s-port-forward' to access the service"

k8s-deploy-pvc:
	@echo "Deploying to Kubernetes (BoltDB with PVC)..."
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/pvc.yaml
	kubectl apply -f k8s/configmap-boltdb.yaml
	kubectl apply -f k8s/deployment-boltdb.yaml
	kubectl apply -f k8s/service.yaml
	@echo "Waiting for deployment..."
	kubectl wait --for=condition=available deployment/todo-service -n todo-app --timeout=60s
	@echo "Deployment complete with persistent storage!"
	@echo "Run 'make k8s-port-forward' to access the service"

k8s-delete:
	@echo "Deleting Kubernetes resources..."
	kubectl delete namespace todo-app --ignore-not-found=true
	@echo "Cleanup complete"

k8s-status:
	@echo "Kubernetes resources status:"
	kubectl get all,pvc -n todo-app

k8s-logs:
	@echo "Streaming logs..."
	kubectl logs -f -n todo-app -l app=todo-service

k8s-describe:
	@echo "Describing deployment..."
	kubectl describe deployment todo-service -n todo-app

k8s-port-forward:
	@echo "Port forwarding to service..."
	@echo "Access at: http://localhost:8080"
	kubectl port-forward -n todo-app svc/todo-service 8080:80

# Test API endpoints (requires service to be running)
test-api:
	@echo "Testing API endpoints..."
	@echo "\n1. Health Check:"
	curl -s http://localhost:8080/health/live
	@echo "\n\n2. Create Todo:"
	curl -s -X POST http://localhost:8080/todos \
		-H "Content-Type: application/json" \
		-d '{"title":"Test from Makefile"}' | jq '.'
	@echo "\n3. List Todos:"
	curl -s http://localhost:8080/todos | jq '.'

# Development helpers
dev:
	@echo "Starting development server with auto-reload..."
	@echo "Install air first: go install github.com/cosmtrek/air@latest"
	air

fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Complete workflow
all: clean build test docker-build

# Demo workflow
demo: docker-build docker-run
	@echo "\n=== Demo Started ==="
	@sleep 2
	@echo "Testing API..."
	@make test-api
	@echo "\n=== Demo Complete ==="
	@echo "Stop with: make docker-stop"
