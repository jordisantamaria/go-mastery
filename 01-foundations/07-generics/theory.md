# 07 - Generics

Generics (Go 1.18+) permiten escribir funciones y tipos que trabajan con **multiples tipos** sin perder type safety. Antes de generics, tenias que usar `interface{}` (perdiendo seguridad) o duplicar codigo.

## Antes vs Despues de generics

```go
// ANTES: duplicar o usar interface{}
func SumInts(nums []int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

func SumFloat64s(nums []float64) float64 {
    total := 0.0
    for _, n := range nums {
        total += n
    }
    return total
}

// DESPUES: una sola funcion generica
func Sum[T int | float64](nums []T) T {
    var total T
    for _, n := range nums {
        total += n
    }
    return total
}

Sum([]int{1, 2, 3})         // 6
Sum([]float64{1.1, 2.2})    // 3.3
```

## Sintaxis basica

### Funcion generica

```go
func FunctionName[T constraint](param T) T {
    // ...
}
```

- **`[T constraint]`** — type parameter `T` con un constraint
- `T` se puede usar como tipo de parametros, return, y variables locales

### Tipo generico

```go
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}

// Uso
intStack := Stack[int]{}
intStack.Push(1)
intStack.Push(2)

strStack := Stack[string]{}
strStack.Push("hello")
```

## Constraints

Un constraint define que operaciones puede hacer el tipo generico.

### Constraints built-in

```go
any         // cualquier tipo (alias de interface{})
comparable  // tipos que soportan == y != (basicos, structs de comparables, etc.)
```

### Constraints como interfaces

```go
// Constraint que requiere un method
type Stringer interface {
    String() string
}

func PrintAll[T Stringer](items []T) {
    for _, item := range items {
        fmt.Println(item.String())
    }
}

// Constraint con union de tipos (Go 1.18+)
type Number interface {
    int | int8 | int16 | int32 | int64 |
    float32 | float64
}

func Sum[T Number](nums []T) T {
    var total T
    for _, n := range nums {
        total += n
    }
    return total
}
```

### El operador ~ (underlying type)

```go
type Celsius float64
type Fahrenheit float64

// Sin ~: Celsius y Fahrenheit NO cumplen el constraint
type StrictFloat interface {
    float64
}

// Con ~: acepta cualquier tipo cuyo underlying type sea float64
type FlexFloat interface {
    ~float64
}

func Double[T FlexFloat](v T) T {
    return v * 2
}

var temp Celsius = 36.6
Double(temp) // OK con ~float64, error sin ~
```

> `~T` significa "cualquier tipo cuyo tipo subyacente sea T". Es necesario para aceptar custom types como `Celsius`.

### cmp.Ordered (constraint de la stdlib)

```go
import "cmp"

// cmp.Ordered incluye todos los tipos que soportan <, >, <=, >=
func Max[T cmp.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}

Max(3, 5)       // 5
Max("abc", "z") // "z"
```

## Patrones comunes

### Map, Filter, Reduce

```go
func Map[T any, U any](items []T, fn func(T) U) []U {
    result := make([]U, len(items))
    for i, item := range items {
        result[i] = fn(item)
    }
    return result
}

func Filter[T any](items []T, pred func(T) bool) []T {
    var result []T
    for _, item := range items {
        if pred(item) {
            result = append(result, item)
        }
    }
    return result
}

func Reduce[T any, U any](items []T, initial U, fn func(U, T) U) U {
    acc := initial
    for _, item := range items {
        acc = fn(acc, item)
    }
    return acc
}

// Uso
names := []string{"Alice", "Bob", "Charlie"}
lengths := Map(names, func(s string) int { return len(s) })
// [5, 3, 7]

long := Filter(names, func(s string) bool { return len(s) > 4 })
// ["Alice", "Charlie"]

total := Reduce([]int{1, 2, 3, 4}, 0, func(acc, n int) int { return acc + n })
// 10
```

### Contains

```go
func Contains[T comparable](items []T, target T) bool {
    for _, item := range items {
        if item == target {
            return true
        }
    }
    return false
}

Contains([]int{1, 2, 3}, 2)      // true
Contains([]string{"a", "b"}, "c") // false
```

### Keys / Values de un map

```go
func Keys[K comparable, V any](m map[K]V) []K {
    keys := make([]K, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}

func Values[K comparable, V any](m map[K]V) []V {
    vals := make([]V, 0, len(m))
    for _, v := range m {
        vals = append(vals, v)
    }
    return vals
}
```

### Generic Set

```go
type Set[T comparable] struct {
    items map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
    return &Set[T]{items: make(map[T]struct{})}
}

func (s *Set[T]) Add(item T) {
    s.items[item] = struct{}{}
}

func (s *Set[T]) Contains(item T) bool {
    _, ok := s.items[item]
    return ok
}

func (s *Set[T]) Len() int {
    return len(s.items)
}
```

## Inferencia de tipos

Go infiere los type parameters en la mayoria de casos:

```go
// No necesitas especificar el tipo
Max(3, 5)          // infiere Max[int]
Contains(nums, 42) // infiere Contains[int]

// A veces necesitas ser explicito
result := Map[string, int](names, func(s string) int { return len(s) })
```

## Cuando usar generics (y cuando NO)

### SI usar generics

- **Funciones utility** que operan sobre slices, maps, channels de cualquier tipo
- **Estructuras de datos** genericas: Stack, Queue, Set, LinkedList
- **Algoritmos** que trabajan con tipos ordenables/comparables
- Cuando estarias **duplicando codigo** identico para diferentes tipos

### NO usar generics

- **Cuando solo necesitas `interface{}`**: si la funcion no opera con el tipo (solo lo pasa), no necesitas generics
- **Cuando hay 1-2 tipos**: es mas claro escribir funciones especificas
- **Para polimorfismo de comportamiento**: usa interfaces clasicas (methods), no generics
- **Prematuramente**: no generalices hasta que tengas al menos 3 usos concretos

```go
// NO: sobrecomplicar por nada
func PrintValue[T any](v T) { fmt.Println(v) }
// Mejor: func PrintValue(v any) { fmt.Println(v) }

// SI: realmente necesitas type safety y evitar duplicacion
func Min[T cmp.Ordered](a, b T) T { ... }
```

## Limitaciones actuales

1. **No method type parameters**: no puedes tener type parameters en methods individuales (solo en el tipo)
   ```go
   // NO compila
   func (s *Stack) Map[U any](fn func(T) U) *Stack[U] { ... }
   // Workaround: funcion libre
   func MapStack[T, U any](s *Stack[T], fn func(T) U) *Stack[U] { ... }
   ```

2. **No specialization**: no puedes tener implementaciones distintas para tipos especificos

3. **No variadic type parameters**: no puedes tener `func F[T1, T2, T3... any](...)`

## Preguntas de entrevista frecuentes

1. **Desde que version de Go existen generics?**
   Go 1.18 (marzo 2022). Antes se usaba `interface{}` o code generation.

2. **Que es un constraint?**
   Una interface que define que tipos pueden usarse como type parameter. Puede requerir methods (como interfaces normales) o tipos concretos con `|`.

3. **Que hace el operador `~`?**
   `~T` significa "cualquier tipo cuyo underlying type sea T". Necesario para aceptar custom types (como `type Celsius float64`).

4. **Cuando NO usarias generics?**
   Cuando solo operas con un tipo concreto, cuando `any` basta (no necesitas type safety), o cuando la abstraccion no simplifica el codigo. Los Go proverbs dicen "a little copying is better than a little dependency".

5. **Diferencia entre `any` constraint y `any` como tipo?**
   `any` como constraint (`func F[T any]`) preserva type safety: el compilador sabe el tipo concreto. `any` como tipo (`func F(v any)`) pierde type safety: necesitas type assertions para operar.
