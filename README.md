# ToDo Microservice - Go + Docker + Kubernetes

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://docker.com)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-Ready-326CE5?style=flat&logo=kubernetes)](https://kubernetes.io)

A production-ready RESTful ToDo API microservice built with Go, containerized with Docker, and deployable to Kubernetes. This project demonstrates clean architecture, modern DevOps practices, and cloud-native deployment strategies.

---

## 📋 Table of Contents

- [Project Overview](#-project-overview)
- [Features](#-features)
- [Architecture](#-architecture)
- [API Specification](#-api-specification)
- [Quick Start](#-quick-start)
- [Local Development](#-local-development)
- [Docker](#-docker)
- [Kubernetes Deployment](#-kubernetes-deployment)
- [Design Decisions](#-design-decisions)
- [Demo Guide](#-demo-guide)
- [Configuration](#-configuration)
- [Testing](#-testing)
- [Troubleshooting](#-troubleshooting)

---

## 🎯 Project Overview

This microservice implements a simple but production-grade **ToDo API** with the following capabilities:

- **CRUD Operations**: Create, Read, Update, and Delete todo items
- **Flexible Storage**: Switch between in-memory and persistent (BoltDB) storage via configuration
- **Health Checks**: Kubernetes-ready liveness and readiness probes
- **Graceful Shutdown**: Properly handles termination signals
- **Resource Management**: CPU/memory limits for optimal cluster performance
- **Cloud-Native**: Designed for containerized deployment in Kubernetes

### Technologies Used

| Technology | Purpose | Version |
|------------|---------|---------|
| **Go** | Backend language | 1.22+ |
| **Chi Router** | HTTP routing | v5.0.12 |
| **BoltDB** | Embedded database (optional) | bbolt v1.3.10 |
| **Docker** | Containerization | Multi-stage build |
| **Kubernetes** | Orchestration | apps/v1 |
| **Alpine Linux** | Base image | latest |

---

## ✨ Features

- ✅ **RESTful API** with standard HTTP methods
- ✅ **Clean Architecture** with repository pattern
- ✅ **Pluggable Storage** (in-memory or BoltDB)
- ✅ **Environment-based Configuration** (12-Factor App)
- ✅ **Health Endpoints** for Kubernetes probes
- ✅ **Graceful Shutdown** with context cancellation
- ✅ **Multi-stage Docker Build** (~15MB final image)
- ✅ **Non-root User** for security
- ✅ **Resource Limits** (CPU/Memory)
- ✅ **Horizontal Scaling** (2+ replicas)
- ✅ **Persistent Storage** (optional PVC)

---

## 🏗️ Architecture

### Project Structure

```
arancia/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Environment variable configuration
│   ├── health/
│   │   └── handler.go           # Health check endpoints
│   ├── httpserver/
│   │   └── router.go            # HTTP server setup & routing
│   └── todo/
│       ├── model.go             # ToDo domain model
│       ├── repository.go        # Storage interface
│       ├── memory_repo.go       # In-memory implementation
│       ├── boltdb_repo.go       # BoltDB implementation
│       └── handler.go           # HTTP request handlers
├── k8s/                         # Kubernetes manifests
│   ├── namespace.yaml           # Namespace isolation
│   ├── configmap.yaml           # In-memory config
│   ├── configmap-boltdb.yaml    # BoltDB config
│   ├── deployment.yaml          # In-memory deployment
│   ├── deployment-boltdb.yaml   # BoltDB deployment with PVC
│   ├── service.yaml             # ClusterIP service
│   ├── pvc.yaml                 # PersistentVolumeClaim (1Gi)
│   └── ingress.yaml             # External access (optional)
├── deployments/                 # Alternative K8s manifests
├── Dockerfile                   # Multi-stage build
├── .dockerignore               # Build context exclusions
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
├── Makefile                    # Build automation
└── README.md                   # This file
```

### Component Diagram

```
┌─────────────────────────────────────────────────────┐
│                   HTTP Client                       │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│                HTTP Server (Chi)                    │
│  ┌──────────────┐  ┌──────────────┐                │
│  │   Health     │  │     ToDo     │                │
│  │   Handler    │  │   Handler    │                │
│  └──────────────┘  └──────┬───────┘                │
└────────────────────────────┼────────────────────────┘
                             │
                             ▼
                   ┌─────────────────┐
                   │   Repository    │ (Interface)
                   └────────┬────────┘
                            │
             ┌──────────────┴──────────────┐
             ▼                             ▼
    ┌────────────────┐            ┌────────────────┐
    │  In-Memory     │            │    BoltDB      │
    │  Repository    │            │  Repository    │
    └────────────────┘            └────────┬───────┘
                                           │
                                           ▼
                                  ┌────────────────┐
                                  │  todos.db      │
                                  │  (Persistent)  │
                                  └────────────────┘
```

---

## 📡 API Specification

### Base URL

- **Local Development**: `http://localhost:8080`
- **Docker**: `http://localhost:8080`
- **Kubernetes (Port-forward)**: `http://localhost:8080`
- **Kubernetes (ClusterIP)**: `http://todo-service.todo-app.svc.cluster.local`
- **Kubernetes (Ingress)**: `http://todo.example.com`

### Endpoints

#### 📌 GET /todos
Retrieve all todo items.

**Request:**
```bash
curl -X GET http://localhost:8080/todos
```

**Response:** `200 OK`
```json
[
  {
    "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "title": "Learn Go programming",
    "completed": false
  },
  {
    "id": "a1b2c3d4-1234-5678-9012-abcdef123456",
    "title": "Deploy to Kubernetes",
    "completed": true
  }
]
```

**Empty list response:**
```json
[]
```

---

#### 📌 POST /todos
Create a new todo item.

**Request:**
```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Complete the challenge"
  }'
```

**Response:** `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Complete the challenge",
  "completed": false
}
```

**Error Response (Missing Title):** `400 Bad Request`
```
title is required
```

---

#### 📌 PUT /todos/{id}
Update an existing todo item. Supports partial updates.

**Request (Update Title Only):**
```bash
curl -X PUT http://localhost:8080/todos/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated task description"
  }'
```

**Request (Mark as Completed):**
```bash
curl -X PUT http://localhost:8080/todos/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "completed": true
  }'
```

**Request (Update Both):**
```bash
curl -X PUT http://localhost:8080/todos/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Completed task",
    "completed": true
  }'
```

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Completed task",
  "completed": true
}
```

**Error Response (Not Found):** `404 Not Found`
```
todo not found
```

---

#### 📌 DELETE /todos/{id}
Delete a todo item.

**Request:**
```bash
curl -X DELETE http://localhost:8080/todos/550e8400-e29b-41d4-a716-446655440000
```

**Response:** `204 No Content`
(No response body)

**Error Response (Not Found):** `404 Not Found`
```
todo not found
```

---

#### 📌 GET /health/live
Liveness probe endpoint for Kubernetes.

**Request:**
```bash
curl http://localhost:8080/health/live
```

**Response:** `200 OK`
```
OK
```

---

#### 📌 GET /health/ready
Readiness probe endpoint for Kubernetes.

**Request:**
```bash
curl http://localhost:8080/health/ready
```

**Response:** `200 OK`
```
OK
```

**Response (Not Ready):** `503 Service Unavailable`
```
Not Ready
```

---

### HTTP Status Codes

| Code | Meaning | Used For |
|------|---------|----------|
| `200 OK` | Success | GET, PUT operations |
| `201 Created` | Resource created | POST operations |
| `204 No Content` | Success, no body | DELETE operations |
| `400 Bad Request` | Invalid input | Missing/invalid data |
| `404 Not Found` | Resource not found | Non-existent ID |
| `500 Internal Server Error` | Server error | Unexpected errors |
| `503 Service Unavailable` | Service not ready | Readiness probe failure |

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.22+** - [Install Go](https://golang.org/dl/)
- **Docker** - [Install Docker](https://docs.docker.com/get-docker/)
- **kubectl** - [Install kubectl](https://kubernetes.io/docs/tasks/tools/)
- **Kubernetes cluster** - [Minikube](https://minikube.sigs.k8s.io/), [Kind](https://kind.sigs.k8s.io/), or cloud provider

### 30-Second Demo

```bash
# 1. Clone repository
git clone <repository-url>
cd arancia

# 2. Run locally
go run cmd/server/main.go

# 3. Test API (in new terminal)
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Todo"}'

curl http://localhost:8080/todos
```

---

## 💻 Local Development

### Option 1: Run with Go

```bash
# Install dependencies
go mod download

# Run with in-memory storage (default)
go run cmd/server/main.go

# Run with BoltDB persistence
STORAGE_BACKEND=boltdb BOLTDB_PATH=./data/todos.db go run cmd/server/main.go

# Run on custom port
PORT=3000 go run cmd/server/main.go
```

### Option 2: Build Binary

```bash
# Build binary
go build -o todo-service ./cmd/server

# Run binary
./todo-service

# With environment variables
PORT=3000 STORAGE_BACKEND=boltdb ./todo-service
```

### Option 3: Using Makefile

```bash
# Run with default settings
make run

# Run with BoltDB
make run-boltdb

# Build binary
make build

# Run tests
make test

# Build Docker image
make docker-build

# Clean build artifacts
make clean
```

### Testing Locally

```bash
# Health check
curl http://localhost:8080/health/live

# Create todos
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Go"}'

curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Build API"}'

# List all todos
curl http://localhost:8080/todos | jq '.'

# Update a todo (replace {id} with actual ID from previous response)
curl -X PUT http://localhost:8080/todos/{id} \
  -H "Content-Type: application/json" \
  -d '{"completed":true}'

# Delete a todo
curl -X DELETE http://localhost:8080/todos/{id}
```

---

## 🐳 Docker

### Build Image

```bash
# Basic build
docker build -t todo-service:latest .

# Build with specific tag
docker build -t todo-service:v1.0.0 .

# Build without cache (force rebuild)
docker build --no-cache -t todo-service:latest .

# Multi-platform build
docker buildx build --platform linux/amd64,linux/arm64 -t todo-service:latest .
```

### Run Container

#### In-Memory Storage (Default)

```bash
# Basic run
docker run --rm -p 8080:8080 todo-service:latest

# Run in background
docker run -d --name todo-api -p 8080:8080 todo-service:latest

# With custom configuration
docker run --rm -p 8080:8080 \
  -e PORT=8080 \
  -e STORAGE_BACKEND=memory \
  todo-service:latest
```

#### With BoltDB Persistence

```bash
# Run with volume mount
docker run --rm -p 8080:8080 \
  -e STORAGE_BACKEND=boltdb \
  -e BOLTDB_PATH=/data/todos.db \
  -v $(pwd)/data:/data \
  todo-service:latest

# Run in background with named volume
docker run -d --name todo-api -p 8080:8080 \
  -e STORAGE_BACKEND=boltdb \
  -e BOLTDB_PATH=/data/todos.db \
  -v todo-data:/data \
  todo-service:latest
```

### Container Management

```bash
# View logs
docker logs -f todo-api

# Check resource usage
docker stats todo-api

# Execute command in container
docker exec -it todo-api sh

# Stop container
docker stop todo-api

# Remove container
docker rm todo-api

# View images
docker images | grep todo-service

# Remove image
docker rmi todo-service:latest
```

### Docker Compose (Optional)

Create `docker-compose.yml`:

```yaml
version: '3.8'
services:
  todo-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - STORAGE_BACKEND=boltdb
      - BOLTDB_PATH=/data/todos.db
    volumes:
      - todo-data:/data
    restart: unless-stopped

volumes:
  todo-data:
```

Usage:
```bash
docker-compose up -d        # Start
docker-compose logs -f      # View logs
docker-compose down         # Stop
docker-compose down -v      # Stop and remove volumes
```

---

## ☸️ Kubernetes Deployment

### Deployment Options

**Option A: In-Memory Storage** (Stateless, no PVC needed)
**Option B: BoltDB with PVC** (Stateful, data persists)

### Option A: In-Memory Deployment

#### Step 1: Prepare Image

```bash
# Build image
docker build -t todo-service:latest .

# For Minikube
minikube image load todo-service:latest

# For Kind
kind load docker-image todo-service:latest

# For cloud/registry
docker tag todo-service:latest <registry>/todo-service:latest
docker push <registry>/todo-service:latest
```

#### Step 2: Deploy to Kubernetes

```bash
# Apply all manifests
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# Or apply all at once
kubectl apply -f k8s/
```

#### Step 3: Verify Deployment

```bash
# Check all resources
kubectl get all -n todo-app

# Check pods
kubectl get pods -n todo-app
kubectl describe pod -n todo-app -l app=todo-service

# Check service
kubectl get svc -n todo-app
kubectl describe svc todo-service -n todo-app

# Check logs
kubectl logs -f -n todo-app -l app=todo-service
```

#### Step 4: Access the Service

```bash
# Port forward
kubectl port-forward -n todo-app svc/todo-service 8080:80

# Test in another terminal
curl http://localhost:8080/health/live
curl http://localhost:8080/todos
```

### Option B: BoltDB with Persistence

#### Step 1: Deploy with PVC

```bash
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Create PVC (1Gi persistent storage)
kubectl apply -f k8s/pvc.yaml

# Create BoltDB ConfigMap
kubectl apply -f k8s/configmap-boltdb.yaml

# Deploy application with volumes
kubectl apply -f k8s/deployment-boltdb.yaml

# Create service
kubectl apply -f k8s/service.yaml
```

#### Step 2: Verify PVC is Bound

```bash
# Check PVC status
kubectl get pvc -n todo-app

# Should show STATUS: Bound
# NAME               STATUS   VOLUME                   CAPACITY   ACCESS MODES
# todo-boltdb-pvc    Bound    pvc-xxxxx-xxxx-xxxx...  1Gi        RWO
```

#### Step 3: Test Persistence

```bash
# Port forward
kubectl port-forward -n todo-app svc/todo-service 8080:80 &

# Create test data
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Persistent Todo"}'

# Verify data exists
curl http://localhost:8080/todos

# Delete pods (simulating crash/restart)
kubectl delete pods -n todo-app -l app=todo-service

# Wait for new pods
kubectl wait --for=condition=ready pod -l app=todo-service -n todo-app

# Port forward to new pod
kubectl port-forward -n todo-app svc/todo-service 8080:80 &

# Verify data persisted
curl http://localhost:8080/todos
# ✅ Data should still be there!
```

### Kubernetes Management

```bash
# Scale deployment
kubectl scale deployment/todo-service --replicas=5 -n todo-app

# Update image
kubectl set image deployment/todo-service \
  todo-service=todo-service:v2.0.0 \
  -n todo-app

# Rollout status
kubectl rollout status deployment/todo-service -n todo-app

# Rollback deployment
kubectl rollout undo deployment/todo-service -n todo-app

# Update ConfigMap
kubectl edit configmap todo-config -n todo-app
kubectl rollout restart deployment/todo-service -n todo-app

# View events
kubectl get events -n todo-app --sort-by='.lastTimestamp'

# Delete everything
kubectl delete namespace todo-app
```

---

## 🧠 Design Decisions

### 1. **Repository Pattern**

**Decision:** Use an interface (`Repository`) to abstract storage operations.

**Rationale:**
- ✅ **Testability**: Easy to mock for unit tests
- ✅ **Flexibility**: Switch storage backends without changing handlers
- ✅ **Maintainability**: Clear separation of concerns
- ✅ **Future-proof**: Can add PostgreSQL, Redis, etc. easily

**Trade-off:** Slight added complexity, but worth it for modularity.

---

### 2. **In-Memory vs BoltDB**

**Decision:** Support both in-memory (default) and BoltDB storage.

**Rationale:**
- 🚀 **In-Memory**: Fast, simple, perfect for demos and testing
- 💾 **BoltDB**: Persistent, embedded (no external DB), single-file storage
- 🎯 **Configurable**: Choose via environment variable

**Trade-off:** BoltDB uses file locking (ReadWriteOnce), limiting multi-replica write scenarios. For production with multiple replicas, consider external DB.

---

### 3. **Health Endpoints**

**Decision:** Implement separate liveness and readiness probes.

**Rationale:**
- ❤️ **Liveness** (`/health/live`): Kubernetes restarts pod if this fails
- 🟢 **Readiness** (`/health/ready`): Kubernetes stops sending traffic if this fails
- 🎯 **Different purposes**: Liveness = "is process alive", Readiness = "can handle requests"

**Implementation:** Currently both return 200 OK. In production, readiness could check DB connectivity, etc.

---

### 4. **Resource Limits**

**Decision:** Set explicit CPU/memory requests and limits.

**Rationale:**
- 📊 **Requests**: Guaranteed resources (100m CPU, 128Mi RAM)
- 🛡️ **Limits**: Maximum allowed (500m CPU, 256Mi RAM)
- ⚖️ **Benefits**: Prevents resource starvation, helps scheduler, enables autoscaling

**Trade-off:** May need tuning based on actual load.

---

### 5. **Graceful Shutdown**

**Decision:** Implement proper signal handling and graceful HTTP server shutdown.

**Rationale:**
- ✅ In-flight requests complete before termination
- ✅ Resources (DB connections) cleaned up properly
- ✅ Zero-downtime deployments

**Implementation:** 5-second timeout for shutdown.

---

### 6. **No Authentication**

**Decision:** No authentication/authorization in this version.

**Rationale:**
- 🎯 **Scope**: Focus on architecture and deployment
- ⏱️ **Simplicity**: Keeps the challenge focused
- 🔮 **Future**: Could add JWT, OAuth, API keys, mTLS, etc.

**Trade-off:** Not production-ready for public internet.

---

### 7. **Error Handling**

**Decision:** Simple error responses with HTTP status codes.

**Rationale:**
- ✅ RESTful: Standard HTTP status codes (400, 404, 500)
- ✅ Clear: Sentinel errors (ErrNotFound, ErrInvalid)
- ⚠️ **Improvement Needed**: Could add structured error responses with error codes

---

### 8. **Chi Router**

**Decision:** Use `chi` router instead of standard `net/http` ServeMux.

**Rationale:**
- ✅ **Lightweight**: Small footprint, fast
- ✅ **Idiomatic**: Works with standard `http.Handler`
- ✅ **Features**: Middleware, URL parameters, sub-routers
- ✅ **Production-ready**: Used by many Go projects

**Alternative:** Could use Gin, Echo, or Gorilla Mux.

---

### 9. **Alpine Base Image**

**Decision:** Use `alpine:latest` for runtime stage (not `scratch` or `distroless`).

**Rationale:**
- ✅ **Small**: ~15MB final image
- ✅ **Debuggable**: Has shell (sh) for troubleshooting
- ✅ **Package manager**: Can install tools (apk)
- ✅ **CA certificates**: Included for HTTPS

**Trade-off:** Slightly larger than distroless (~10MB) or scratch (~8MB), but much easier to debug.

---

### 10. **Multiple Kubernetes Manifest Variants**

**Decision:** Provide both standard and BoltDB-specific manifests.

**Rationale:**
- 📁 **Clarity**: Clear separation between stateless and stateful deployments
- 🎯 **Choice**: Users can pick what they need
- 📚 **Educational**: Shows both patterns

**Files:**
- `k8s/deployment.yaml` - In-memory (no PVC)
- `k8s/deployment-boltdb.yaml` - BoltDB with PVC
- `k8s/configmap.yaml` - In-memory config
- `k8s/configmap-boltdb.yaml` - BoltDB config

---

## 🎬 Demo Guide

### Complete Demo Script

This script demonstrates building, containerizing, and deploying the microservice.

```bash
#!/bin/bash
set -e

echo "=== ToDo Microservice Demo ==="
echo

# Step 1: Local Development
echo "Step 1: Running locally..."
go run cmd/server/main.go &
SERVER_PID=$!
sleep 2

echo "Testing API locally..."
curl -X POST http://localhost:8080/todos -H "Content-Type: application/json" -d '{"title":"Demo Todo"}'
curl http://localhost:8080/todos

kill $SERVER_PID
echo

# Step 2: Docker Build
echo "Step 2: Building Docker image..."
docker build -t todo-service:demo .
echo "Image size:"
docker images todo-service:demo
echo

# Step 3: Docker Run
echo "Step 3: Running in Docker..."
docker run -d --name todo-demo -p 8080:8080 todo-service:demo
sleep 2

echo "Testing containerized API..."
curl http://localhost:8080/health/live
curl -X POST http://localhost:8080/todos -H "Content-Type: application/json" -d '{"title":"Containerized Todo"}'
curl http://localhost:8080/todos

docker stop todo-demo
docker rm todo-demo
echo

# Step 4: Kubernetes Deployment
echo "Step 4: Deploying to Kubernetes..."

# Load image to cluster (minikube example)
minikube image load todo-service:demo

# Deploy
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

echo "Waiting for deployment..."
kubectl wait --for=condition=available deployment/todo-service -n todo-app --timeout=60s

echo "Testing in Kubernetes..."
kubectl port-forward -n todo-app svc/todo-service 8080:80 &
PF_PID=$!
sleep 2

curl http://localhost:8080/health/live
curl -X POST http://localhost:8080/todos -H "Content-Type: application/json" -d '{"title":"Kubernetes Todo"}'
curl http://localhost:8080/todos

kill $PF_PID

echo
echo "=== Demo Complete! ==="
echo "Cleanup: kubectl delete namespace todo-app"
```

Save as `demo.sh` and run:
```bash
chmod +x demo.sh
./demo.sh
```

---

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description | Example |
|----------|---------|-------------|---------|
| `PORT` | `8080` | HTTP server port | `3000`, `8080` |
| `STORAGE_BACKEND` | `memory` | Storage type | `memory`, `boltdb` |
| `BOLTDB_PATH` | `/data/todos.db` | BoltDB file path | `./data/todos.db`, `/var/lib/todos.db` |

### Kubernetes ConfigMap

In-memory configuration (`k8s/configmap.yaml`):
```yaml
data:
  PORT: "8080"
  STORAGE_BACKEND: "memory"
  BOLTDB_PATH: "/data/todos.db"
```

BoltDB configuration (`k8s/configmap-boltdb.yaml`):
```yaml
data:
  PORT: "8080"
  STORAGE_BACKEND: "boltdb"
  BOLTDB_PATH: "/data/todos.db"
```

### Updating Configuration

```bash
# Edit ConfigMap
kubectl edit configmap todo-config -n todo-app

# Or patch it
kubectl patch configmap todo-config -n todo-app \
  --patch '{"data":{"STORAGE_BACKEND":"boltdb"}}'

# Restart pods to pick up changes
kubectl rollout restart deployment/todo-service -n todo-app
```

---

## 🧪 Testing

### Manual Testing with curl

```bash
# Complete workflow test
./test-api.sh
```

Create `test-api.sh`:
```bash
#!/bin/bash
BASE_URL="http://localhost:8080"

echo "1. Health Check"
curl $BASE_URL/health/live
echo -e "\n"

echo "2. Create Todos"
TODO1=$(curl -s -X POST $BASE_URL/todos -H "Content-Type: application/json" -d '{"title":"Task 1"}')
TODO1_ID=$(echo $TODO1 | jq -r '.id')
echo "Created: $TODO1_ID"

TODO2=$(curl -s -X POST $BASE_URL/todos -H "Content-Type: application/json" -d '{"title":"Task 2"}')
TODO2_ID=$(echo $TODO2 | jq -r '.id')
echo "Created: $TODO2_ID"

echo -e "\n3. List All"
curl -s $BASE_URL/todos | jq '.'

echo -e "\n4. Update Todo"
curl -s -X PUT $BASE_URL/todos/$TODO1_ID -H "Content-Type: application/json" -d '{"completed":true}' | jq '.'

echo -e "\n5. Delete Todo"
curl -i -X DELETE $BASE_URL/todos/$TODO2_ID

echo -e "\n6. List After Delete"
curl -s $BASE_URL/todos | jq '.'
```

### Unit Testing

Add tests in `internal/todo/repository_test.go`:

```go
package todo_test

import (
    "testing"
    "github.com/arancia/todo-service/internal/todo"
)

func TestMemoryRepository(t *testing.T) {
    repo := todo.NewMemoryRepository()
    
    // Test Create
    item := &todo.ToDo{
        ID:        "test-id",
        Title:     "Test",
        Completed: false,
    }
    err := repo.Create(item)
    if err != nil {
        t.Fatalf("Create failed: %v", err)
    }
    
    // Test GetByID
    retrieved, err := repo.GetByID("test-id")
    if err != nil {
        t.Fatalf("GetByID failed: %v", err)
    }
    if retrieved.Title != "Test" {
        t.Errorf("Expected title 'Test', got '%s'", retrieved.Title)
    }
}
```

Run tests:
```bash
go test ./internal/...
```

---

## 🐛 Troubleshooting

### Common Issues

#### **1. Pods not starting**

```bash
# Check pod status
kubectl get pods -n todo-app

# View detailed pod information
kubectl describe pod -n todo-app <pod-name>

# Check pod logs
kubectl logs -n todo-app <pod-name>

# Check events
kubectl get events -n todo-app --sort-by='.lastTimestamp'
```

**Common Causes:**
- Image pull errors (use `imagePullPolicy: IfNotPresent` for local images)
- Resource limits too low
- ConfigMap not found
- PVC not bound (for BoltDB deployment)

---

#### **2. Health checks failing**

```bash
# Test health endpoint directly
kubectl exec -n todo-app <pod-name> -- wget -O- http://localhost:8080/health/live

# Check probe configuration
kubectl describe pod -n todo-app <pod-name> | grep -A 10 "Liveness\|Readiness"
```

**Common Causes:**
- Wrong port in probe configuration
- InitialDelaySeconds too short
- Application not starting fast enough

---

#### **3. Service not accessible**

```bash
# Check service endpoints
kubectl get endpoints -n todo-app

# Verify service selector matches pod labels
kubectl get svc todo-service -n todo-app -o yaml | grep selector -A 2
kubectl get pods -n todo-app --show-labels
```

**Common Causes:**
- Service selector doesn't match pod labels
- Port mismatch between service and container
- Pods not ready

---

#### **4. PVC not binding (BoltDB)**

```bash
# Check PVC status
kubectl get pvc -n todo-app
kubectl describe pvc todo-boltdb-pvc -n todo-app

# Check available storage classes
kubectl get storageclass
```

**Common Causes:**
- No default storage class configured
- Insufficient storage available
- Storage class doesn't exist

**Solution:**
```bash
# For minikube, enable storage provisioner
minikube addons enable storage-provisioner

# For kind, storage provisioner is enabled by default
```

---

#### **5. Image not found**

```bash
# For local Kubernetes, load image
minikube image load todo-service:latest
# or
kind load docker-image todo-service:latest

# Verify image is available
minikube image ls | grep todo-service
# or
docker exec -it kind-control-plane crictl images | grep todo-service
```

---

#### **6. Port forward not working**

```bash
# Kill existing port-forward processes
pkill -f "port-forward"

# Try different port
kubectl port-forward -n todo-app svc/todo-service 9090:80

# Access via pod directly
POD=$(kubectl get pod -n todo-app -l app=todo-service -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward -n todo-app $POD 8080:8080
```

---

### Debugging Tips

```bash
# Get shell access to pod
kubectl exec -it -n todo-app <pod-name> -- sh

# Inside pod:
# - Check if process is running
ps aux | grep todo-service

# - Check if port is listening
netstat -tlnp | grep 8080

# - Test health endpoint
wget -O- http://localhost:8080/health/live

# - Check environment variables
env | grep STORAGE
```

---

## 📚 Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Chi Router](https://github.com/go-chi/chi)
- [BoltDB](https://github.com/etcd-io/bbolt)
- [12-Factor App](https://12factor.net/)

---

## 📄 License

This project is provided as-is for evaluation purposes.

---

## 👤 Author

**Rafael** - Kubernetes Microservice Challenge

---

## 🙏 Acknowledgments

- Go community for excellent tooling
- Kubernetes for cloud-native orchestration
- Docker for containerization standards
- Open-source contributors

---

**Happy Coding!** 🚀
