# 06 - Concurrency

> "Do not communicate by sharing memory; instead, share memory by communicating." — Go Proverbs

Concurrency is Go's most powerful feature and the most asked about in interviews. Go was designed from the ground up for concurrency.

**Concurrency vs Parallelism**:
- **Concurrency**: structuring a program as independent tasks that can execute in any order
- **Parallelism**: executing multiple tasks simultaneously on multiple CPUs

Go gives you concurrency. The runtime decides whether there is also parallelism (depends on available cores).

## Goroutines

A goroutine is a **lightweight thread** managed by the Go runtime (not by the OS):

```go
func sayHello(name string) {
    fmt.Printf("Hello, %s!\n", name)
}

func main() {
    go sayHello("World")  // launches goroutine — does NOT wait
    go sayHello("Go")     // launches another

    time.Sleep(time.Millisecond) // without this, main exits before they execute
}
```

### Goroutines vs OS Threads

| | Goroutine | OS Thread |
|---|---|---|
| **Initial memory** | ~2 KB (stack grows dynamically) | ~1 MB (fixed) |
| **Creation** | ~microseconds | ~milliseconds |
| **Scheduling** | Go runtime (user-space) | OS kernel |
| **Typical count** | Thousands or millions | Hundreds |
| **Context switch** | ~nanoseconds | ~microseconds |

> You can launch **100,000 goroutines** without problems. Trying the same with OS threads would crash your system.

### The GMP model (important for interviews)

Go's scheduler uses the **G-M-P** model:

```
G = Goroutine       (the task)
M = Machine/Thread  (actual OS thread)
P = Processor       (execution context, default = num CPUs)

    P0          P1          P2          P3
    |           |           |           |
    M0          M1          M2          M3
    |           |           |           |
   [G1]       [G2]       [G3]       [G4]
   [G5]       [G6]       [G7]       [G8]   <- local run queues
    ...         ...
                    [G9, G10, G11...]      <- global run queue
```

- Each **P** has a local queue of goroutines
- When a queue is empty, the P "steals" work from another P (**work stealing**)
- `GOMAXPROCS` controls how many Ps there are (default = num CPUs)

## Channels

Channels are the primary communication mechanism between goroutines:

```go
// Create a channel
ch := make(chan int)    // unbuffered
ch := make(chan int, 5) // buffered (capacity 5)

// Send
ch <- 42

// Receive
value := <-ch

// Close (the sender closes, NEVER the receiver)
close(ch)
```

### Unbuffered channels (synchronous)

```go
ch := make(chan int) // no buffer

go func() {
    ch <- 42  // BLOCKS until someone receives
}()

value := <-ch  // BLOCKS until someone sends
fmt.Println(value) // 42
```

An unbuffered channel **synchronizes** sender and receiver: both block until the other is ready. It is like a handshake.

### Buffered channels (asynchronous up to the limit)

```go
ch := make(chan int, 3) // buffer of 3

ch <- 1  // does not block (1/3)
ch <- 2  // does not block (2/3)
ch <- 3  // does not block (3/3)
// ch <- 4  // BLOCKS — buffer full, waits for someone to receive

fmt.Println(<-ch) // 1 (FIFO)
```

### When to use each

| Unbuffered | Buffered |
|---|---|
| Guaranteed synchronization | Decoupling sender/receiver |
| Sender knows the receiver received | Sender can continue without waiting |
| Ideal for signals and handshakes | Ideal for rate limiting, batching |

### Directionality (channel types)

```go
func producer(out chan<- int) {  // can only SEND
    out <- 42
}

func consumer(in <-chan int) {   // can only RECEIVE
    value := <-in
}

// The compiler verifies that you don't use a send-only channel to receive
```

> Always specify the direction in function signatures. It is free documentation and the compiler verifies it.

### Iterating over a channel (range)

```go
ch := make(chan int)

go func() {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch) // IMPORTANT: close so that range terminates
}()

for value := range ch {
    fmt.Println(value) // 0, 1, 2, 3, 4
}
// The loop terminates when the channel is closed and emptied
```

### Comma-ok pattern with channels

```go
value, ok := <-ch
if !ok {
    fmt.Println("channel closed")
}
```

## Select

`select` is like a switch for channels. It waits for **any** to be ready:

```go
select {
case msg := <-ch1:
    fmt.Println("Received from ch1:", msg)
case msg := <-ch2:
    fmt.Println("Received from ch2:", msg)
case ch3 <- "hello":
    fmt.Println("Sent to ch3")
default:
    fmt.Println("No channel ready") // does not block
}
```

- If multiple channels are ready, **one is chosen at random** (non-deterministic)
- Without `default`, it blocks until one is ready
- With `default`, it never blocks (useful for polling)

### Select with timeout

```go
select {
case result := <-ch:
    fmt.Println("Got result:", result)
case <-time.After(3 * time.Second):
    fmt.Println("Timeout!")
}
```

### Select for done/quit signal

```go
func worker(done <-chan struct{}, jobs <-chan int) {
    for {
        select {
        case <-done:
            fmt.Println("Worker stopping")
            return
        case job := <-jobs:
            fmt.Println("Processing job", job)
        }
    }
}
```

> `chan struct{}` is idiomatic for signals (carries no data, takes 0 bytes).

## sync.WaitGroup

Wait for a group of goroutines to finish:

```go
var wg sync.WaitGroup

for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done() // decrements on exit
        fmt.Println("Worker", id)
    }(i)
}

wg.Wait() // blocks until counter reaches 0
```

- `Add(n)` — increments the counter
- `Done()` — decrements (equivalent to `Add(-1)`)
- `Wait()` — blocks until the counter is 0

> **Rule**: call `Add` BEFORE launching the goroutine, not inside. If you call Add inside the goroutine, there is a race condition with Wait.

## sync.Mutex and sync.RWMutex

To protect shared data:

```go
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}
```

### RWMutex — multiple readers, a single writer

```go
type SafeCache struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *SafeCache) Get(key string) (string, bool) {
    c.mu.RLock()         // multiple goroutines can read at once
    defer c.mu.RUnlock()
    val, ok := c.data[key]
    return val, ok
}

func (c *SafeCache) Set(key, value string) {
    c.mu.Lock()          // exclusive — blocks readers and writers
    defer c.mu.Unlock()
    c.data[key] = value
}
```

> Use `RWMutex` when you have **many more reads than writes**. For everything else, use a regular `Mutex`.

### Mutex vs Channels — when to use each

| Mutex | Channel |
|---|---|
| Protect shared data | Communicate between goroutines |
| Cache, counters, maps | Pipelines, signals, results |
| "Guard this" | "Pass this" |

> **Practical rule**: if the operation is "sharing state", use Mutex. If it is "passing data between goroutines", use channels.

## sync.Once

Execute something **exactly once** (thread-safe):

```go
var once sync.Once
var instance *Database

func GetDB() *Database {
    once.Do(func() {
        instance = connectToDatabase() // only executes once
    })
    return instance
}
```

Useful for singletons, lazy initialization, and one-time setup.

## context.Context

Context propagates **deadlines, cancellation, and values** across goroutines:

```go
// Create with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel() // ALWAYS call cancel to release resources

// Create with manual cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Check cancellation
select {
case <-ctx.Done():
    fmt.Println("Cancelled:", ctx.Err())
case result := <-doWork(ctx):
    fmt.Println("Result:", result)
}
```

### Context in functions (standard pattern)

```go
// Context is ALWAYS the first parameter
func fetchData(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    // If the context is cancelled, the request is aborted automatically
    resp, err := http.DefaultClient.Do(req)
    // ...
}
```

> **Rule**: Context always goes as the **first parameter**, never in a struct. This is a strong convention in Go.

### Context values (use sparingly)

```go
type contextKey string

const userIDKey contextKey = "userID"

// Store
ctx := context.WithValue(parentCtx, userIDKey, "user-123")

// Read
if userID, ok := ctx.Value(userIDKey).(string); ok {
    fmt.Println("User:", userID)
}
```

> Use context values only for **request-scoped** data (user ID, trace ID, etc). Never for passing dependencies or config.

## Race conditions

A race condition occurs when multiple goroutines access shared data without synchronization:

```go
// BUG: race condition
counter := 0
for i := 0; i < 1000; i++ {
    go func() {
        counter++ // read + write is NOT atomic
    }()
}
// counter can be any value < 1000
```

### Detect with -race

```bash
go test -race ./...
go run -race main.go
```

The race detector is **essential** during development. It detects concurrent access without protection.

### Solutions to race conditions

```go
// Solution 1: Mutex
var mu sync.Mutex
mu.Lock()
counter++
mu.Unlock()

// Solution 2: Atomic (faster for simple operations)
var counter int64
atomic.AddInt64(&counter, 1)

// Solution 3: Channel (send updates to a controller goroutine)
ch := make(chan int)
go func() {
    count := 0
    for delta := range ch {
        count += delta
    }
}()
ch <- 1
```

## Concurrency patterns

### Fan-out / Fan-in

```go
// Fan-out: distribute work among N workers
func fanOut(jobs <-chan int, numWorkers int) []<-chan int {
    workers := make([]<-chan int, numWorkers)
    for i := 0; i < numWorkers; i++ {
        workers[i] = worker(jobs)
    }
    return workers
}

// Fan-in: combine results from N channels into one
func fanIn(channels ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    merged := make(chan int)

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for val := range c {
                merged <- val
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(merged)
    }()

    return merged
}
```

### Worker Pool

```go
func workerPool(numWorkers int, jobs <-chan int, results chan<- int) {
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for job := range jobs {
                results <- process(job) // each worker consumes from the same channel
            }
        }(i)
    }
    go func() {
        wg.Wait()
        close(results)
    }()
}
```

### Pipeline

```go
// Each stage is a function that reads from a channel and writes to another
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

func filter(in <-chan int, pred func(int) bool) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            if pred(n) {
                out <- n
            }
        }
        close(out)
    }()
    return out
}

// Compose: generate -> square -> filter
pipeline := filter(square(generate(1, 2, 3, 4, 5)), func(n int) bool {
    return n > 10
})
for v := range pipeline {
    fmt.Println(v) // 16, 25
}
```

### Graceful Shutdown

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()

    go runServer(ctx)

    <-ctx.Done() // wait for Ctrl+C
    fmt.Println("Shutting down...")

    // Give time for cleanup
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    server.Shutdown(shutdownCtx)
}
```

## Deadlocks

A deadlock occurs when goroutines block each other waiting for one another:

```go
// Classic deadlock: unbuffered channel without receiver
ch := make(chan int)
ch <- 1 // blocks forever — nobody receives
// fatal error: all goroutines are asleep - deadlock!

// Deadlock by lock ordering
// Goroutine 1: Lock(A), Lock(B)
// Goroutine 2: Lock(B), Lock(A)
// -> They block each other
```

Solution: always acquire locks in the **same order** in all goroutines.

## errgroup (golang.org/x/sync/errgroup)

Run goroutines that can fail and collect the first error:

```go
import "golang.org/x/sync/errgroup"

g, ctx := errgroup.WithContext(context.Background())

g.Go(func() error {
    return fetchUsers(ctx)
})

g.Go(func() error {
    return fetchOrders(ctx)
})

if err := g.Wait(); err != nil {
    // err is the first error that occurred
    // ctx is cancelled automatically when there is an error
    log.Fatal(err)
}
```

## Common interview questions

1. **What is a goroutine and how does it differ from a thread?**
   A goroutine is a lightweight thread managed by the Go runtime (~2KB stack, scheduling in user-space). An OS thread takes ~1MB and is managed by the kernel. You can have millions of goroutines but only hundreds of threads.

2. **Explain the GMP model of Go's scheduler.**
   G=Goroutine, M=Machine (OS thread), P=Processor (context). Each P has a local queue of Gs. Ms execute Gs through Ps. When a queue is empty, a P can "steal" work from another (work stealing). GOMAXPROCS controls how many Ps there are.

3. **Difference between buffered and unbuffered channel?**
   Unbuffered: sender and receiver block until the other is ready (synchronization). Buffered: sender only blocks if the buffer is full, receiver only blocks if it is empty (decoupling).

4. **When would you use Mutex vs Channel?**
   Mutex: to protect shared data (caches, counters, maps). Channel: to communicate data between goroutines (pipelines, results, signals). "Sharing state" -> Mutex. "Passing data" -> Channel.

5. **What is a race condition and how do you detect it?**
   Concurrent access to shared data without synchronization. Detected with `go test -race` or `go run -race`. Solutions: Mutex, atomic operations, or channels.

6. **What happens if you write to a closed channel?**
   **Panic**. Reading from a closed channel returns the zero value immediately. That is why **only the sender should close** the channel.

7. **What is context.Context and what is it used for?**
   Propagates deadlines, cancellation, and request-scoped values through the goroutine tree. It is always the first parameter. Used for HTTP timeouts, cancelling long operations, and passing data like trace IDs.

8. **Explain the fan-out/fan-in pattern.**
   Fan-out: distribute work from one channel among N workers (goroutines). Fan-in: combine results from N channels into a single channel. Allows parallelizing CPU-bound or I/O-bound work.

9. **How would you implement a graceful shutdown?**
   Capture the OS signal (SIGINT/SIGTERM) with signal.NotifyContext, propagate cancellation via context, give goroutines time to finish with a deadline, and close resources (DB, HTTP server) in an orderly fashion.

10. **What is a goroutine leak and how is it prevented?**
    A goroutine that blocks forever (waiting on a channel that nobody closes, a lock that nobody releases). It is prevented with context cancellation, timeouts, and ensuring every channel has a close path.
