# 03 - Functions & Closures

## Basic functions

```go
func add(a int, b int) int {
    return a + b
}

// Parameters of the same type can be grouped
func add(a, b int) int {
    return a + b
}
```

- **Lowercase name** (`add`) = private (only visible within the package)
- **Uppercase name** (`Add`) = public (visible from other packages)

> This visibility rule applies to EVERYTHING in Go: functions, types, variables, constants, struct fields.

## Multiple return values

Go allows returning multiple values. It is the basis of error handling:

```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}

// The caller MUST handle both values
result, err := divide(10, 3)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result)
```

> **Fundamental pattern**: the last return position is `error`. If `err != nil`, the rest of the values are invalid.

## Named return values

You can name return values. They are initialized with zero values:

```go
func divide(a, b float64) (result float64, err error) {
    if b == 0 {
        err = fmt.Errorf("division by zero")
        return // "naked return" — returns result=0, err=<the error>
    }
    result = a / b
    return
}
```

> **Tip**: named returns are useful for documenting what a function returns, but **naked returns** (return without values) make the code less readable. Use them only in short functions.

## Variadic functions

Functions that accept a variable number of arguments:

```go
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

// Call
sum(1, 2, 3)       // 6
sum(1, 2, 3, 4, 5) // 15

// Pass a slice with ...
nums := []int{1, 2, 3}
sum(nums...)        // 6
```

- `nums` inside the function is an `[]int` (slice)
- The variadic parameter must be the **last one**
- `fmt.Println` and `append` are variadic

## Functions as first-class citizens

In Go, functions are values. You can assign them to variables, pass them as arguments, and return them:

```go
// Assign to variable
greet := func(name string) string {
    return "Hello, " + name
}
fmt.Println(greet("Jordi"))

// Function type
type MathFunc func(int, int) int

func apply(fn MathFunc, a, b int) int {
    return fn(a, b)
}

result := apply(func(a, b int) int { return a + b }, 3, 4) // 7
```

## Closures

A closure is a function that **captures variables from its outer scope**:

```go
func counter() func() int {
    count := 0
    return func() int {
        count++ // captures 'count' from the outer scope
        return count
    }
}

c := counter()
fmt.Println(c()) // 1
fmt.Println(c()) // 2
fmt.Println(c()) // 3

// Each call to counter() creates an independent closure
c2 := counter()
fmt.Println(c2()) // 1 — its own counter
```

The closure **maintains a reference** to the variable, not a copy. If the variable changes outside, the closure sees the change.

### Classic trap: closure in a loop

```go
// BUG: all goroutines see the same value of i
for i := 0; i < 5; i++ {
    go func() {
        fmt.Println(i) // probably prints "5" five times
    }()
}

// SOLUTION 1: pass as argument
for i := 0; i < 5; i++ {
    go func(n int) {
        fmt.Println(n) // correct: each goroutine has its own copy
    }(i)
}

// SOLUTION 2 (Go 1.22+): the loop variable has per-iteration scope
// In Go 1.22+, the first example already works correctly
```

> **Classic interview question**: "What does this code print?" with a closure in a loop. Knowing this bug and the solutions is fundamental.

## Anonymous functions

Functions without a name, useful for callbacks and inline operations:

```go
// Execute immediately (IIFE)
result := func(a, b int) int {
    return a * b
}(3, 4) // 12

// As callback
numbers := []int{5, 3, 8, 1, 9}
sort.Slice(numbers, func(i, j int) bool {
    return numbers[i] < numbers[j]
})
```

## Defer with functions

`defer` works with functions (anonymous or named):

```go
func process() {
    fmt.Println("start")

    // defer with anonymous function
    defer func() {
        fmt.Println("cleanup")
    }()

    fmt.Println("working...")
    // Output: start, working..., cleanup
}
```

### Defer to measure time (common pattern)

```go
func measureTime(name string) func() {
    start := time.Now()
    return func() {
        fmt.Printf("%s took %v\n", name, time.Since(start))
    }
}

func doWork() {
    defer measureTime("doWork")()
    // ... expensive work ...
}
```

> Note the `()` at the end: `defer measureTime("doWork")()`. First `measureTime` executes (which saves the start time), and then defer executes the returned function on exit.

## Pattern: Functional Options

Very popular pattern in Go for constructors with many options:

```go
type Server struct {
    host    string
    port    int
    timeout time.Duration
}

// Option is a function that modifies the Server
type Option func(*Server)

func WithPort(port int) Option {
    return func(s *Server) {
        s.port = port
    }
}

func WithTimeout(d time.Duration) Option {
    return func(s *Server) {
        s.timeout = d
    }
}

func NewServer(host string, opts ...Option) *Server {
    s := &Server{
        host:    host,
        port:    8080,           // default
        timeout: 30 * time.Second, // default
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Clean and extensible usage:
s := NewServer("localhost",
    WithPort(9090),
    WithTimeout(5 * time.Second),
)
```

> This pattern appears in many popular Go libraries (gRPC, zap logger, etc). It is **very likely** you will be asked about it in an interview.

## Pattern: Middleware / Function wrapping

Functions that wrap other functions to add behavior:

```go
type Handler func(string) string

func withLogging(h Handler) Handler {
    return func(input string) string {
        fmt.Printf("Input: %s\n", input)
        result := h(input)
        fmt.Printf("Output: %s\n", result)
        return result
    }
}

func toUpper(s string) string {
    return strings.ToUpper(s)
}

// Compose
handler := withLogging(toUpper)
handler("hello") // Input: hello, Output: HELLO
```

## init()

Each package can have `init()` functions that execute **automatically** when the package is imported:

```go
var config map[string]string

func init() {
    config = make(map[string]string)
    config["env"] = "development"
}
```

- They execute before `main()`
- There can be multiple `init()` per file
- They execute in declaration order
- **Avoid using them** for complex logic — they make code hard to test

## Common interview questions

1. **Does Go support function overloading?**
   No. Each function name must be unique within the package. Use descriptive names or variadic params.

2. **What is a closure and how does it handle memory?**
   A closure is a function that captures variables from its outer scope. Captured variables are moved to the heap (escape analysis) and live as long as the closure exists.

3. **Difference between value receiver and pointer receiver in methods?**
   Value receiver works with a copy (does not modify the original). Pointer receiver works with a reference (modifies the original and avoids copies). We will see this in detail in module 04.

4. **What is the functional options pattern?**
   A pattern that uses closures and variadic params to create flexible and extensible constructors. Each option is a function that modifies the configuration.

5. **When would you use init() and when not?**
   Use: registering drivers (database/sql), setup of complex constants. Do not use: business logic, I/O, external dependencies. It makes code hard to test and execution order unpredictable.

6. **What happens with closures in a for loop?**
   All closures share the same loop variable. In Go <1.22, when they execute, they all see the last value. Solution: pass the variable as an argument or use Go 1.22+ where each iteration has its own variable.
