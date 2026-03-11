# Concurrency Puzzles — Traps and Patterns

Practical exercises to master concurrency in Go. Each puzzle presents code with a concurrency bug. Your task is to identify the problem, explain it, and propose the correct solution.

---

## Table of Contents

1. [Common Concurrency Traps](#common-concurrency-traps)
2. [Puzzle 1: Race Condition Detection](#puzzle-1-race-condition-detection)
3. [Puzzle 2: Deadlock](#puzzle-2-deadlock)
4. [Puzzle 3: Goroutine Leak](#puzzle-3-goroutine-leak)
5. [Puzzle 4: Channel Direction Bug](#puzzle-4-channel-direction-bug)
6. [Puzzle 5: Closure in Loop](#puzzle-5-closure-in-loop)
7. [Puzzle 6: Select Priority](#puzzle-6-select-priority)
8. [Puzzle 7: Context Cancellation](#puzzle-7-context-cancellation)
9. [Puzzle 8: WaitGroup Misuse](#puzzle-8-waitgroup-misuse)

---

## Common Concurrency Traps

Before the puzzles, these are the most frequent traps that cause concurrency bugs in Go:

### 1. Data Races
Accessing shared memory without synchronization. Go has the race detector (`go run -race`) that detects this at runtime, but it only covers code paths that are actually executed.

### 2. Goroutine Leaks
Goroutines that remain blocked forever waiting on a channel or lock that is never released. They are the equivalent of memory leaks but for goroutines. Tools like `goleak` help detect them in tests.

### 3. Deadlocks
Two or more goroutines waiting for each other, creating a dependency cycle. The Go runtime detects when **all** goroutines are blocked (`fatal error: all goroutines are asleep`), but does not detect partial deadlocks.

### 4. Starvation
A goroutine never gets access to a resource because other goroutines monopolize it. Common with `sync.Mutex` under high contention.

### 5. Premature Closure
Closing a channel or cancelling a context before all consumers have finished.

---

## Puzzle 1: Race Condition Detection

### Buggy Code

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    counter := 0
    var wg sync.WaitGroup

    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter++ // DATA RACE: concurrent access without synchronization
        }()
    }

    wg.Wait()
    fmt.Println("Counter:", counter) // Unpredictable result, almost never 1000
}
```

### Question
What is wrong with this code? How would you fix it? Provide at least three different ways.

### Problem Explanation
1000 goroutines access the `counter` variable simultaneously without synchronization. `counter++` is not atomic — it involves reading, incrementing, and writing. Two goroutines can read the same value, increment it, and write the same result, losing an increment.

Running with `go run -race main.go` reports the data race.

### Fix 1: sync.Mutex

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    counter := 0
    var mu sync.Mutex
    var wg sync.WaitGroup

    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            mu.Lock()
            counter++
            mu.Unlock()
        }()
    }

    wg.Wait()
    fmt.Println("Counter:", counter) // Always 1000
}
```

### Fix 2: Channel

```go
package main

import "fmt"

func main() {
    counter := 0
    done := make(chan struct{})
    increment := make(chan struct{}, 1000)

    // A single goroutine manages the counter (confinement)
    go func() {
        for range increment {
            counter++
        }
        done <- struct{}{}
    }()

    for i := 0; i < 1000; i++ {
        increment <- struct{}{}
    }
    close(increment)
    <-done
    fmt.Println("Counter:", counter) // Always 1000
}
```

### Fix 3: sync/atomic

```go
package main

import (
    "fmt"
    "sync"
    "sync/atomic"
)

func main() {
    var counter atomic.Int64
    var wg sync.WaitGroup

    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter.Add(1) // Atomic operation, no lock
        }()
    }

    wg.Wait()
    fmt.Println("Counter:", counter.Load()) // Always 1000
}
```

### Takeaway
For simple counters, `sync/atomic` is the most efficient option. For more complex logic, `sync.Mutex`. The channel pattern (confinement) is preferable when you can design the data flow without shared state.

---

## Puzzle 2: Deadlock

### Buggy Code

```go
package main

import "fmt"

func main() {
    ch1 := make(chan int)
    ch2 := make(chan int)

    // Goroutine A: sends to ch1, then receives from ch2
    go func() {
        ch1 <- 1     // blocks waiting for receiver on ch1
        val := <-ch2  // never reaches here
        fmt.Println("A received:", val)
    }()

    // Goroutine B: sends to ch2, then receives from ch1
    go func() {
        ch2 <- 2     // blocks waiting for receiver on ch2
        val := <-ch1  // never reaches here
        fmt.Println("B received:", val)
    }()

    // Main waits (without select or synchronization)
    select {}
}
```

### Question
Why does this code produce a deadlock? How would you fix it?

### Problem Explanation
Both goroutines try to **send** on unbuffered channels before **receiving**:
- Goroutine A blocks on `ch1 <- 1` waiting for someone to receive from ch1.
- Goroutine B blocks on `ch2 <- 2` waiting for someone to receive from ch2.
- Nobody receives from either channel. Circular deadlock.

In this case the runtime detects that all goroutines are asleep and throws: `fatal error: all goroutines are asleep - deadlock!`

### Fixed Code

```go
package main

import "fmt"

func main() {
    ch1 := make(chan int)
    ch2 := make(chan int)

    // Goroutine A: receives from ch2, then sends to ch1
    go func() {
        val := <-ch2 // first receives
        fmt.Println("A received:", val)
        ch1 <- 1 // then sends
    }()

    // Goroutine B: sends to ch2, then receives from ch1
    go func() {
        ch2 <- 2     // sends to ch2 (A receives it)
        val := <-ch1  // receives from ch1 (A sends it)
        fmt.Println("B received:", val)
    }()

    // We need to wait for them to finish
    // A simple way: use a done channel
    done := make(chan struct{}, 2)
    // (In real code we would use sync.WaitGroup, here we simplify)

    ch1Bis := make(chan int)
    ch2Bis := make(chan int)

    go func() {
        val := <-ch2Bis
        fmt.Println("A received:", val)
        ch1Bis <- 1
        done <- struct{}{}
    }()

    go func() {
        ch2Bis <- 2
        val := <-ch1Bis
        fmt.Println("B received:", val)
        done <- struct{}{}
    }()

    <-done
    <-done
}
```

**Cleaner alternative with buffered channels:**

```go
package main

import "fmt"

func main() {
    ch1 := make(chan int, 1) // buffered — send does not block
    ch2 := make(chan int, 1) // buffered — send does not block

    done := make(chan struct{})

    go func() {
        ch1 <- 1
        val := <-ch2
        fmt.Println("A received:", val)
        done <- struct{}{}
    }()

    go func() {
        ch2 <- 2
        val := <-ch1
        fmt.Println("B received:", val)
        done <- struct{}{}
    }()

    <-done
    <-done
}
```

### Takeaway
To avoid deadlocks: (1) avoid circular dependencies between channels, (2) use buffered channels when appropriate, (3) establish a consistent acquire/release order, (4) use `select` with `default` or timeout as an escape.

---

## Puzzle 3: Goroutine Leak

### Buggy Code

```go
package main

import (
    "fmt"
    "net/http"
    "time"
)

func fetchFirst(urls []string) string {
    ch := make(chan string)

    for _, url := range urls {
        go func(u string) {
            resp, err := http.Get(u)
            if err != nil {
                return
            }
            defer resp.Body.Close()
            ch <- u // The losing goroutines remain blocked here FOREVER
        }(url)
    }

    return <-ch // Only receives the first, the rest are leaked
}

func main() {
    result := fetchFirst([]string{
        "https://httpbin.org/delay/1",
        "https://httpbin.org/delay/2",
        "https://httpbin.org/delay/3",
    })
    fmt.Println("First result:", result)
    time.Sleep(5 * time.Second) // The leaked goroutines are still alive
}
```

### Question
Where is the goroutine leak? How would you prevent it?

### Problem Explanation
N goroutines are launched, but only **one** result is consumed from channel `ch` (which is unbuffered). The N-1 remaining goroutines remain permanently blocked on `ch <- u` because nobody else receives from the channel. These goroutines never finish and their memory is never freed.

### Fixed Code

```go
package main

import (
    "context"
    "fmt"
    "net/http"
)

func fetchFirst(ctx context.Context, urls []string) string {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel() // Cancel all remaining goroutines

    ch := make(chan string, len(urls)) // Buffer for all goroutines

    for _, url := range urls {
        go func(u string) {
            req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
            if err != nil {
                return
            }
            resp, err := http.DefaultClient.Do(req)
            if err != nil {
                return // Context cancelled or network error
            }
            defer resp.Body.Close()

            select {
            case ch <- u:
            case <-ctx.Done():
                return // Don't block if we already have a result
            }
        }(url)
    }

    return <-ch
}

func main() {
    result := fetchFirst(context.Background(), []string{
        "https://httpbin.org/delay/1",
        "https://httpbin.org/delay/2",
        "https://httpbin.org/delay/3",
    })
    fmt.Println("First result:", result)
    // All goroutines will terminate thanks to context cancellation
}
```

**Key points of the fix:**
1. Buffered channel (`len(urls)`) so that sends do not block.
2. `context.WithCancel` to cancel pending HTTP requests.
3. `select` with `ctx.Done()` so goroutines do not get stuck.

### Takeaway
Always make sure every goroutine has a way to terminate. Use `context.Context` for cancellation, channels with adequate buffer, and `select` with exit cases. The `go.uber.org/goleak` library can detect leaks in tests.

---

## Puzzle 4: Channel Direction Bug

### Buggy Code

```go
package main

import "fmt"

func producer(ch chan int) {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch)
}

func consumer(ch chan int) {
    for val := range ch {
        fmt.Println(val)
    }
    ch <- 99 // BUG: the consumer is sending to the channel after consuming
}

func main() {
    ch := make(chan int)

    go producer(ch)
    consumer(ch) // panic: send on closed channel
}
```

### Question
What is the bug? How would you fix it using channel directions?

### Problem Explanation
The consumer sends a value to the channel (`ch <- 99`) after the producer has closed it. Sending to a closed channel causes `panic: send on closed channel`. Additionally, conceptually, a consumer should not send to the input channel.

The real problem is that there are no type restrictions on the channels — both functions receive `chan int`, which allows both sending and receiving.

### Fixed Code

```go
package main

import "fmt"

// ch <-chan int: can only RECEIVE (receive-only)
// ch chan<- int: can only SEND (send-only)

func producer(ch chan<- int) { // send-only: can only send
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch) // close is only valid on send-only or bidirectional
}

func consumer(ch <-chan int) { // receive-only: can only receive
    for val := range ch {
        fmt.Println(val)
    }
    // ch <- 99 // COMPILE ERROR: cannot send on receive-only channel
}

func main() {
    ch := make(chan int)

    go producer(ch) // chan int is automatically converted to chan<- int
    consumer(ch)    // chan int is automatically converted to <-chan int
}
```

### Takeaway
Always use channel directions in function signatures. The compiler prevents errors at compile time. Rule: `chan<-` for the producer (send-only), `<-chan` for the consumer (receive-only). Only `close()` from the producer side.

---

## Puzzle 5: Closure in Loop

### Buggy Code

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    var wg sync.WaitGroup
    values := []string{"a", "b", "c", "d", "e"}

    for _, v := range values {
        wg.Add(1)
        go func() {
            defer wg.Done()
            fmt.Println(v) // BUG: all goroutines capture the SAME variable v
        }()
    }

    wg.Wait()
}
```

### Question
What is the probable output? Why? How do you fix it?

### Problem Explanation
**In Go < 1.22**: the `v` variable from the range loop is a **single variable** that is reused on each iteration. All goroutines capture a pointer to the same variable. When the goroutines execute, `v` already has the last value (`"e"`), so the typical output is:

```
e
e
e
e
e
```

**In Go 1.22+**: Go changed the semantics of loop variables. Now each iteration creates a new variable, so this code works correctly. However, it is fundamental to understand the problem for:
- Maintaining legacy code.
- Interviews (classic question).
- Other languages with similar closures.

### Fixed Code (compatible with all versions)

**Fix 1: Parameter in the goroutine**
```go
for _, v := range values {
    wg.Add(1)
    go func(val string) { // val is a local copy
        defer wg.Done()
        fmt.Println(val)
    }(v) // v is passed as argument (copied)
}
```

**Fix 2: Local variable in the loop (pre-Go 1.22)**
```go
for _, v := range values {
    v := v // shadow: creates a new local variable per iteration
    wg.Add(1)
    go func() {
        defer wg.Done()
        fmt.Println(v) // captures the local variable, not the loop's
    }()
}
```

**Fix 3: Go 1.22+ (no changes needed)**
```go
// With Go 1.22+, the original code works correctly.
// Each iteration of the loop creates a new variable v.
for _, v := range values {
    wg.Add(1)
    go func() {
        defer wg.Done()
        fmt.Println(v) // OK in Go 1.22+
    }()
}
```

### Takeaway
In Go < 1.22, closures inside loops capture the loop variable by reference (they share the same variable). Always pass as a parameter or shadow with `v := v`. In Go 1.22+ this was resolved at the language level, but it remains a fundamental interview question.

---

## Puzzle 6: Select Priority

### Buggy Code

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch1 := make(chan string, 1)
    ch2 := make(chan string, 1)

    ch1 <- "one"
    ch2 <- "two"

    // Both channels have data ready
    select {
    case msg := <-ch1:
        fmt.Println("Received from ch1:", msg)
    case msg := <-ch2:
        fmt.Println("Received from ch2:", msg)
    }

    // Question: does ch1 always execute because it comes first?
}
```

### Question
Which case executes? Is it deterministic?

### Problem Explanation
**It is not deterministic**. When multiple `select` cases are ready simultaneously, Go chooses one **at random** (uniform pseudo-random). There is no priority based on the order of cases.

Running this code multiple times will give different results:
```
$ go run main.go
Received from ch1: one
$ go run main.go
Received from ch2: two
$ go run main.go
Received from ch1: one
```

**Design reason**: to avoid starvation. If the first case always had priority, channels listed after it might never be served.

### How to Implement Real Priority

If you need priority, use nested `select`:

```go
func prioritySelect(high <-chan string, low <-chan string) string {
    // First try the high-priority channel
    select {
    case msg := <-high:
        return msg
    default:
    }

    // If high is not ready, wait for either
    select {
    case msg := <-high:
        return msg
    case msg := <-low:
        return msg
    }
}
```

**Robust alternative with loop:**
```go
func processWithPriority(ctx context.Context, high, low <-chan string) {
    for {
        select {
        case <-ctx.Done():
            return
        case msg := <-high:
            fmt.Println("HIGH:", msg)
        default:
            select {
            case <-ctx.Done():
                return
            case msg := <-high:
                fmt.Println("HIGH:", msg)
            case msg := <-low:
                fmt.Println("LOW:", msg)
            }
        }
    }
}
```

### Takeaway
`select` with multiple ready cases chooses at random — there is no priority by order. For real priority, use nested `select` with `default`. Keep in mind that the priority pattern is not perfect under high load — there may be cases where a low-priority message is processed before a high-priority one.

---

## Puzzle 7: Context Cancellation

### Buggy Code

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func longRunningTask(ctx context.Context) string {
    // Simulates heavy work that does NOT check the context
    result := ""
    for i := 0; i < 10; i++ {
        time.Sleep(500 * time.Millisecond) // Does not check ctx.Done()
        result += fmt.Sprintf("step-%d ", i)
    }
    return result
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    result := longRunningTask(ctx)
    fmt.Println("Result:", result)
    // The task takes 5 seconds even though the timeout is 1 second!
}
```

### Question
Why doesn't the context cancellation work?

### Problem Explanation
The `ctx` context is passed as a parameter but **is never checked inside the function**. `context.Context` is cooperative — it does not magically kill goroutines. The code inside the function must actively check `ctx.Done()` or `ctx.Err()` to react to cancellation.

### Fixed Code

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func longRunningTask(ctx context.Context) (string, error) {
    result := ""
    for i := 0; i < 10; i++ {
        // Check cancellation BEFORE each step
        select {
        case <-ctx.Done():
            return result, ctx.Err() // Return what we have + error
        default:
        }

        time.Sleep(500 * time.Millisecond)
        result += fmt.Sprintf("step-%d ", i)
    }
    return result, nil
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    result, err := longRunningTask(ctx)
    if err != nil {
        fmt.Println("Cancelled:", err)
        fmt.Println("Partial result:", result)
        return
    }
    fmt.Println("Complete result:", result)
}
```

**More idiomatic version using select with timer:**
```go
func longRunningTask(ctx context.Context) (string, error) {
    result := ""
    for i := 0; i < 10; i++ {
        select {
        case <-ctx.Done():
            return result, ctx.Err()
        case <-time.After(500 * time.Millisecond):
            result += fmt.Sprintf("step-%d ", i)
        }
    }
    return result, nil
}
```

### Takeaway
`context.Context` is cooperative. Every function that receives a context must check it periodically. Key patterns: (1) `select` with `ctx.Done()` in loops, (2) pass the context to downstream functions (HTTP requests, DB queries), (3) return `ctx.Err()` when cancellation is detected.

---

## Puzzle 8: WaitGroup Misuse

### Buggy Code

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    var wg sync.WaitGroup

    for i := 0; i < 5; i++ {
        go func(n int) {
            wg.Add(1) // BUG: Add inside the goroutine
            defer wg.Done()
            fmt.Println("Worker:", n)
        }(i)
    }

    wg.Wait() // May finish BEFORE all goroutines call Add
    fmt.Println("All finished") // Lie — maybe not
}
```

### Question
What is wrong? Why is the output inconsistent?

### Problem Explanation
`wg.Add(1)` is called **inside** the goroutine, but `wg.Wait()` is called in main. There is a race condition:

1. Main launches the goroutines and reaches `wg.Wait()`.
2. If no goroutine has executed `wg.Add(1)` yet, the WaitGroup counter is 0.
3. `wg.Wait()` returns immediately (counter == 0).
4. Main exits, killing goroutines that haven't even started.

The result is that some (or all) goroutines print nothing.

### Fixed Code

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    var wg sync.WaitGroup

    for i := 0; i < 5; i++ {
        wg.Add(1) // CORRECT: Add BEFORE launching the goroutine
        go func(n int) {
            defer wg.Done()
            fmt.Println("Worker:", n)
        }(i)
    }

    wg.Wait() // Now waits for all 5 goroutines to finish
    fmt.Println("All finished")
}
```

### WaitGroup Rules

```go
// CORRECT: Add before go
wg.Add(1)
go worker(&wg)

// INCORRECT: Add inside the goroutine
go func() {
    wg.Add(1) // race with wg.Wait()
    defer wg.Done()
    // ...
}()

// CORRECT: Add with total before the loop
wg.Add(len(items))
for _, item := range items {
    go process(item, &wg)
}

// INCORRECT: Done without prior Add (panic: negative WaitGroup counter)
wg.Done() // panic!

// INCORRECT: reusing WaitGroup before Wait returns
wg.Wait()
wg.Add(1) // OK only if Wait has fully returned
```

### Takeaway
**Always call `wg.Add()` before launching the goroutine**, never inside. The correct pattern is: Add -> go func -> defer Done. Alternatively, do a single `wg.Add(n)` with the total before the loop.

---

## Summary of Patterns

| Puzzle | Bug | Main Solution |
|---|---|---|
| 1. Race Condition | Concurrent access without sync | `sync.Mutex`, `atomic`, or confinement |
| 2. Deadlock | Circular dependency in channels | Change order, use buffer, or `select` |
| 3. Goroutine Leak | Goroutines blocked on sends | Adequate buffer + `context.Context` |
| 4. Channel Direction | Consumer sending to channel | Use `<-chan` and `chan<-` in signatures |
| 5. Closure in Loop | Shared loop variable | Parameter, shadow, or Go 1.22+ |
| 6. Select Priority | Assuming order in `select` | Nested `select` for real priority |
| 7. Context Cancellation | Not checking `ctx.Done()` | Periodic `select` with `ctx.Done()` |
| 8. WaitGroup | `Add` inside goroutine | `Add` before `go func()` |
