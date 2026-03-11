# 09 - Standard Library Deep Dive

Go's standard library is one of its greatest strengths. Unlike other languages where you need external dependencies for basic tasks, Go includes robust and well-designed packages for HTTP, I/O, JSON, concurrency, logging, and more. In this module we explore the most important packages in depth.

---

## net/http

The `net/http` package provides a complete HTTP server and client, ready for production.

### Basic server

```go
http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %s!", r.URL.Query().Get("name"))
})

log.Fatal(http.ListenAndServe(":8080", nil))
```

- `http.HandleFunc` registers a handler on the `DefaultServeMux`
- `http.ListenAndServe` starts the server. Use `log.Fatal` because if it fails, it is critical
- `w http.ResponseWriter` — writes the response
- `r *http.Request` — contains all the request info (headers, body, URL, method...)

### http.ServeMux with improved routing (Go 1.22+)

Go 1.22 introduced **improved pattern matching** in `http.ServeMux`:

```go
mux := http.NewServeMux()

// Method + path
mux.HandleFunc("GET /api/users", listUsers)
mux.HandleFunc("POST /api/users", createUser)

// Path parameters with {name}
mux.HandleFunc("GET /api/users/{id}", getUser)
mux.HandleFunc("DELETE /api/users/{id}", deleteUser)

// Access the path parameter
func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")  // new in Go 1.22
    fmt.Fprintf(w, "User ID: %s", id)
}
```

**Go 1.22+ matching rules:**
- `"GET /path"` — only accepts GET requests to `/path`
- `"/path"` — accepts any method
- `"/path/{param}"` — captures the segment as a path parameter
- `"/path/{param...}"` — captures the rest of the path (wildcard)
- More specific patterns take priority over general ones

### The Handler interface

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Any type that implements `ServeHTTP` is a Handler. `http.HandlerFunc` is an adapter:

```go
// HandlerFunc converts a function into a Handler
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
    f(w, r)
}
```

### Middleware pattern

A middleware is a function that wraps a Handler to add functionality:

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)  // call the actual handler
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}

func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                http.Error(w, "Internal Server Error", 500)
                log.Printf("panic recovered: %v", err)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// Chain middlewares
handler := recoveryMiddleware(loggingMiddleware(mux))
http.ListenAndServe(":8080", handler)
```

### http.Client with timeout

```go
client := &http.Client{
    Timeout: 10 * time.Second,
}

resp, err := client.Get("https://api.example.com/data")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, err := io.ReadAll(resp.Body)
```

> **Important**: ALWAYS configure a timeout on `http.Client`. The default client has no timeout, which can cause goroutine leaks.

### Graceful shutdown

```go
srv := &http.Server{Addr: ":8080", Handler: mux}

// Run server in a goroutine
go func() {
    if err := srv.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatalf("server error: %v", err)
    }
}()

// Wait for shutdown signal
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

// Graceful shutdown with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
srv.Shutdown(ctx)
```

---

## io

The `io` package defines the fundamental I/O interfaces in Go. The entire ecosystem is built on `io.Reader` and `io.Writer`.

### Reader and Writer

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}
```

- `Read` fills the buffer `p` with data. Returns `n` bytes read and `io.EOF` when finished.
- `Write` writes `p` to the destination. Returns `n` bytes written.

Implement `Reader`: `os.File`, `strings.Reader`, `bytes.Buffer`, `http.Response.Body`, `net.Conn`...

Implement `Writer`: `os.File`, `bytes.Buffer`, `http.ResponseWriter`, `os.Stdout`...

### io.Copy

```go
// Copy from src to dst, returns bytes copied
n, err := io.Copy(dst, src)

// Practical examples
io.Copy(os.Stdout, resp.Body)          // print response to terminal
io.Copy(file, resp.Body)               // download file
io.Copy(hash, file)                    // calculate file hash
```

### io.MultiReader

Concatenates multiple Readers into a single one:

```go
header := strings.NewReader("--- HEADER ---\n")
body := strings.NewReader("body content\n")
footer := strings.NewReader("--- FOOTER ---\n")

combined := io.MultiReader(header, body, footer)
io.Copy(os.Stdout, combined)
// Output:
// --- HEADER ---
// body content
// --- FOOTER ---
```

### io.TeeReader

Reads from a Reader and simultaneously writes to a Writer (like the Unix `tee` command):

```go
var buf bytes.Buffer
tee := io.TeeReader(resp.Body, &buf)

// When reading from tee, it also writes to buf
io.Copy(os.Stdout, tee)  // prints to terminal
// buf now contains a copy of the body
```

Useful for: logging the body of a request while processing it, calculating a hash while copying a file, etc.

### Interface composition

```go
// ReadWriter combines Reader and Writer
type ReadWriter interface {
    Reader
    Writer
}

// ReadCloser is a Reader that needs to be closed (e.g., resp.Body)
type ReadCloser interface {
    Reader
    Closer
}

// io.ReadAll reads all content from a Reader
data, err := io.ReadAll(reader)
```

---

## encoding/json

The `encoding/json` package is the most used for serializing/deserializing data in Go.

### Marshal and Unmarshal

```go
// Struct -> JSON (Marshal)
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

user := User{Name: "Alice", Email: "alice@example.com", Age: 30}
data, err := json.Marshal(user)
// {"name":"Alice","email":"alice@example.com","age":30}

// JSON -> Struct (Unmarshal)
var decoded User
err = json.Unmarshal(data, &decoded)
```

### Struct tags

```go
type Config struct {
    Host     string `json:"host"`              // rename field
    Port     int    `json:"port,omitempty"`     // omit if zero value
    Debug    bool   `json:"debug"`
    Internal string `json:"-"`                  // always ignore
    Password string `json:"password,omitempty"` // omit if empty
}

// omitempty omits the field if it is: 0, "", false, nil, empty slice/map
```

### json.MarshalIndent (pretty print)

```go
data, err := json.MarshalIndent(user, "", "  ")
// {
//   "name": "Alice",
//   "email": "alice@example.com",
//   "age": 30
// }
```

### json.Encoder and json.Decoder

For working with streams (files, HTTP bodies) instead of []byte:

```go
// Encoder: writes JSON to a Writer
file, _ := os.Create("data.json")
encoder := json.NewEncoder(file)
encoder.SetIndent("", "  ")
encoder.Encode(user)  // writes to the file

// Decoder: reads JSON from a Reader
file, _ := os.Open("data.json")
decoder := json.NewDecoder(file)
var user User
decoder.Decode(&user)
```

> **Encoder/Decoder vs Marshal/Unmarshal**: Use Encoder/Decoder for streams (files, HTTP). Use Marshal/Unmarshal for []byte in memory.

### Custom MarshalJSON

```go
type Status int

const (
    StatusActive Status = iota
    StatusInactive
    StatusBanned
)

func (s Status) MarshalJSON() ([]byte, error) {
    var str string
    switch s {
    case StatusActive:
        str = "active"
    case StatusInactive:
        str = "inactive"
    case StatusBanned:
        str = "banned"
    default:
        str = "unknown"
    }
    return json.Marshal(str)
}

func (s *Status) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err != nil {
        return err
    }
    switch str {
    case "active":
        *s = StatusActive
    case "inactive":
        *s = StatusInactive
    case "banned":
        *s = StatusBanned
    default:
        return fmt.Errorf("unknown status: %s", str)
    }
    return nil
}
```

### json.RawMessage

Delays parsing of part of the JSON:

```go
type Event struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"` // not parsed yet
}

var event Event
json.Unmarshal(data, &event)

// Now we parse the payload according to the type
switch event.Type {
case "user_created":
    var user User
    json.Unmarshal(event.Payload, &user)
case "order_placed":
    var order Order
    json.Unmarshal(event.Payload, &order)
}
```

---

## context

The `context` package allows propagating cancellations, deadlines, and values through a chain of calls. It is fundamental in HTTP servers and concurrent operations.

### Creating contexts

```go
// Root context — never cancelled
ctx := context.Background()

// Context that can be cancelled manually
ctx, cancel := context.WithCancel(parent)
defer cancel()  // ALWAYS defer cancel to avoid leaks

// Context with timeout (cancelled automatically)
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()

// Context with absolute deadline
deadline := time.Now().Add(5 * time.Second)
ctx, cancel := context.WithDeadline(parent, deadline)
defer cancel()

// Context with a value (use sparingly)
ctx := context.WithValue(parent, "requestID", "abc-123")
```

### Propagation in HTTP

```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()  // get context from the request

    // If the client closes the connection, ctx is cancelled automatically

    result, err := fetchFromDB(ctx)  // propagate context
    if err != nil {
        if ctx.Err() == context.Canceled {
            return  // the client is gone, no point in responding
        }
        http.Error(w, err.Error(), 500)
        return
    }
    json.NewEncoder(w).Encode(result)
}

func fetchFromDB(ctx context.Context) (Data, error) {
    // Use ctx to cancel the query if the context expires
    select {
    case <-ctx.Done():
        return Data{}, ctx.Err()
    case result := <-doQuery():
        return result, nil
    }
}
```

### Best practices with context

1. **`context.Context` always as the first parameter**: `func DoSomething(ctx context.Context, ...)`
2. **Do not store context in structs**: pass it as a parameter
3. **ALWAYS `defer cancel()`** after `WithCancel/WithTimeout/WithDeadline`
4. **`context.WithValue` only for request-scoped data**: request ID, auth token, trace ID. NEVER to pass function parameters
5. **`WithValue` keys should be private types** to avoid collisions:
   ```go
   type contextKey string
   const requestIDKey contextKey = "requestID"
   ctx = context.WithValue(ctx, requestIDKey, "abc-123")
   ```

---

## os / filepath

### Reading and writing files

```go
// Read entire file
data, err := os.ReadFile("config.json")

// Write entire file
err := os.WriteFile("output.txt", []byte("hello"), 0644)

// Open file for reading
file, err := os.Open("data.txt")
defer file.Close()

// Create file for writing
file, err := os.Create("output.txt")
defer file.Close()
file.WriteString("hello\n")

// Open with specific options
file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
```

### os.Args and environment variables

```go
// Command line arguments
fmt.Println(os.Args[0])  // program name
fmt.Println(os.Args[1:]) // arguments

// Environment variables
home := os.Getenv("HOME")
port, exists := os.LookupEnv("PORT")
os.Setenv("MY_VAR", "value")
```

### filepath.Walk / filepath.WalkDir

```go
// WalkDir is more efficient (Go 1.16+, does not Stat each entry)
err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
    if err != nil {
        return err
    }
    if !d.IsDir() && filepath.Ext(path) == ".go" {
        fmt.Println(path)
    }
    return nil
})
```

### Directory operations

```go
// Create directory (and parents)
os.MkdirAll("path/to/dir", 0755)

// Temporary files
tmpFile, err := os.CreateTemp("", "prefix-*.txt")
defer os.Remove(tmpFile.Name())  // clean up when done

// Temporary directory
tmpDir, err := os.MkdirTemp("", "myapp-*")
defer os.RemoveAll(tmpDir)

// filepath utilities
filepath.Join("path", "to", "file.txt")  // "path/to/file.txt"
filepath.Ext("file.tar.gz")              // ".gz"
filepath.Base("/path/to/file.txt")       // "file.txt"
filepath.Dir("/path/to/file.txt")        // "/path/to"
filepath.Abs("relative/path")            // "/full/path/relative/path"
```

---

## flag

The `flag` package parses command line arguments.

### Basic usage

```go
// Define flags
host := flag.String("host", "localhost", "server host")
port := flag.Int("port", 8080, "server port")
verbose := flag.Bool("verbose", false, "enable verbose logging")

// Parse (mandatory)
flag.Parse()

fmt.Printf("host=%s port=%d verbose=%t\n", *host, *port, *verbose)
fmt.Println("remaining args:", flag.Args())
```

### flag.StringVar (without pointer)

```go
var config struct {
    Host string
    Port int
}

flag.StringVar(&config.Host, "host", "localhost", "server host")
flag.IntVar(&config.Port, "port", 8080, "server port")
flag.Parse()
```

### Subcommands pattern

```go
// go run main.go serve --port 9090
// go run main.go migrate --steps 5

serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
servePort := serveCmd.Int("port", 8080, "port")

migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)
migrateSteps := migrateCmd.Int("steps", 1, "migration steps")

if len(os.Args) < 2 {
    fmt.Println("expected 'serve' or 'migrate' subcommand")
    os.Exit(1)
}

switch os.Args[1] {
case "serve":
    serveCmd.Parse(os.Args[2:])
    fmt.Printf("Serving on port %d\n", *servePort)
case "migrate":
    migrateCmd.Parse(os.Args[2:])
    fmt.Printf("Migrating %d steps\n", *migrateSteps)
default:
    fmt.Printf("unknown subcommand: %s\n", os.Args[1])
    os.Exit(1)
}
```

---

## log/slog (Go 1.21+)

`slog` is the official structured logging package. It replaces the `log` package for modern applications.

### Basic usage

```go
slog.Info("server started", "port", 8080)
// 2024/01/15 10:30:00 INFO server started port=8080

slog.Warn("disk almost full", "usage", 95.5, "path", "/data")
slog.Error("connection failed", "err", err, "host", "db.example.com")
```

### Levels

```go
slog.Debug("debug message")  // not shown by default
slog.Info("info message")
slog.Warn("warn message")
slog.Error("error message")
```

### Handlers: JSON and Text

```go
// Text handler (default, key=value format)
textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,  // show all levels
})

// JSON handler (for production / log aggregation)
jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
})

// Set as default logger
logger := slog.New(jsonHandler)
slog.SetDefault(logger)

// Now slog.Info produces JSON:
// {"time":"2024-01-15T10:30:00Z","level":"INFO","msg":"request","method":"GET"}
```

### With (logger with fixed fields)

```go
// Create logger with fields included in every message
reqLogger := slog.With("requestID", "abc-123", "userID", 42)

reqLogger.Info("processing request")
// INFO processing request requestID=abc-123 userID=42

reqLogger.Error("failed", "err", err)
// ERROR failed requestID=abc-123 userID=42 err=...
```

### Group (group fields)

```go
slog.Info("request completed",
    slog.Group("request",
        slog.String("method", "GET"),
        slog.String("path", "/api/users"),
    ),
    slog.Group("response",
        slog.Int("status", 200),
        slog.Duration("latency", 42*time.Millisecond),
    ),
)
// With JSON handler:
// {"msg":"request completed","request":{"method":"GET","path":"/api/users"},"response":{"status":200,"latency":"42ms"}}
```

---

## time

### Duration

```go
d := 5 * time.Second
d = 100 * time.Millisecond
d = 2*time.Hour + 30*time.Minute

// Convert
d.Seconds()      // float64
d.Milliseconds() // int64
d.String()        // "2h30m0s"

// Parse duration from string
d, err := time.ParseDuration("1h30m")
```

### Timers and Tickers

```go
// Timer: fires once
timer := time.NewTimer(2 * time.Second)
<-timer.C  // wait
// or cancel before: timer.Stop()

// time.After: shortcut for one-shot timer
select {
case result := <-ch:
    fmt.Println(result)
case <-time.After(5 * time.Second):
    fmt.Println("timeout")
}

// Ticker: fires repeatedly
ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()  // ALWAYS stop to avoid leaks
for t := range ticker.C {
    fmt.Println("tick at", t)
}
```

### Format and Parse

Go uses a unique **reference time**: `Mon Jan 2 15:04:05 MST 2006` (1/2 3:04:05 PM 2006).

```go
now := time.Now()

// Format
now.Format("2006-01-02")                    // "2024-01-15"
now.Format("2006-01-02 15:04:05")           // "2024-01-15 10:30:00"
now.Format(time.RFC3339)                    // "2024-01-15T10:30:00Z"
now.Format("02/Jan/2006 03:04 PM")          // "15/Jan/2024 10:30 AM"

// Parse
t, err := time.Parse("2006-01-02", "2024-01-15")
t, err := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")

// Parse with timezone
t, err := time.ParseInLocation("2006-01-02 15:04", "2024-01-15 10:30",
    time.FixedZone("CET", 1*60*60))
```

### Time zones

```go
loc, err := time.LoadLocation("Europe/Madrid")
madridTime := now.In(loc)

utc := now.UTC()

// Compare times
t1.Before(t2)
t1.After(t2)
t1.Equal(t2)
t1.Sub(t2)   // returns Duration
t1.Add(d)    // adds Duration
```

---

## strings / strconv / fmt

### strings.Builder

Efficient for building strings iteratively (avoids allocations):

```go
var b strings.Builder
for i := 0; i < 1000; i++ {
    fmt.Fprintf(&b, "item %d, ", i)
}
result := b.String()
```

> Concatenating with `+` in a loop creates a new string each time (O(n^2)). `strings.Builder` is O(n).

### strings.Cut (Go 1.18+)

Splits a string by the first separator found:

```go
before, after, found := strings.Cut("host:8080", ":")
// before="host", after="8080", found=true

before, after, found = strings.Cut("noport", ":")
// before="noport", after="", found=false

// Cleaner than strings.Split when you only need 2 parts
// Before: parts := strings.SplitN(s, ":", 2)
// Now: host, port, _ := strings.Cut(s, ":")
```

### Common strings functions

```go
strings.Contains("hello world", "world")   // true
strings.HasPrefix("hello", "he")           // true
strings.HasSuffix("hello.go", ".go")       // true
strings.TrimSpace("  hello  ")             // "hello"
strings.ToUpper("hello")                   // "HELLO"
strings.ToLower("HELLO")                   // "hello"
strings.Replace("aaa", "a", "b", 2)        // "bba"
strings.ReplaceAll("aaa", "a", "b")        // "bbb"
strings.Split("a,b,c", ",")               // ["a", "b", "c"]
strings.Join([]string{"a","b","c"}, "-")   // "a-b-c"
strings.Repeat("ab", 3)                   // "ababab"
strings.Count("hello", "l")               // 2
strings.Fields("  hello  world  ")         // ["hello", "world"]
strings.Map(unicode.ToUpper, "hello")      // "HELLO" (rune by rune)
```

### strconv

Conversions between strings and basic types:

```go
// String -> int
n, err := strconv.Atoi("42")         // 42, nil
n, err := strconv.Atoi("abc")        // 0, error

// Int -> string
s := strconv.Itoa(42)                // "42"

// Parsing with more control
i, err := strconv.ParseInt("FF", 16, 64)  // 255 (hex)
f, err := strconv.ParseFloat("3.14", 64)  // 3.14
b, err := strconv.ParseBool("true")       // true

// Formatting with more control
s := strconv.FormatFloat(3.14159, 'f', 2, 64)  // "3.14"
s := strconv.FormatInt(255, 16)                 // "ff"
```

### fmt verbs

```go
type Point struct{ X, Y int }
p := Point{1, 2}

fmt.Printf("%v\n", p)    // {1 2}            — default format
fmt.Printf("%+v\n", p)   // {X:1 Y:2}        — with field names
fmt.Printf("%#v\n", p)   // main.Point{X:1, Y:2} — full Go syntax
fmt.Printf("%T\n", p)    // main.Point        — type

fmt.Sprintf("%d", 42)     // "42"              — decimal integer
fmt.Sprintf("%x", 255)    // "ff"              — hexadecimal
fmt.Sprintf("%b", 8)      // "1000"            — binary
fmt.Sprintf("%f", 3.14)   // "3.140000"        — float
fmt.Sprintf("%.2f", 3.14) // "3.14"            — 2 decimal places
fmt.Sprintf("%q", "hi")   // "\"hi\""          — string with quotes
fmt.Sprintf("%p", &p)     // "0xc0000b4000"    — pointer
```

---

## sync (review + Pool)

> The basic primitives (`WaitGroup`, `Mutex`, `RWMutex`, `Once`) were covered in module 06. Here we do a quick review and dive deeper into `sync.Pool`.

### Quick review

```go
// WaitGroup — wait for N goroutines
var wg sync.WaitGroup
wg.Add(1)
go func() { defer wg.Done(); /* work */ }()
wg.Wait()

// Mutex — mutual exclusion
var mu sync.Mutex
mu.Lock()
// critical section
mu.Unlock()

// RWMutex — multiple readers, one writer
var rw sync.RWMutex
rw.RLock()   // read
rw.RUnlock()
rw.Lock()    // write
rw.Unlock()

// Once — execute exactly once
var once sync.Once
once.Do(func() { /* initialization */ })

// Map — concurrent map (without external Mutex)
var m sync.Map
m.Store("key", "value")
v, ok := m.Load("key")
m.Range(func(k, v any) bool { return true })
```

### sync.Pool — object reuse

`sync.Pool` reuses temporary objects to reduce pressure on the garbage collector:

```go
var bufPool = sync.Pool{
    New: func() any {
        return new(bytes.Buffer)
    },
}

func processRequest(data []byte) string {
    buf := bufPool.Get().(*bytes.Buffer)
    buf.Reset()  // IMPORTANT: clean before using
    defer bufPool.Put(buf)  // return to pool

    buf.Write(data)
    buf.WriteString(" processed")
    return buf.String()
}
```

**When to use `sync.Pool`:**
- Objects expensive to create that are used and discarded frequently
- Temporary buffers in hot paths
- Reusable encoders/decoders

**When NOT to use `sync.Pool`:**
- For objects that need to persist (the GC can clean the pool at any time)
- If object creation is cheap
- As a general cache (use `sync.Map` or a real cache)

```go
// Practical example: pool of JSON encoders
var encoderPool = sync.Pool{
    New: func() any {
        return &bytes.Buffer{}
    },
}

func jsonResponse(w http.ResponseWriter, data any) {
    buf := encoderPool.Get().(*bytes.Buffer)
    buf.Reset()
    defer encoderPool.Put(buf)

    enc := json.NewEncoder(buf)
    enc.Encode(data)

    w.Header().Set("Content-Type", "application/json")
    w.Write(buf.Bytes())
}
```

---

## Common interview questions

### 1. Why is it important to configure a timeout on http.Client?

The default `http.Client` has no timeout. If the remote server does not respond, the goroutine making the request will be blocked indefinitely, causing a goroutine leak. Always configure `Timeout` or use a context with timeout.

### 2. What is the difference between json.Marshal and json.Encoder?

`json.Marshal` converts to `[]byte` in memory. `json.Encoder` writes directly to an `io.Writer` (stream). Use Encoder for files and HTTP responses (more efficient, does not need an intermediate buffer). Use Marshal when you need the JSON as `[]byte` or string.

### 3. Why must context.Context be the first parameter?

It is a strong convention in Go. The context carries request-scoped metadata (deadlines, cancellation, values). Putting it first makes clear that the function is cancellable and facilitates the propagation pattern. It should not be stored in structs.

### 4. What happens if you don't call the cancel() function from WithTimeout/WithCancel?

A resource leak occurs. The internal resources of the context (timers, goroutines) are not released until the parent context is cancelled. `defer cancel()` immediately after creating the context is the correct practice.

### 5. What is the advantage of io.Copy over io.ReadAll + Write?

`io.Copy` processes data in streaming with an internal buffer (usually 32KB). It does not need to load everything into memory. For large files, `io.ReadAll` would cause excessive memory usage, while `io.Copy` maintains constant usage.

### 6. What is strings.Builder for and why is it better than concatenating with +?

`strings.Builder` accumulates bytes in an internal buffer and only builds the final string once. Concatenating with `+` creates a new string (immutable) on each operation, copying all the previous content, resulting in O(n^2) complexity. `strings.Builder` is O(n).

### 7. How does Go 1.22's improved routing in http.ServeMux work?

Go 1.22 added support for: HTTP methods in the pattern (`"GET /path"`), path parameters (`"/users/{id}"`), and wildcards (`"/files/{path...}"`). More specific patterns take priority. Before 1.22, you needed an external router like gorilla/mux or chi for this functionality.

### 8. What is sync.Pool and when would you use it?

`sync.Pool` is a cache of temporary objects that reduces pressure on the garbage collector by reusing objects instead of creating and discarding them. It is ideal for temporary buffers in hot paths (like JSON encoders in an HTTP server). Objects in the pool can be cleaned by the GC at any time, so it should not be used as a persistent cache.

### 9. How would you implement graceful shutdown in an HTTP server?

Use `http.Server.Shutdown(ctx)`: it stops accepting new connections, waits for active connections to finish (or the context to expire), and then closes. Typically you listen for OS signals (SIGINT/SIGTERM) to trigger the shutdown, and pass a context with timeout to avoid waiting indefinitely.
