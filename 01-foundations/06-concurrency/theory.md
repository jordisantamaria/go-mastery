# 06 - Concurrency

> "Do not communicate by sharing memory; instead, share memory by communicating." — Go Proverbs

Concurrencia es la caracteristica mas potente de Go y la mas preguntada en entrevistas. Go fue diseñado desde cero para concurrencia.

**Concurrencia vs Paralelismo**:
- **Concurrencia**: estructurar un programa como tareas independientes que se pueden ejecutar en cualquier orden
- **Paralelismo**: ejecutar multiples tareas simultaneamente en multiples CPUs

Go te da concurrencia. El runtime decide si ademas hay paralelismo (depende de los cores disponibles).

## Goroutines

Una goroutine es un **hilo ligero** gestionado por el runtime de Go (no por el OS):

```go
func sayHello(name string) {
    fmt.Printf("Hello, %s!\n", name)
}

func main() {
    go sayHello("World")  // lanza goroutine — NO espera
    go sayHello("Go")     // lanza otra

    time.Sleep(time.Millisecond) // sin esto, main termina antes de que ejecuten
}
```

### Goroutines vs OS Threads

| | Goroutine | OS Thread |
|---|---|---|
| **Memoria inicial** | ~2 KB (stack crece dinamicamente) | ~1 MB (fijo) |
| **Creacion** | ~microsegundos | ~milisegundos |
| **Scheduling** | Go runtime (user-space) | OS kernel |
| **Cantidad tipica** | Miles o millones | Cientos |
| **Context switch** | ~nanosegundos | ~microsegundos |

> Puedes lanzar **100.000 goroutines** sin problemas. Intentar lo mismo con OS threads crashearia tu sistema.

### El modelo GMP (importante para entrevista)

El scheduler de Go usa el modelo **G-M-P**:

```
G = Goroutine       (la tarea)
M = Machine/Thread  (OS thread real)
P = Processor       (contexto de ejecucion, por defecto = num CPUs)

    P0          P1          P2          P3
    |           |           |           |
    M0          M1          M2          M3
    |           |           |           |
   [G1]       [G2]       [G3]       [G4]
   [G5]       [G6]       [G7]       [G8]   <- run queues locales
    ...         ...
                    [G9, G10, G11...]      <- global run queue
```

- Cada **P** tiene una cola local de goroutines
- Cuando una cola se vacia, el P "roba" trabajo de otro P (**work stealing**)
- `GOMAXPROCS` controla cuantos P hay (default = num CPUs)

## Channels

Los channels son el mecanismo principal de comunicacion entre goroutines:

```go
// Crear un channel
ch := make(chan int)    // unbuffered
ch := make(chan int, 5) // buffered (capacidad 5)

// Enviar
ch <- 42

// Recibir
value := <-ch

// Cerrar (el sender cierra, NUNCA el receiver)
close(ch)
```

### Unbuffered channels (sincronos)

```go
ch := make(chan int) // sin buffer

go func() {
    ch <- 42  // BLOQUEA hasta que alguien reciba
}()

value := <-ch  // BLOQUEA hasta que alguien envie
fmt.Println(value) // 42
```

Un unbuffered channel **sincroniza** sender y receiver: ambos se bloquean hasta que el otro esta listo. Es como un handshake.

### Buffered channels (asincronos hasta el limite)

```go
ch := make(chan int, 3) // buffer de 3

ch <- 1  // no bloquea (1/3)
ch <- 2  // no bloquea (2/3)
ch <- 3  // no bloquea (3/3)
// ch <- 4  // BLOQUEA — buffer lleno, espera a que alguien reciba

fmt.Println(<-ch) // 1 (FIFO)
```

### Cuando usar cada uno

| Unbuffered | Buffered |
|---|---|
| Sincronizacion garantizada | Desacoplamiento sender/receiver |
| El sender sabe que el receiver recibio | El sender puede seguir sin esperar |
| Ideal para senales y handshakes | Ideal para rate limiting, batching |

### Direccionalidad (channel types)

```go
func producer(out chan<- int) {  // solo puede ENVIAR
    out <- 42
}

func consumer(in <-chan int) {   // solo puede RECIBIR
    value := <-in
}

// El compilador verifica que no uses un send-only channel para recibir
```

> Siempre especifica la direccion en las firmas de funciones. Es documentacion gratis y el compilador la verifica.

### Iterar sobre un channel (range)

```go
ch := make(chan int)

go func() {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch) // IMPORTANTE: close para que range termine
}()

for value := range ch {
    fmt.Println(value) // 0, 1, 2, 3, 4
}
// El loop termina cuando el channel se cierra y se vacia
```

### Comma-ok pattern con channels

```go
value, ok := <-ch
if !ok {
    fmt.Println("channel cerrado")
}
```

## Select

`select` es como un switch para channels. Espera a que **alguno** este listo:

```go
select {
case msg := <-ch1:
    fmt.Println("Received from ch1:", msg)
case msg := <-ch2:
    fmt.Println("Received from ch2:", msg)
case ch3 <- "hello":
    fmt.Println("Sent to ch3")
default:
    fmt.Println("No channel ready") // no bloquea
}
```

- Si multiples channels estan listos, **elige uno al azar** (no deterministico)
- Sin `default`, bloquea hasta que alguno este listo
- Con `default`, no bloquea nunca (util para polling)

### Select con timeout

```go
select {
case result := <-ch:
    fmt.Println("Got result:", result)
case <-time.After(3 * time.Second):
    fmt.Println("Timeout!")
}
```

### Select para done/quit signal

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

> `chan struct{}` es idiomatico para senales (no lleva datos, ocupa 0 bytes).

## sync.WaitGroup

Esperar a que un grupo de goroutines termine:

```go
var wg sync.WaitGroup

for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done() // decrementa al salir
        fmt.Println("Worker", id)
    }(i)
}

wg.Wait() // bloquea hasta que el contador llega a 0
```

- `Add(n)` — incrementa el contador
- `Done()` — decrementa (equivalente a `Add(-1)`)
- `Wait()` — bloquea hasta que el contador sea 0

> **Regla**: llama `Add` ANTES de lanzar la goroutine, no dentro. Si llamas Add dentro de la goroutine, hay una race condition con Wait.

## sync.Mutex y sync.RWMutex

Para proteger datos compartidos:

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

### RWMutex — multiples readers, un solo writer

```go
type SafeCache struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *SafeCache) Get(key string) (string, bool) {
    c.mu.RLock()         // multiples goroutines pueden leer a la vez
    defer c.mu.RUnlock()
    val, ok := c.data[key]
    return val, ok
}

func (c *SafeCache) Set(key, value string) {
    c.mu.Lock()          // exclusivo — bloquea readers y writers
    defer c.mu.Unlock()
    c.data[key] = value
}
```

> Usa `RWMutex` cuando tienes **muchas mas lecturas que escrituras**. Para el resto, `Mutex` normal.

### Mutex vs Channels — cuando usar cada uno

| Mutex | Channel |
|---|---|
| Proteger un dato compartido | Comunicar entre goroutines |
| Cache, contadores, mapas | Pipelines, senales, resultados |
| "Guarda esto" | "Pasa esto" |

> **Regla practica**: si la operacion es "compartir estado", usa Mutex. Si es "pasar datos entre goroutines", usa channels.

## sync.Once

Ejecutar algo **exactamente una vez** (thread-safe):

```go
var once sync.Once
var instance *Database

func GetDB() *Database {
    once.Do(func() {
        instance = connectToDatabase() // solo se ejecuta 1 vez
    })
    return instance
}
```

Util para singletons, inicializacion lazy, y setup one-time.

## context.Context

Context propaga **deadlines, cancellation, y valores** a traves de goroutines:

```go
// Crear con timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel() // SIEMPRE llamar cancel para liberar recursos

// Crear con cancelacion manual
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Verificar cancelacion
select {
case <-ctx.Done():
    fmt.Println("Cancelled:", ctx.Err())
case result := <-doWork(ctx):
    fmt.Println("Result:", result)
}
```

### Context en funciones (patron estandar)

```go
// Context SIEMPRE es el primer parametro
func fetchData(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    // Si el context se cancela, el request se aborta automaticamente
    resp, err := http.DefaultClient.Do(req)
    // ...
}
```

> **Regla**: Context va siempre como **primer parametro**, nunca en un struct. Esto es una convencion fuerte en Go.

### Context values (usar con moderacion)

```go
type contextKey string

const userIDKey contextKey = "userID"

// Guardar
ctx := context.WithValue(parentCtx, userIDKey, "user-123")

// Leer
if userID, ok := ctx.Value(userIDKey).(string); ok {
    fmt.Println("User:", userID)
}
```

> Usa context values solo para datos **request-scoped** (user ID, trace ID, etc). Nunca para pasar dependencias o config.

## Race conditions

Una race condition ocurre cuando multiples goroutines acceden a datos compartidos sin sincronizacion:

```go
// BUG: race condition
counter := 0
for i := 0; i < 1000; i++ {
    go func() {
        counter++ // lectura + escritura NO atomica
    }()
}
// counter puede ser cualquier valor < 1000
```

### Detectar con -race

```bash
go test -race ./...
go run -race main.go
```

El race detector es **imprescindible** durante desarrollo. Detecta accesos concurrentes sin proteccion.

### Soluciones a race conditions

```go
// Solucion 1: Mutex
var mu sync.Mutex
mu.Lock()
counter++
mu.Unlock()

// Solucion 2: Atomic (mas rapido para operaciones simples)
var counter int64
atomic.AddInt64(&counter, 1)

// Solucion 3: Channel (enviar updates a una goroutine controladora)
ch := make(chan int)
go func() {
    count := 0
    for delta := range ch {
        count += delta
    }
}()
ch <- 1
```

## Patrones de concurrencia

### Fan-out / Fan-in

```go
// Fan-out: distribuir trabajo entre N workers
func fanOut(jobs <-chan int, numWorkers int) []<-chan int {
    workers := make([]<-chan int, numWorkers)
    for i := 0; i < numWorkers; i++ {
        workers[i] = worker(jobs)
    }
    return workers
}

// Fan-in: combinar resultados de N channels en uno
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
                results <- process(job) // cada worker consume del mismo channel
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
// Cada stage es una funcion que lee de un channel y escribe a otro
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

// Componer: generate -> square -> filter
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

    <-ctx.Done() // espera Ctrl+C
    fmt.Println("Shutting down...")

    // Dar tiempo para cleanup
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    server.Shutdown(shutdownCtx)
}
```

## Deadlocks

Un deadlock ocurre cuando goroutines se bloquean mutuamente esperando una a la otra:

```go
// Deadlock clasico: unbuffered channel sin receiver
ch := make(chan int)
ch <- 1 // bloquea para siempre — nadie recibe
// fatal error: all goroutines are asleep - deadlock!

// Deadlock por lock ordering
// Goroutine 1: Lock(A), Lock(B)
// Goroutine 2: Lock(B), Lock(A)
// -> Se bloquean mutuamente
```

Solucion: siempre adquirir locks en el **mismo orden** en todas las goroutines.

## errgroup (golang.org/x/sync/errgroup)

Ejecutar goroutines que pueden fallar y recoger el primer error:

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
    // err es el primer error que ocurrio
    // ctx se cancela automaticamente cuando hay error
    log.Fatal(err)
}
```

## Preguntas de entrevista frecuentes

1. **Que es una goroutine y como se diferencia de un thread?**
   Una goroutine es un hilo ligero gestionado por el runtime de Go (~2KB stack, scheduling en user-space). Un OS thread ocupa ~1MB y es gestionado por el kernel. Puedes tener millones de goroutines pero solo cientos de threads.

2. **Explica el modelo GMP del scheduler de Go.**
   G=Goroutine, M=Machine (OS thread), P=Processor (contexto). Cada P tiene una cola local de Gs. Los Ms ejecutan Gs a traves de Ps. Cuando una cola se vacia, un P puede "robar" trabajo de otro (work stealing). GOMAXPROCS controla cuantos Ps hay.

3. **Diferencia entre buffered y unbuffered channel?**
   Unbuffered: sender y receiver se bloquean hasta que el otro esta listo (sincronizacion). Buffered: sender solo bloquea si el buffer esta lleno, receiver solo bloquea si esta vacio (desacoplamiento).

4. **Cuando usarias Mutex vs Channel?**
   Mutex: para proteger datos compartidos (caches, contadores, mapas). Channel: para comunicar datos entre goroutines (pipelines, resultados, senales). "Compartir estado" -> Mutex. "Pasar datos" -> Channel.

5. **Que es una race condition y como la detectas?**
   Acceso concurrente a datos compartidos sin sincronizacion. Se detecta con `go test -race` o `go run -race`. Soluciones: Mutex, atomic operations, o channels.

6. **Que pasa si escribes a un channel cerrado?**
   **Panic**. Leer de un channel cerrado devuelve el zero value inmediatamente. Por eso, **solo el sender debe cerrar** el channel.

7. **Que es context.Context y para que se usa?**
   Propaga deadlines, cancelacion, y valores request-scoped a traves del arbol de goroutines. Siempre es el primer parametro. Se usa para timeouts en HTTP, cancelar operaciones largas, y pasar datos como trace IDs.

8. **Explica el patron fan-out/fan-in.**
   Fan-out: distribuir trabajo de un channel entre N workers (goroutines). Fan-in: combinar los resultados de N channels en un solo channel. Permite paralelizar CPU-bound o I/O-bound work.

9. **Como harias un graceful shutdown?**
   Capturar la senal OS (SIGINT/SIGTERM) con signal.NotifyContext, propagar la cancelacion via context, dar tiempo a las goroutines para terminar con un deadline, y cerrar recursos (DB, HTTP server) ordenadamente.

10. **Que es un goroutine leak y como se previene?**
    Una goroutine que se bloquea para siempre (esperando un channel que nadie cierra, un lock que nadie libera). Se previene con context cancellation, timeouts, y asegurandose de que todo channel tenga un close path.
