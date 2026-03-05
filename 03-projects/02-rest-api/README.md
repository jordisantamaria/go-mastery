# Proyecto 2: Finance Tracker REST API

API REST para tracking de finanzas personales (ingresos y gastos). Demuestra clean architecture, desarrollo backend profesional y dominio profundo de la stdlib de Go.

**Sin dependencias externas** — solo stdlib de Go (`net/http`, `crypto`, `encoding/json`, etc.).

## Arquitectura

```
02-rest-api/
├── cmd/
│   └── api/
│       └── main.go              # Entrypoint: wiring de dependencias, servidor
├── internal/
│   ├── model/
│   │   ├── transaction.go       # Tipo de dominio Transaction
│   │   └── user.go              # Tipo de dominio User
│   ├── handler/
│   │   ├── transaction.go       # Handlers HTTP para transacciones (CRUD)
│   │   ├── auth.go              # Handlers de registro y login
│   │   └── health.go            # Health check
│   ├── service/
│   │   ├── transaction.go       # Lógica de negocio de transacciones
│   │   └── auth.go              # Lógica de autenticación (JWT + hashing)
│   ├── repository/
│   │   ├── interfaces.go        # Interfaces del repositorio
│   │   ├── memory.go            # Implementación in-memory (maps + RWMutex)
│   │   └── memory_test.go       # Tests del repositorio
│   └── middleware/
│       ├── auth.go              # Middleware JWT (extrae user del token)
│       ├── logging.go           # Log estructurado de cada petición
│       ├── recovery.go          # Recuperación de panics
│       └── cors.go              # Cabeceras CORS
├── pkg/
│   └── jwt/
│       ├── jwt.go               # Implementación JWT HS256 (sin librerías)
│       └── jwt_test.go          # Tests del JWT
├── go.mod
└── README.md
```

### Flujo de dependencias

```
Handler → Service → Repository (interfaz)
                         ↑
                    MemoryStore (implementación)
```

Los handlers nunca acceden directamente al repositorio. El servicio contiene toda la lógica de negocio y validación. El repositorio es una interfaz que se puede intercambiar (memoria, PostgreSQL, etc.).

## API Endpoints

| Método | Ruta                      | Auth | Descripción                          |
|--------|---------------------------|------|--------------------------------------|
| POST   | `/api/auth/register`      | No   | Registrar nuevo usuario              |
| POST   | `/api/auth/login`         | No   | Login, devuelve JWT                  |
| GET    | `/api/transactions`       | Sí   | Listar transacciones (con filtros)   |
| POST   | `/api/transactions`       | Sí   | Crear transacción                    |
| GET    | `/api/transactions/{id}`  | Sí   | Obtener una transacción              |
| PUT    | `/api/transactions/{id}`  | Sí   | Actualizar transacción               |
| DELETE | `/api/transactions/{id}`  | Sí   | Eliminar transacción                 |
| GET    | `/api/health`             | No   | Health check                         |

### Filtros de listado

```
GET /api/transactions?type=income&category=food&from=2024-01-01&to=2024-12-31&page=1&limit=10
```

| Parámetro  | Tipo   | Descripción                        |
|------------|--------|------------------------------------|
| `type`     | string | `income` o `expense`               |
| `category` | string | Categoría (food, transport, etc.)  |
| `from`     | string | Fecha inicio (YYYY-MM-DD)          |
| `to`       | string | Fecha fin (YYYY-MM-DD)             |
| `page`     | int    | Página (default: 1)                |
| `limit`    | int    | Resultados por página (default: 10)|

## Cómo ejecutar

```bash
# Desde la raíz del proyecto
cd 03-projects/02-rest-api

# Compilar
go build ./...

# Ejecutar (puerto 8080 por defecto)
go run ./cmd/api/

# Con configuración personalizada
PORT=3000 JWT_SECRET=mi-clave-secreta go run ./cmd/api/
```

## Ejemplos con curl

### Registrar usuario

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "maria@example.com",
    "password": "secreto123",
    "name": "María García"
  }'
```

Respuesta (201):
```json
{
  "id": "a1b2c3d4...",
  "email": "maria@example.com",
  "name": "María García"
}
```

### Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "maria@example.com",
    "password": "secreto123"
  }'
```

Respuesta (200):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "a1b2c3d4...",
    "email": "maria@example.com",
    "name": "María García"
  }
}
```

### Crear transacción

```bash
curl -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <tu-token-jwt>" \
  -d '{
    "type": "expense",
    "amount": 45.50,
    "category": "food",
    "description": "Cena en restaurante",
    "date": "2024-06-15"
  }'
```

### Listar transacciones con filtros

```bash
# Todas las transacciones
curl http://localhost:8080/api/transactions \
  -H "Authorization: Bearer <tu-token-jwt>"

# Solo gastos de comida del último mes
curl "http://localhost:8080/api/transactions?type=expense&category=food&from=2024-06-01&to=2024-06-30" \
  -H "Authorization: Bearer <tu-token-jwt>"
```

Respuesta (200):
```json
{
  "data": [
    {
      "id": "tx-abc123",
      "user_id": "a1b2c3d4...",
      "type": "expense",
      "amount": 45.50,
      "category": "food",
      "description": "Cena en restaurante",
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

### Actualizar transacción

```bash
curl -X PUT http://localhost:8080/api/transactions/<id> \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <tu-token-jwt>" \
  -d '{
    "amount": 50.00,
    "category": "restaurant"
  }'
```

### Eliminar transacción

```bash
curl -X DELETE http://localhost:8080/api/transactions/<id> \
  -H "Authorization: Bearer <tu-token-jwt>"
```

### Health check

```bash
curl http://localhost:8080/api/health
```

Respuesta (200):
```json
{
  "status": "ok",
  "uptime": "2h30m15s",
  "version": "1.0.0"
}
```

## Cómo ejecutar los tests

```bash
# Todos los tests
go test ./...

# Con verbose
go test -v ./...

# Solo tests del JWT
go test -v ./pkg/jwt/

# Solo tests del repositorio
go test -v ./internal/repository/

# Con cobertura
go test -cover ./...
```

## Patrones que demuestra

### Clean Architecture
- **Separación de capas**: handler (HTTP) → service (negocio) → repository (datos)
- **Inversión de dependencias**: los servicios dependen de interfaces, no de implementaciones
- **Domain models**: los tipos del dominio no dependen de la capa HTTP ni de la base de datos

### Stdlib Mastery
- **Go 1.22+ ServeMux**: routing basado en método (`"GET /api/transactions/{id}"`)
- **JWT desde cero**: implementación HS256 con `crypto/hmac` + `crypto/sha256`
- **Password hashing**: SHA-256 con salt aleatorio usando `crypto/rand`
- **Graceful shutdown**: `signal.Notify` + `server.Shutdown` con timeout

### Patrones de diseño
- **Dependency Injection**: constructores que reciben interfaces (`NewTransactionService(repo)`)
- **Middleware Chain**: Recovery → CORS → Logging → Auth (composición funcional)
- **Repository Pattern**: interfaz + implementación intercambiable
- **Constructor Pattern**: `NewXxxService`, `NewXxxHandler`, `NewXxxRepository`

### Concurrencia
- **sync.RWMutex**: el store en memoria usa read-write locks para acceso concurrente seguro
- **context.Context**: propagación de contexto en toda la cadena de llamadas

## Cómo extender con base de datos real

El proyecto está diseñado para intercambiar el store fácilmente gracias a las interfaces:

1. **Crear implementación PostgreSQL**:
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
   // ... implementar el resto de métodos de la interfaz
   ```

2. **Cambiar el wiring en main.go**:
   ```go
   // Antes (in-memory):
   txRepo := repository.NewMemoryTransactionRepository()

   // Después (PostgreSQL):
   db, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))
   txRepo := repository.NewPostgresTransactionRepository(db)
   ```

3. **El resto del código no cambia** — los servicios y handlers siguen funcionando exactamente igual porque dependen de la interfaz, no de la implementación.

## Variables de entorno

| Variable     | Default                                      | Descripción              |
|-------------|----------------------------------------------|--------------------------|
| `PORT`      | `8080`                                       | Puerto del servidor      |
| `JWT_SECRET`| `super-secret-key-cambiar-en-produccion`     | Clave secreta para JWT   |
