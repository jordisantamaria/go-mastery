# Project 2: Finance Tracker REST API

REST API for personal finance tracking (income and expenses). Demonstrates clean architecture, professional backend development, and deep mastery of Go's stdlib.

**No external dependencies** — only Go's stdlib (`net/http`, `crypto`, `encoding/json`, etc.).

## Architecture

```
02-rest-api/
├── cmd/
│   └── api/
│       └── main.go              # Entrypoint: dependency wiring, server
├── internal/
│   ├── model/
│   │   ├── transaction.go       # Transaction domain type
│   │   └── user.go              # User domain type
│   ├── handler/
│   │   ├── transaction.go       # HTTP handlers for transactions (CRUD)
│   │   ├── auth.go              # Registration and login handlers
│   │   └── health.go            # Health check
│   ├── service/
│   │   ├── transaction.go       # Transaction business logic
│   │   └── auth.go              # Authentication logic (JWT + hashing)
│   ├── repository/
│   │   ├── interfaces.go        # Repository interfaces
│   │   ├── memory.go            # In-memory implementation (maps + RWMutex)
│   │   └── memory_test.go       # Repository tests
│   └── middleware/
│       ├── auth.go              # JWT middleware (extracts user from token)
│       ├── logging.go           # Structured logging of each request
│       ├── recovery.go          # Panic recovery
│       └── cors.go              # CORS headers
├── pkg/
│   └── jwt/
│       ├── jwt.go               # JWT HS256 implementation (no libraries)
│       └── jwt_test.go          # JWT tests
├── go.mod
└── README.md
```

### Dependency Flow

```
Handler → Service → Repository (interface)
                         ↑
                    MemoryStore (implementation)
```

Handlers never access the repository directly. The service contains all business logic and validation. The repository is an interface that can be swapped (memory, PostgreSQL, etc.).

## API Endpoints

| Method | Route                      | Auth | Description                          |
|--------|---------------------------|------|--------------------------------------|
| POST   | `/api/auth/register`      | No   | Register new user                    |
| POST   | `/api/auth/login`         | No   | Login, returns JWT                   |
| GET    | `/api/transactions`       | Yes  | List transactions (with filters)     |
| POST   | `/api/transactions`       | Yes  | Create transaction                   |
| GET    | `/api/transactions/{id}`  | Yes  | Get a transaction                    |
| PUT    | `/api/transactions/{id}`  | Yes  | Update transaction                   |
| DELETE | `/api/transactions/{id}`  | Yes  | Delete transaction                   |
| GET    | `/api/health`             | No   | Health check                         |

### Listing Filters

```
GET /api/transactions?type=income&category=food&from=2024-01-01&to=2024-12-31&page=1&limit=10
```

| Parameter  | Type   | Description                        |
|------------|--------|------------------------------------|
| `type`     | string | `income` or `expense`              |
| `category` | string | Category (food, transport, etc.)   |
| `from`     | string | Start date (YYYY-MM-DD)            |
| `to`       | string | End date (YYYY-MM-DD)              |
| `page`     | int    | Page (default: 1)                  |
| `limit`    | int    | Results per page (default: 10)     |

## How to Run

```bash
# From the project root
cd 03-projects/02-rest-api

# Build
go build ./...

# Run (port 8080 by default)
go run ./cmd/api/

# With custom configuration
PORT=3000 JWT_SECRET=my-secret-key go run ./cmd/api/
```

## Examples with curl

### Register User

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "maria@example.com",
    "password": "secret123",
    "name": "Maria Garcia"
  }'
```

Response (201):
```json
{
  "id": "a1b2c3d4...",
  "email": "maria@example.com",
  "name": "Maria Garcia"
}
```

### Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "maria@example.com",
    "password": "secret123"
  }'
```

Response (200):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "a1b2c3d4...",
    "email": "maria@example.com",
    "name": "Maria Garcia"
  }
}
```

### Create Transaction

```bash
curl -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -d '{
    "type": "expense",
    "amount": 45.50,
    "category": "food",
    "description": "Dinner at restaurant",
    "date": "2024-06-15"
  }'
```

### List Transactions with Filters

```bash
# All transactions
curl http://localhost:8080/api/transactions \
  -H "Authorization: Bearer <your-jwt-token>"

# Only food expenses from the last month
curl "http://localhost:8080/api/transactions?type=expense&category=food&from=2024-06-01&to=2024-06-30" \
  -H "Authorization: Bearer <your-jwt-token>"
```

Response (200):
```json
{
  "data": [
    {
      "id": "tx-abc123",
      "user_id": "a1b2c3d4...",
      "type": "expense",
      "amount": 45.50,
      "category": "food",
      "description": "Dinner at restaurant",
      "date": "2024-06-15T00:00:00Z",
      "created_at": "2024-06-15T20:30:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10,
  "total_pages": 1
}
```

### Update Transaction

```bash
curl -X PUT http://localhost:8080/api/transactions/<id> \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -d '{
    "amount": 50.00,
    "category": "restaurant"
  }'
```

### Delete Transaction

```bash
curl -X DELETE http://localhost:8080/api/transactions/<id> \
  -H "Authorization: Bearer <your-jwt-token>"
```

### Health Check

```bash
curl http://localhost:8080/api/health
```

Response (200):
```json
{
  "status": "ok",
  "uptime": "2h30m15s",
  "version": "1.0.0"
}
```

## How to Run Tests

```bash
# All tests
go test ./...

# With verbose
go test -v ./...

# Only JWT tests
go test -v ./pkg/jwt/

# Only repository tests
go test -v ./internal/repository/

# With coverage
go test -cover ./...
```

## Patterns Demonstrated

### Clean Architecture
- **Layer separation**: handler (HTTP) -> service (business) -> repository (data)
- **Dependency inversion**: services depend on interfaces, not implementations
- **Domain models**: domain types do not depend on the HTTP layer or the database

### Stdlib Mastery
- **Go 1.22+ ServeMux**: method-based routing (`"GET /api/transactions/{id}"`)
- **JWT from scratch**: HS256 implementation with `crypto/hmac` + `crypto/sha256`
- **Password hashing**: SHA-256 with random salt using `crypto/rand`
- **Graceful shutdown**: `signal.Notify` + `server.Shutdown` with timeout

### Design Patterns
- **Dependency Injection**: constructors that receive interfaces (`NewTransactionService(repo)`)
- **Middleware Chain**: Recovery -> CORS -> Logging -> Auth (functional composition)
- **Repository Pattern**: interface + swappable implementation
- **Constructor Pattern**: `NewXxxService`, `NewXxxHandler`, `NewXxxRepository`

### Concurrency
- **sync.RWMutex**: the in-memory store uses read-write locks for safe concurrent access
- **context.Context**: context propagation throughout the call chain

## How to Extend with a Real Database

The project is designed to easily swap the store thanks to interfaces:

1. **Create a PostgreSQL implementation**:
   ```go
   // internal/repository/postgres.go
   type PostgresTransactionRepository struct {
       db *sql.DB
   }

   func (r *PostgresTransactionRepository) Create(ctx context.Context, tx *model.Transaction) error {
       _, err := r.db.ExecContext(ctx,
           "INSERT INTO transactions (id, user_id, type, amount, category, description, date, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)",
           tx.ID, tx.UserID, tx.Type, tx.Amount, tx.Category, tx.Description, tx.Date, tx.CreatedAt,
       )
       return err
   }
   // ... implement the rest of the interface methods
   ```

2. **Change the wiring in main.go**:
   ```go
   // Before (in-memory):
   txRepo := repository.NewMemoryTransactionRepository()

   // After (PostgreSQL):
   db, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))
   txRepo := repository.NewPostgresTransactionRepository(db)
   ```

3. **The rest of the code does not change** — services and handlers continue to work exactly the same because they depend on the interface, not the implementation.

## Environment Variables

| Variable     | Default                                      | Description              |
|-------------|----------------------------------------------|--------------------------|
| `PORT`      | `8080`                                       | Server port              |
| `JWT_SECRET`| `super-secret-key-change-in-production`      | Secret key for JWT       |
