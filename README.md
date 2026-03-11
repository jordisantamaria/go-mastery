# Go Mastery

Complete Go learning path: from fundamentals to production microservices.

Each module combines **theory**, **executable code**, and **exercises with tests** to verify the solutions.

## Roadmap

```
Week 1-2     Fundamentals + Project 1 (CLI)
Week 3-4     Concurrency + Testing + Interview Prep
Week 5-6     Project 2 (REST API) + Standard Library
Week 7-8     Project 3 (Pipeline) + Concurrency Puzzles
Week 9-10    Project 4 (Microservices) + System Design
```

## Structure

### 01 - Foundations

Theory + short exercises with tests. Each module has:
- `theory.md` — Concise explanation with examples
- `examples/` — Commented executable code (`go run`)
- `exercises/` — Challenges with tests (`go test`)

| # | Module | Key Topics |
|---|--------|------------|
| 01 | [Syntax & Types](01-foundations/01-syntax-types/) | Variables, types, zero values, slices, maps, pointers |
| 02 | [Control Flow](01-foundations/02-control-flow/) | if/else, switch, for, range, defer, labels |
| 03 | [Functions & Closures](01-foundations/03-functions-closures/) | Multiple returns, variadic, closures, first-class functions |
| 04 | [Structs & Interfaces](01-foundations/04-structs-interfaces/) | Composition, embedding, implicit interfaces, polymorphism |
| 05 | [Error Handling](01-foundations/05-error-handling/) | error interface, sentinel errors, wrapping, errors.Is/As |
| 06 | [Concurrency](01-foundations/06-concurrency/) | Goroutines, channels, select, sync, context, errgroup |
| 07 | [Generics](01-foundations/07-generics/) | Type parameters, constraints, when to use them (and when not to) |
| 08 | [Testing](01-foundations/08-testing/) | Table-driven tests, benchmarks, fuzzing, mocks with interfaces |
| 09 | [Stdlib Deep Dive](01-foundations/09-stdlib-deep-dive/) | net/http, io, context, encoding/json, os, flag |

### 02 - Interview Prep

| Section | Content |
|---------|---------|
| [Language Internals](02-interview-prep/language-internals/) | GC, scheduler GMP, memory model, escape analysis, stack vs heap |
| [Concurrency Puzzles](02-interview-prep/concurrency-puzzles/) | Race conditions, deadlocks, classic patterns |
| [System Design](02-interview-prep/system-design/) | Go design patterns for system design interviews |
| [Coding Challenges](02-interview-prep/coding-challenges/) | Classic problems solved idiomatically in Go |

### 03 - Projects (Portfolio)

| # | Project | Demonstrates | Stack |
|---|---------|-------------|-------|
| 01 | [CLI Task Manager](03-projects/01-cli-tool/) | Idiomatic Go, testing, packaging | Cobra, BoltDB |
| 02 | [Finance Tracker API](03-projects/02-rest-api/) | Clean architecture, middleware, auth | Chi, PostgreSQL, JWT |
| 03 | [Data Pipeline](03-projects/03-concurrent-pipeline/) | Goroutines, channels, fan-out/fan-in | Workers, rate limiting |
| 04 | [Microservices Platform](03-projects/04-microservices/) | Distributed systems, observability | gRPC, Docker, Prometheus |

## How to use this repo

```bash
# Run an example
go run 01-foundations/01-syntax-types/examples/variables.go

# Run the tests for an exercise
go test ./01-foundations/01-syntax-types/exercises/...

# Run all tests
go test ./...
```

## Requirements

- Go 1.21+ (`go version`)
- Editor with Go support (VS Code + Go extension recommended)
