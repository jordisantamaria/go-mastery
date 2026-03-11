# System Design — Preparation with Go

Preparation guide for system design interviews focused on Go: when to choose Go, architecture patterns, microservices, scalability, and design questions with solutions.

---

## Table of Contents

1. [When to Choose Go](#when-to-choose-go)
2. [Architecture Patterns in Go](#architecture-patterns-in-go)
3. [Microservices with Go](#microservices-with-go)
4. [Scalability Patterns](#scalability-patterns)
5. [Database Patterns](#database-patterns)
6. [Observability](#observability)
7. [System Design Questions](#system-design-questions)

---

## When to Choose Go

### Comparison with Other Languages

| Criterion | Go | Java | Python | Node.js | Rust |
|---|---|---|---|---|---|
| Concurrency | Goroutines (excellent) | Threads + Virtual Threads | asyncio (limited) | Event loop (single-thread) | async/await (excellent) |
| Compile speed | Very fast (~seconds) | Slow (minutes) | N/A (interpreted) | N/A (interpreted) | Very slow (minutes) |
| Runtime performance | High | High (mature JVM) | Low | Medium | Very high (zero-cost abstractions) |
| Binary size | Small (static) | Large (requires JVM) | N/A | N/A (requires Node) | Small (static) |
| Learning curve | Low | High | Very low | Low | Very high |
| Ecosystem | Good (networking, CLI) | Huge (enterprise) | Huge (ML, scripting) | Large (web) | Growing |
| Memory management | GC (short pauses) | GC (configurable) | GC | GC | Ownership (no GC) |
| Deployment | A single static binary | JAR/WAR + JVM | Interpreter + deps | Node + deps | A single static binary |

### Go's Strengths

1. **Native concurrency**: goroutines and channels are first-class citizens. Creating thousands of goroutines is trivial and cheap.
2. **Fast compilation**: extremely short feedback loop. A large project compiles in seconds.
3. **Simple deployment**: a single static binary with no dependencies. Ideal for containers (Docker images of ~10MB with scratch/distroless).
4. **Excellent standard library**: `net/http`, `encoding/json`, `crypto`, `testing` — you can build a complete web service without external dependencies.
5. **Trivial cross-compilation**: `GOOS=linux GOARCH=amd64 go build` generates a Linux binary from macOS.
6. **Predictable performance**: no JIT warmup, no long GC pauses, no significant runtime overhead.
7. **Integrated tooling**: `go fmt`, `go vet`, `go test`, `go build`, profiler (`pprof`), race detector.

### Go's Weaknesses

1. **Limited generics**: added in Go 1.18, but still no methods with type parameters, no specialization.
2. **No real enums**: simulated with `iota` and custom types, but no native exhaustive checking.
3. **Verbose error handling**: repetitive `if err != nil`. No `try/catch`, no `?` operator (like Rust).
4. **No inheritance**: composition over inheritance is the approach, but it can be verbose for complex hierarchies.
5. **Smaller ecosystem**: fewer libraries than Java/Python for specific domains (ML, data science).
6. **No macros/advanced metaprogramming**: `go generate` exists but is limited compared to Rust macros or Java annotations.

### Go Is Ideal For

- **APIs and web services**: the `net/http` package is production-ready without frameworks.
- **Microservices**: small binaries, fast startup, low memory consumption.
- **CLI tools**: cross-compilation, single binary, `cobra`/`urfave/cli` libraries.
- **Infrastructure tools**: Docker, Kubernetes, Terraform, Prometheus — all written in Go.
- **Network services**: proxies, load balancers, API gateways.
- **Data pipelines**: concurrent processing with goroutines and channels.

### Go Is NOT Ideal For

- **Machine Learning**: Python dominates (TensorFlow, PyTorch).
- **Desktop applications**: better with Electron, Qt, or native frameworks.
- **Strict real-time systems**: the GC introduces non-deterministic latency (use Rust/C).
- **Quick scripting**: Python is more concise for one-off scripts.

---

## Architecture Patterns in Go

### Clean Architecture

Organizes code in concentric layers with the dependency rule: inner layers do not know about outer ones.

```
project/
├── cmd/
│   └── api/
│       └── main.go              # entry point, wiring
├── internal/
│   ├── domain/                   # business entities (no dependencies)
│   │   ├── user.go
│   │   └── order.go
│   ├── usecase/                  # business logic (depends on domain)
│   │   ├── user_service.go
│   │   └── order_service.go
│   ├── adapter/                  # implementations (depends on domain)
│   │   ├── postgres/
│   │   │   ├── user_repo.go
│   │   │   └── order_repo.go
│   │   └── redis/
│   │       └── cache.go
│   └── port/                     # interfaces (defined by the consumer)
│       ├── repository.go         # repository interfaces
│       └── cache.go              # cache interface
├── pkg/                          # reusable libraries
│   └── httputil/
└── go.mod
```

**Key principle in Go**: interfaces are defined where they are **used**, not where they are implemented.

```go
// internal/port/repository.go
// Interfaces are defined by the CONSUMER (usecase layer)
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*domain.User, error)
    Create(ctx context.Context, user *domain.User) error
}

// internal/usecase/user_service.go
type UserService struct {
    repo port.UserRepository // depends on the interface, not the implementation
}

func NewUserService(repo port.UserRepository) *UserService {
    return &UserService{repo: repo}
}
```

### Hexagonal Architecture

Similar to Clean Architecture but emphasizes "ports and adapters":

- **Ports**: interfaces that define how the domain interacts with the outside.
- **Adapters**: concrete implementations (HTTP handler, PostgreSQL repo, Redis cache).
- **Driving adapters**: enter the domain (HTTP handlers, CLI, gRPC).
- **Driven adapters**: the domain reaches the outside (databases, external APIs, message queues).

### Repository Pattern

```go
// The repository abstracts data access
type OrderRepository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
    FindByUserID(ctx context.Context, userID uuid.UUID) ([]*Order, error)
    Save(ctx context.Context, order *Order) error
    Delete(ctx context.Context, id uuid.UUID) error
}

// PostgreSQL implementation
type PostgresOrderRepository struct {
    db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
    return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*Order, error) {
    var order Order
    err := r.db.QueryRowContext(ctx,
        "SELECT id, user_id, total, status FROM orders WHERE id = $1", id,
    ).Scan(&order.ID, &order.UserID, &order.Total, &order.Status)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, ErrOrderNotFound
    }
    return &order, err
}
```

### Service Layer Pattern

```go
type OrderService struct {
    orders  OrderRepository
    users   UserRepository
    events  EventPublisher
    logger  *slog.Logger
}

func NewOrderService(
    orders OrderRepository,
    users UserRepository,
    events EventPublisher,
    logger *slog.Logger,
) *OrderService {
    return &OrderService{
        orders: orders,
        users:  users,
        events: events,
        logger: logger,
    }
}

func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error) {
    // Validation
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }

    // Verify that the user exists
    user, err := s.users.FindByID(ctx, req.UserID)
    if err != nil {
        return nil, fmt.Errorf("finding user: %w", err)
    }

    // Business logic
    order := &Order{
        ID:     uuid.New(),
        UserID: user.ID,
        Total:  req.Total,
        Status: OrderStatusPending,
    }

    // Persist
    if err := s.orders.Save(ctx, order); err != nil {
        return nil, fmt.Errorf("saving order: %w", err)
    }

    // Publish event
    s.events.Publish(ctx, OrderCreatedEvent{OrderID: order.ID})

    return order, nil
}
```

### Dependency Injection Without Frameworks

Go does not need DI frameworks — injection is done manually in `main()`:

```go
func main() {
    // Infrastructure
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    redisClient := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_URL")})
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    // Repositories (driven adapters)
    userRepo := postgres.NewUserRepository(db)
    orderRepo := postgres.NewOrderRepository(db)
    cache := rediscache.NewCache(redisClient)

    // Services (use cases)
    userService := usecase.NewUserService(userRepo, cache, logger)
    orderService := usecase.NewOrderService(orderRepo, userRepo, cache, logger)

    // HTTP handlers (driving adapters)
    userHandler := httphandler.NewUserHandler(userService)
    orderHandler := httphandler.NewOrderHandler(orderService)

    // Router
    mux := http.NewServeMux()
    userHandler.Register(mux)
    orderHandler.Register(mux)

    // Server
    srv := &http.Server{Addr: ":8080", Handler: mux}
    log.Fatal(srv.ListenAndServe())
}
```

**Advantage**: no magic, no reflection, no annotations. The dependency graph is explicit and verifiable at compile time.

---

## Microservices with Go

### Defining Service Boundaries

Principles for dividing services:

1. **Bounded Context (DDD)**: each service handles a clear business context.
2. **Data ownership**: each service owns its database.
3. **Decoupling**: services communicate through contracts, not implementations.
4. **Size**: large enough to have business value, small enough for a team to maintain.

### Inter-Service Communication

#### REST (HTTP/JSON)
```go
// Server
mux := http.NewServeMux()
mux.HandleFunc("GET /api/users/{id}", getUserHandler)

// Client
resp, err := http.Get(fmt.Sprintf("http://user-service/api/users/%s", id))
```
- **Pros**: simple, universal, easy to debug.
- **Cons**: JSON serialization overhead, no typed contract.

#### gRPC (Protocol Buffers)
```protobuf
// user.proto
service UserService {
    rpc GetUser(GetUserRequest) returns (User);
}
```
```go
// gRPC client
client := pb.NewUserServiceClient(conn)
user, err := client.GetUser(ctx, &pb.GetUserRequest{Id: "123"})
```
- **Pros**: strong typing, efficient (binary), streaming, code generation.
- **Cons**: harder to debug, requires extra tooling.

#### Message Queues (NATS, RabbitMQ, Kafka)
```go
// Producer
nc.Publish("orders.created", orderJSON)

// Consumer
nc.Subscribe("orders.created", func(msg *nats.Msg) {
    var order Order
    json.Unmarshal(msg.Data, &order)
    processOrder(order)
})
```
- **Pros**: temporal decoupling, resilience, event-driven.
- **Cons**: operational complexity, eventual consistency.

**Selection guide:**
| Use Case | Protocol |
|---|---|
| Synchronous request-response | REST or gRPC |
| High internal performance | gRPC |
| Asynchronous events | Message queues |
| Fire-and-forget notifications | Message queues |

### Circuit Breaker Pattern

Prevents failure cascades when a downstream service is down:

```go
type CircuitBreaker struct {
    mu            sync.Mutex
    failureCount  int
    threshold     int
    state         State // Closed, Open, HalfOpen
    lastFailure   time.Time
    resetTimeout  time.Duration
}

type State int

const (
    StateClosed   State = iota // Functioning normally
    StateOpen                   // Rejecting requests (service down)
    StateHalfOpen              // Testing if it recovered
)

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mu.Lock()
    if cb.state == StateOpen {
        if time.Since(cb.lastFailure) > cb.resetTimeout {
            cb.state = StateHalfOpen
        } else {
            cb.mu.Unlock()
            return ErrCircuitOpen
        }
    }
    cb.mu.Unlock()

    err := fn()

    cb.mu.Lock()
    defer cb.mu.Unlock()

    if err != nil {
        cb.failureCount++
        cb.lastFailure = time.Now()
        if cb.failureCount >= cb.threshold {
            cb.state = StateOpen
        }
        return err
    }

    cb.failureCount = 0
    cb.state = StateClosed
    return nil
}
```

In production, use libraries like `sony/gobreaker` or `afex/hystrix-go`.

### Health Checks and Readiness Probes

```go
// Liveness: is the process alive?
mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
})

// Readiness: can it receive traffic?
mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
    // Check critical dependencies
    if err := db.PingContext(r.Context()); err != nil {
        http.Error(w, "database unavailable", http.StatusServiceUnavailable)
        return
    }
    if err := redisClient.Ping(r.Context()).Err(); err != nil {
        http.Error(w, "cache unavailable", http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ready"))
})
```

For Kubernetes:
```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
readinessProbe:
  httpGet:
    path: /readyz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

---

## Scalability Patterns

### Worker Pool Pattern

For processing work in parallel with concurrency control:

```go
func WorkerPool[T any, R any](
    ctx context.Context,
    numWorkers int,
    jobs <-chan T,
    process func(context.Context, T) (R, error),
) <-chan Result[R] {
    results := make(chan Result[R], numWorkers)

    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                select {
                case <-ctx.Done():
                    return
                case job, ok := <-jobs:
                    if !ok {
                        return
                    }
                    val, err := process(ctx, job)
                    results <- Result[R]{Value: val, Err: err}
                }
            }
        }()
    }

    go func() {
        wg.Wait()
        close(results)
    }()

    return results
}

type Result[T any] struct {
    Value T
    Err   error
}
```

### Rate Limiting

**Token Bucket** (the most common):

```go
// Using the standard package (golang.org/x/time/rate)
limiter := rate.NewLimiter(rate.Limit(100), 10) // 100 req/s, burst of 10

func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

**Per-client rate limiting:**
```go
type ClientLimiter struct {
    mu       sync.Mutex
    clients  map[string]*rate.Limiter
    rate     rate.Limit
    burst    int
}

func (cl *ClientLimiter) GetLimiter(clientID string) *rate.Limiter {
    cl.mu.Lock()
    defer cl.mu.Unlock()

    limiter, exists := cl.clients[clientID]
    if !exists {
        limiter = rate.NewLimiter(cl.rate, cl.burst)
        cl.clients[clientID] = limiter
    }
    return limiter
}
```

### Connection Pooling

```go
// HTTP client with connection pool
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}

// IMPORTANT: never create an http.Client per request — reuse it
```

### Caching Strategies

```go
// Simple in-memory cache with sync.Map
type Cache struct {
    data sync.Map
}

type cacheEntry struct {
    value     any
    expiresAt time.Time
}

func (c *Cache) Get(key string) (any, bool) {
    entry, ok := c.data.Load(key)
    if !ok {
        return nil, false
    }
    e := entry.(cacheEntry)
    if time.Now().After(e.expiresAt) {
        c.data.Delete(key)
        return nil, false
    }
    return e.value, true
}

func (c *Cache) Set(key string, value any, ttl time.Duration) {
    c.data.Store(key, cacheEntry{
        value:     value,
        expiresAt: time.Now().Add(ttl),
    })
}
```

### Graceful Shutdown

```go
func main() {
    srv := &http.Server{Addr: ":8080", Handler: mux}

    // Channel for server errors
    serverErr := make(chan error, 1)
    go func() {
        serverErr <- srv.ListenAndServe()
    }()

    // Wait for shutdown signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    select {
    case err := <-serverErr:
        log.Fatal("server error:", err)
    case sig := <-quit:
        log.Printf("shutdown signal received: %v", sig)
    }

    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("forced shutdown:", err)
    }

    log.Println("server stopped gracefully")
}
```

**Zero-downtime deploys**: the new process starts and passes health checks before the old one receives the shutdown signal. Kubernetes handles this natively with rolling deployments.

---

## Database Patterns

### Connection Pooling with sql.DB

```go
db, err := sql.Open("postgres", connString)
if err != nil {
    log.Fatal(err)
}

// Pool configuration — CRITICAL for production
db.SetMaxOpenConns(25)              // Maximum active connections
db.SetMaxIdleConns(10)              // Idle connections in the pool
db.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection
db.SetConnMaxIdleTime(1 * time.Minute) // Maximum idle time before closing
```

**Configuration rules:**
| Parameter | Rule | Reason |
|---|---|---|
| `MaxOpenConns` | Lower than the DB limit | Avoid saturating the DB |
| `MaxIdleConns` | ~50% of MaxOpenConns | Balance between reuse and memory |
| `ConnMaxLifetime` | < firewall/LB timeout | Avoid broken connections |
| `ConnMaxIdleTime` | Less than ConnMaxLifetime | Free idle connections |

### Transactions

```go
func (r *OrderRepository) CreateWithItems(
    ctx context.Context,
    order *Order,
    items []*OrderItem,
) error {
    tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelReadCommitted,
    })
    if err != nil {
        return fmt.Errorf("beginning transaction: %w", err)
    }
    // defer Rollback is safe — it does nothing if Commit was already called
    defer tx.Rollback()

    // Insert order
    _, err = tx.ExecContext(ctx,
        "INSERT INTO orders (id, user_id, total) VALUES ($1, $2, $3)",
        order.ID, order.UserID, order.Total,
    )
    if err != nil {
        return fmt.Errorf("inserting order: %w", err)
    }

    // Insert items
    for _, item := range items {
        _, err = tx.ExecContext(ctx,
            "INSERT INTO order_items (id, order_id, product_id, qty) VALUES ($1, $2, $3, $4)",
            item.ID, order.ID, item.ProductID, item.Quantity,
        )
        if err != nil {
            return fmt.Errorf("inserting item: %w", err)
        }
    }

    return tx.Commit()
}
```

### Isolation Levels

| Level | Dirty Read | Non-repeatable Read | Phantom Read | Use |
|---|---|---|---|---|
| Read Uncommitted | Yes | Yes | Yes | Almost never |
| Read Committed | No | Yes | Yes | Default PostgreSQL |
| Repeatable Read | No | No | Yes* | Reports |
| Serializable | No | No | No | Critical transactions |

*PostgreSQL prevents phantom reads in Repeatable Read.

### Migration Strategies

```go
// Example with golang-migrate
import "github.com/golang-migrate/migrate/v4"

m, err := migrate.New(
    "file://migrations",
    "postgres://localhost:5432/mydb?sslmode=disable",
)
if err != nil {
    log.Fatal(err)
}

// Apply all pending migrations
if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
    log.Fatal(err)
}
```

### Repository with Interfaces

```go
// Define interface (in the package that uses it)
type UserRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, opts ListOptions) ([]*User, int, error)
}

// Mock for tests
type MockUserRepository struct {
    users map[uuid.UUID]*User
}

func (m *MockUserRepository) GetByID(_ context.Context, id uuid.UUID) (*User, error) {
    user, ok := m.users[id]
    if !ok {
        return nil, ErrNotFound
    }
    return user, nil
}
```

---

## Observability

### Structured Logging with slog

```go
// Configuration
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

// Usage
slog.Info("order created",
    slog.String("order_id", order.ID.String()),
    slog.String("user_id", order.UserID.String()),
    slog.Float64("total", order.Total),
    slog.Duration("latency", elapsed),
)

// JSON output:
// {"time":"2025-01-15T10:30:00Z","level":"INFO","msg":"order created",
//  "order_id":"abc-123","user_id":"user-456","total":99.99,"latency":"15ms"}

// Logger with fixed fields (ideal for requests)
requestLogger := logger.With(
    slog.String("request_id", requestID),
    slog.String("method", r.Method),
    slog.String("path", r.URL.Path),
)
```

### Metrics with Prometheus

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

// Middleware
func metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

        next.ServeHTTP(wrapped, r)

        duration := time.Since(start).Seconds()
        httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(wrapped.statusCode)).Inc()
        httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
    })
}
```

### Distributed Tracing with OpenTelemetry

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("order-service")

func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error) {
    ctx, span := tracer.Start(ctx, "OrderService.CreateOrder")
    defer span.End()

    // The span is automatically propagated through the context
    user, err := s.users.GetByID(ctx, req.UserID) // creates sub-span
    if err != nil {
        span.RecordError(err)
        return nil, err
    }

    span.SetAttributes(
        attribute.String("user.id", user.ID.String()),
        attribute.Float64("order.total", req.Total),
    )

    // ... rest of the logic
}
```

### Health Check Endpoints

```go
type HealthChecker struct {
    checks map[string]func(ctx context.Context) error
}

func (h *HealthChecker) AddCheck(name string, check func(ctx context.Context) error) {
    h.checks[name] = check
}

func (h *HealthChecker) Handler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()

        results := make(map[string]string)
        healthy := true

        for name, check := range h.checks {
            if err := check(ctx); err != nil {
                results[name] = "unhealthy: " + err.Error()
                healthy = false
            } else {
                results[name] = "healthy"
            }
        }

        status := http.StatusOK
        if !healthy {
            status = http.StatusServiceUnavailable
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(status)
        json.NewEncoder(w).Encode(results)
    }
}
```

---

## System Design Questions

### Question 1: Design a URL Shortener

**Requirements**: given a long URL, generate a short URL. Given a short URL, redirect to the original.

**Main components:**

```
┌──────────┐     ┌──────────────┐     ┌───────────┐
│  Client  │────►│  API Server  │────►│ PostgreSQL│
│          │◄────│  (Go)        │◄────│           │
└──────────┘     └──────┬───────┘     └───────────┘
                        │
                   ┌────▼────┐
                   │  Redis  │  (cache for popular URLs)
                   └─────────┘
```

**Implementation notes in Go:**

```go
// Short ID generation with base62
const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func encodeBase62(num uint64) string {
    if num == 0 {
        return string(base62Chars[0])
    }
    var result []byte
    for num > 0 {
        result = append(result, base62Chars[num%62])
        num /= 62
    }
    slices.Reverse(result)
    return string(result)
}

// Redirect handler — high performance
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
    shortCode := r.PathValue("code")

    // 1. Look up in cache (Redis)
    if url, err := h.cache.Get(r.Context(), shortCode); err == nil {
        http.Redirect(w, r, url, http.StatusMovedPermanently)
        return
    }

    // 2. Look up in DB
    url, err := h.repo.GetByShortCode(r.Context(), shortCode)
    if err != nil {
        http.NotFound(w, r)
        return
    }

    // 3. Save in cache for future requests
    h.cache.Set(r.Context(), shortCode, url.OriginalURL, 24*time.Hour)

    http.Redirect(w, r, url.OriginalURL, http.StatusMovedPermanently)
}
```

**Scalability**: use Snowflake IDs or a distributed counter to generate unique IDs. DB sharding by short code hash. CDN for redirects of very popular URLs.

---

### Question 2: Design a Rate Limiter

**Requirements**: limit requests per client (API key or IP).

**Token Bucket Algorithm in Go:**

```go
type TokenBucket struct {
    mu         sync.Mutex
    tokens     float64
    maxTokens  float64
    refillRate float64    // tokens per second
    lastRefill time.Time
}

func NewTokenBucket(maxTokens, refillRate float64) *TokenBucket {
    return &TokenBucket{
        tokens:     maxTokens,
        maxTokens:  maxTokens,
        refillRate: refillRate,
        lastRefill: time.Now(),
    }
}

func (tb *TokenBucket) Allow() bool {
    tb.mu.Lock()
    defer tb.mu.Unlock()

    now := time.Now()
    elapsed := now.Sub(tb.lastRefill).Seconds()
    tb.tokens = min(tb.maxTokens, tb.tokens+elapsed*tb.refillRate)
    tb.lastRefill = now

    if tb.tokens >= 1 {
        tb.tokens--
        return true
    }
    return false
}
```

**For a distributed system**: use Redis with Lua scripts for atomicity:
```go
// Sliding window counter with Redis
func (rl *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {
    pipe := rl.redis.Pipeline()
    now := time.Now().UnixMilli()
    windowStart := now - rl.windowMs

    pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
    pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
    countCmd := pipe.ZCard(ctx, key)
    pipe.Expire(ctx, key, time.Duration(rl.windowMs)*time.Millisecond)

    _, err := pipe.Exec(ctx)
    if err != nil {
        return false, err
    }
    return countCmd.Val() <= int64(rl.maxRequests), nil
}
```

---

### Question 3: Design a Chat System

**Requirements**: real-time chat between users. Support for rooms and direct messages.

**Architecture:**

```
┌──────────┐  WebSocket  ┌──────────────┐  Pub/Sub  ┌───────┐
│ Client A │────────────►│  Go Server   │◄─────────►│ Redis │
│          │◄────────────│  (goroutine  │            │ or    │
└──────────┘             │   per conn)  │            │ NATS  │
                         └──────┬───────┘            └───────┘
┌──────────┐                    │
│ Client B │◄──────────────────►│
└──────────┘                    │
                          ┌─────▼──────┐
                          │ PostgreSQL │  (message persistence)
                          └────────────┘
```

**Go is ideal because**: one goroutine per WebSocket connection is trivial and efficient. With 10K connections, you only need ~10K goroutines (~80MB of stacks).

```go
type Hub struct {
    rooms      map[string]map[*Client]struct{}
    mu         sync.RWMutex
    register   chan *Client
    unregister chan *Client
    broadcast  chan Message
}

type Client struct {
    conn   *websocket.Conn
    roomID string
    send   chan []byte
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            if h.rooms[client.roomID] == nil {
                h.rooms[client.roomID] = make(map[*Client]struct{})
            }
            h.rooms[client.roomID][client] = struct{}{}
            h.mu.Unlock()

        case client := <-h.unregister:
            h.mu.Lock()
            if clients, ok := h.rooms[client.roomID]; ok {
                delete(clients, client)
                close(client.send)
            }
            h.mu.Unlock()

        case msg := <-h.broadcast:
            h.mu.RLock()
            for client := range h.rooms[msg.RoomID] {
                select {
                case client.send <- msg.Data:
                default:
                    close(client.send)
                    delete(h.rooms[msg.RoomID], client)
                }
            }
            h.mu.RUnlock()
        }
    }
}
```

**Multi-server scalability**: Redis Pub/Sub or NATS to propagate messages between server instances.

---

### Question 4: Design a Task Queue

**Requirements**: asynchronous task system with retries, priorities, and concurrent processing.

```go
type Task struct {
    ID       string
    Payload  []byte
    Priority int
    Retries  int
    MaxRetry int
}

type TaskQueue struct {
    tasks   chan Task
    workers int
    handler func(context.Context, Task) error
    logger  *slog.Logger
}

func NewTaskQueue(workers, bufferSize int, handler func(context.Context, Task) error) *TaskQueue {
    return &TaskQueue{
        tasks:   make(chan Task, bufferSize),
        workers: workers,
        handler: handler,
        logger:  slog.Default(),
    }
}

func (tq *TaskQueue) Start(ctx context.Context) {
    var wg sync.WaitGroup

    for i := 0; i < tq.workers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            tq.logger.Info("worker started", slog.Int("worker_id", workerID))

            for {
                select {
                case <-ctx.Done():
                    tq.logger.Info("worker stopping", slog.Int("worker_id", workerID))
                    return
                case task, ok := <-tq.tasks:
                    if !ok {
                        return
                    }
                    if err := tq.processWithRetry(ctx, task); err != nil {
                        tq.logger.Error("task failed permanently",
                            slog.String("task_id", task.ID),
                            slog.Any("error", err),
                        )
                    }
                }
            }
        }(i)
    }

    wg.Wait()
}

func (tq *TaskQueue) processWithRetry(ctx context.Context, task Task) error {
    var lastErr error
    for attempt := 0; attempt <= task.MaxRetry; attempt++ {
        if err := tq.handler(ctx, task); err != nil {
            lastErr = err
            backoff := time.Duration(1<<attempt) * time.Second // exponential backoff
            tq.logger.Warn("task failed, retrying",
                slog.String("task_id", task.ID),
                slog.Int("attempt", attempt+1),
                slog.Duration("backoff", backoff),
            )
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(backoff):
            }
            continue
        }
        return nil
    }
    return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (tq *TaskQueue) Submit(task Task) error {
    select {
    case tq.tasks <- task:
        return nil
    default:
        return errors.New("task queue is full")
    }
}
```

**Go is ideal because**: workers are goroutines, the queue is a buffered channel, and coordination is done with `select` and `context`. No heavy frameworks needed.

---

### Question 5: Design a Cache with TTL

**Requirements**: in-memory cache with per-entry expiration time. Must be thread-safe.

```go
type TTLCache[K comparable, V any] struct {
    mu      sync.RWMutex
    items   map[K]*cacheItem[V]
    onEvict func(K, V) // optional callback on expiration
}

type cacheItem[V any] struct {
    value     V
    expiresAt time.Time
    timer     *time.Timer
}

func NewTTLCache[K comparable, V any]() *TTLCache[K, V] {
    return &TTLCache[K, V]{
        items: make(map[K]*cacheItem[V]),
    }
}

func (c *TTLCache[K, V]) Set(key K, value V, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // If it already exists, cancel the previous timer
    if existing, ok := c.items[key]; ok {
        existing.timer.Stop()
    }

    item := &cacheItem[V]{
        value:     value,
        expiresAt: time.Now().Add(ttl),
    }

    // Timer for auto-eviction
    item.timer = time.AfterFunc(ttl, func() {
        c.mu.Lock()
        defer c.mu.Unlock()

        // Verify it was not rewritten
        if current, ok := c.items[key]; ok && current == item {
            delete(c.items, key)
            if c.onEvict != nil {
                c.onEvict(key, value)
            }
        }
    })

    c.items[key] = item
}

func (c *TTLCache[K, V]) Get(key K) (V, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    item, ok := c.items[key]
    if !ok {
        var zero V
        return zero, false
    }

    if time.Now().After(item.expiresAt) {
        var zero V
        return zero, false
    }

    return item.value, true
}

func (c *TTLCache[K, V]) Delete(key K) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if item, ok := c.items[key]; ok {
        item.timer.Stop()
        delete(c.items, key)
    }
}

func (c *TTLCache[K, V]) Len() int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return len(c.items)
}
```

**Scalability considerations:**
- For high concurrency: shard the cache into N maps with independent locks (reduces contention).
- To avoid memory pressure: limit the number of entries (LRU eviction).
- For distributed systems: migrate to Redis/Memcached.

**Go is ideal because**: `sync.RWMutex` for concurrency, `time.AfterFunc` for efficient timers, generics for type safety. All with the standard library.
