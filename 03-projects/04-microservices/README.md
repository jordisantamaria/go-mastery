# Proyecto 4: Microservices Platform

Plataforma de microservicios que demuestra arquitectura distribuida en Go usando solo la biblioteca estandar. Usa `net/rpc` en lugar de gRPC para mantener el proyecto sin dependencias externas, demostrando los mismos patrones: comunicacion entre servicios, API gateway, health checks y apagado graceful.

## Arquitectura

```
┌─────────────────────────────────────────────────────────────────┐
│                        Cliente (curl/browser)                    │
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
│                                    RPC  │ (valida usuario)       │
│                                         │                        │
│              ┌──────────────────────────▼┐                       │
│              │       UserService         │                       │
│              └───────────────────────────┘                       │
└─────────────────────────────────────────────────────────────────┘
```

**Flujo de datos:**
1. El cliente envia peticiones REST al API Gateway (:8080)
2. El Gateway traduce REST a llamadas RPC hacia los servicios backend
3. OrderService valida la existencia del usuario llamando a UserService via RPC
4. Cada servicio tiene almacenamiento en memoria con sync.RWMutex

## Estructura del proyecto

```
04-microservices/
├── go.mod                      # Modulo Go (sin dependencias externas)
├── Dockerfile                  # Multi-stage build para los 3 servicios
├── docker-compose.yml          # Orquestacion de contenedores
├── cmd/
│   ├── userservice/main.go     # Entrypoint del servicio de usuarios
│   ├── orderservice/main.go    # Entrypoint del servicio de pedidos
│   └── gateway/main.go         # Entrypoint del API Gateway
├── internal/
│   ├── userservice/            # Logica del servicio de usuarios
│   ├── orderservice/           # Logica del servicio de pedidos
│   └── gateway/                # Logica del gateway REST → RPC
├── pkg/
│   ├── model/                  # Tipos compartidos (User, Order)
│   ├── health/                 # Health checks
│   ├── discovery/              # Registro de servicios
│   └── middleware/             # Middlewares HTTP (logging, recovery)
└── scripts/
    └── test.sh                 # Script para probar todos los endpoints
```

## Como ejecutar

### Opcion 1: Localmente (3 terminales)

**Terminal 1 - UserService:**
```bash
go run ./cmd/userservice
# Salida: {"level":"INFO","msg":"UserService iniciado","addr":":50051"}
```

**Terminal 2 - OrderService:**
```bash
go run ./cmd/orderservice
# Salida: {"level":"INFO","msg":"OrderService iniciado","addr":":50052","user_service_addr":"localhost:50051"}
```

**Terminal 3 - API Gateway:**
```bash
go run ./cmd/gateway
# Salida: {"level":"INFO","msg":"API Gateway iniciado","addr":":8080",...}
```

### Opcion 2: Docker Compose

```bash
docker compose up --build
```

Para detener:
```bash
docker compose down
```

### Variables de entorno

| Variable             | Defecto          | Descripcion                       |
|----------------------|------------------|-----------------------------------|
| `USER_SERVICE_ADDR`  | `:50051`         | Direccion del UserService         |
| `ORDER_SERVICE_ADDR` | `:50052`         | Direccion del OrderService        |
| `GATEWAY_ADDR`       | `:8080`          | Direccion del API Gateway         |

## API REST

### Usuarios

```bash
# Crear usuario
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Ana Garcia","email":"ana@example.com"}'

# Listar usuarios
curl http://localhost:8080/api/users

# Obtener usuario por ID
curl http://localhost:8080/api/users/{id}

# Actualizar usuario
curl -X PUT http://localhost:8080/api/users/{id} \
  -H "Content-Type: application/json" \
  -d '{"name":"Ana Garcia Martinez","email":"ana.garcia@example.com"}'

# Eliminar usuario
curl -X DELETE http://localhost:8080/api/users/{id}
```

### Pedidos

```bash
# Crear pedido (valida que el usuario exista)
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "{user_id}",
    "items": [
      {"product_id":"p1","name":"Teclado","quantity":1,"price":49.99},
      {"product_id":"p2","name":"Raton","quantity":2,"price":19.99}
    ]
  }'

# Obtener pedido por ID
curl http://localhost:8080/api/orders/{id}

# Listar pedidos de un usuario
curl http://localhost:8080/api/users/{user_id}/orders

# Actualizar estado del pedido
curl -X PUT http://localhost:8080/api/orders/{id}/status \
  -H "Content-Type: application/json" \
  -d '{"status":"confirmed"}'
```

**Transiciones de estado validas:**
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

## Script de pruebas

```bash
chmod +x scripts/test.sh
./scripts/test.sh
# O con un host personalizado:
./scripts/test.sh http://mi-host:9090
```

## Tests unitarios

```bash
go test ./...
```

Los tests cubren:
- **UserService:** CRUD completo, validaciones, emails duplicados
- **OrderService:** Creacion con validacion de usuario (mock), transiciones de estado
- **Gateway:** Tests de integracion que levantan servicios RPC reales en puertos aleatorios

## Patrones demostrados

| Patron                         | Implementacion                                            |
|--------------------------------|-----------------------------------------------------------|
| **API Gateway**                | REST HTTP que traduce a llamadas RPC                      |
| **Comunicacion inter-servicio**| OrderService valida usuarios via RPC al UserService       |
| **Service Discovery**          | Registro en memoria (pkg/discovery)                       |
| **Health Checks**              | Verificacion de conectividad a servicios backend          |
| **Graceful Shutdown**          | Captura SIGINT/SIGTERM, cierra listeners ordenadamente    |
| **Middleware chain**           | Logging y recovery aplicados al gateway                   |
| **Dependency Injection**       | UserValidator como interfaz inyectable en OrderService    |
| **Maquina de estados**         | Transiciones de estado validadas en pedidos               |
| **Multi-stage Docker build**   | Imagen final de ~15MB con binarios estaticos              |
| **Logging estructurado**       | slog con formato JSON para observabilidad                 |
| **Configuracion por entorno**  | Variables de entorno con defaults sensatos                |

## Equivalencia con gRPC en produccion

Este proyecto usa `net/rpc` de la stdlib para evitar dependencias externas. En un entorno de produccion se usaria gRPC. Esta tabla muestra la equivalencia:

| Concepto              | Este proyecto (net/rpc)          | Produccion (gRPC)                          |
|-----------------------|----------------------------------|--------------------------------------------|
| **Definicion de API** | Structs Go como argumentos       | Protocol Buffers (.proto)                  |
| **Serializacion**     | encoding/gob (binario Go)        | Protocol Buffers (multi-lenguaje)          |
| **Transporte**        | TCP directo                      | HTTP/2 con multiplexing                    |
| **Streaming**         | No soportado                     | Streaming unidireccional y bidireccional   |
| **Code generation**   | No necesario                     | protoc genera stubs cliente/servidor       |
| **Interceptors**      | No tiene (se implementa manual)  | Interceptors unarios y de streaming        |
| **Load balancing**    | Manual                           | gRPC tiene LB integrado                    |
| **Metadata**          | No tiene                         | Headers/trailers para auth, tracing        |
| **Deadlines**         | Manual con context               | Integrado en el protocolo                  |
| **Multi-lenguaje**    | Solo Go                          | Go, Java, Python, C++, etc.               |

### Cuando elegir cada uno

**net/rpc es suficiente cuando:**
- Todos los servicios estan en Go
- No necesitas streaming
- Proyecto personal, prototipos o aprendizaje
- Quieres cero dependencias externas

**gRPC es necesario cuando:**
- Servicios en multiples lenguajes
- Necesitas streaming bidireccional
- Requieres un contrato de API formal (.proto)
- Necesitas interceptors avanzados (auth, tracing, metricas)
- Entorno de produccion con multiples equipos

## Lo que aprenderas con este proyecto

- Descomponer una aplicacion en microservicios independientes
- Comunicacion sincrona entre servicios via RPC
- Traducir una API REST publica a llamadas RPC internas
- Implementar health checks para monitoreo
- Apagado graceful en servicios concurrentes
- Docker multi-stage builds para imagenes minimas
- Inyeccion de dependencias para testing (mock del UserValidator)
- Configuracion por variables de entorno (12-Factor App)
