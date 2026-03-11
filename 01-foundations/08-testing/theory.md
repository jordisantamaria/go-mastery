# 08 - Testing

Go has testing as a first-class citizen. You don't need external frameworks — the `testing` package from the stdlib is powerful and the community uses it universally.

## Basics

### Conventions

- Test files: `xxx_test.go` (Go automatically excludes them from the build)
- Test functions: `TestXxx(t *testing.T)` (starts with `Test` + uppercase)
- Same package as the tested code (or `package xxx_test` for black-box testing)

```go
// math.go
package math

func Add(a, b int) int {
    return a + b
}

// math_test.go
package math

import "testing"

func TestAdd(t *testing.T) {
    got := Add(2, 3)
    if got != 5 {
        t.Errorf("Add(2, 3) = %d, want 5", got)
    }
}
```

### Running tests

```bash
go test ./...                    # all tests in the project
go test ./pkg/...                # tests in a package and sub-packages
go test -v ./...                 # verbose — shows each test
go test -run TestAdd ./...       # only tests matching the regex
go test -race ./...              # with race detector
go test -count=1 ./...           # without cache
go test -short ./...             # skip tests marked as long
go test -cover ./...             # show coverage
go test -coverprofile=cover.out  # generate coverage file
go tool cover -html=cover.out    # open coverage in the browser
```

## Table-driven tests (THE Go pattern)

The most important pattern. **Every interview expects you to know it**:

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 2, 3, 5},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
        {"mixed", -1, 5, 4},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.expected {
                t.Errorf("Add(%d, %d) = %d, want %d",
                    tt.a, tt.b, got, tt.expected)
            }
        })
    }
}
```

### Why table-driven?

- **Easy to add cases**: just a new line in the table
- **Named**: each case has a descriptive name
- **Subtests**: `t.Run` allows running individual cases (`-run TestAdd/positive`)
- **DRY**: the assertion logic is written only once
- **Standard**: the entire Go ecosystem uses it

## Subtests with t.Run

```go
func TestUser(t *testing.T) {
    t.Run("creation", func(t *testing.T) {
        u := NewUser("Jordi")
        if u.Name != "Jordi" {
            t.Error("wrong name")
        }
    })

    t.Run("validation", func(t *testing.T) {
        t.Run("empty name", func(t *testing.T) {
            err := ValidateUser("")
            if err == nil {
                t.Error("expected error for empty name")
            }
        })

        t.Run("valid name", func(t *testing.T) {
            err := ValidateUser("Jordi")
            if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    })
}
```

Run a specific subtest:
```bash
go test -run TestUser/validation/empty_name ./...
```

## t.Helper()

Marks a function as a helper — errors report the caller's line, not the helper's:

```go
func assertEqual(t *testing.T, got, want int) {
    t.Helper() // without this, the error points to this line instead of the test
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}

func TestSomething(t *testing.T) {
    assertEqual(t, Add(1, 2), 3) // the error will point HERE, not inside assertEqual
}
```

## t.Parallel()

Run tests in parallel:

```go
func TestSlowA(t *testing.T) {
    t.Parallel() // mark as parallelizable
    time.Sleep(time.Second)
    // ...
}

func TestSlowB(t *testing.T) {
    t.Parallel()
    time.Sleep(time.Second)
    // ...
}
// Both run at the same time — total ~1s instead of ~2s
```

> **Caution**: parallel tests must not share mutable state. Each one must be independent.

### Parallel in table-driven tests

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"case1", 1, 2, 3},
        {"case2", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // each subtest runs in parallel
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("got %d, want %d", got, tt.want)
            }
        })
    }
}
```

## t.Cleanup()

Register cleanup functions that run at the end of the test:

```go
func TestWithDB(t *testing.T) {
    db := setupTestDB(t)
    t.Cleanup(func() {
        db.Close()  // runs at the end, no matter what
    })

    // use db...
}
```

## TestMain — global setup/teardown

```go
func TestMain(m *testing.M) {
    // Global setup (before all tests)
    db := setupDatabase()

    code := m.Run() // run all tests

    // Global teardown (after all tests)
    db.Close()

    os.Exit(code)
}
```

## t.Skip — skip tests conditionally

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    // long test...
}
```

```bash
go test -short ./...  # skips tests marked with t.Skip in short mode
```

## Mocking with interfaces (no frameworks)

In Go, mocking is done with **interfaces**, not magic frameworks:

```go
// The real code depends on an interface
type UserRepository interface {
    GetByID(id int) (*User, error)
    Save(user *User) error
}

type UserService struct {
    repo UserRepository
}

func (s *UserService) GetUser(id int) (*User, error) {
    return s.repo.GetByID(id)
}

// In tests, you create a mock that implements the interface
type mockUserRepo struct {
    users map[int]*User
    err   error
}

func (m *mockUserRepo) GetByID(id int) (*User, error) {
    if m.err != nil {
        return nil, m.err
    }
    user, ok := m.users[id]
    if !ok {
        return nil, ErrNotFound
    }
    return user, nil
}

func (m *mockUserRepo) Save(user *User) error {
    return m.err
}

// Test
func TestGetUser(t *testing.T) {
    mock := &mockUserRepo{
        users: map[int]*User{
            1: {ID: 1, Name: "Jordi"},
        },
    }

    service := &UserService{repo: mock}

    user, err := service.GetUser(1)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.Name != "Jordi" {
        t.Errorf("got name %q, want Jordi", user.Name)
    }

    // Test error case
    _, err = service.GetUser(999)
    if !errors.Is(err, ErrNotFound) {
        t.Errorf("expected ErrNotFound, got %v", err)
    }
}
```

> This pattern is **fundamental**. Dependency injection via interfaces + mock structs = testable without frameworks.

## Benchmarks

Measure the performance of functions:

```go
func BenchmarkAdd(b *testing.B) {
    for b.Loop() {  // Go 1.24+ — the runtime controls iterations
        Add(2, 3)
    }
}

// Benchmark with setup
func BenchmarkSort(b *testing.B) {
    data := generateRandomSlice(10000)
    b.ResetTimer() // don't count the setup

    for b.Loop() {
        sorted := make([]int, len(data))
        copy(sorted, data)
        sort.Ints(sorted)
    }
}
```

```bash
go test -bench=. ./...                 # run benchmarks
go test -bench=BenchmarkAdd ./...      # specific benchmark
go test -bench=. -benchmem ./...       # include allocations
go test -bench=. -count=5 ./...        # 5 iterations for stability
```

Output:
```
BenchmarkAdd-8    1000000000    0.3 ns/op    0 B/op    0 allocs/op
```

## Fuzzing (Go 1.18+)

Find bugs with random inputs:

```go
func FuzzReverse(f *testing.F) {
    // Seed corpus — initial examples
    f.Add("hello")
    f.Add("world")
    f.Add("")

    f.Fuzz(func(t *testing.T, s string) {
        reversed := Reverse(s)
        doubleReversed := Reverse(reversed)
        if s != doubleReversed {
            t.Errorf("Reverse(Reverse(%q)) = %q, want %q", s, doubleReversed, s)
        }
    })
}
```

```bash
go test -fuzz=FuzzReverse ./...          # run fuzzing
go test -fuzz=FuzzReverse -fuzztime=30s  # limit time
```

> Fuzzing is powerful for finding edge cases you would not think of: unusual Unicode strings, extreme numbers, empty inputs.

## Build tags for integration tests

```go
//go:build integration

package mypackage

func TestDatabaseIntegration(t *testing.T) {
    // only runs with: go test -tags=integration ./...
}
```

```bash
go test ./...                    # does NOT run integration tests
go test -tags=integration ./...  # DOES run them
```

## Golden files (snapshot testing)

```go
func TestRenderHTML(t *testing.T) {
    got := RenderHTML(data)

    golden := filepath.Join("testdata", t.Name()+".golden")

    if *update {  // flag -update to regenerate
        os.WriteFile(golden, []byte(got), 0644)
    }

    want, _ := os.ReadFile(golden)
    if got != string(want) {
        t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
    }
}
```

## Common mistakes in testing

### t.Fatal vs t.Error

```go
t.Error("this fails but the test CONTINUES")    // reports error, continues
t.Fatal("this fails and the test STOPS")         // reports error, stops immediately
t.Errorf("with format: got %d", got)             // like Error but with format
t.Fatalf("with format: %v", err)                 // like Fatal but with format
```

- Use `t.Fatal` when the rest of the test makes no sense without that value
- Use `t.Error` when you want to report multiple failures

## Common interview questions

1. **What is a table-driven test and why is it the standard in Go?**
   A test that defines a table of cases (struct slice), iterates over them, and runs each one as a subtest with t.Run. It is the standard because it is DRY, easy to extend, and each case is isolated and named.

2. **How do you mock dependencies in Go?**
   With interfaces. The code depends on an interface, and in tests you pass a mock implementation (a struct that satisfies it). No frameworks are needed.

3. **What is t.Helper() and when do you use it?**
   Marks a function as a test helper so that errors report the caller's line, not the helper's. Used in reusable assertion functions.

4. **Difference between t.Error and t.Fatal?**
   `t.Error` reports the failure and continues. `t.Fatal` reports and stops the test immediately. Use Fatal when a failure makes the rest of the test meaningless.

5. **How do you run benchmarks in Go?**
   `go test -bench=. ./...`. Benchmark functions use `testing.B` and the `b.Loop()` method (Go 1.24+) or `for i := 0; i < b.N; i++`. With `-benchmem` it shows allocations.

6. **What is fuzzing and when is it useful?**
   Testing with automatically generated random inputs. Finds edge cases (Unicode, extreme numbers, long strings) that a programmer would not think of. Available since Go 1.18.

7. **How do you separate unit tests from integration tests?**
   With build tags (`//go:build integration`) and running with `go test -tags=integration`. Or with `testing.Short()` and the `-short` flag.
