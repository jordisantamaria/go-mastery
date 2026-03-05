# Concurrency Puzzles — Trampas y Patrones

Ejercicios practicos para dominar la concurrencia en Go. Cada puzzle presenta codigo con un bug de concurrencia. Tu tarea es identificar el problema, explicarlo, y proponer la solucion correcta.

---

## Tabla de Contenidos

1. [Trampas Comunes de Concurrencia](#trampas-comunes-de-concurrencia)
2. [Puzzle 1: Race Condition Detection](#puzzle-1-race-condition-detection)
3. [Puzzle 2: Deadlock](#puzzle-2-deadlock)
4. [Puzzle 3: Goroutine Leak](#puzzle-3-goroutine-leak)
5. [Puzzle 4: Channel Direction Bug](#puzzle-4-channel-direction-bug)
6. [Puzzle 5: Closure in Loop](#puzzle-5-closure-in-loop)
7. [Puzzle 6: Select Priority](#puzzle-6-select-priority)
8. [Puzzle 7: Context Cancellation](#puzzle-7-context-cancellation)
9. [Puzzle 8: WaitGroup Misuse](#puzzle-8-waitgroup-misuse)

---

## Trampas Comunes de Concurrencia

Antes de los puzzles, estas son las trampas mas frecuentes que causan bugs de concurrencia en Go:

### 1. Data Races
Acceder a memoria compartida sin sincronizacion. Go tiene el race detector (`go run -race`) que detecta esto en runtime, pero solo cubre las rutas de codigo que realmente se ejecutan.

### 2. Goroutine Leaks
Goroutines que quedan bloqueadas para siempre esperando en un canal o lock que nunca se libera. Son el equivalente a memory leaks pero para goroutines. Herramientas como `goleak` ayudan a detectarlas en tests.

### 3. Deadlocks
Dos o mas goroutines esperandose mutuamente, creando un ciclo de dependencias. El runtime de Go detecta cuando **todas** las goroutines estan bloqueadas (`fatal error: all goroutines are asleep`), pero no detecta deadlocks parciales.

### 4. Starvation
Una goroutine nunca obtiene acceso a un recurso porque otras goroutines lo acaparan. Comun con `sync.Mutex` bajo alta contencion.

### 5. Premature Closure
Cerrar un canal o cancelar un context antes de que todos los consumidores hayan terminado.

---

## Puzzle 1: Race Condition Detection

### Codigo con Bug

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
            counter++ // DATA RACE: acceso concurrente sin sincronizacion
        }()
    }

    wg.Wait()
    fmt.Println("Counter:", counter) // Resultado impredecible, casi nunca 1000
}
```

### Pregunta
Que esta mal en este codigo? Como lo arreglarias? Proporciona al menos tres formas diferentes.

### Explicacion del Problema
1000 goroutines acceden a la variable `counter` simultaneamente sin sincronizacion. `counter++` no es atomico — implica leer, incrementar y escribir. Dos goroutines pueden leer el mismo valor, incrementarlo, y escribir el mismo resultado, perdiendo un incremento.

Ejecutar con `go run -race main.go` reporta el data race.

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
    fmt.Println("Counter:", counter) // Siempre 1000
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

    // Una sola goroutine gestiona el counter (confinamiento)
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
    fmt.Println("Counter:", counter) // Siempre 1000
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
            counter.Add(1) // Operacion atomica, sin lock
        }()
    }

    wg.Wait()
    fmt.Println("Counter:", counter.Load()) // Siempre 1000
}
```

### Takeaway
Para contadores simples, `sync/atomic` es la opcion mas eficiente. Para logica mas compleja, `sync.Mutex`. El patron de canal (confinamiento) es preferible cuando se puede disenar el flujo de datos sin estado compartido.

---

## Puzzle 2: Deadlock

### Codigo con Bug

```go
package main

import "fmt"

func main() {
    ch1 := make(chan int)
    ch2 := make(chan int)

    // Goroutine A: envia a ch1, luego recibe de ch2
    go func() {
        ch1 <- 1     // bloquea esperando receptor en ch1
        val := <-ch2  // nunca llega aqui
        fmt.Println("A recibio:", val)
    }()

    // Goroutine B: envia a ch2, luego recibe de ch1
    go func() {
        ch2 <- 2     // bloquea esperando receptor en ch2
        val := <-ch1  // nunca llega aqui
        fmt.Println("B recibio:", val)
    }()

    // Main espera (sin select o sincronizacion)
    select {}
}
```

### Pregunta
Por que este codigo genera un deadlock? Como lo arreglarias?

### Explicacion del Problema
Ambas goroutines intentan **enviar** en canales unbuffered antes de **recibir**:
- Goroutine A bloquea en `ch1 <- 1` esperando que alguien reciba de ch1.
- Goroutine B bloquea en `ch2 <- 2` esperando que alguien reciba de ch2.
- Nadie recibe de ninguno de los dos canales. Deadlock circular.

En este caso el runtime detecta que todas las goroutines estan dormidas y lanza: `fatal error: all goroutines are asleep - deadlock!`

### Codigo Corregido

```go
package main

import "fmt"

func main() {
    ch1 := make(chan int)
    ch2 := make(chan int)

    // Goroutine A: recibe de ch2, luego envia a ch1
    go func() {
        val := <-ch2 // primero recibe
        fmt.Println("A recibio:", val)
        ch1 <- 1 // luego envia
    }()

    // Goroutine B: envia a ch2, luego recibe de ch1
    go func() {
        ch2 <- 2     // envia a ch2 (A lo recibe)
        val := <-ch1  // recibe de ch1 (A lo envia)
        fmt.Println("B recibio:", val)
    }()

    // Necesitamos esperar a que terminen
    // Una forma simple: usar un done channel
    done := make(chan struct{}, 2)
    // (En un codigo real usariamos sync.WaitGroup, aqui simplificamos)

    ch1Bis := make(chan int)
    ch2Bis := make(chan int)

    go func() {
        val := <-ch2Bis
        fmt.Println("A recibio:", val)
        ch1Bis <- 1
        done <- struct{}{}
    }()

    go func() {
        ch2Bis <- 2
        val := <-ch1Bis
        fmt.Println("B recibio:", val)
        done <- struct{}{}
    }()

    <-done
    <-done
}
```

**Alternativa mas limpia con select:**

```go
package main

import "fmt"

func main() {
    ch1 := make(chan int, 1) // buffered — el send no bloquea
    ch2 := make(chan int, 1) // buffered — el send no bloquea

    done := make(chan struct{})

    go func() {
        ch1 <- 1
        val := <-ch2
        fmt.Println("A recibio:", val)
        done <- struct{}{}
    }()

    go func() {
        ch2 <- 2
        val := <-ch1
        fmt.Println("B recibio:", val)
        done <- struct{}{}
    }()

    <-done
    <-done
}
```

### Takeaway
Para evitar deadlocks: (1) evitar dependencias circulares entre canales, (2) usar canales con buffer cuando es apropiado, (3) establecer un orden consistente de acquire/release, (4) usar `select` con `default` o timeout como escape.

---

## Puzzle 3: Goroutine Leak

### Codigo con Bug

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
            ch <- u // Las goroutines perdedoras quedan bloqueadas aqui PARA SIEMPRE
        }(url)
    }

    return <-ch // Solo recibe el primero, las demas goroutines quedan leakeadas
}

func main() {
    result := fetchFirst([]string{
        "https://httpbin.org/delay/1",
        "https://httpbin.org/delay/2",
        "https://httpbin.org/delay/3",
    })
    fmt.Println("Primer resultado:", result)
    time.Sleep(5 * time.Second) // Las goroutines leakeadas siguen vivas
}
```

### Pregunta
Donde esta el leak de goroutines? Como lo prevenirias?

### Explicacion del Problema
Se lanzan N goroutines, pero solo se consume **un** resultado del canal `ch` (que es unbuffered). Las N-1 goroutines restantes quedan bloqueadas permanentemente en `ch <- u` porque nadie mas recibe del canal. Estas goroutines nunca terminan y su memoria nunca se libera.

### Codigo Corregido

```go
package main

import (
    "context"
    "fmt"
    "net/http"
)

func fetchFirst(ctx context.Context, urls []string) string {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel() // Cancela todas las goroutines restantes

    ch := make(chan string, len(urls)) // Buffer para todas las goroutines

    for _, url := range urls {
        go func(u string) {
            req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
            if err != nil {
                return
            }
            resp, err := http.DefaultClient.Do(req)
            if err != nil {
                return // Context cancelado o error de red
            }
            defer resp.Body.Close()

            select {
            case ch <- u:
            case <-ctx.Done():
                return // No bloquear si ya tenemos resultado
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
    fmt.Println("Primer resultado:", result)
    // Todas las goroutines terminaran gracias a context cancellation
}
```

**Puntos clave del fix:**
1. Canal con buffer (`len(urls)`) para que los sends no bloqueen.
2. `context.WithCancel` para cancelar las requests HTTP pendientes.
3. `select` con `ctx.Done()` para que las goroutines no se queden bloqueadas.

### Takeaway
Siempre asegurate de que cada goroutine tiene una forma de terminar. Usa `context.Context` para cancelacion, canales con buffer adecuado, y `select` con casos de salida. La libreria `go.uber.org/goleak` puede detectar leaks en tests.

---

## Puzzle 4: Channel Direction Bug

### Codigo con Bug

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
    ch <- 99 // BUG: el consumer esta enviando al canal despues de consumir
}

func main() {
    ch := make(chan int)

    go producer(ch)
    consumer(ch) // panic: send on closed channel
}
```

### Pregunta
Cual es el bug? Como lo arreglarias usando channel directions?

### Explicacion del Problema
El consumer envia un valor al canal (`ch <- 99`) despues de que el producer lo ha cerrado. Enviar a un canal cerrado causa `panic: send on closed channel`. Ademas, conceptualmente, un consumer no deberia enviar al canal de entrada.

El problema real es que no hay restricciones de tipo en los canales — ambas funciones reciben `chan int`, que permite tanto enviar como recibir.

### Codigo Corregido

```go
package main

import "fmt"

// ch <-chan int: solo puede RECIBIR (receive-only)
// ch chan<- int: solo puede ENVIAR (send-only)

func producer(ch chan<- int) { // send-only: solo puede enviar
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch) // close solo es valido en send-only o bidireccional
}

func consumer(ch <-chan int) { // receive-only: solo puede recibir
    for val := range ch {
        fmt.Println(val)
    }
    // ch <- 99 // ERROR DE COMPILACION: no se puede enviar en receive-only channel
}

func main() {
    ch := make(chan int)

    go producer(ch) // chan int se convierte automaticamente a chan<- int
    consumer(ch)    // chan int se convierte automaticamente a <-chan int
}
```

### Takeaway
Siempre usa channel directions en las firmas de funciones. El compilador previene errores en tiempo de compilacion. Regla: `chan<-` para el productor (send-only), `<-chan` para el consumidor (receive-only). Solo `close()` desde el lado del productor.

---

## Puzzle 5: Closure in Loop

### Codigo con Bug

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
            fmt.Println(v) // BUG: todas las goroutines capturan la MISMA variable v
        }()
    }

    wg.Wait()
}
```

### Pregunta
Cual es el output probable? Por que? Como se arregla?

### Explicacion del Problema
**En Go < 1.22**: la variable `v` del range loop es una **unica variable** que se reutiliza en cada iteracion. Todas las goroutines capturan un puntero a la misma variable. Cuando las goroutines se ejecutan, `v` ya tiene el ultimo valor (`"e"`), asi que la salida tipica es:

```
e
e
e
e
e
```

**En Go 1.22+**: Go cambio la semantica de loop variables. Ahora cada iteracion crea una nueva variable, asi que este codigo funciona correctamente. Sin embargo, es fundamental entender el problema para:
- Mantener codigo legacy.
- Entrevistas (pregunta clasica).
- Otros lenguajes con closures similares.

### Codigo Corregido (compatible con todas las versiones)

**Fix 1: Parametro en la goroutine**
```go
for _, v := range values {
    wg.Add(1)
    go func(val string) { // val es una copia local
        defer wg.Done()
        fmt.Println(val)
    }(v) // se pasa v como argumento (se copia)
}
```

**Fix 2: Variable local en el loop (pre-Go 1.22)**
```go
for _, v := range values {
    v := v // shadow: crea una nueva variable local por iteracion
    wg.Add(1)
    go func() {
        defer wg.Done()
        fmt.Println(v) // captura la variable local, no la del loop
    }()
}
```

**Fix 3: Go 1.22+ (sin cambios necesarios)**
```go
// Con Go 1.22+, el codigo original funciona correctamente.
// Cada iteracion del loop crea una nueva variable v.
for _, v := range values {
    wg.Add(1)
    go func() {
        defer wg.Done()
        fmt.Println(v) // OK en Go 1.22+
    }()
}
```

### Takeaway
En Go < 1.22, las closures dentro de loops capturan la variable del loop por referencia (comparten la misma variable). Siempre pasar como parametro o hacer shadow con `v := v`. En Go 1.22+ esto se resolvio a nivel del lenguaje, pero sigue siendo una pregunta de entrevista fundamental.

---

## Puzzle 6: Select Priority

### Codigo con Bug

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch1 := make(chan string, 1)
    ch2 := make(chan string, 1)

    ch1 <- "uno"
    ch2 <- "dos"

    // Ambos canales tienen datos listos
    select {
    case msg := <-ch1:
        fmt.Println("Recibido de ch1:", msg)
    case msg := <-ch2:
        fmt.Println("Recibido de ch2:", msg)
    }

    // Pregunta: siempre se ejecuta ch1 porque esta primero?
}
```

### Pregunta
Cual case se ejecuta? Es determinista?

### Explicacion del Problema
**No es determinista**. Cuando multiples cases de un `select` estan listos simultaneamente, Go elige uno **al azar** (pseudo-aleatorio uniforme). No hay prioridad basada en el orden de los cases.

Ejecutar este codigo multiples veces dara resultados diferentes:
```
$ go run main.go
Recibido de ch1: uno
$ go run main.go
Recibido de ch2: dos
$ go run main.go
Recibido de ch1: uno
```

**Razon de diseno**: evitar starvation. Si el primer case siempre tuviera prioridad, los canales listados despues podrian nunca ser atendidos.

### Como Implementar Prioridad Real

Si necesitas prioridad, usa `select` anidados:

```go
func prioritySelect(high <-chan string, low <-chan string) string {
    // Primero intentar el canal de alta prioridad
    select {
    case msg := <-high:
        return msg
    default:
    }

    // Si high no esta listo, esperar cualquiera
    select {
    case msg := <-high:
        return msg
    case msg := <-low:
        return msg
    }
}
```

**Alternativa robusta con loop:**
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
`select` con multiples cases listos elige al azar — no hay prioridad por orden. Para prioridad real, usa `select` anidados con `default`. Ten en cuenta que el patron de prioridad no es perfecto bajo alta carga — puede haber casos donde se procese un mensaje de baja prioridad antes que uno de alta.

---

## Puzzle 7: Context Cancellation

### Codigo con Bug

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func longRunningTask(ctx context.Context) string {
    // Simula trabajo pesado que NO revisa el contexto
    result := ""
    for i := 0; i < 10; i++ {
        time.Sleep(500 * time.Millisecond) // No revisa ctx.Done()
        result += fmt.Sprintf("paso-%d ", i)
    }
    return result
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    result := longRunningTask(ctx)
    fmt.Println("Resultado:", result)
    // La tarea toma 5 segundos aunque el timeout es de 1 segundo!
}
```

### Pregunta
Por que la cancelacion del contexto no funciona?

### Explicacion del Problema
El contexto `ctx` se pasa como parametro pero **nunca se revisa dentro de la funcion**. `context.Context` es cooperativo — no mata goroutines magicamente. El codigo dentro de la funcion debe revisar activamente `ctx.Done()` o `ctx.Err()` para reaccionar a la cancelacion.

### Codigo Corregido

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
        // Revisar cancelacion ANTES de cada paso
        select {
        case <-ctx.Done():
            return result, ctx.Err() // Retornar lo que se tiene + error
        default:
        }

        time.Sleep(500 * time.Millisecond)
        result += fmt.Sprintf("paso-%d ", i)
    }
    return result, nil
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    result, err := longRunningTask(ctx)
    if err != nil {
        fmt.Println("Cancelado:", err)
        fmt.Println("Resultado parcial:", result)
        return
    }
    fmt.Println("Resultado completo:", result)
}
```

**Version mas idiomatica usando select con timer:**
```go
func longRunningTask(ctx context.Context) (string, error) {
    result := ""
    for i := 0; i < 10; i++ {
        select {
        case <-ctx.Done():
            return result, ctx.Err()
        case <-time.After(500 * time.Millisecond):
            result += fmt.Sprintf("paso-%d ", i)
        }
    }
    return result, nil
}
```

### Takeaway
`context.Context` es cooperativo. Toda funcion que recibe un context debe revisarlo periodicamente. Patrones clave: (1) `select` con `ctx.Done()` en loops, (2) pasar el context a funciones downstream (HTTP requests, DB queries), (3) retornar `ctx.Err()` cuando se detecta cancelacion.

---

## Puzzle 8: WaitGroup Misuse

### Codigo con Bug

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
            wg.Add(1) // BUG: Add dentro de la goroutine
            defer wg.Done()
            fmt.Println("Worker:", n)
        }(i)
    }

    wg.Wait() // Puede terminar ANTES de que todas las goroutines hagan Add
    fmt.Println("Todos terminaron") // Mentira — puede que no
}
```

### Pregunta
Que esta mal? Por que el output es inconsistente?

### Explicacion del Problema
`wg.Add(1)` se llama **dentro** de la goroutine, pero `wg.Wait()` se llama en main. Hay una race condition:

1. Main lanza las goroutines y llega a `wg.Wait()`.
2. Si ninguna goroutine ha ejecutado `wg.Add(1)` todavia, el counter del WaitGroup es 0.
3. `wg.Wait()` retorna inmediatamente (counter == 0).
4. Main termina, matando a las goroutines que ni siquiera empezaron.

El resultado es que algunas (o todas) las goroutines no imprimen nada.

### Codigo Corregido

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    var wg sync.WaitGroup

    for i := 0; i < 5; i++ {
        wg.Add(1) // CORRECTO: Add ANTES de lanzar la goroutine
        go func(n int) {
            defer wg.Done()
            fmt.Println("Worker:", n)
        }(i)
    }

    wg.Wait() // Ahora espera a que las 5 goroutines terminen
    fmt.Println("Todos terminaron")
}
```

### Reglas del WaitGroup

```go
// CORRECTO: Add antes de go
wg.Add(1)
go worker(&wg)

// INCORRECTO: Add dentro de la goroutine
go func() {
    wg.Add(1) // race con wg.Wait()
    defer wg.Done()
    // ...
}()

// CORRECTO: Add con el total antes del loop
wg.Add(len(items))
for _, item := range items {
    go process(item, &wg)
}

// INCORRECTO: Done sin Add previo (panic: negative WaitGroup counter)
wg.Done() // panic!

// INCORRECTO: reusar WaitGroup antes de que Wait retorne
wg.Wait()
wg.Add(1) // OK solo si Wait ya retorno completamente
```

### Takeaway
**Siempre llamar a `wg.Add()` antes de lanzar la goroutine**, nunca dentro. El patron correcto es: Add -> go func -> defer Done. Alternativamente, hacer un solo `wg.Add(n)` con el total antes del loop.

---

## Resumen de Patrones

| Puzzle | Bug | Solucion Principal |
|---|---|---|
| 1. Race Condition | Acceso concurrente sin sync | `sync.Mutex`, `atomic`, o confinamiento |
| 2. Deadlock | Dependencia circular en canales | Cambiar orden, usar buffer, o `select` |
| 3. Goroutine Leak | Goroutines bloqueadas en sends | Buffer adecuado + `context.Context` |
| 4. Channel Direction | Consumer enviando a canal | Usar `<-chan` y `chan<-` en firmas |
| 5. Closure in Loop | Variable compartida del loop | Parametro, shadow, o Go 1.22+ |
| 6. Select Priority | Asumir orden en `select` | `select` anidado para prioridad real |
| 7. Context Cancellation | No revisar `ctx.Done()` | `select` periodico con `ctx.Done()` |
| 8. WaitGroup | `Add` dentro de goroutine | `Add` antes de `go func()` |
