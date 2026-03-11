# 05 - Error Handling

In Go, **errors are values**, not exceptions. There is no `try/catch`. This is a deliberate design decision that forces the programmer to handle errors explicitly.

## The error interface

```go
type error interface {
    Error() string
}
```

Any type that has an `Error() string` method is an error. That simple.

## Creating errors

### errors.New — simple errors

```go
import "errors"

func validate(age int) error {
    if age < 0 {
        return errors.New("age cannot be negative")
    }
    return nil  // nil = no error
}
```

### fmt.Errorf — formatted errors

```go
func validate(name string) error {
    if len(name) < 2 {
        return fmt.Errorf("name %q is too short (min 2 chars)", name)
    }
    return nil
}
```

## The basic pattern: if err != nil

```go
result, err := doSomething()
if err != nil {
    return err  // propagate the error
}
// use result...
```

> You will see `if err != nil` **hundreds of times** in any Go project. It is the idiomatic way to handle errors. Do not try to avoid it — embrace it.

### Return early (the pattern)

```go
// BAD — unnecessary nesting
func process(id int) error {
    user, err := getUser(id)
    if err == nil {
        orders, err := getOrders(user.ID)
        if err == nil {
            // process...
            return nil
        }
        return err
    }
    return err
}

// GOOD — return early
func process(id int) error {
    user, err := getUser(id)
    if err != nil {
        return err
    }

    orders, err := getOrders(user.ID)
    if err != nil {
        return err
    }

    // process...
    return nil
}
```

## Error wrapping

Since Go 1.13, you can **wrap** errors to add context without losing the original error:

```go
func getUser(id int) (*User, error) {
    row := db.QueryRow("SELECT ...")
    if err := row.Scan(&user); err != nil {
        return nil, fmt.Errorf("getUser(%d): %w", id, err)
        //                                    ^^ %w wraps the error
    }
    return &user, nil
}
```

`%w` (wrap) is different from `%v`:
- **`%w`**: wraps the error — it can be unwrapped with `errors.Is`/`errors.As`
- **`%v`**: only formats the message — the original error is lost

## Sentinel errors

Predefined errors that are compared by identity:

```go
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrInvalidInput = errors.New("invalid input")
)

func getUser(id int) (*User, error) {
    // ...
    if user == nil {
        return nil, ErrNotFound
    }
    return user, nil
}

// Check
err := getUser(42)
if errors.Is(err, ErrNotFound) {
    // handle "not found"
}
```

> Convention: sentinel errors start with `Err` (e.g., `ErrNotFound`, `io.EOF`).

## errors.Is and errors.As

### errors.Is — compare with sentinel error (through wrapping)

```go
// Works even if the error is wrapped
err := fmt.Errorf("database query: %w", ErrNotFound)

errors.Is(err, ErrNotFound)  // true! — unwraps automatically
err == ErrNotFound           // false — not the same object
```

### errors.As — extract a specific error type

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s - %s", e.Field, e.Message)
}

// Check and extract
var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Println("Invalid field:", valErr.Field)
    fmt.Println("Message:", valErr.Message)
}
```

- `errors.Is`: **"is this error?"** — compares identity
- `errors.As`: **"is it of this type?"** — extracts concrete type

## Custom error types

For errors with extra information:

```go
type HTTPError struct {
    StatusCode int
    Message    string
    URL        string
}

func (e *HTTPError) Error() string {
    return fmt.Sprintf("HTTP %d: %s (url: %s)", e.StatusCode, e.Message, e.URL)
}

func fetchData(url string) error {
    resp, err := http.Get(url)
    if err != nil {
        return fmt.Errorf("fetch %s: %w", url, err)
    }
    if resp.StatusCode != 200 {
        return &HTTPError{
            StatusCode: resp.StatusCode,
            Message:    resp.Status,
            URL:        url,
        }
    }
    return nil
}

// Usage
err := fetchData("https://api.example.com/data")
var httpErr *HTTPError
if errors.As(err, &httpErr) {
    if httpErr.StatusCode == 404 {
        // handle not found
    }
}
```

## Pattern: error handling across multiple operations

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("open: %w", err)
    }
    defer f.Close()

    data, err := io.ReadAll(f)
    if err != nil {
        return fmt.Errorf("read: %w", err)
    }

    if err := validate(data); err != nil {
        return fmt.Errorf("validate: %w", err)
    }

    if err := save(data); err != nil {
        return fmt.Errorf("save: %w", err)
    }

    return nil
}
```

Each step adding context with `%w` creates an error "stack trace":
`"save: validate: invalid format"` — you know exactly where it failed.

## panic vs error

| | `error` | `panic` |
|---|---|---|
| **Usage** | Expected situations | Unrecoverable bugs |
| **Examples** | File does not exist, invalid input | Index out of range, nil pointer |
| **Recoverable?** | Yes, with if err != nil | Yes, with recover() (but not recommended) |
| **Flow** | Propagates via return | Unwinds the entire call stack |

> **Rule**: if you can anticipate the error, use `error`. `panic` is for programmer bugs or impossible corrupted state.

### panic in practice

```go
// Acceptable: initialization that MUST work
func mustCompile(pattern string) *regexp.Regexp {
    re, err := regexp.Compile(pattern)
    if err != nil {
        panic(fmt.Sprintf("invalid regex %q: %v", pattern, err))
    }
    return re
}

// Acceptable: Must* functions in the stdlib (template.Must, regexp.MustCompile)
var emailRegex = regexp.MustCompile(`^[a-z]+@[a-z]+\.[a-z]+$`)
```

## Common interview questions

1. **Why doesn't Go have exceptions?**
   To force explicit error handling. Exceptions create invisible control flows and make it hard to know what can fail. In Go, every error is visible in the function signature.

2. **Difference between errors.Is and errors.As?**
   `errors.Is` compares identity (sentinel errors). `errors.As` extracts a concrete type. Both work through error wrapping.

3. **What is error wrapping and why does it matter?**
   `fmt.Errorf("context: %w", err)` wraps an error adding context. It allows knowing WHERE it failed (via context) and WHAT failed (via errors.Is/As on the original error).

4. **When would you use panic?**
   Only for programmer bugs (impossible state). Never for I/O errors, validation, or user input. Exceptions: `Must*` functions that run during init.

5. **What is the problem with verbose error handling in Go?**
   `if err != nil` is repeated a lot, which can seem excessive. But this verbosity is deliberate: it makes explicit what can fail and forces you to decide how to handle it. The tradeoff is clarity over brevity.

6. **How would you make an error that contains additional information?**
   By creating a custom error type (struct with Error() method). You can include fields like StatusCode, Field, Timestamp, etc. It is extracted with errors.As.
