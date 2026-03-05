# 09 - Standard Library Deep Dive

La standard library de Go es una de sus mayores fortalezas. A diferencia de otros lenguajes donde necesitas dependencias externas para tareas basicas, Go incluye paquetes robustos y bien disenados para HTTP, I/O, JSON, concurrencia, logging, y mas. En este modulo exploramos los paquetes mas importantes en profundidad.

---

## net/http

El paquete `net/http` proporciona un servidor y cliente HTTP completo, listo para produccion.

### Servidor basico

```go
http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hola, %s!", r.URL.Query().Get("name"))
})

log.Fatal(http.ListenAndServe(":8080", nil))
```

- `http.HandleFunc` registra un handler en el `DefaultServeMux`
- `http.ListenAndServe` inicia el servidor. Usa `log.Fatal` porque si falla, es critico
- `w http.ResponseWriter` — escribe la respuesta
- `r *http.Request` — contiene toda la info del request (headers, body, URL, method...)

### http.ServeMux con routing mejorado (Go 1.22+)

Go 1.22 introdujo **pattern matching mejorado** en `http.ServeMux`:

```go
mux := http.NewServeMux()

// Metodo + path
mux.HandleFunc("GET /api/users", listUsers)
mux.HandleFunc("POST /api/users", createUser)

// Path parameters con {name}
mux.HandleFunc("GET /api/users/{id}", getUser)
mux.HandleFunc("DELETE /api/users/{id}", deleteUser)

// Acceder al path parameter
func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")  // nuevo en Go 1.22
    fmt.Fprintf(w, "User ID: %s", id)
}
```

**Reglas de matching Go 1.22+:**
- `"GET /path"` — solo acepta GET requests a `/path`
- `"/path"` — acepta cualquier metodo
- `"/path/{param}"` — captura el segmento como path parameter
- `"/path/{param...}"` — captura el resto del path (wildcard)
- Patrones mas especificos tienen prioridad sobre los generales

### La interfaz Handler

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Cualquier tipo que implemente `ServeHTTP` es un Handler. `http.HandlerFunc` es un adaptador:

```go
// HandlerFunc convierte una funcion en un Handler
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
    f(w, r)
}
```

### Patron Middleware

Un middleware es una funcion que envuelve un Handler para agregar funcionalidad:

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)  // llamar al handler real
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

// Encadenar middlewares
handler := recoveryMiddleware(loggingMiddleware(mux))
http.ListenAndServe(":8080", handler)
```

### http.Client con timeout

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

> **Importante**: SIEMPRE configura un timeout en `http.Client`. El cliente por defecto no tiene timeout, lo que puede causar goroutine leaks.

### Graceful shutdown

```go
srv := &http.Server{Addr: ":8080", Handler: mux}

// Correr servidor en goroutine
go func() {
    if err := srv.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatalf("server error: %v", err)
    }
}()

// Esperar signal para shutdown
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

// Graceful shutdown con timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
srv.Shutdown(ctx)
```

---

## io

El paquete `io` define las interfaces fundamentales de I/O en Go. Todo el ecosistema se construye sobre `io.Reader` y `io.Writer`.

### Reader y Writer

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}
```

- `Read` llena el buffer `p` con datos. Devuelve `n` bytes leidos y `io.EOF` al terminar.
- `Write` escribe `p` al destino. Devuelve `n` bytes escritos.

Implementan `Reader`: `os.File`, `strings.Reader`, `bytes.Buffer`, `http.Response.Body`, `net.Conn`...

Implementan `Writer`: `os.File`, `bytes.Buffer`, `http.ResponseWriter`, `os.Stdout`...

### io.Copy

```go
// Copia de src a dst, retorna bytes copiados
n, err := io.Copy(dst, src)

// Ejemplos practicos
io.Copy(os.Stdout, resp.Body)          // imprimir response al terminal
io.Copy(file, resp.Body)               // descargar archivo
io.Copy(hash, file)                    // calcular hash de archivo
```

### io.MultiReader

Concatena multiples Readers en uno solo:

```go
header := strings.NewReader("--- HEADER ---\n")
body := strings.NewReader("contenido del body\n")
footer := strings.NewReader("--- FOOTER ---\n")

combined := io.MultiReader(header, body, footer)
io.Copy(os.Stdout, combined)
// Output:
// --- HEADER ---
// contenido del body
// --- FOOTER ---
```

### io.TeeReader

Lee de un Reader y simultaneamente escribe a un Writer (como el comando `tee` de Unix):

```go
var buf bytes.Buffer
tee := io.TeeReader(resp.Body, &buf)

// Al leer de tee, tambien se escribe a buf
io.Copy(os.Stdout, tee)  // imprime al terminal
// buf ahora contiene una copia del body
```

Util para: loggear el body de un request mientras lo procesas, calcular un hash mientras copias un archivo, etc.

### Composicion de interfaces

```go
// ReadWriter combina Reader y Writer
type ReadWriter interface {
    Reader
    Writer
}

// ReadCloser es un Reader que necesita cerrarse (ej: resp.Body)
type ReadCloser interface {
    Reader
    Closer
}

// io.ReadAll lee todo el contenido de un Reader
data, err := io.ReadAll(reader)
```

---

## encoding/json

El paquete `encoding/json` es el mas usado para serializar/deserializar datos en Go.

### Marshal y Unmarshal

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
    Host     string `json:"host"`              // renombrar campo
    Port     int    `json:"port,omitempty"`     // omitir si es zero value
    Debug    bool   `json:"debug"`
    Internal string `json:"-"`                  // ignorar siempre
    Password string `json:"password,omitempty"` // omitir si vacio
}

// omitempty omite el campo si es: 0, "", false, nil, slice/map vacio
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

### json.Encoder y json.Decoder

Para trabajar con streams (archivos, HTTP bodies) en lugar de []byte:

```go
// Encoder: escribe JSON a un Writer
file, _ := os.Create("data.json")
encoder := json.NewEncoder(file)
encoder.SetIndent("", "  ")
encoder.Encode(user)  // escribe al archivo

// Decoder: lee JSON de un Reader
file, _ := os.Open("data.json")
decoder := json.NewDecoder(file)
var user User
decoder.Decode(&user)
```

> **Encoder/Decoder vs Marshal/Unmarshal**: Usa Encoder/Decoder para streams (archivos, HTTP). Usa Marshal/Unmarshal para []byte en memoria.

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

Retrasa el parsing de parte del JSON:

```go
type Event struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"` // no se parsea todavia
}

var event Event
json.Unmarshal(data, &event)

// Ahora parseamos el payload segun el tipo
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

El paquete `context` permite propagar cancelaciones, deadlines, y valores a traves de una cadena de llamadas. Es fundamental en servidores HTTP y operaciones concurrentes.

### Creacion de contexts

```go
// Context raiz — nunca se cancela
ctx := context.Background()

// Context que puede cancelarse manualmente
ctx, cancel := context.WithCancel(parent)
defer cancel()  // SIEMPRE defer cancel para evitar leaks

// Context con timeout (se cancela automaticamente)
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()

// Context con deadline absoluto
deadline := time.Now().Add(5 * time.Second)
ctx, cancel := context.WithDeadline(parent, deadline)
defer cancel()

// Context con un valor (usar con moderacion)
ctx := context.WithValue(parent, "requestID", "abc-123")
```

### Propagacion en HTTP

```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()  // obtener context del request

    // Si el cliente cierra la conexion, ctx se cancela automaticamente

    result, err := fetchFromDB(ctx)  // propagar context
    if err != nil {
        if ctx.Err() == context.Canceled {
            return  // el cliente ya no esta, no vale la pena responder
        }
        http.Error(w, err.Error(), 500)
        return
    }
    json.NewEncoder(w).Encode(result)
}

func fetchFromDB(ctx context.Context) (Data, error) {
    // Usar ctx para cancelar la query si el context expira
    select {
    case <-ctx.Done():
        return Data{}, ctx.Err()
    case result := <-doQuery():
        return result, nil
    }
}
```

### Buenas practicas con context

1. **`context.Context` siempre como primer parametro**: `func DoSomething(ctx context.Context, ...)`
2. **No guardar context en structs**: pasarlo como parametro
3. **SIEMPRE `defer cancel()`** despues de `WithCancel/WithTimeout/WithDeadline`
4. **`context.WithValue` solo para request-scoped data**: request ID, auth token, trace ID. NUNCA para pasar parametros de funcion
5. **Las keys de `WithValue` deben ser tipos privados** para evitar colisiones:
   ```go
   type contextKey string
   const requestIDKey contextKey = "requestID"
   ctx = context.WithValue(ctx, requestIDKey, "abc-123")
   ```

---

## os / filepath

### Lectura y escritura de archivos

```go
// Leer archivo completo
data, err := os.ReadFile("config.json")

// Escribir archivo completo
err := os.WriteFile("output.txt", []byte("hello"), 0644)

// Abrir archivo para leer
file, err := os.Open("data.txt")
defer file.Close()

// Crear archivo para escribir
file, err := os.Create("output.txt")
defer file.Close()
file.WriteString("hello\n")

// Abrir con opciones especificas
file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
```

### os.Args y variables de entorno

```go
// Argumentos de la linea de comandos
fmt.Println(os.Args[0])  // nombre del programa
fmt.Println(os.Args[1:]) // argumentos

// Variables de entorno
home := os.Getenv("HOME")
port, exists := os.LookupEnv("PORT")
os.Setenv("MY_VAR", "value")
```

### filepath.Walk / filepath.WalkDir

```go
// WalkDir es mas eficiente (Go 1.16+, no hace Stat en cada entrada)
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

### Operaciones de directorio

```go
// Crear directorio (y padres)
os.MkdirAll("path/to/dir", 0755)

// Archivos temporales
tmpFile, err := os.CreateTemp("", "prefix-*.txt")
defer os.Remove(tmpFile.Name())  // limpiar al terminar

// Directorio temporal
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

El paquete `flag` parsea argumentos de la linea de comandos.

### Uso basico

```go
// Definir flags
host := flag.String("host", "localhost", "server host")
port := flag.Int("port", 8080, "server port")
verbose := flag.Bool("verbose", false, "enable verbose logging")

// Parsear (obligatorio)
flag.Parse()

fmt.Printf("host=%s port=%d verbose=%t\n", *host, *port, *verbose)
fmt.Println("args restantes:", flag.Args())
```

### flag.StringVar (sin puntero)

```go
var config struct {
    Host string
    Port int
}

flag.StringVar(&config.Host, "host", "localhost", "server host")
flag.IntVar(&config.Port, "port", 8080, "server port")
flag.Parse()
```

### Patron subcommands

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

`slog` es el paquete de structured logging oficial. Reemplaza al paquete `log` para aplicaciones modernas.

### Uso basico

```go
slog.Info("servidor iniciado", "port", 8080)
// 2024/01/15 10:30:00 INFO servidor iniciado port=8080

slog.Warn("disco casi lleno", "usage", 95.5, "path", "/data")
slog.Error("fallo conexion", "err", err, "host", "db.example.com")
```

### Niveles

```go
slog.Debug("mensaje debug")  // no se muestra por defecto
slog.Info("mensaje info")
slog.Warn("mensaje warn")
slog.Error("mensaje error")
```

### Handlers: JSON y Text

```go
// Text handler (por defecto, formato key=value)
textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,  // mostrar todos los niveles
})

// JSON handler (para produccion / log aggregation)
jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
})

// Establecer como logger por defecto
logger := slog.New(jsonHandler)
slog.SetDefault(logger)

// Ahora slog.Info produce JSON:
// {"time":"2024-01-15T10:30:00Z","level":"INFO","msg":"request","method":"GET"}
```

### With (logger con campos fijos)

```go
// Crear logger con campos que se incluyen en cada mensaje
reqLogger := slog.With("requestID", "abc-123", "userID", 42)

reqLogger.Info("procesando request")
// INFO procesando request requestID=abc-123 userID=42

reqLogger.Error("fallo", "err", err)
// ERROR fallo requestID=abc-123 userID=42 err=...
```

### Group (agrupar campos)

```go
slog.Info("request completado",
    slog.Group("request",
        slog.String("method", "GET"),
        slog.String("path", "/api/users"),
    ),
    slog.Group("response",
        slog.Int("status", 200),
        slog.Duration("latency", 42*time.Millisecond),
    ),
)
// Con JSON handler:
// {"msg":"request completado","request":{"method":"GET","path":"/api/users"},"response":{"status":200,"latency":"42ms"}}
```

---

## time

### Duration

```go
d := 5 * time.Second
d = 100 * time.Millisecond
d = 2*time.Hour + 30*time.Minute

// Convertir
d.Seconds()      // float64
d.Milliseconds() // int64
d.String()        // "2h30m0s"

// Parsear duration de string
d, err := time.ParseDuration("1h30m")
```

### Timers y Tickers

```go
// Timer: se dispara una vez
timer := time.NewTimer(2 * time.Second)
<-timer.C  // esperar
// o cancelar antes: timer.Stop()

// time.After: shortcut para timer de un solo uso
select {
case result := <-ch:
    fmt.Println(result)
case <-time.After(5 * time.Second):
    fmt.Println("timeout")
}

// Ticker: se dispara repetidamente
ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()  // SIEMPRE stop para evitar leaks
for t := range ticker.C {
    fmt.Println("tick at", t)
}
```

### Format y Parse

Go usa un **reference time** unico: `Mon Jan 2 15:04:05 MST 2006` (1/2 3:04:05 PM 2006).

```go
now := time.Now()

// Formatear
now.Format("2006-01-02")                    // "2024-01-15"
now.Format("2006-01-02 15:04:05")           // "2024-01-15 10:30:00"
now.Format(time.RFC3339)                    // "2024-01-15T10:30:00Z"
now.Format("02/Jan/2006 03:04 PM")          // "15/Jan/2024 10:30 AM"

// Parsear
t, err := time.Parse("2006-01-02", "2024-01-15")
t, err := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")

// Parsear con timezone
t, err := time.ParseInLocation("2006-01-02 15:04", "2024-01-15 10:30",
    time.FixedZone("CET", 1*60*60))
```

### Zonas horarias

```go
loc, err := time.LoadLocation("Europe/Madrid")
madridTime := now.In(loc)

utc := now.UTC()

// Comparar tiempos
t1.Before(t2)
t1.After(t2)
t1.Equal(t2)
t1.Sub(t2)   // devuelve Duration
t1.Add(d)    // suma Duration
```

---

## strings / strconv / fmt

### strings.Builder

Eficiente para construir strings iterativamente (evita allocations):

```go
var b strings.Builder
for i := 0; i < 1000; i++ {
    fmt.Fprintf(&b, "item %d, ", i)
}
result := b.String()
```

> Concatenar con `+` en un loop crea un nuevo string cada vez (O(n^2)). `strings.Builder` es O(n).

### strings.Cut (Go 1.18+)

Divide un string por el primer separador encontrado:

```go
before, after, found := strings.Cut("host:8080", ":")
// before="host", after="8080", found=true

before, after, found = strings.Cut("noport", ":")
// before="noport", after="", found=false

// Mas limpio que strings.Split cuando solo necesitas 2 partes
// Antes: parts := strings.SplitN(s, ":", 2)
// Ahora: host, port, _ := strings.Cut(s, ":")
```

### Funciones comunes de strings

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
strings.Map(unicode.ToUpper, "hello")      // "HELLO" (rune por rune)
```

### strconv

Conversiones entre strings y tipos basicos:

```go
// String -> int
n, err := strconv.Atoi("42")         // 42, nil
n, err := strconv.Atoi("abc")        // 0, error

// Int -> string
s := strconv.Itoa(42)                // "42"

// Parsing con mas control
i, err := strconv.ParseInt("FF", 16, 64)  // 255 (hex)
f, err := strconv.ParseFloat("3.14", 64)  // 3.14
b, err := strconv.ParseBool("true")       // true

// Formatting con mas control
s := strconv.FormatFloat(3.14159, 'f', 2, 64)  // "3.14"
s := strconv.FormatInt(255, 16)                 // "ff"
```

### fmt verbs

```go
type Point struct{ X, Y int }
p := Point{1, 2}

fmt.Printf("%v\n", p)    // {1 2}            — formato por defecto
fmt.Printf("%+v\n", p)   // {X:1 Y:2}        — con nombres de campos
fmt.Printf("%#v\n", p)   // main.Point{X:1, Y:2} — sintaxis Go completa
fmt.Printf("%T\n", p)    // main.Point        — tipo

fmt.Sprintf("%d", 42)     // "42"              — entero decimal
fmt.Sprintf("%x", 255)    // "ff"              — hexadecimal
fmt.Sprintf("%b", 8)      // "1000"            — binario
fmt.Sprintf("%f", 3.14)   // "3.140000"        — float
fmt.Sprintf("%.2f", 3.14) // "3.14"            — 2 decimales
fmt.Sprintf("%q", "hi")   // "\"hi\""          — string con comillas
fmt.Sprintf("%p", &p)     // "0xc0000b4000"    — puntero
```

---

## sync (repaso + Pool)

> Las primitivas basicas (`WaitGroup`, `Mutex`, `RWMutex`, `Once`) se cubrieron en el modulo 06. Aqui hacemos un repaso rapido y profundizamos en `sync.Pool`.

### Repaso rapido

```go
// WaitGroup — esperar N goroutines
var wg sync.WaitGroup
wg.Add(1)
go func() { defer wg.Done(); /* trabajo */ }()
wg.Wait()

// Mutex — exclusion mutua
var mu sync.Mutex
mu.Lock()
// seccion critica
mu.Unlock()

// RWMutex — multiples lectores, un escritor
var rw sync.RWMutex
rw.RLock()   // leer
rw.RUnlock()
rw.Lock()    // escribir
rw.Unlock()

// Once — ejecutar exactamente una vez
var once sync.Once
once.Do(func() { /* inicializacion */ })

// Map — map concurrente (sin Mutex externo)
var m sync.Map
m.Store("key", "value")
v, ok := m.Load("key")
m.Range(func(k, v any) bool { return true })
```

### sync.Pool — reuso de objetos

`sync.Pool` reutiliza objetos temporales para reducir presion sobre el garbage collector:

```go
var bufPool = sync.Pool{
    New: func() any {
        return new(bytes.Buffer)
    },
}

func processRequest(data []byte) string {
    buf := bufPool.Get().(*bytes.Buffer)
    buf.Reset()  // IMPORTANTE: limpiar antes de usar
    defer bufPool.Put(buf)  // devolver al pool

    buf.Write(data)
    buf.WriteString(" processed")
    return buf.String()
}
```

**Cuando usar `sync.Pool`:**
- Objetos costosos de crear que se usan y descartan frecuentemente
- Buffers temporales en hot paths
- Encoders/decoders reutilizables

**Cuando NO usar `sync.Pool`:**
- Para objetos que necesitan persistir (el GC puede limpiar el pool en cualquier momento)
- Si la creacion del objeto es barata
- Como cache general (usa `sync.Map` o un cache real)

```go
// Ejemplo practico: pool de JSON encoders
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

## Preguntas de entrevista frecuentes

### 1. Por que es importante configurar un timeout en http.Client?

El `http.Client` por defecto no tiene timeout. Si el servidor remoto no responde, la goroutine que hace el request se quedara bloqueada indefinidamente, causando un goroutine leak. Siempre configura `Timeout` o usa un context con timeout.

### 2. Cual es la diferencia entre json.Marshal y json.Encoder?

`json.Marshal` convierte a `[]byte` en memoria. `json.Encoder` escribe directamente a un `io.Writer` (stream). Usa Encoder para archivos y HTTP responses (mas eficiente, no necesita buffer intermedio). Usa Marshal cuando necesitas el JSON como `[]byte` o string.

### 3. Por que context.Context debe ser el primer parametro?

Es una convencion fuerte en Go. El context lleva metadata request-scoped (deadlines, cancelacion, valores). Ponerlo primero hace claro que la funcion es cancelable y facilita el patron de propagacion. No se debe guardar en structs.

### 4. Que pasa si no llamas a la funcion cancel() de WithTimeout/WithCancel?

Se produce un resource leak. Los recursos internos del context (timers, goroutines) no se liberan hasta que el context padre se cancela. `defer cancel()` inmediatamente despues de crear el context es la practica correcta.

### 5. Cual es la ventaja de io.Copy sobre io.ReadAll + Write?

`io.Copy` procesa los datos en streaming con un buffer interno (normalmente 32KB). No necesita cargar todo en memoria. Para archivos grandes, `io.ReadAll` causaria un uso excesivo de memoria, mientras que `io.Copy` mantiene uso constante.

### 6. Para que sirve strings.Builder y por que es mejor que concatenar con +?

`strings.Builder` acumula bytes en un buffer interno y solo construye el string final una vez. Concatenar con `+` crea un nuevo string (immutable) en cada operacion, copiando todo el contenido anterior, resultando en complejidad O(n^2). `strings.Builder` es O(n).

### 7. Como funciona el routing mejorado de Go 1.22 en http.ServeMux?

Go 1.22 anadio soporte para: metodos HTTP en el patron (`"GET /path"`), path parameters (`"/users/{id}"`), y wildcards (`"/files/{path...}"`). Los patrones mas especificos tienen prioridad. Antes de 1.22, necesitabas un router externo como gorilla/mux o chi para esta funcionalidad.

### 8. Que es sync.Pool y cuando lo usarias?

`sync.Pool` es un cache de objetos temporales que reduce la presion sobre el garbage collector reutilizando objetos en lugar de crearlos y descartarlos. Es ideal para buffers temporales en hot paths (como encoders JSON en un servidor HTTP). Los objetos en el pool pueden ser limpiados por el GC en cualquier momento, asi que no se debe usar como cache persistente.

### 9. Como implementarias graceful shutdown en un servidor HTTP?

Usa `http.Server.Shutdown(ctx)`: para de aceptar nuevas conexiones, espera a que las conexiones activas terminen (o el context expire), y luego cierra. Tipicamente escuchas senales del OS (SIGINT/SIGTERM) para disparar el shutdown, y pasas un context con timeout para no esperar indefinidamente.
