# Language Internals — Go In Depth

Exhaustive guide to Go internals for technical interview preparation. Each section covers the theory and implementation details that interviewers commonly ask about.

---

## Table of Contents

1. [Garbage Collector](#garbage-collector)
2. [Scheduler (GMP Model)](#scheduler-gmp-model)
3. [Memory Model](#memory-model)
4. [Escape Analysis](#escape-analysis)
5. [Stack vs Heap](#stack-vs-heap)
6. [Interface Internals](#interface-internals)
7. [Slice Internals](#slice-internals)
8. [Map Internals](#map-internals)
9. [Interview Questions](#interview-questions)

---

## Garbage Collector

### Tri-Color Mark-and-Sweep Algorithm

Go's GC uses a **concurrent** tri-color mark-and-sweep algorithm. Objects are classified into three colors:

- **White**: not yet visited. At the end of the marking cycle, white objects are considered garbage.
- **Gray**: visited, but their references have not yet been examined.
- **Black**: visited and all their references have been examined.

**Process:**

1. Initially all objects are white.
2. Root objects (stack, globals) are marked as gray.
3. A gray object is taken, its references are examined (and marked as gray), and the original object becomes black.
4. This repeats until no gray objects remain.
5. All remaining white objects are garbage and can be freed.

### Write Barrier

Since the GC runs **concurrently** with the program (mutator), there is a risk that the mutator modifies pointers while the GC is marking. The **write barrier** is a piece of code that executes every time the mutator writes a pointer, ensuring that the GC does not miss live objects.

Go uses a **hybrid write barrier** (since Go 1.8):
- Combines the Dijkstra write barrier (marks the new referenced object) with the Yuasa write barrier (marks the old referenced object).
- This eliminates the need to re-scan stacks during marking, reducing latency.

### GC Phases

1. **Mark Setup (STW)**: brief stop-the-world pause to activate the write barrier and prepare marking. All goroutines must reach a safe point.
2. **Marking (concurrent)**: traverses the object graph marking live objects. Runs in parallel with the program using up to 25% of CPU resources by default.
3. **Mark Termination (STW)**: second STW pause to deactivate the write barrier and perform final cleanup.
4. **Sweeping (concurrent)**: frees the memory of unmarked objects. Occurs incrementally as memory is needed.

### Tuning: GOGC and GOMEMLIMIT

**GOGC** (default 100):
- Controls GC frequency.
- Value = percentage of heap growth before triggering a new cycle.
- `GOGC=100`: the GC triggers when the heap doubles since the last cycle.
- `GOGC=200`: tolerates 3x the live heap size before collecting.
- `GOGC=off`: disables the GC (dangerous in production).

**GOMEMLIMIT** (Go 1.19+):
- Sets a soft limit on the total memory of the Go process.
- The GC becomes more aggressive when approaching the limit.
- Solves the classic problem: with high GOGC, the program can consume too much memory.
- Recommended pattern in production: `GOGC=off` + `GOMEMLIMIT=XGiB` when you want to maximize throughput with a fixed memory budget.

```go
// Example: configure via environment variables
// GOGC=100 GOMEMLIMIT=512MiB ./my-service

// Or programmatically:
import "runtime/debug"

func init() {
    debug.SetGCPercent(100)
    debug.SetMemoryLimit(512 << 20) // 512 MiB
}
```

### Latency vs Throughput

| Strategy | Latency | Throughput | Memory |
|---|---|---|---|
| Low GOGC (e.g. 50) | Lower (frequent but fast GC) | Lower (more time in GC) | Lower |
| High GOGC (e.g. 200) | Higher (less frequent GC but more work) | Higher (less overhead) | Higher |
| GOGC=off + GOMEMLIMIT | Variable | Maximized | Controlled |

**Rule of thumb**: for latency-sensitive services (APIs), keep GOGC at default and adjust GOMEMLIMIT. For batch processing, increase GOGC.

---

## Scheduler (GMP Model)

### The Three Components

- **G (Goroutine)**: lightweight unit of work. Contains the stack, instruction pointer, and metadata. Initial stack size: ~2-8 KB (varies by version).
- **M (Machine/OS Thread)**: operating system thread that executes Go code. Each M needs a P to execute goroutines.
- **P (Processor)**: logical scheduling resource. Contains a local run queue of goroutines and the local memory cache (mcache). The number of Ps is controlled by `GOMAXPROCS`.

```
                ┌─────────┐
                │ Global  │
                │Run Queue│
                └────┬────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
   ┌────▼────┐  ┌────▼────┐  ┌────▼────┐
   │   P0    │  │   P1    │  │   P2    │
   │Local RQ │  │Local RQ │  │Local RQ │
   └────┬────┘  └────┬────┘  └────┬────┘
        │            │            │
   ┌────▼────┐  ┌────▼────┐  ┌────▼────┐
   │   M0    │  │   M1    │  │   M2    │
   │(thread) │  │(thread) │  │(thread) │
   └─────────┘  └─────────┘  └─────────┘
```

### Work Stealing

When a P's local run queue is empty, it tries to steal goroutines from other Ps:

1. First checks the global run queue (1/61 of the time for fairness).
2. Then tries to steal half of the run queue from a random P.
3. If no work is found, checks the netpoller.
4. If there is still no work, the M parks and releases the P.

### Preemption

**Cooperative Preemption (before Go 1.14):**
- Goroutines only yielded the processor at safe points: function calls, channel operations, memory allocation, etc.
- Problem: a `for {}` loop without function calls would block the P indefinitely.

**Asynchronous Preemption (Go 1.14+):**
- The runtime sends signals (SIGURG on Unix) to Ms to force preemption.
- Solves the problem of non-cooperating goroutines.
- The runtime can pause a goroutine at any safe point in the code.

### GOMAXPROCS

- Controls how many Ps (and therefore how many OS threads execute goroutines in parallel) there are.
- Default = number of logical CPUs.
- `runtime.GOMAXPROCS(n)` allows changing it at runtime.
- **Do not confuse with the total number of threads** — there can be more Ms than Ps (for example, Ms blocked in syscalls).

### Goroutine Lifecycle

```
              ┌──────────┐
    go f() ──►│ Runnable │◄──── I/O ready, timer, chan recv
              └────┬─────┘
                   │ scheduler
              ┌────▼─────┐
              │ Running  │──── executing on an M/P
              └────┬─────┘
                   │
          ┌────────┼────────┐
          │        │        │
     ┌────▼───┐ ┌──▼───┐ ┌─▼──────┐
     │Waiting │ │Syscall│ │ Dead   │
     │(chan,  │ │(OS)   │ │(return)│
     │ mutex) │ └───┬───┘ └────────┘
     └────┬───┘     │
          │         │ syscall finishes
          └─────────┘
```

### Netpoller

- Mechanism for non-blocking I/O integrated into the scheduler.
- Uses `epoll` (Linux), `kqueue` (macOS), `IOCP` (Windows).
- When a goroutine does network I/O, instead of blocking an M:
  1. The file descriptor is registered with the netpoller.
  2. The goroutine transitions to "waiting" state.
  3. The M is free to execute another goroutine.
  4. When the I/O is ready, the netpoller wakes the goroutine and puts it on the run queue.

---

## Memory Model

### Happens-Before Relationships

Go's memory model defines when a read of a variable can **observe** a value written by another goroutine. It is based on **happens-before** relationships:

> If event A **happens-before** event B, then A is observable by B.

Main guarantees:

1. **Initialization**: the `init()` function of package A happens-before any function of a package that imports A.
2. **Goroutine creation**: the `go f()` statement happens-before the start of `f()`'s execution.
3. **Goroutine termination**: the return of a goroutine has **no** happens-before guarantee with respect to any event in another goroutine (unless explicit synchronization is used).
4. **Channels**: a send happens-before the corresponding receive completes.
5. **Buffered channels**: the receive of element i happens-before the send of element i+C completes (C = channel capacity).
6. **close(ch)**: happens-before a receive that returns a zero value due to a closed channel.
7. **sync.Mutex**: `Unlock()` happens-before any subsequent `Lock()`.
8. **sync.Once**: the return of `f()` in `once.Do(f)` happens-before any return of `once.Do`.

### sync/atomic Guarantees

The `sync/atomic` package provides atomic operations that also establish happens-before relationships:

```go
var x, y atomic.Int64

// Goroutine 1
x.Store(1)  // A
y.Store(1)  // B (A happens-before B)

// Goroutine 2
if y.Load() == 1 { // C
    // x.Load() is guaranteed to be 1 here
    // because A happens-before B, and B happens-before C
    fmt.Println(x.Load()) // 1
}
```

Since Go 1.19, atomic operations are **sequentially consistent**, meaning the total order of atomic operations is consistent with the program order of each goroutine.

### Channel Guarantees

```go
// A send on an unbuffered channel happens-before
// the corresponding receive completes.
ch := make(chan int)

go func() {
    x = 42          // A
    ch <- struct{}{} // B — send happens-before receive completes
}()

<-ch               // C — receive completes after the send
fmt.Println(x)     // D — x is guaranteed to be 42
```

### When NOT to Use Shared Memory

Prefer channels over shared memory when:

- Communication is between goroutines with a clear data flow.
- You need to transfer **ownership** of data.
- The pattern is producer-consumer, fan-out/fan-in, or pipeline.

Use mutex/atomics when:
- Protecting a shared cache.
- Maintaining a simple counter.
- Access is fast and contention is low.

> **"Do not communicate by sharing memory; instead, share memory by communicating."** — Go Proverb

---

## Escape Analysis

### What It Is

The Go compiler decides whether a variable can be allocated on the stack (cheap) or must escape to the heap (more expensive, requires GC). This process is called **escape analysis**.

### How to See It

```bash
go build -gcflags='-m' ./...
# More verbose output:
go build -gcflags='-m -m' ./...
```

Example output:
```
./main.go:10:6: can inline newUser
./main.go:11:9: &User{} escapes to heap
./main.go:15:2: moved to heap: x
```

### Common Escape Scenarios

#### 1. Returning a pointer to a local variable
```go
func newUser(name string) *User {
    u := User{Name: name} // escapes to heap — a pointer is returned
    return &u
}
```

#### 2. Conversion to interface
```go
func logValue(v any) {
    fmt.Println(v) // the argument may escape because 'any' is an interface
}

func main() {
    x := 42
    logValue(x) // x escapes to heap due to the interface conversion
}
```

#### 3. Closures that capture variables
```go
func makeCounter() func() int {
    count := 0 // escapes to heap — captured by the closure
    return func() int {
        count++
        return count
    }
}
```

#### 4. Slices that grow beyond their initial capacity
```go
func grow() []int {
    s := make([]int, 0)
    for i := 0; i < 1000; i++ {
        s = append(s, i) // may escape if the compiler cannot determine the size
    }
    return s
}
```

#### 5. Sending a pointer through a channel
```go
ch := make(chan *User)
u := &User{Name: "Ana"} // escapes — sent via channel, indeterminate lifetime
ch <- u
```

### Why It Matters for Performance

- **Stack**: allocation = moving the stack pointer (extremely fast, ~1 instruction).
- **Heap**: allocation = requesting memory from the allocator + will be tracked by the GC.
- Reducing heap escapes = less GC pressure = lower latency.
- **Tip**: pass structs by value (not pointer) when they are small and do not need mutation.

---

## Stack vs Heap

### Goroutine Stacks

- **Initial size**: typically 2-8 KB (depends on Go version).
- **Growth**: when the stack is full, Go allocates a new stack of **double** the size and copies all contents (copy stack, not segmented stacks since Go 1.4).
- **Shrinking**: the GC can reduce stacks that are using much less than allocated (typically when only 1/4 is used).

### Stack Copying

When a stack grows:
1. A new memory block of double the size is allocated.
2. All contents of the old stack are copied to the new one.
3. **All pointers** pointing to the old stack are updated (this is why Go does not easily allow pointers to the stack from C/assembly).
4. The old stack is freed.

This is O(n) with respect to the stack size, but occurs infrequently thanks to exponential growth.

### When Allocations Go to the Heap

- The variable escapes the function's scope (see Escape Analysis).
- The compiler cannot determine the size at compile time.
- Objects too large for the stack.
- Variables shared between goroutines (through pointers).

### Implications for GC Pressure

```
Stack allocation:
  - Automatically freed when the function returns
  - Does not involve the GC
  - Extremely fast

Heap allocation:
  - Requires the GC to track the object
  - Can cause longer STW pauses if there is a lot of garbage
  - Slower than stack allocation
```

**Strategies to reduce GC pressure:**
1. Use `sync.Pool` for temporary objects that are frequently reused.
2. Pre-allocate slices with known capacity: `make([]T, 0, expectedSize)`.
3. Pass small structs by value instead of by pointer.
4. Avoid unnecessary conversions to `interface{}`.
5. Reuse buffers with `bytes.Buffer` or `[]byte` pools.

---

## Interface Internals

### iface vs eface

Internally, Go represents interfaces in two ways:

**eface** (empty interface / `any` / `interface{}`):
```go
type eface struct {
    _type *_type // pointer to type info
    data  unsafe.Pointer // pointer to the data
}
```

**iface** (interface with methods):
```go
type iface struct {
    tab  *itab          // pointer to the method table + type info
    data unsafe.Pointer // pointer to the data
}

type itab struct {
    inter *interfacetype // interface type
    _type *_type         // concrete type
    hash  uint32         // type hash (for fast type assertions)
    _     [4]byte
    fun   [1]uintptr     // array of function pointers (methods)
}
```

**Key point**: both are 2 words (16 bytes on 64-bit). The itab is cached after its first creation.

### Interface Satisfaction at Compile Time

Go verifies that a type satisfies an interface **at compile time**:

```go
type Writer interface {
    Write([]byte) (int, error)
}

type MyWriter struct{}

func (m MyWriter) Write(p []byte) (int, error) {
    return len(p), nil
}

// Compile-time verification (common pattern):
var _ Writer = MyWriter{}       // OK
var _ Writer = (*MyWriter)(nil) // OK — verifies that *MyWriter implements Writer
```

If the type does not implement the interface, the compiler gives an error **immediately**.

### The Classic Gotcha: Nil Interface vs Interface with Nil Value

This is one of the most commonly asked errors in interviews:

```go
type MyError struct{}

func (e *MyError) Error() string { return "error" }

func getError() error {
    var err *MyError = nil // nil pointer to MyError
    return err             // CAUTION: returns iface{tab: *itab(MyError), data: nil}
}

func main() {
    err := getError()
    if err != nil {
        // THIS EXECUTES! err is not nil.
        // err is an interface with tab != nil (it knows the type)
        // but data == nil (the value is nil)
        fmt.Println("error:", err)
    }
}
```

**Explanation**:
- An interface is `nil` **only** when both `tab` and `data` are nil.
- When you assign a typed nil pointer to an interface, the interface has type information (tab != nil), so the interface is NOT nil.

**Solution**: return `nil` directly, not a typed nil pointer:
```go
func getError() error {
    var err *MyError = nil
    if err == nil {
        return nil // returns a nil interface (tab=nil, data=nil)
    }
    return err
}
```

### Type Assertion and Type Switch

**Type assertion** — extracts the concrete value:
```go
var w Writer = MyWriter{}
mw := w.(MyWriter)          // panics if w is not MyWriter
mw, ok := w.(MyWriter)      // ok=false if not MyWriter, no panic
```

**Internally**: compares the hash in the itab with the hash of the requested type. If they match, returns the data pointer. It is O(1).

**Type switch** — common and efficient pattern:
```go
switch v := i.(type) {
case string:
    fmt.Println("string:", v)
case int:
    fmt.Println("int:", v)
default:
    fmt.Println("other type")
}
```

---

## Slice Internals

### SliceHeader

A slice in Go is a three-field descriptor:

```go
type SliceHeader struct {
    Data uintptr // pointer to the underlying array
    Len  int     // number of current elements
    Cap  int     // total capacity of the underlying array
}
```

```
slice := []int{1, 2, 3, 4, 5}

SliceHeader:
┌──────┬─────┬─────┐
│ Data │ Len │ Cap │
│  *───┼──5──┼──5──│
└──┼───┴─────┴─────┘
   │
   ▼
┌───┬───┬───┬───┬───┐
│ 1 │ 2 │ 3 │ 4 │ 5 │  (underlying array)
└───┴───┴───┴───┴───┘
```

### Append Behavior and Reallocation

```go
s := make([]int, 0, 4) // len=0, cap=4

s = append(s, 1, 2, 3)    // len=3, cap=4 (same array)
s = append(s, 4)           // len=4, cap=4 (same array)
s = append(s, 5)           // len=5, cap=8 (NEW array, copies data)
```

**Growth strategy** (simplified, varies by version):
- If cap < 256: double the capacity.
- If cap >= 256: grow ~25% + a bit more (formula adjusted since Go 1.18).

**Important**: when append causes reallocation, the new slice points to a different array. Slices that shared the original array **do not see the changes**.

### Slice Tricks

```go
// Remove element at index i (does not preserve order — O(1)):
s[i] = s[len(s)-1]
s = s[:len(s)-1]

// Remove element at index i (preserves order — O(n)):
s = append(s[:i], s[i+1:]...)

// Insert element at index i:
s = append(s[:i], append([]T{elem}, s[i:]...)...)
// Better alternative (no intermediate allocation):
s = append(s, zero) // grow by 1
copy(s[i+1:], s[i:])
s[i] = elem

// Copy a slice (independent from the original):
copia := make([]int, len(s))
copy(copia, s)
// Or with Go 1.21+:
copia := slices.Clone(s)

// Filter in-place:
n := 0
for _, v := range s {
    if keep(v) {
        s[n] = v
        n++
    }
}
s = s[:n]
```

### Memory Leaks with Slices

**Classic problem**: when sub-slicing, the entire underlying array remains in memory:

```go
func getFirstThree(data []byte) []byte {
    return data[:3] // DANGER: retains reference to the entire array
}

// If 'data' is 1GB, that 1GB cannot be freed.
```

**Solution**: copy the needed data:
```go
func getFirstThree(data []byte) []byte {
    result := make([]byte, 3)
    copy(result, data[:3])
    return result
    // Or with Go 1.21+:
    // return bytes.Clone(data[:3])
}
```

---

## Map Internals

### Hash Table with Buckets

A `map[K]V` in Go is a hash table implemented with buckets:

- Each bucket stores up to **8 key-value pairs**.
- The hash of the key determines which bucket it goes to.
- The top 8 bits of the hash (tophash) are stored in the bucket for fast comparison.
- If a bucket is full, overflow buckets are chained.

```
map[string]int

┌─────────────────────────┐
│  Buckets Array          │
├─────────┬───────────────┤
│Bucket 0 │ tophash[8]    │
│         │ keys[8]       │
│         │ values[8]     │
│         │ overflow *    │
├─────────┼───────────────┤
│Bucket 1 │ ...           │
├─────────┼───────────────┤
│  ...    │               │
└─────────┴───────────────┘
```

**Layout note**: keys and values are stored separately (first all keys, then all values) to avoid unnecessary padding. For example, `map[int64]int8` — without this optimization, each pair would need 16 bytes (due to padding); with it, 8 keys of 8 bytes + 8 values of 1 byte.

### Growth and Evacuation

When the load factor exceeds ~6.5 (average of 6.5 elements per bucket), the map grows:

1. A new bucket array of **double** the size is allocated.
2. The data is **not** copied immediately (unlike stacks).
3. **Incremental evacuation** is used: each insert/delete operation migrates some old buckets to the new array.
4. During evacuation, both the old and new arrays coexist.

This distributes the growth cost over time, avoiding latency spikes.

### Maps Are Not Safe for Concurrent Access

```go
m := make(map[string]int)

// THIS CAUSES A DATA RACE (and possible crash):
go func() { m["a"] = 1 }()
go func() { m["b"] = 2 }()
```

**Reason**: the runtime detects concurrent writes to the map and throws a **fatal error** (not a recoverable panic — the program dies):

```
fatal error: concurrent map writes
```

**Solutions**:
1. `sync.Mutex` or `sync.RWMutex` to protect the map.
2. `sync.Map` for specific cases (see below).
3. Confinement pattern: each goroutine has its own map.

**sync.Map** is optimal for:
- Keys written once and read many times.
- Goroutines working on disjoint sets of keys.
- **Not** a general replacement for map + mutex.

### Randomized Iteration Order

```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
for k, v := range m {
    fmt.Println(k, v) // different order each execution
}
```

**By design**: Go randomizes the iteration order of maps to prevent code from depending on a particular order. This is implemented by choosing a random starting offset for the iteration.

---

## Interview Questions

### Question 1: Explain Go's GMP scheduler model.
**Answer**: G is a goroutine (lightweight unit of work with its own stack). M is an OS thread that executes code. P is a logical processor with a local run queue. The number of Ps is controlled by GOMAXPROCS. The scheduler assigns Gs to Ms through Ps. When a P has no work, it steals goroutines from other Ps (work stealing). This allows thousands of goroutines to run efficiently on few OS threads.

### Question 2: What happens when a goroutine makes a blocking syscall?
**Answer**: The M executing the goroutine blocks with it. The P detaches from the blocked M and looks for another free M (or creates a new one) to continue executing goroutines. When the syscall finishes, the M tries to reacquire a P. If there are no free Ps, the goroutine goes to the global run queue and the M parks.

### Question 3: What is the difference between a nil interface and an interface containing a nil value?
**Answer**: An interface in Go is two words: (type, value). A nil interface has both nil. An interface containing a nil pointer has type != nil and value == nil. Comparing with `!= nil` returns `true` for the latter because the interface "knows" what type it contains. This is a common source of bugs, especially when returning `error`.

### Question 4: How does escape analysis work and why does it matter?
**Answer**: The compiler analyzes whether a variable can live only on the stack or needs to escape to the heap. Stack allocations are practically free (just moving the stack pointer) and do not involve the GC. Heap allocations require the allocator and the GC tracks them. Reducing escapes improves performance. It can be inspected with `go build -gcflags='-m'`.

### Question 5: What happens internally when you append to a slice that is full?
**Answer**: If `len == cap`, append allocates a new underlying array with greater capacity (typically double if small, ~25% more if large), copies the elements from the old array to the new one, and returns a new SliceHeader pointing to the new array. The old array will eventually be collected by the GC if there are no more references.

### Question 6: Why are Go maps not safe for concurrent access?
**Answer**: For performance. Adding internal synchronization would penalize all uses, even single-threaded ones. Go opts to leave synchronization to the programmer. The runtime detects concurrent writes to the map and throws a fatal error (not a panic) to prevent silent data corruption.

### Question 7: What is the write barrier and what is it for?
**Answer**: The write barrier is code that executes every time the mutator (program) writes a pointer. It is necessary because the GC runs concurrently — without it, the GC might not see new references created during marking and could collect live objects. Go uses a hybrid write barrier that combines the Dijkstra and Yuasa techniques.

### Question 8: How does Go handle goroutine stack growth?
**Answer**: Go uses "copy stacks" (since Go 1.4). When a goroutine needs more stack, a new block of double the size is allocated, all contents are copied, all pointers pointing to the old stack are updated, and the old one is freed. This is O(n) but infrequent thanks to exponential growth.

### Question 9: Explain the phases of Go's Garbage Collector.
**Answer**: (1) Mark Setup: brief STW pause to activate the write barrier. (2) Marking: concurrent phase that traverses the object graph using the tri-color algorithm. Uses ~25% of CPU. (3) Mark Termination: brief STW pause to deactivate the write barrier. (4) Sweeping: concurrent and incremental freeing of unmarked objects.

### Question 10: What is GOMEMLIMIT and when would you use it?
**Answer**: GOMEMLIMIT (Go 1.19+) is a soft memory limit. The GC becomes more aggressive when approaching the limit. It is useful when the memory budget is known (e.g., a container with 512MB). It can be combined with GOGC=off to maximize throughput: the GC only runs when approaching the limit, not based on growth percentage.

### Question 11: What is the difference between iface and eface?
**Answer**: eface is the representation of `interface{}` (any): it has a pointer to the type and a pointer to the data. iface is for interfaces with methods: it has a pointer to an itab (which contains the method table + type information) and a pointer to the data. The itab is cached to avoid recalculating it on each assignment.

### Question 12: How would you avoid a memory leak with slices?
**Answer**: When sub-slicing, the entire underlying array remains in memory. If the original array is large and you only need a small portion, you must copy the data to a new slice with `copy()` or `slices.Clone()`. This also applies when removing elements: set `s[i] = zeroValue` before truncating to avoid retaining references to objects that should be collected.

### Question 13: What happens-before guarantees do channels provide?
**Answer**: (1) A send happens-before the corresponding receive completes. (2) Closing a channel happens-before a receive that returns a zero value. (3) In unbuffered channels, the receive happens-before the send completes. (4) In buffered channels with cap C, the receive of element k happens-before the send of element k+C completes.

### Question 14: What is the netpoller and how does it integrate with the scheduler?
**Answer**: The netpoller uses OS mechanisms (epoll/kqueue/IOCP) for non-blocking network I/O. When a goroutine does network I/O, the file descriptor is registered with the netpoller, the goroutine moves to "waiting", and the M is freed. When the I/O is ready, the goroutine is put on the run queue. This allows thousands of goroutines to do concurrent I/O without needing thousands of threads.

### Question 15: How does asynchronous preemption work in Go 1.14+?
**Answer**: Before Go 1.14, goroutines only yielded the processor at cooperative points (function calls, channel operations). A `for {}` loop could block a P indefinitely. Since Go 1.14, the runtime uses OS signals (SIGURG on Unix) to interrupt goroutines at any safe point. The sysmon thread detects goroutines that have been running for >10ms without yielding and sends the signal.

### Question 16: What is the difference between sync.Mutex and sync.RWMutex? When to use each?
**Answer**: `sync.Mutex` allows a single access (read or write). `sync.RWMutex` allows multiple simultaneous readers but only one writer. Use RWMutex when reads are much more frequent than writes (typical ratio >10:1). If writes are frequent, the overhead of RWMutex does not justify its complexity and a simple Mutex is preferable.

### Question 17: What is a data race and how does it differ from a race condition?
**Answer**: A **data race** is when two goroutines access the same variable, at least one writes, and there is no synchronization. It is undefined behavior. A **race condition** is a logical bug where the result depends on the execution order. You can have race conditions without data races (using synchronization but with incorrect logic). Go detects data races with `go run -race`.
