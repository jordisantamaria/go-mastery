# Interview Prep

Material de preparacion para entrevistas tecnicas de Go.

## Secciones

### [Language Internals](language-internals/)
Preguntas sobre el runtime de Go: garbage collector, scheduler GMP model, memory model, escape analysis, stack vs heap allocation.

### [Concurrency Puzzles](concurrency-puzzles/)
Ejercicios practicos de concurrencia: detectar race conditions, resolver deadlocks, implementar patrones clasicos.

### [System Design](system-design/)
Como aplicar Go en problemas de system design: cuando elegir Go, patrones de microservicios, escalabilidad.

### [Coding Challenges](coding-challenges/)
Problemas de algoritmos y estructuras de datos resueltos de forma idiomatica en Go (no "Java traducido a Go").

## Como Usar

Recomendacion de estudio:

1. **Language Internals**: leer la teoria y practicar explicando cada concepto en voz alta (como si estuvieras en una entrevista). Intentar responder las 17 preguntas sin mirar las respuestas.
2. **Concurrency Puzzles**: para cada puzzle, primero analizar el codigo roto y encontrar el bug antes de leer la explicacion.
3. **System Design**: estudiar los patrones y luego practicar las 5 preguntas de diseno en una pizarra o papel.
4. **Coding Challenges**: implementar las soluciones en `challenges.go` y ejecutar `go test -v` para verificar.
