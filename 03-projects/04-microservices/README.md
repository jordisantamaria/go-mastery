# Proyecto 4: Microservices Platform

Plataforma de microservicios con gRPC, Docker y observability. El proyecto mas avanzado del portfolio.

## Servicios
- **User Service** — gestion de usuarios (gRPC)
- **Order Service** — gestion de pedidos (gRPC)
- **API Gateway** — REST gateway para clientes externos
- **Notification Service** — notificaciones async via message queue

## Stack
- gRPC + Protocol Buffers
- Docker + Docker Compose
- PostgreSQL (por servicio)
- Redis (cache + pub/sub)
- Prometheus + Grafana (metricas)
- OpenTelemetry (tracing distribuido)

## Lo que aprenderas
- Comunicacion entre servicios con gRPC
- Protocol Buffers y code generation
- Service discovery
- Distributed tracing
- Health checks y graceful shutdown
- Docker multi-stage builds

> En progreso — semanas 9-10 del roadmap.
