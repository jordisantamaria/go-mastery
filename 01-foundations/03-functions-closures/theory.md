# 03 - Functions & Closures

## Funciones basicas

```go
func add(a int, b int) int {
    return a + b
}

// Parametros del mismo tipo se pueden agrupar
func add(a, b int) int {
    return a + b
}
```

- **Nombre en minuscula** (`add`) = privado (solo visible dentro del package)
- **Nombre en mayuscula** (`Add`) = publico (visible desde otros packages)

> Esta regla de visibilidad aplica a TODO en Go: funciones, tipos, variables, constantes, campos de struct.

## Multiple return values

Go permite devolver multiples valores. Es la base del error handling:

```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}

// El caller DEBE manejar ambos valores
result, err := divide(10, 3)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result)
```

> **Patron fundamental**: la ultima posicion del return es `error`. Si `err != nil`, el resto de valores son invalidos.

## Named return values

Puedes nombrar los valores de retorno. Se inicializan con zero values:

```go
func divide(a, b float64) (result float64, err error) {
    if b == 0 {
        err = fmt.Errorf("division by zero")
        return // "naked return" — devuelve result=0, err=<el error>
    }
    result = a / b
    return
}
```

> **Consejo**: los named returns son utiles para documentar que devuelve una funcion, pero los **naked returns** (return sin valores) hacen el codigo menos legible. Usalos solo en funciones cortas.

## Variadic functions

Funciones que aceptan un numero variable de argumentos:

```go
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

// Llamar
sum(1, 2, 3)       // 6
sum(1, 2, 3, 4, 5) // 15

// Pasar un slice con ...
nums := []int{1, 2, 3}
sum(nums...)        // 6
```

- `nums` dentro de la funcion es un `[]int` (slice)
- El parametro variadic debe ser el **ultimo**
- `fmt.Println` y `append` son variadic

## Funciones como first-class citizens

En Go, las funciones son valores. Puedes asignarlas a variables, pasarlas como argumento, y devolverlas:

```go
// Asignar a variable
greet := func(name string) string {
    return "Hola, " + name
}
fmt.Println(greet("Jordi"))

// Tipo de funcion
type MathFunc func(int, int) int

func apply(fn MathFunc, a, b int) int {
    return fn(a, b)
}

result := apply(func(a, b int) int { return a + b }, 3, 4) // 7
```

## Closures

Una closure es una funcion que **captura variables de su scope externo**:

```go
func counter() func() int {
    count := 0
    return func() int {
        count++ // captura 'count' del scope externo
        return count
    }
}

c := counter()
fmt.Println(c()) // 1
fmt.Println(c()) // 2
fmt.Println(c()) // 3

// Cada llamada a counter() crea una closure independiente
c2 := counter()
fmt.Println(c2()) // 1 — su propio contador
```

La closure **mantiene una referencia** a la variable, no una copia. Si la variable cambia fuera, la closure ve el cambio.

### Trampa clasica: closure en loop

```go
// BUG: todas las goroutines ven el mismo valor de i
for i := 0; i < 5; i++ {
    go func() {
        fmt.Println(i) // probablemente imprime "5" cinco veces
    }()
}

// SOLUCION 1: pasar como argumento
for i := 0; i < 5; i++ {
    go func(n int) {
        fmt.Println(n) // correcto: cada goroutine tiene su copia
    }(i)
}

// SOLUCION 2 (Go 1.22+): la variable del loop tiene scope por iteracion
// En Go 1.22+, el primer ejemplo ya funciona correctamente
```

> **Pregunta de entrevista clasica**: "Que imprime este codigo?" con una closure en un loop. Conocer este bug y las soluciones es fundamental.

## Anonymous functions (funciones anonimas)

Funciones sin nombre, utiles para callbacks y operaciones inline:

```go
// Ejecutar inmediatamente (IIFE)
result := func(a, b int) int {
    return a * b
}(3, 4) // 12

// Como callback
numbers := []int{5, 3, 8, 1, 9}
sort.Slice(numbers, func(i, j int) bool {
    return numbers[i] < numbers[j]
})
```

## Defer con funciones

`defer` trabaja con funciones (anonimas o nombradas):

```go
func process() {
    fmt.Println("start")

    // defer con funcion anonima
    defer func() {
        fmt.Println("cleanup")
    }()

    fmt.Println("working...")
    // Output: start, working..., cleanup
}
```

### Defer para medir tiempo (patron comun)

```go
func measureTime(name string) func() {
    start := time.Now()
    return func() {
        fmt.Printf("%s took %v\n", name, time.Since(start))
    }
}

func doWork() {
    defer measureTime("doWork")()
    // ... trabajo costoso ...
}
```

> Nota el `()` al final: `defer measureTime("doWork")()`. Primero se ejecuta `measureTime` (que guarda el tiempo inicial), y luego el defer ejecuta la funcion devuelta al salir.

## Patron: Functional Options

Patron muy popular en Go para constructores con muchas opciones:

```go
type Server struct {
    host    string
    port    int
    timeout time.Duration
}

// Option es una funcion que modifica el Server
type Option func(*Server)

func WithPort(port int) Option {
    return func(s *Server) {
        s.port = port
    }
}

func WithTimeout(d time.Duration) Option {
    return func(s *Server) {
        s.timeout = d
    }
}

func NewServer(host string, opts ...Option) *Server {
    s := &Server{
        host:    host,
        port:    8080,           // default
        timeout: 30 * time.Second, // default
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Uso limpio y extensible:
s := NewServer("localhost",
    WithPort(9090),
    WithTimeout(5 * time.Second),
)
```

> Este patron aparece en muchas librerias populares de Go (gRPC, zap logger, etc). Es **muy probable** que te pregunten por el en entrevista.

## Patron: Middleware / Function wrapping

Funciones que envuelven otras funciones para anyadir comportamiento:

```go
type Handler func(string) string

func withLogging(h Handler) Handler {
    return func(input string) string {
        fmt.Printf("Input: %s\n", input)
        result := h(input)
        fmt.Printf("Output: %s\n", result)
        return result
    }
}

func toUpper(s string) string {
    return strings.ToUpper(s)
}

// Componer
handler := withLogging(toUpper)
handler("hello") // Input: hello, Output: HELLO
```

## init()

Cada package puede tener funciones `init()` que se ejecutan **automaticamente** al importar el package:

```go
var config map[string]string

func init() {
    config = make(map[string]string)
    config["env"] = "development"
}
```

- Se ejecutan antes de `main()`
- Pueden haber multiples `init()` por archivo
- Se ejecutan en orden de declaracion
- **Evita usarlas** para logica compleja — hacen el codigo dificil de testear

## Preguntas de entrevista frecuentes

1. **Go soporta sobrecarga de funciones (overloading)?**
   No. Cada nombre de funcion debe ser unico dentro del package. Usa nombres descriptivos o variadic params.

2. **Que es una closure y como maneja la memoria?**
   Una closure es una funcion que captura variables de su scope externo. Las variables capturadas se mueven al heap (escape analysis) y viven mientras la closure exista.

3. **Diferencia entre value receiver y pointer receiver en methods?**
   Value receiver trabaja con copia (no modifica el original). Pointer receiver trabaja con referencia (modifica el original y evita copias). Lo veremos en detalle en modulo 04.

4. **Que es el functional options pattern?**
   Un patron que usa closures y variadic params para crear constructores flexibles y extensibles. Cada opcion es una funcion que modifica la configuracion.

5. **Cuando usarias init() y cuando no?**
   Usar: registrar drivers (database/sql), setup de constantes complejas. No usar: logica de negocio, I/O, dependencias externas. Hace el codigo dificil de testear y el orden de ejecucion poco predecible.

6. **Que pasa con closures en un for loop?**
   Todas las closures comparten la misma variable del loop. En Go <1.22, al ejecutarse, todas ven el ultimo valor. Solucion: pasar la variable como argumento o usar Go 1.22+ donde cada iteracion tiene su propia variable.
