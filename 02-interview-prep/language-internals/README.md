# Language Internals — Go en Profundidad

Guia exhaustiva de los internos de Go para preparacion de entrevistas tecnicas. Cada seccion cubre la teoria y los detalles de implementacion que los entrevistadores suelen preguntar.

---

## Tabla de Contenidos

1. [Garbage Collector](#garbage-collector)
2. [Scheduler (Modelo GMP)](#scheduler-modelo-gmp)
3. [Memory Model](#memory-model)
4. [Escape Analysis](#escape-analysis)
5. [Stack vs Heap](#stack-vs-heap)
6. [Interface Internals](#interface-internals)
7. [Slice Internals](#slice-internals)
8. [Map Internals](#map-internals)
9. [Preguntas de Entrevista](#preguntas-de-entrevista)

---

## Garbage Collector

### Algoritmo Tri-Color Mark-and-Sweep

El GC de Go utiliza un algoritmo de marcado y barrido tricolor **concurrente**. Los objetos se clasifican en tres colores:

- **Blanco**: no visitado todavia. Al final del ciclo de marcado, los objetos blancos se consideran basura.
- **Gris**: visitado, pero sus referencias aun no han sido examinadas.
- **Negro**: visitado y todas sus referencias han sido examinadas.

**Proceso:**

1. Inicialmente todos los objetos son blancos.
2. Los objetos raiz (stack, globals) se marcan como grises.
3. Se toma un objeto gris, se examinan sus referencias (que se marcan como grises), y el objeto original pasa a negro.
4. Se repite hasta que no queden objetos grises.
5. Todos los objetos blancos restantes son basura y se pueden liberar.

### Write Barrier

Como el GC corre **concurrentemente** con el programa (mutator), existe el riesgo de que el mutator modifique punteros mientras el GC esta marcando. El **write barrier** es un fragmento de codigo que se ejecuta cada vez que el mutator escribe un puntero, asegurando que el GC no pierda objetos vivos.

Go usa un **hybrid write barrier** (desde Go 1.8):
- Combina el Dijkstra write barrier (marca el nuevo objeto referenciado) con el Yuasa write barrier (marca el viejo objeto referenciado).
- Esto elimina la necesidad de re-escanear stacks durante el marcado, reduciendo la latencia.

### Fases del GC

1. **Mark Setup (STW)**: breve pausa stop-the-world para activar el write barrier y preparar el marcado. Todos los goroutines deben alcanzar un safe point.
2. **Marking (concurrent)**: recorre el grafo de objetos marcando los vivos. Corre en paralelo con el programa usando hasta 25% de los recursos de CPU por defecto.
3. **Mark Termination (STW)**: segunda pausa STW para desactivar el write barrier y hacer limpieza final.
4. **Sweeping (concurrent)**: libera la memoria de los objetos no marcados. Ocurre de forma incremental conforme se necesita memoria.

### Tuning: GOGC y GOMEMLIMIT

**GOGC** (por defecto 100):
- Controla la frecuencia del GC.
- Valor = porcentaje de crecimiento del heap antes de disparar un nuevo ciclo.
- `GOGC=100`: el GC se dispara cuando el heap crece al doble desde el ultimo ciclo.
- `GOGC=200`: tolera 3x el tamano del heap vivo antes de recolectar.
- `GOGC=off`: desactiva el GC (peligroso en produccion).

**GOMEMLIMIT** (Go 1.19+):
- Establece un limite suave de memoria total del proceso Go.
- El GC se vuelve mas agresivo cuando se acerca al limite.
- Resuelve el problema clasico: con GOGC alto, el programa puede consumir demasiada memoria.
- Patron recomendado en produccion: `GOGC=off` + `GOMEMLIMIT=XGiB` cuando se quiere maximizar throughput con un presupuesto de memoria fijo.

```go
// Ejemplo: configurar via variables de entorno
// GOGC=100 GOMEMLIMIT=512MiB ./mi-servicio

// O programaticamente:
import "runtime/debug"

func init() {
    debug.SetGCPercent(100)
    debug.SetMemoryLimit(512 << 20) // 512 MiB
}
```

### Latencia vs Throughput

| Estrategia | Latencia | Throughput | Memoria |
|---|---|---|---|
| GOGC bajo (e.g. 50) | Menor (GC frecuente pero rapido) | Menor (mas tiempo en GC) | Menor |
| GOGC alto (e.g. 200) | Mayor (GC menos frecuente pero mas trabajo) | Mayor (menos overhead) | Mayor |
| GOGC=off + GOMEMLIMIT | Variable | Maximizado | Controlada |

**Regla practica**: para servicios sensibles a latencia (APIs), mantener GOGC por defecto y ajustar GOMEMLIMIT. Para batch processing, aumentar GOGC.

---

## Scheduler (Modelo GMP)

### Los Tres Componentes

- **G (Goroutine)**: unidad de trabajo ligera. Contiene el stack, el instruction pointer y metadata. Tamano inicial de stack: ~2-8 KB (varia por version).
- **M (Machine/OS Thread)**: hilo del sistema operativo que ejecuta codigo Go. Cada M necesita un P para ejecutar goroutines.
- **P (Processor)**: recurso logico de scheduling. Contiene una run queue local de goroutines y el cache de memoria local (mcache). La cantidad de Ps se controla con `GOMAXPROCS`.

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

Cuando la run queue local de un P esta vacia, intenta robar goroutines de otros Ps:

1. Primero revisa la global run queue (1/61 del tiempo para fairness).
2. Luego intenta robar la mitad de la run queue de otro P aleatorio.
3. Si no encuentra trabajo, revisa el netpoller.
4. Si aun no hay trabajo, el M se estaciona (park) y deja el P libre.

### Preemption (Apropiacion)

**Cooperative Preemption (antes de Go 1.14):**
- Las goroutines solo cedian el procesador en puntos seguros: llamadas a funciones, operaciones de canal, asignacion de memoria, etc.
- Problema: un loop `for {}` sin llamadas a funciones bloqueaba el P indefinidamente.

**Asynchronous Preemption (Go 1.14+):**
- El runtime envia senales (SIGURG en Unix) a los Ms para forzar la preemption.
- Resuelve el problema de goroutines que no cooperan.
- El runtime puede pausar una goroutine en cualquier punto seguro del codigo.

### GOMAXPROCS

- Controla cuantos Ps (y por tanto cuantos hilos de OS ejecutan goroutines en paralelo) hay.
- Por defecto = numero de CPUs logicas.
- `runtime.GOMAXPROCS(n)` permite cambiarlo en runtime.
- **No confundir con el numero de hilos totales** — puede haber mas Ms que Ps (por ejemplo, Ms bloqueados en syscalls).

### Ciclo de Vida de una Goroutine

```
              ┌──────────┐
    go f() ──►│ Runnable │◄──── I/O listo, timer, chan recv
              └────┬─────┘
                   │ scheduler
              ┌────▼─────┐
              │ Running  │──── ejecutandose en un M/P
              └────┬─────┘
                   │
          ┌────────┼────────┐
          │        │        │
     ┌────▼───┐ ┌──▼───┐ ┌─▼──────┐
     │Waiting │ │Syscall│ │ Dead   │
     │(chan,  │ │(OS)   │ │(return)│
     │ mutex) │ └───┬───┘ └────────┘
     └────┬───┘     │
          │         │ syscall termina
          └─────────┘
```

### Netpoller

- Mecanismo para I/O no bloqueante integrado en el scheduler.
- Usa `epoll` (Linux), `kqueue` (macOS), `IOCP` (Windows).
- Cuando una goroutine hace I/O de red, en lugar de bloquear un M:
  1. Se registra el file descriptor con el netpoller.
  2. La goroutine pasa a estado "waiting".
  3. El M queda libre para ejecutar otra goroutine.
  4. Cuando la I/O esta lista, el netpoller despierta la goroutine y la pone en la run queue.

---

## Memory Model

### Relaciones Happens-Before

El memory model de Go define cuando una lectura de una variable puede **observar** un valor escrito por otra goroutine. Se basa en relaciones **happens-before**:

> Si evento A **happens-before** evento B, entonces A es observable por B.

Garantias principales:

1. **Inicializacion**: la funcion `init()` del paquete A happens-before cualquier funcion de un paquete que importa A.
2. **Creacion de goroutine**: la sentencia `go f()` happens-before el inicio de la ejecucion de `f()`.
3. **Fin de goroutine**: el retorno de una goroutine **no** tiene garantia de happens-before respecto a ningun evento en otra goroutine (a menos que se use sincronizacion explicita).
4. **Canales**: un send happens-before el correspondiente receive se completa.
5. **Canales con buffer**: el receive del elemento i happens-before el send del elemento i+C se completa (C = capacidad del canal).
6. **close(ch)**: happens-before un receive que retorna zero value por canal cerrado.
7. **sync.Mutex**: `Unlock()` happens-before cualquier `Lock()` posterior.
8. **sync.Once**: el retorno de `f()` en `once.Do(f)` happens-before cualquier retorno de `once.Do`.

### Garantias de sync/atomic

El paquete `sync/atomic` provee operaciones atomicas que tambien establecen relaciones happens-before:

```go
var x, y atomic.Int64

// Goroutine 1
x.Store(1)  // A
y.Store(1)  // B (A happens-before B)

// Goroutine 2
if y.Load() == 1 { // C
    // x.Load() esta garantizado a ser 1 aqui
    // porque A happens-before B, y B happens-before C
    fmt.Println(x.Load()) // 1
}
```

Desde Go 1.19, las operaciones atomicas son **sequentially consistent**, lo que significa que el orden total de operaciones atomicas es consistente con el orden del programa de cada goroutine.

### Garantias de Canales

```go
// Un send en un unbuffered channel happens-before
// el receive correspondiente se completa.
ch := make(chan int)

go func() {
    x = 42          // A
    ch <- struct{}{} // B — send happens-before receive completes
}()

<-ch               // C — receive completa despues del send
fmt.Println(x)     // D — x es garantizado 42
```

### Cuando NO Usar Memoria Compartida

Preferir canales sobre memoria compartida cuando:

- La comunicacion es entre goroutines con flujo claro de datos.
- Se necesita transferir **ownership** de datos.
- El patron es productor-consumidor, fan-out/fan-in, o pipeline.

Usar mutex/atomics cuando:
- Se protege un cache compartido.
- Se mantiene un contador simple.
- El acceso es rapido y la contencion es baja.

> **"Do not communicate by sharing memory; instead, share memory by communicating."** — Proverbio Go

---

## Escape Analysis

### Que es

El compilador de Go decide si una variable se puede alojar en el stack (barato) o si debe escapar al heap (mas caro, requiere GC). Este proceso se llama **escape analysis**.

### Como Verlo

```bash
go build -gcflags='-m' ./...
# Output mas verbose:
go build -gcflags='-m -m' ./...
```

Ejemplo de output:
```
./main.go:10:6: can inline newUser
./main.go:11:9: &User{} escapes to heap
./main.go:15:2: moved to heap: x
```

### Escenarios Comunes de Escape

#### 1. Retornar un puntero a variable local
```go
func newUser(name string) *User {
    u := User{Name: name} // escapa al heap — se retorna un puntero
    return &u
}
```

#### 2. Conversion a interface
```go
func logValue(v any) {
    fmt.Println(v) // el argumento puede escapar porque 'any' es interface
}

func main() {
    x := 42
    logValue(x) // x escapa al heap por la conversion a interface
}
```

#### 3. Closures que capturan variables
```go
func makeCounter() func() int {
    count := 0 // escapa al heap — capturada por el closure
    return func() int {
        count++
        return count
    }
}
```

#### 4. Slices que crecen mas alla de su capacidad inicial
```go
func grow() []int {
    s := make([]int, 0)
    for i := 0; i < 1000; i++ {
        s = append(s, i) // puede escapar si el compilador no puede determinar el tamano
    }
    return s
}
```

#### 5. Enviar un puntero a traves de un canal
```go
ch := make(chan *User)
u := &User{Name: "Ana"} // escapa — enviada por canal, lifetime indeterminado
ch <- u
```

### Por Que Importa para el Rendimiento

- **Stack**: asignacion = mover el stack pointer (extremadamente rapido, ~1 instruccion).
- **Heap**: asignacion = pedir memoria al allocator + sera rastreado por el GC.
- Reducir escapes al heap = menos presion en el GC = menor latencia.
- **Tip**: pasar structs por valor (no puntero) cuando son pequenos y no necesitan mutacion.

---

## Stack vs Heap

### Stacks de Goroutines

- **Tamano inicial**: tipicamente 2-8 KB (depende de la version de Go).
- **Crecimiento**: cuando el stack se llena, Go asigna un nuevo stack de **el doble** del tamano y copia todo el contenido (copy stack, no segmented stacks desde Go 1.4).
- **Shrinking**: el GC puede reducir stacks que estan usando mucho menos de lo asignado (tipicamente cuando solo se usa 1/4).

### Stack Copying

Cuando un stack crece:
1. Se asigna un nuevo bloque de memoria del doble del tamano.
2. Se copia todo el contenido del stack viejo al nuevo.
3. Se actualizan **todos los punteros** que apuntan al stack viejo (por eso Go no permite punteros a stack desde C/assembly facilmente).
4. Se libera el stack viejo.

Esto es O(n) respecto al tamano del stack, pero ocurre con poca frecuencia gracias al crecimiento exponencial.

### Cuando las Asignaciones Van al Heap

- La variable escapa del scope de la funcion (ver Escape Analysis).
- El compilador no puede determinar el tamano en tiempo de compilacion.
- Objetos demasiado grandes para el stack.
- Variables compartidas entre goroutines (a traves de punteros).

### Implicaciones para la Presion del GC

```
Stack allocation:
  - Automaticamente liberada cuando la funcion retorna
  - No involucra al GC
  - Extremadamente rapida

Heap allocation:
  - Requiere que el GC rastree el objeto
  - Puede causar pausas STW mas largas si hay mucha basura
  - Mas lenta que stack allocation
```

**Estrategias para reducir presion del GC:**
1. Usar `sync.Pool` para objetos temporales que se reusan frecuentemente.
2. Pre-alocar slices con capacidad conocida: `make([]T, 0, expectedSize)`.
3. Pasar structs pequenos por valor en lugar de por puntero.
4. Evitar conversiones innecesarias a `interface{}`.
5. Reusar buffers con `bytes.Buffer` o `[]byte` pools.

---

## Interface Internals

### iface vs eface

Internamente, Go representa las interfaces de dos formas:

**eface** (empty interface / `any` / `interface{}`):
```go
type eface struct {
    _type *_type // puntero a la info del tipo
    data  unsafe.Pointer // puntero a los datos
}
```

**iface** (interface con metodos):
```go
type iface struct {
    tab  *itab          // puntero a la tabla de metodos + info de tipos
    data unsafe.Pointer // puntero a los datos
}

type itab struct {
    inter *interfacetype // tipo de la interface
    _type *_type         // tipo concreto
    hash  uint32         // hash del tipo (para type assertions rapidos)
    _     [4]byte
    fun   [1]uintptr     // array de punteros a funciones (metodos)
}
```

**Punto clave**: ambas son 2 words (16 bytes en 64-bit). La itab se cachea despues de la primera creacion.

### Satisfaccion de Interfaces en Compile Time

Go verifica que un tipo satisface una interface **en tiempo de compilacion**:

```go
type Writer interface {
    Write([]byte) (int, error)
}

type MyWriter struct{}

func (m MyWriter) Write(p []byte) (int, error) {
    return len(p), nil
}

// Verificacion en compile time (patron comun):
var _ Writer = MyWriter{}       // OK
var _ Writer = (*MyWriter)(nil) // OK — verifica que *MyWriter implementa Writer
```

Si el tipo no implementa la interface, el compilador da error **inmediatamente**.

### El Gotcha Clasico: Nil Interface vs Interface con Valor Nil

Este es uno de los errores mas preguntados en entrevistas:

```go
type MyError struct{}

func (e *MyError) Error() string { return "error" }

func getError() error {
    var err *MyError = nil // puntero nil a MyError
    return err             // CUIDADO: retorna iface{tab: *itab(MyError), data: nil}
}

func main() {
    err := getError()
    if err != nil {
        // ESTO SE EJECUTA! err no es nil.
        // err es una interface con tab != nil (conoce el tipo)
        // pero data == nil (el valor es nil)
        fmt.Println("error:", err)
    }
}
```

**Explicacion**:
- Una interface es `nil` **solo** cuando tanto `tab` como `data` son nil.
- Cuando asignas un puntero nil tipado a una interface, la interface tiene informacion de tipo (tab != nil), asi que la interface NO es nil.

**Solucion**: retornar `nil` directamente, no un puntero nil tipado:
```go
func getError() error {
    var err *MyError = nil
    if err == nil {
        return nil // retorna una interface nil (tab=nil, data=nil)
    }
    return err
}
```

### Type Assertion y Type Switch

**Type assertion** — extrae el valor concreto:
```go
var w Writer = MyWriter{}
mw := w.(MyWriter)          // panic si w no es MyWriter
mw, ok := w.(MyWriter)      // ok=false si no es MyWriter, no panic
```

**Internamente**: compara el hash en la itab con el hash del tipo solicitado. Si coincide, retorna el puntero data. Es O(1).

**Type switch** — patron comun y eficiente:
```go
switch v := i.(type) {
case string:
    fmt.Println("string:", v)
case int:
    fmt.Println("int:", v)
default:
    fmt.Println("otro tipo")
}
```

---

## Slice Internals

### SliceHeader

Un slice en Go es un descriptor de tres campos:

```go
type SliceHeader struct {
    Data uintptr // puntero al array subyacente
    Len  int     // numero de elementos actuales
    Cap  int     // capacidad total del array subyacente
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
│ 1 │ 2 │ 3 │ 4 │ 5 │  (array subyacente)
└───┴───┴───┴───┴───┘
```

### Comportamiento de Append y Reasignacion

```go
s := make([]int, 0, 4) // len=0, cap=4

s = append(s, 1, 2, 3)    // len=3, cap=4 (mismo array)
s = append(s, 4)           // len=4, cap=4 (mismo array)
s = append(s, 5)           // len=5, cap=8 (NUEVO array, copia datos)
```

**Estrategia de crecimiento** (simplificada, varia por version):
- Si cap < 256: duplicar la capacidad.
- Si cap >= 256: crecer ~25% + un poco mas (formula ajustada desde Go 1.18).

**Importante**: cuando append causa reasignacion, el nuevo slice apunta a un array diferente. Slices que compartian el array original **no ven los cambios**.

### Slice Tricks

```go
// Eliminar elemento en indice i (no preserva orden — O(1)):
s[i] = s[len(s)-1]
s = s[:len(s)-1]

// Eliminar elemento en indice i (preserva orden — O(n)):
s = append(s[:i], s[i+1:]...)

// Insertar elemento en indice i:
s = append(s[:i], append([]T{elem}, s[i:]...)...)
// Mejor alternativa (sin allocation intermedia):
s = append(s, zero) // crece por 1
copy(s[i+1:], s[i:])
s[i] = elem

// Copiar un slice (independiente del original):
copia := make([]int, len(s))
copy(copia, s)
// O con Go 1.21+:
copia := slices.Clone(s)

// Filtrar in-place:
n := 0
for _, v := range s {
    if keep(v) {
        s[n] = v
        n++
    }
}
s = s[:n]
```

### Memory Leaks con Slices

**Problema clasico**: al hacer sub-slice, el array subyacente completo permanece en memoria:

```go
func getFirstThree(data []byte) []byte {
    return data[:3] // PELIGRO: retiene referencia al array completo
}

// Si 'data' tiene 1GB, esos 1GB no se pueden liberar.
```

**Solucion**: copiar los datos necesarios:
```go
func getFirstThree(data []byte) []byte {
    result := make([]byte, 3)
    copy(result, data[:3])
    return result
    // O con Go 1.21+:
    // return bytes.Clone(data[:3])
}
```

---

## Map Internals

### Tabla Hash con Buckets

Un `map[K]V` en Go es una tabla hash implementada con buckets:

- Cada bucket almacena hasta **8 pares key-value**.
- El hash de la key determina en que bucket va.
- Los top 8 bits del hash (tophash) se almacenan en el bucket para comparacion rapida.
- Si un bucket esta lleno, se encadenan overflow buckets.

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

**Nota de layout**: keys y values se almacenan separados (primero todas las keys, luego todos los values) para evitar padding innecesario. Por ejemplo, `map[int64]int8` — sin esta optimizacion, cada par necesitaria 16 bytes (por padding); con ella, 8 keys de 8 bytes + 8 values de 1 byte.

### Crecimiento y Evacuacion

Cuando el factor de carga excede ~6.5 (promedio de 6.5 elementos por bucket), el map crece:

1. Se asigna un nuevo array de buckets del **doble** del tamano.
2. Los datos **no** se copian inmediatamente (a diferencia de los stacks).
3. Se usa **evacuacion incremental**: cada operacion de insert/delete migra algunos buckets viejos al nuevo array.
4. Durante la evacuacion, tanto el array viejo como el nuevo coexisten.

Esto distribuye el costo del crecimiento a lo largo del tiempo, evitando picos de latencia.

### Maps No Son Seguros para Acceso Concurrente

```go
m := make(map[string]int)

// ESTO CAUSA DATA RACE (y posible crash):
go func() { m["a"] = 1 }()
go func() { m["b"] = 2 }()
```

**Razon**: el runtime detecta escrituras concurrentes al map y lanza un **fatal error** (no es un panic recuperable — el programa muere):

```
fatal error: concurrent map writes
```

**Soluciones**:
1. `sync.Mutex` o `sync.RWMutex` para proteger el map.
2. `sync.Map` para casos especificos (ver abajo).
3. Patron de confinamiento: cada goroutine tiene su propio map.

**sync.Map** es optimo para:
- Keys escritas una vez y leidas muchas veces.
- Goroutines que trabajan en sets de keys disjuntos.
- **No** es un reemplazo general para map + mutex.

### Aleatorizacion del Orden de Iteracion

```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
for k, v := range m {
    fmt.Println(k, v) // orden diferente cada ejecucion
}
```

**Por diseno**: Go aleatoriza el orden de iteracion de maps para evitar que el codigo dependa de un orden particular. Esto se implementa eligiendo un offset de inicio aleatorio para la iteracion.

---

## Preguntas de Entrevista

### Pregunta 1: Explica el modelo GMP del scheduler de Go.
**Respuesta**: G es una goroutine (unidad de trabajo ligera con su propio stack). M es un OS thread que ejecuta codigo. P es un procesador logico con una run queue local. La cantidad de Ps se controla con GOMAXPROCS. El scheduler asigna Gs a Ms a traves de Ps. Cuando un P no tiene trabajo, roba goroutines de otros Ps (work stealing). Esto permite que miles de goroutines corran eficientemente sobre pocos hilos del OS.

### Pregunta 2: Que pasa cuando un goroutine hace una syscall bloqueante?
**Respuesta**: El M que ejecuta la goroutine se bloquea con ella. El P se desasocia del M bloqueado y busca otro M libre (o crea uno nuevo) para seguir ejecutando goroutines. Cuando la syscall termina, el M intenta reacquirir un P. Si no hay Ps libres, la goroutine va a la global run queue y el M se estaciona.

### Pregunta 3: Cual es la diferencia entre una interface nil y una interface que contiene un valor nil?
**Respuesta**: Una interface en Go son dos palabras: (type, value). Una interface nil tiene ambos nil. Una interface que contiene un puntero nil tiene type != nil y value == nil. Comparar con `!= nil` da `true` para la segunda porque la interface "sabe" que tipo contiene. Esto es una fuente comun de bugs, especialmente al retornar `error`.

### Pregunta 4: Como funciona el escape analysis y por que importa?
**Respuesta**: El compilador analiza si una variable puede vivir solo en el stack o necesita escapar al heap. Las asignaciones en stack son practicamente gratis (solo mover el stack pointer) y no involucran al GC. Las del heap requieren el allocator y el GC las rastrea. Reducir escapes mejora el rendimiento. Se puede inspeccionar con `go build -gcflags='-m'`.

### Pregunta 5: Que pasa internamente cuando haces append a un slice que esta lleno?
**Respuesta**: Si `len == cap`, append asigna un nuevo array subyacente con mayor capacidad (tipicamente el doble si es pequeno, ~25% mas si es grande), copia los elementos del array viejo al nuevo, y retorna un nuevo SliceHeader apuntando al nuevo array. El array viejo eventualmente sera recolectado por el GC si no hay mas referencias.

### Pregunta 6: Por que los maps de Go no son seguros para acceso concurrente?
**Respuesta**: Por rendimiento. Agregar sincronizacion interna penalizaria todos los usos, incluso los single-threaded. Go opta por dejar la sincronizacion al programador. El runtime detecta escrituras concurrentes al map y lanza un fatal error (no un panic) para evitar corrupcion silenciosa de datos.

### Pregunta 7: Que es el write barrier y para que sirve?
**Respuesta**: El write barrier es codigo que se ejecuta cada vez que el mutator (programa) escribe un puntero. Es necesario porque el GC corre concurrentemente — sin el, el GC podria no ver nuevas referencias creadas durante el marcado y recolectar objetos vivos. Go usa un hybrid write barrier que combina las tecnicas de Dijkstra y Yuasa.

### Pregunta 8: Como maneja Go el crecimiento de stacks de goroutines?
**Respuesta**: Go usa "copy stacks" (desde Go 1.4). Cuando una goroutine necesita mas stack, se asigna un nuevo bloque del doble del tamano, se copia todo el contenido, se actualizan todos los punteros que apuntan al stack viejo, y se libera el viejo. Esto es O(n) pero infrecuente gracias al crecimiento exponencial.

### Pregunta 9: Explica las fases del Garbage Collector de Go.
**Respuesta**: (1) Mark Setup: breve pausa STW para activar el write barrier. (2) Marking: fase concurrente donde se recorre el grafo de objetos usando el algoritmo tricolor. Usa ~25% de CPU. (3) Mark Termination: breve pausa STW para desactivar el write barrier. (4) Sweeping: liberacion concurrente e incremental de objetos no marcados.

### Pregunta 10: Que es GOMEMLIMIT y cuando lo usarias?
**Respuesta**: GOMEMLIMIT (Go 1.19+) es un limite suave de memoria. El GC se vuelve mas agresivo al acercarse al limite. Es util cuando se conoce el presupuesto de memoria (ej: contenedor con 512MB). Se puede combinar con GOGC=off para maximizar throughput: el GC solo corre cuando se acerca al limite, no por porcentaje de crecimiento.

### Pregunta 11: Cual es la diferencia entre iface y eface?
**Respuesta**: eface es la representacion de `interface{}` (any): tiene un puntero al tipo y un puntero a los datos. iface es para interfaces con metodos: tiene un puntero a una itab (que contiene la tabla de metodos + informacion de tipos) y un puntero a los datos. La itab se cachea para evitar recalcularla en cada asignacion.

### Pregunta 12: Como evitarias una memory leak con slices?
**Respuesta**: Al hacer sub-slicing, el array subyacente completo permanece en memoria. Si el array original es grande y solo necesitas una pequena porcion, debes copiar los datos a un nuevo slice con `copy()` o `slices.Clone()`. Tambien aplica cuando se eliminan elementos: hacer `s[i] = zeroValue` antes de truncar para evitar retener referencias a objetos que deberian ser recolectados.

### Pregunta 13: Que garantias de happens-before proveen los canales?
**Respuesta**: (1) Un send happens-before el receive correspondiente se completa. (2) El close de un canal happens-before un receive que retorna zero value. (3) En unbuffered channels, el receive happens-before el send se completa. (4) En buffered channels con cap C, el receive del elemento k happens-before el send del elemento k+C se completa.

### Pregunta 14: Que es el netpoller y como se integra con el scheduler?
**Respuesta**: El netpoller usa mecanismos del OS (epoll/kqueue/IOCP) para I/O de red no bloqueante. Cuando una goroutine hace I/O de red, se registra el file descriptor con el netpoller, la goroutine pasa a "waiting", y el M queda libre. Cuando la I/O esta lista, la goroutine se pone en la run queue. Esto permite que miles de goroutines hagan I/O concurrente sin necesitar miles de threads.

### Pregunta 15: Como funciona la preemption asincrona en Go 1.14+?
**Respuesta**: Antes de Go 1.14, las goroutines solo cedian el procesador en puntos cooperativos (llamadas a funciones, operaciones de canal). Un loop `for {}` podia bloquear un P indefinidamente. Desde Go 1.14, el runtime usa senales del OS (SIGURG en Unix) para interrumpir goroutines en cualquier punto seguro. El sysmon thread detecta goroutines que llevan >10ms sin ceder y envia la senal.

### Pregunta 16: Cual es la diferencia entre sync.Mutex y sync.RWMutex? Cuando usar cada uno?
**Respuesta**: `sync.Mutex` permite un solo acceso (lectura o escritura). `sync.RWMutex` permite multiples lectores simultaneos pero solo un escritor. Usar RWMutex cuando las lecturas son mucho mas frecuentes que las escrituras (ratio tipico >10:1). Si las escrituras son frecuentes, el overhead de RWMutex no justifica su complejidad y un Mutex simple es preferible.

### Pregunta 17: Que es un data race y como se diferencia de una race condition?
**Respuesta**: Un **data race** es cuando dos goroutines acceden a la misma variable, al menos una escribe, y no hay sincronizacion. Es undefined behavior. Una **race condition** es un bug logico donde el resultado depende del orden de ejecucion. Se puede tener race conditions sin data races (usando sincronizacion pero con logica incorrecta). Go detecta data races con `go run -race`.
