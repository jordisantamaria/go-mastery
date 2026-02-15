# Proyecto 2: Finance Tracker API

REST API para tracking de gastos e ingresos. Demuestra clean architecture y desarrollo backend profesional.

## Stack
- [Chi](https://github.com/go-chi/chi) — router ligero
- PostgreSQL + [pgx](https://github.com/jackc/pgx) — driver nativo
- JWT para autenticacion
- Docker para desarrollo local

## Features
- CRUD de transacciones (ingresos/gastos)
- Autenticacion con JWT
- Filtros por fecha, categoria, tipo
- Paginacion
- Middleware: logging, auth, recovery, CORS

## Arquitectura
```
cmd/api/          — entrypoint
internal/
  handler/        — HTTP handlers
  service/        — logica de negocio
  repository/     — acceso a datos
  model/          — domain types
  middleware/      — middleware HTTP
```

## Lo que aprenderas
- Clean architecture en Go
- Dependency injection sin frameworks
- Middleware chain pattern
- Database migrations
- Integration testing

> En progreso — semanas 5-6 del roadmap.
