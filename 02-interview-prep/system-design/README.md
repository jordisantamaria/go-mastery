# System Design — Preparacion con Go

Guia de preparacion para entrevistas de system design enfocada en Go: cuando elegir Go, patrones de arquitectura, microservicios, escalabilidad, y preguntas de diseno con soluciones.

---

## Tabla de Contenidos

1. [Cuando Elegir Go](#cuando-elegir-go)
2. [Patrones de Arquitectura en Go](#patrones-de-arquitectura-en-go)
3. [Microservicios con Go](#microservicios-con-go)
4. [Patrones de Escalabilidad](#patrones-de-escalabilidad)
5. [Patrones de Base de Datos](#patrones-de-base-de-datos)
6. [Observabilidad](#observabilidad)
7. [Preguntas de System Design](#preguntas-de-system-design)

---

## Cuando Elegir Go

### Comparacion con Otros Lenguajes

| Criterio | Go | Java | Python | Node.js | Rust |
|---|---|---|---|---|---|
| Concurrencia | Goroutines (excelente) | Threads + Virtual Threads | asyncio (limitado) | Event loop (single-thread) | async/await (excelente) |
| Velocidad compilacion | Muy rapida (~segundos) | Lenta (minutos) | N/A (interpretado) | N/A (interpretado) | Muy lenta (minutos) |
| Rendimiento runtime | Alto | Alto (JVM maduro) | Bajo | Medio | Muy alto (zero-cost abstractions) |
| Tamano binario | Pequeno (estatico) | Grande (requiere JVM) | N/A | N/A (requiere Node) | Pequeno (estatico) |
| Curva aprendizaje | Baja | Alta | Muy baja | Baja | Muy alta |
| Ecosistema | Bueno (networking, CLI) | Enorme (enterprise) | Enorme (ML, scripting) | Grande (web) | Creciendo |
| Gestion memoria | GC (pausas cortas) | GC (configurable) | GC | GC | Ownership (sin GC) |
| Deploy | Un binario estatico | JAR/WAR + JVM | Interprete + deps | Node + deps | Un binario estatico |

### Fortalezas de Go

1. **Concurrencia nativa**: goroutines y canales son ciudadanos de primera clase. Crear miles de goroutines es trivial y barato.
2. **Compilacion rapida**: feedback loop extremadamente corto. Un proyecto grande compila en segundos.
3. **Deployment simple**: un unico binario estatico sin dependencias. Ideal para contenedores (imagenes Docker de ~10MB con scratch/distroless).
4. **Standard library excelente**: `net/http`, `encoding/json`, `crypto`, `testing` — se puede construir un servicio web completo sin dependencias externas.
5. **Cross-compilation trivial**: `GOOS=linux GOARCH=amd64 go build` genera un binario para Linux desde macOS.
6. **Performance predecible**: sin JIT warmup, sin GC pauses largos, sin runtime overhead significativo.
7. **Tooling integrado**: `go fmt`, `go vet`, `go test`, `go build`, profiler (`pprof`), race detector.

### Debilidades de Go

1. **Generics limitados**: anadidos en Go 1.18, pero aun sin metodos con type parameters, sin especializacion.
2. **Sin enums reales**: se simulan con `iota` y tipos personalizados, pero sin exhaustive checking nativo.
3. **Error handling verboso**: `if err != nil` repetitivo. Sin `try/catch`, sin `?` operator (como Rust).
4. **Sin herencia**: composicion sobre herencia es el enfoque, pero puede ser verboso para jerarquias complejas.
5. **Ecosistema mas pequeno**: menos librerias que Java/Python para dominios especificos (ML, data science).
6. **Sin macros/metaprogramacion avanzada**: `go generate` existe pero es limitado comparado con macros de Rust o annotations de Java.

### Go Es Ideal Para

- **APIs y servicios web**: el paquete `net/http` es de produccion sin frameworks.
- **Microservicios**: binarios pequenos, startup rapido, bajo consumo de memoria.
- **CLI tools**: compilacion cruzada, binario unico, libreria `cobra`/`urfave/cli`.
- **Infrastructure tools**: Docker, Kubernetes, Terraform, Prometheus — todos escritos en Go.
- **Network services**: proxies, load balancers, API gateways.
- **Data pipelines**: procesamiento concurrente con goroutines y canales.

### Go NO Es Ideal Para

- **Machine Learning**: Python domina (TensorFlow, PyTorch).
- **Aplicaciones de escritorio**: mejor Electron, Qt, o nativas.
- **Sistemas de tiempo real estricto**: el GC introduce latencia no determinista (usar Rust/C).
- **Scripting rapido**: Python es mas conciso para scripts de una vez.

---

## Patrones de Arquitectura en Go

### Clean Architecture

Organiza el codigo en capas concentricas con la regla de dependencia: las capas internas no conocen las externas.

```
project/
├── cmd/
│   └── api/
│       └── main.go              # punto de entrada, wiring
├── internal/
│   ├── domain/                   # entidades de negocio (sin dependencias)
│   │   ├── user.go
│   │   └── order.go
│   ├── usecase/                  # logica de negocio (depende de domain)
│   │   ├── user_service.go
│   │   └── order_service.go
│   ├── adapter/                  # implementaciones (depende de domain)
│   │   ├── postgres/
│   │   │   ├── user_repo.go
│   │   │   └── order_repo.go
│   │   └── redis/
│   │       └── cache.go
│   └── port/                     # interfaces (definidas por quien las usa)
│       ├── repository.go         # interfaces de repositorio
│       └── cache.go              # interface de cache
├── pkg/                          # librerias reutilizables
│   └── httputil/
└── go.mod
```

**Principio clave en Go**: las interfaces se definen donde se **usan**, no donde se implementan.

```go
// internal/port/repository.go
// Las interfaces las define el CONSUMIDOR (usecase layer)
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*domain.User, error)
    Create(ctx context.Context, user *domain.User) error
}

// internal/usecase/user_service.go
type UserService struct {
    repo port.UserRepository // depende de la interface, no de la implementacion
}

func NewUserService(repo port.UserRepository) *UserService {
    return &UserService{repo: repo}
}
```

### Hexagonal Architecture

Similar a Clean Architecture pero enfatiza los "ports and adapters":

- **Ports**: interfaces que definen como el dominio interactua con el exterior.
- **Adapters**: implementaciones concretas (HTTP handler, PostgreSQL repo, Redis cache).
- **Driving adapters**: entran al dominio (HTTP handlers, CLI, gRPC).
- **Driven adapters**: el dominio sale al exterior (databases, APIs externas, message queues).

### Repository Pattern

```go
// El repositorio abstrae el acceso a datos
type OrderRepository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
    FindByUserID(ctx context.Context, userID uuid.UUID) ([]*Order, error)
    Save(ctx context.Context, order *Order) error
    Delete(ctx context.Context, id uuid.UUID) error
}

// Implementacion con PostgreSQL
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
    // Validacion
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }

    // Verificar que el usuario existe
    user, err := s.users.FindByID(ctx, req.UserID)
    if err != nil {
        return nil, fmt.Errorf("finding user: %w", err)
    }

    // Logica de negocio
    order := &Order{
        ID:     uuid.New(),
        UserID: user.ID,
        Total:  req.Total,
        Status: OrderStatusPending,
    }

    // Persistir
    if err := s.orders.Save(ctx, order); err != nil {
        return nil, fmt.Errorf("saving order: %w", err)
    }

    // Publicar evento
    s.events.Publish(ctx, OrderCreatedEvent{OrderID: order.ID})

    return order, nil
}
```

### Dependency Injection Sin Frameworks

Go no necesita frameworks de DI — la inyeccion se hace manualmente en `main()`:

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

**Ventaja**: sin magia, sin reflection, sin annotations. El grafo de dependencias es explicito y verificable en compile time.

---

## Microservicios con Go

### Definicion de Service Boundaries

Principios para dividir servicios:

1. **Bounded Context (DDD)**: cada servicio maneja un contexto de negocio claro.
2. **Ownership de datos**: cada servicio es dueno de su base de datos.
3. **Desacoplamiento**: los servicios se comunican por contratos, no por implementacion.
4. **Tamano**: lo suficientemente grande para tener valor de negocio, lo suficientemente pequeno para que un equipo lo mantenga.

### Comunicacion entre Servicios

#### REST (HTTP/JSON)
```go
// Servidor
mux := http.NewServeMux()
mux.HandleFunc("GET /api/users/{id}", getUserHandler)

// Cliente
resp, err := http.Get(fmt.Sprintf("http://user-service/api/users/%s", id))
```
- **Pros**: simple, universal, facil de debugear.
- **Contras**: overhead de serializacion JSON, sin contrato tipado.

#### gRPC (Protocol Buffers)
```protobuf
// user.proto
service UserService {
    rpc GetUser(GetUserRequest) returns (User);
}
```
```go
// Cliente gRPC
client := pb.NewUserServiceClient(conn)
user, err := client.GetUser(ctx, &pb.GetUserRequest{Id: "123"})
```
- **Pros**: tipado fuerte, eficiente (binary), streaming, generacion de codigo.
- **Contras**: mas complejo de debugear, requiere tooling extra.

#### Message Queues (NATS, RabbitMQ, Kafka)
```go
// Productor
nc.Publish("orders.created", orderJSON)

// Consumidor
nc.Subscribe("orders.created", func(msg *nats.Msg) {
    var order Order
    json.Unmarshal(msg.Data, &order)
    processOrder(order)
})
```
- **Pros**: desacoplamiento temporal, resiliencia, event-driven.
- **Contras**: complejidad operacional, eventual consistency.

**Guia de eleccion:**
| Caso | Protocolo |
|---|---|
| Request-response sincronico | REST o gRPC |
| Alta performance interna | gRPC |
| Eventos asincronos | Message queues |
| Notificaciones fire-and-forget | Message queues |

### Circuit Breaker Pattern

Previene cascadas de fallos cuando un servicio downstream esta caido:

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
    StateClosed   State = iota // Funcionando normal
    StateOpen                   // Rechazando requests (servicio caido)
    StateHalfOpen              // Probando si se recupero
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

En produccion, usar librerias como `sony/gobreaker` o `afex/hystrix-go`.

### Health Checks y Readiness Probes

```go
// Liveness: el proceso esta vivo?
mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
})

// Readiness: puede recibir trafico?
mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
    // Verificar dependencias criticas
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

Para Kubernetes:
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

## Patrones de Escalabilidad

### Worker Pool Pattern

Para procesar trabajo en paralelo con control de concurrencia:

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

**Token Bucket** (el mas comun):

```go
// Usando el paquete estandar (golang.org/x/time/rate)
limiter := rate.NewLimiter(rate.Limit(100), 10) // 100 req/s, burst de 10

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
// HTTP client con pool de conexiones
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}

// IMPORTANTE: nunca crear http.Client por request — reusar
```

### Caching Strategies

```go
// In-memory cache simple con sync.Map
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

    // Canal para errores del servidor
    serverErr := make(chan error, 1)
    go func() {
        serverErr <- srv.ListenAndServe()
    }()

    // Esperar senal de shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    select {
    case err := <-serverErr:
        log.Fatal("server error:", err)
    case sig := <-quit:
        log.Printf("shutdown signal received: %v", sig)
    }

    // Graceful shutdown con timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("forced shutdown:", err)
    }

    log.Println("server stopped gracefully")
}
```

**Zero-downtime deploys**: el nuevo proceso arranca y pasa health checks antes de que el viejo reciba la senal de shutdown. Kubernetes maneja esto nativamente con rolling deployments.

---

## Patrones de Base de Datos

### Connection Pooling con sql.DB

```go
db, err := sql.Open("postgres", connString)
if err != nil {
    log.Fatal(err)
}

// Configuracion del pool — CRITICO para produccion
db.SetMaxOpenConns(25)              // Conexiones activas maximas
db.SetMaxIdleConns(10)              // Conexiones idle en el pool
db.SetConnMaxLifetime(5 * time.Minute) // Tiempo maximo de vida de una conexion
db.SetConnMaxIdleTime(1 * time.Minute) // Tiempo maximo idle antes de cerrar
```

**Reglas de configuracion:**
| Parametro | Regla | Razon |
|---|---|---|
| `MaxOpenConns` | Menor que el limite de la DB | Evitar saturar la DB |
| `MaxIdleConns` | ~50% de MaxOpenConns | Balance entre reusar y memoria |
| `ConnMaxLifetime` | < timeout del firewall/LB | Evitar conexiones rotas |
| `ConnMaxIdleTime` | Menor que ConnMaxLifetime | Liberar conexiones idle |

### Transacciones

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
    // defer Rollback es seguro — no hace nada si ya se hizo Commit
    defer tx.Rollback()

    // Insertar orden
    _, err = tx.ExecContext(ctx,
        "INSERT INTO orders (id, user_id, total) VALUES ($1, $2, $3)",
        order.ID, order.UserID, order.Total,
    )
    if err != nil {
        return fmt.Errorf("inserting order: %w", err)
    }

    // Insertar items
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

### Niveles de Aislamiento

| Nivel | Dirty Read | Non-repeatable Read | Phantom Read | Uso |
|---|---|---|---|---|
| Read Uncommitted | Si | Si | Si | Casi nunca |
| Read Committed | No | Si | Si | Default PostgreSQL |
| Repeatable Read | No | No | Si* | Reportes |
| Serializable | No | No | No | Transacciones criticas |

*PostgreSQL previene phantom reads en Repeatable Read.

### Migration Strategies

```go
// Ejemplo con golang-migrate
import "github.com/golang-migrate/migrate/v4"

m, err := migrate.New(
    "file://migrations",
    "postgres://localhost:5432/mydb?sslmode=disable",
)
if err != nil {
    log.Fatal(err)
}

// Aplicar todas las migraciones pendientes
if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
    log.Fatal(err)
}
```

### Repository con Interfaces

```go
// Definir interface (en el paquete que la usa)
type UserRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, opts ListOptions) ([]*User, int, error)
}

// Mock para tests
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

## Observabilidad

### Structured Logging con slog

```go
// Configuracion
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

// Uso
slog.Info("order created",
    slog.String("order_id", order.ID.String()),
    slog.String("user_id", order.UserID.String()),
    slog.Float64("total", order.Total),
    slog.Duration("latency", elapsed),
)

// Output JSON:
// {"time":"2025-01-15T10:30:00Z","level":"INFO","msg":"order created",
//  "order_id":"abc-123","user_id":"user-456","total":99.99,"latency":"15ms"}

// Logger con campos fijos (ideal para requests)
requestLogger := logger.With(
    slog.String("request_id", requestID),
    slog.String("method", r.Method),
    slog.String("path", r.URL.Path),
)
```

### Metricas con Prometheus

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total de requests HTTP",
        },
        []string{"method", "path", "status"},
    )

    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "Duracion de requests HTTP",
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

### Distributed Tracing con OpenTelemetry

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("order-service")

func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error) {
    ctx, span := tracer.Start(ctx, "OrderService.CreateOrder")
    defer span.End()

    // El span se propaga automaticamente a traves del context
    user, err := s.users.GetByID(ctx, req.UserID) // crea sub-span
    if err != nil {
        span.RecordError(err)
        return nil, err
    }

    span.SetAttributes(
        attribute.String("user.id", user.ID.String()),
        attribute.Float64("order.total", req.Total),
    )

    // ... resto de la logica
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

## Preguntas de System Design

### Pregunta 1: Disena un URL Shortener

**Requisitos**: dado un URL largo, generar un URL corto. Dado un URL corto, redirigir al original.

**Componentes principales:**

```
┌──────────┐     ┌──────────────┐     ┌───────────┐
│  Client  │────►│  API Server  │────►│ PostgreSQL│
│          │◄────│  (Go)        │◄────│           │
└──────────┘     └──────┬───────┘     └───────────┘
                        │
                   ┌────▼────┐
                   │  Redis  │  (cache de URLs populares)
                   └─────────┘
```

**Notas de implementacion en Go:**

```go
// Generacion de ID corto con base62
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

// Handler de redireccion — alta performance
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
    shortCode := r.PathValue("code")

    // 1. Buscar en cache (Redis)
    if url, err := h.cache.Get(r.Context(), shortCode); err == nil {
        http.Redirect(w, r, url, http.StatusMovedPermanently)
        return
    }

    // 2. Buscar en DB
    url, err := h.repo.GetByShortCode(r.Context(), shortCode)
    if err != nil {
        http.NotFound(w, r)
        return
    }

    // 3. Guardar en cache para futuras requests
    h.cache.Set(r.Context(), shortCode, url.OriginalURL, 24*time.Hour)

    http.Redirect(w, r, url.OriginalURL, http.StatusMovedPermanently)
}
```

**Escalabilidad**: usar Snowflake IDs o counter distribuido para generar IDs unicos. Sharding de la DB por hash del short code. CDN para redirects de URLs muy populares.

---

### Pregunta 2: Disena un Rate Limiter

**Requisitos**: limitar requests por cliente (API key o IP).

**Algoritmo Token Bucket en Go:**

```go
type TokenBucket struct {
    mu         sync.Mutex
    tokens     float64
    maxTokens  float64
    refillRate float64    // tokens por segundo
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

**Para sistema distribuido**: usar Redis con Lua scripts para atomicidad:
```go
// Sliding window counter con Redis
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

### Pregunta 3: Disena un Sistema de Chat

**Requisitos**: chat en tiempo real entre usuarios. Soporte para salas (rooms) y mensajes directos.

**Arquitectura:**

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
                          │ PostgreSQL │  (persistencia de mensajes)
                          └────────────┘
```

**Go es ideal porque**: una goroutine por conexion WebSocket es trivial y eficiente. Con 10K conexiones, solo necesitas ~10K goroutines (~80MB de stacks).

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

**Escalabilidad multi-servidor**: Redis Pub/Sub o NATS para propagar mensajes entre instancias del servidor.

---

### Pregunta 4: Disena una Task Queue

**Requisitos**: sistema de tareas asincronas con reintentos, prioridades, y procesamiento concurrente.

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

**Go es ideal porque**: los workers son goroutines, la queue es un canal con buffer, y la coordinacion se hace con `select` y `context`. Sin frameworks pesados.

---

### Pregunta 5: Disena un Cache con TTL

**Requisitos**: cache en memoria con tiempo de expiracion por entrada. Debe ser thread-safe.

```go
type TTLCache[K comparable, V any] struct {
    mu      sync.RWMutex
    items   map[K]*cacheItem[V]
    onEvict func(K, V) // callback opcional al expirar
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

    // Si ya existe, cancelar el timer anterior
    if existing, ok := c.items[key]; ok {
        existing.timer.Stop()
    }

    item := &cacheItem[V]{
        value:     value,
        expiresAt: time.Now().Add(ttl),
    }

    // Timer para auto-eviccion
    item.timer = time.AfterFunc(ttl, func() {
        c.mu.Lock()
        defer c.mu.Unlock()

        // Verificar que no fue reescrito
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

**Consideraciones de escalabilidad:**
- Para alta concurrencia: shardear el cache en N mapas con locks independientes (reduce la contencion).
- Para evitar memory pressure: limitar el numero de entradas (LRU eviction).
- Para sistemas distribuidos: migrar a Redis/Memcached.

**Go es ideal porque**: `sync.RWMutex` para concurrencia, `time.AfterFunc` para timers eficientes, generics para type safety. Todo con la standard library.
