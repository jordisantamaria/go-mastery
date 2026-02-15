# Go Mastery

Ruta de aprendizaje completa de Go: desde los fundamentos hasta microservicios en produccion.

Cada modulo combina **teoria**, **codigo ejecutable** y **ejercicios con tests** para verificar las soluciones.

## Roadmap

```
Semana 1-2    Fundamentos + Proyecto 1 (CLI)
Semana 3-4    Concurrencia + Testing + Interview Prep
Semana 5-6    Proyecto 2 (REST API) + Standard Library
Semana 7-8    Proyecto 3 (Pipeline) + Puzzles de concurrencia
Semana 9-10   Proyecto 4 (Microservicios) + System Design
```

## Estructura

### 01 - Foundations

Teoria + ejercicios cortos con tests. Cada modulo tiene:
- `theory.md` — Explicacion concisa con ejemplos
- `examples/` — Codigo comentado ejecutable (`go run`)
- `exercises/` — Retos con tests (`go test`)

| # | Modulo | Temas clave |
|---|--------|-------------|
| 01 | [Syntax & Types](01-foundations/01-syntax-types/) | Variables, tipos, zero values, slices, maps, punteros |
| 02 | [Control Flow](01-foundations/02-control-flow/) | if/else, switch, for, range, defer, labels |
| 03 | [Functions & Closures](01-foundations/03-functions-closures/) | Multiple returns, variadic, closures, first-class functions |
| 04 | [Structs & Interfaces](01-foundations/04-structs-interfaces/) | Composicion, embedding, interfaces implicitas, polimorfismo |
| 05 | [Error Handling](01-foundations/05-error-handling/) | error interface, sentinel errors, wrapping, errors.Is/As |
| 06 | [Concurrency](01-foundations/06-concurrency/) | Goroutines, channels, select, sync, context, errgroup |
| 07 | [Generics](01-foundations/07-generics/) | Type parameters, constraints, cuando usarlos (y cuando no) |
| 08 | [Testing](01-foundations/08-testing/) | Table-driven tests, benchmarks, fuzzing, mocks con interfaces |
| 09 | [Stdlib Deep Dive](01-foundations/09-stdlib-deep-dive/) | net/http, io, context, encoding/json, os, flag |

### 02 - Interview Prep

| Seccion | Contenido |
|---------|-----------|
| [Language Internals](02-interview-prep/language-internals/) | GC, scheduler GMP, memory model, escape analysis, stack vs heap |
| [Concurrency Puzzles](02-interview-prep/concurrency-puzzles/) | Race conditions, deadlocks, patrones clasicos |
| [System Design](02-interview-prep/system-design/) | Patrones de diseno en Go para system design interviews |
| [Coding Challenges](02-interview-prep/coding-challenges/) | Problemas clasicos resueltos de forma idiomatica en Go |

### 03 - Projects (Portfolio)

| # | Proyecto | Demuestra | Stack |
|---|----------|-----------|-------|
| 01 | [CLI Task Manager](03-projects/01-cli-tool/) | Go idiomatico, testing, packaging | Cobra, BoltDB |
| 02 | [Finance Tracker API](03-projects/02-rest-api/) | Clean architecture, middleware, auth | Chi, PostgreSQL, JWT |
| 03 | [Data Pipeline](03-projects/03-concurrent-pipeline/) | Goroutines, channels, fan-out/fan-in | Workers, rate limiting |
| 04 | [Microservices Platform](03-projects/04-microservices/) | Distributed systems, observability | gRPC, Docker, Prometheus |

## Como usar este repo

```bash
# Ejecutar un ejemplo
go run 01-foundations/01-syntax-types/examples/variables.go

# Ejecutar los tests de un ejercicio
go test ./01-foundations/01-syntax-types/exercises/...

# Ejecutar todos los tests
go test ./...
```

## Requisitos

- Go 1.21+ (`go version`)
- Editor con soporte Go (VS Code + Go extension recomendado)
