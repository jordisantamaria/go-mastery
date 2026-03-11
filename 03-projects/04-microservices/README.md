# Project 4: Microservices Platform

Microservices platform that demonstrates distributed architecture in Go using only the standard library. Uses `net/rpc` instead of gRPC to keep the project free of external dependencies, demonstrating the same patterns: inter-service communication, API gateway, health checks, and graceful shutdown.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client (curl/browser)                     │
│                              │                                   │
│                         HTTP REST                                │
│                              │                                   │
│                    ┌─────────▼──────────┐                        │
│                    │   API Gateway      │                        │
│                    │   :8080            │                        │
│                    │   REST → RPC       │                        │
│                    └──────┬───────┬─────┘                        │
│                           │       │                              │
│                    RPC    │       │   RPC                         │
│                           │       │                              │
│              ┌────────────▼┐   ┌──▼─────────────┐               │
│              │ UserService │   │ OrderService    │               │
│              │ :50051      │   │ :50052          │               │
│              │ (net/rpc)   │   │ (net/rpc)       │               │
│              └─────────────┘   └────────┬────────┘               │
│                                         │                        │
│                                    RPC  │ (validates user)       │
│                                         │                        │
│              ┌──────────────────────────▼┐                       │
│              │       UserService         │                       │
│              └───────────────────────────┘                       │
└─────────────────────────────────────────────────────────────────┘
```

**Data flow:**
1. The client sends REST requests to the API Gateway (:8080)
2. The Gateway translates REST to RPC calls to the backend services
3. OrderService validates the user's existence by calling UserService via RPC
4. Each service has in-memory storage with sync.RWMutex

## Project Structure

```
04-microservices/
├── go.mod                      # Go module (no external dependencies)
├── Dockerfile                  # Multi-stage build for all 3 services
├── docker-compose.yml          # Container orchestration
├── cmd/
│   ├── userservice/main.go     # User service entrypoint
│   ├── orderservice/main.go    # Order service entrypoint
│   └── gateway/main.go         # API Gateway entrypoint
├── internal/
│   ├── userservice/            # User service logic
│   ├── orderservice/           # Order service logic
│   └── gateway/                # Gateway logic REST → RPC
├── pkg/
│   ├── model/                  # Shared types (User, Order)
│   ├── health/                 # Health checks
│   ├── discovery/              # Service registry
│   └── middleware/             # HTTP middlewares (logging, recovery)
└── scripts/
    └── test.sh                 # Script to test all endpoints
```

## How to Run

### Option 1: Locally (3 terminals)

**Terminal 1 - UserService:**
```bash
go run ./cmd/userservice
# Output: {"level":"INFO","msg":"UserService started","addr":":50051"}
```

**Terminal 2 - OrderService:**
```bash
go run ./cmd/orderservice
# Output: {"level":"INFO","msg":"OrderService started","addr":":50052","user_service_addr":"localhost:50051"}
```

**Terminal 3 - API Gateway:**
```bash
go run ./cmd/gateway
# Output: {"level":"INFO","msg":"API Gateway started","addr":":8080",...}
```

### Option 2: Docker Compose

```bash
docker compose up --build
```

To stop:
```bash
docker compose down
```

### Environment Variables

| Variable             | Default          | Description                       |
|----------------------|------------------|-----------------------------------|
| `USER_SERVICE_ADDR`  | `:50051`         | UserService address               |
| `ORDER_SERVICE_ADDR` | `:50052`         | OrderService address              |
| `GATEWAY_ADDR`       | `:8080`          | API Gateway address               |

## REST API

### Users

```bash
# Create user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Ana Garcia","email":"ana@example.com"}'

# List users
curl http://localhost:8080/api/users

# Get user by ID
curl http://localhost:8080/api/users/{id}

# Update user
curl -X PUT http://localhost:8080/api/users/{id} \
  -H "Content-Type: application/json" \
  -d '{"name":"Ana Garcia Martinez","email":"ana.garcia@example.com"}'

# Delete user
curl -X DELETE http://localhost:8080/api/users/{id}
```

### Orders

```bash
# Create order (validates that the user exists)
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "{user_id}",
    "items": [
      {"product_id":"p1","name":"Keyboard","quantity":1,"price":49.99},
      {"product_id":"p2","name":"Mouse","quantity":2,"price":19.99}
    ]
  }'

# Get order by ID
curl http://localhost:8080/api/orders/{id}

# List orders for a user
curl http://localhost:8080/api/users/{user_id}/orders

# Update order status
curl -X PUT http://localhost:8080/api/orders/{id}/status \
  -H "Content-Type: application/json" \
  -d '{"status":"confirmed"}'
```

**Valid state transitions:**
```
pending → confirmed → shipped → delivered
pending → cancelled
confirmed → cancelled
```

### Health Check

```bash
curl http://localhost:8080/health
# {"status":"healthy","services":[{"service":"user-service","status":"healthy"},{"service":"order-service","status":"healthy"}]}
```

## Test Script

```bash
chmod +x scripts/test.sh
./scripts/test.sh
# Or with a custom host:
./scripts/test.sh http://my-host:9090
```

## Unit Tests

```bash
go test ./...
```

Tests cover:
- **UserService:** full CRUD, validations, duplicate emails
- **OrderService:** creation with user validation (mock), state transitions
- **Gateway:** integration tests that spin up real RPC services on random ports

## Patterns Demonstrated

| Pattern                         | Implementation                                            |
|--------------------------------|-----------------------------------------------------------|
| **API Gateway**                | REST HTTP that translates to RPC calls                    |
| **Inter-service communication**| OrderService validates users via RPC to UserService       |
| **Service Discovery**          | In-memory registry (pkg/discovery)                        |
| **Health Checks**              | Backend service connectivity verification                 |
| **Graceful Shutdown**          | Captures SIGINT/SIGTERM, closes listeners in order        |
| **Middleware chain**           | Logging and recovery applied to the gateway               |
| **Dependency Injection**       | UserValidator as injectable interface in OrderService      |
| **State machine**              | Validated state transitions on orders                     |
| **Multi-stage Docker build**   | Final image of ~15MB with static binaries                 |
| **Structured logging**         | slog with JSON format for observability                   |
| **Environment configuration**  | Environment variables with sensible defaults              |

## Equivalence with gRPC in Production

This project uses `net/rpc` from the stdlib to avoid external dependencies. In a production environment, gRPC would be used. This table shows the equivalence:

| Concept              | This project (net/rpc)           | Production (gRPC)                          |
|-----------------------|----------------------------------|--------------------------------------------|
| **API definition**    | Go structs as arguments          | Protocol Buffers (.proto)                  |
| **Serialization**     | encoding/gob (Go binary)         | Protocol Buffers (multi-language)          |
| **Transport**         | Direct TCP                       | HTTP/2 with multiplexing                   |
| **Streaming**         | Not supported                    | Unidirectional and bidirectional streaming  |
| **Code generation**   | Not needed                       | protoc generates client/server stubs       |
| **Interceptors**      | None (implemented manually)      | Unary and streaming interceptors           |
| **Load balancing**    | Manual                           | gRPC has built-in LB                       |
| **Metadata**          | None                             | Headers/trailers for auth, tracing         |
| **Deadlines**         | Manual with context              | Built into the protocol                    |
| **Multi-language**    | Go only                          | Go, Java, Python, C++, etc.               |

### When to Choose Each

**net/rpc is sufficient when:**
- All services are in Go
- You do not need streaming
- Personal project, prototypes, or learning
- You want zero external dependencies

**gRPC is necessary when:**
- Services in multiple languages
- You need bidirectional streaming
- You require a formal API contract (.proto)
- You need advanced interceptors (auth, tracing, metrics)
- Production environment with multiple teams

## What You Will Learn from This Project

- Decomposing an application into independent microservices
- Synchronous inter-service communication via RPC
- Translating a public REST API to internal RPC calls
- Implementing health checks for monitoring
- Graceful shutdown in concurrent services
- Docker multi-stage builds for minimal images
- Dependency injection for testing (UserValidator mock)
- Configuration via environment variables (12-Factor App)
