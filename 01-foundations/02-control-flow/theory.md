# 02 - Control Flow

## if / else

In Go, **there are no parentheses** around the condition. Braces are **mandatory**:

```go
x := 10
if x > 5 {
    fmt.Println("greater than 5")
} else if x > 0 {
    fmt.Println("positive")
} else {
    fmt.Println("zero or negative")
}
```

### if with initial statement (idiomatic)

You can declare a variable inside the `if` — its scope is limited to the block:

```go
if err := doSomething(); err != nil {
    fmt.Println("error:", err)
    return
}
// err does not exist out here
```

> This pattern is **extremely common** in Go. You will see it in every function that handles errors.

## switch

Much more powerful than in other languages. **No `break` needed** (no fall-through by default):

```go
day := "Monday"
switch day {
case "Monday", "Tuesday", "Wednesday", "Thursday", "Friday":
    fmt.Println("Weekday")
case "Saturday", "Sunday":
    fmt.Println("Weekend")
default:
    fmt.Println("Invalid day")
}
```

### switch without expression (replaces long if/else)

```go
score := 85
switch {
case score >= 90:
    fmt.Println("A")
case score >= 80:
    fmt.Println("B")
case score >= 70:
    fmt.Println("C")
default:
    fmt.Println("F")
}
```

### Type switch (important for interfaces)

```go
func describe(i interface{}) string {
    switch v := i.(type) {
    case int:
        return fmt.Sprintf("integer: %d", v)
    case string:
        return fmt.Sprintf("string: %s", v)
    case bool:
        return fmt.Sprintf("bool: %t", v)
    default:
        return fmt.Sprintf("unknown type: %T", v)
    }
}
```

### fallthrough (rare, but exists)

```go
switch 1 {
case 1:
    fmt.Println("one")
    fallthrough // forces execution of the next case
case 2:
    fmt.Println("two") // executes even though case != 2
}
```

## for (the only loop in Go)

Go only has `for`. There is no `while` or `do-while`:

```go
// classic for
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// "while" loop
count := 0
for count < 5 {
    count++
}

// infinite loop
for {
    // use break to exit
    break
}
```

### range — iterating over collections

```go
// Slice
nums := []int{10, 20, 30}
for i, v := range nums {
    fmt.Printf("index %d: value %d\n", i, v)
}

// Values only (discard index)
for _, v := range nums {
    fmt.Println(v)
}

// Indices only
for i := range nums {
    fmt.Println(i)
}

// Map
m := map[string]int{"a": 1, "b": 2}
for key, value := range m {
    fmt.Printf("%s: %d\n", key, value)
}

// String (iterates by runes, not bytes)
for i, r := range "Hola 🌍" {
    fmt.Printf("byte %d: rune %c\n", i, r)
}
```

### break and continue

```go
for i := 0; i < 100; i++ {
    if i%2 == 0 {
        continue // skip to the next iteration
    }
    if i > 10 {
        break // exit the loop
    }
    fmt.Println(i) // 1, 3, 5, 7, 9
}
```

### Labels (break/continue in nested loops)

```go
outer:
for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
        if i == 1 && j == 1 {
            break outer // exits BOTH loops
        }
        fmt.Printf("(%d, %d) ", i, j)
    }
}
```

## defer

`defer` schedules a function to execute **when the current function exits** (after the return). They execute in **LIFO** (last in, first out) order:

```go
func readFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close() // executes when readFile exits, no matter what

    // work with the file...
    return nil
}
```

### defer LIFO

```go
func main() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
    // Output: 3, 2, 1
}
```

### defer evaluates arguments immediately

```go
x := 10
defer fmt.Println(x) // prints 10, not 20
x = 20
```

> Arguments are evaluated when `defer` is declared, not when it executes.

## panic and recover

`panic` stops normal execution. `recover` catches it (only inside `defer`):

```go
func safeDiv(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic recovered: %v", r)
        }
    }()

    return a / b, nil // if b == 0, Go panics
}
```

> **Rule**: do not use `panic` for normal errors. Only for truly unrecoverable situations (bug in the program, corrupted state). Use `error` for everything else.

## Common interview questions

1. **Does Go have a while loop?**
   No. `for` covers all cases: `for condition {}` is the equivalent of while.

2. **What happens if you use defer inside a loop?**
   Defers accumulate and all execute when the function exits (not when the loop exits). This can cause memory leaks if the loop is long. Solution: extract the loop body into a separate function.

3. **In what order do defers execute?**
   LIFO — the last defer declared is the first to execute.

4. **When to use panic vs error?**
   `error` for normal flow (file not found, invalid input, etc). `panic` only for unrecoverable bugs (index out of range, nil pointer in an impossible place).

5. **Why doesn't Go have fall-through by default in switch?**
   Because implicit fall-through (like in C/Java) is a source of bugs. If you need it, you use `fallthrough` explicitly.
