# Interview Prep

Preparation material for Go technical interviews.

## Sections

### [Language Internals](language-internals/)
Questions about the Go runtime: garbage collector, scheduler GMP model, memory model, escape analysis, stack vs heap allocation.

### [Concurrency Puzzles](concurrency-puzzles/)
Practical concurrency exercises: detecting race conditions, solving deadlocks, implementing classic patterns.

### [System Design](system-design/)
How to apply Go in system design problems: when to choose Go, microservices patterns, scalability.

### [Coding Challenges](coding-challenges/)
Algorithm and data structure problems solved idiomatically in Go (not "Java translated to Go").

## How to Use

Recommended study approach:

1. **Language Internals**: read the theory and practice explaining each concept out loud (as if you were in an interview). Try to answer the 17 questions without looking at the answers.
2. **Concurrency Puzzles**: for each puzzle, first analyze the broken code and find the bug before reading the explanation.
3. **System Design**: study the patterns and then practice the 5 design questions on a whiteboard or paper.
4. **Coding Challenges**: implement the solutions in `challenges.go` and run `go test -v` to verify.
