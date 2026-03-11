# 07 - Generics

Generics (Go 1.18+) allow writing functions and types that work with **multiple types** without losing type safety. Before generics, you had to use `interface{}` (losing safety) or duplicate code.

## Before vs After generics

```go
// BEFORE: duplicate or use interface{}
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

// AFTER: a single generic function
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

## Basic syntax

### Generic function

```go
func FunctionName[T constraint](param T) T {
    // ...
}
```

- **`[T constraint]`** — type parameter `T` with a constraint
- `T` can be used as the type for parameters, return values, and local variables

### Generic type

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

// Usage
intStack := Stack[int]{}
intStack.Push(1)
intStack.Push(2)

strStack := Stack[string]{}
strStack.Push("hello")
```

## Constraints

A constraint defines what operations the generic type can perform.

### Built-in constraints

```go
any         // any type (alias for interface{})
comparable  // types that support == and != (basic types, structs of comparables, etc.)
```

### Constraints as interfaces

```go
// Constraint that requires a method
type Stringer interface {
    String() string
}

func PrintAll[T Stringer](items []T) {
    for _, item := range items {
        fmt.Println(item.String())
    }
}

// Constraint with type union (Go 1.18+)
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

### The ~ operator (underlying type)

```go
type Celsius float64
type Fahrenheit float64

// Without ~: Celsius and Fahrenheit do NOT satisfy the constraint
type StrictFloat interface {
    float64
}

// With ~: accepts any type whose underlying type is float64
type FlexFloat interface {
    ~float64
}

func Double[T FlexFloat](v T) T {
    return v * 2
}

var temp Celsius = 36.6
Double(temp) // OK with ~float64, error without ~
```

> `~T` means "any type whose underlying type is T". It is necessary to accept custom types like `Celsius`.

### cmp.Ordered (stdlib constraint)

```go
import "cmp"

// cmp.Ordered includes all types that support <, >, <=, >=
func Max[T cmp.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}

Max(3, 5)       // 5
Max("abc", "z") // "z"
```

## Common patterns

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

// Usage
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

### Keys / Values of a map

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

## Type inference

Go infers type parameters in most cases:

```go
// You don't need to specify the type
Max(3, 5)          // infers Max[int]
Contains(nums, 42) // infers Contains[int]

// Sometimes you need to be explicit
result := Map[string, int](names, func(s string) int { return len(s) })
```

## When to use generics (and when NOT)

### DO use generics

- **Utility functions** that operate on slices, maps, channels of any type
- **Generic data structures**: Stack, Queue, Set, LinkedList
- **Algorithms** that work with ordered/comparable types
- When you would be **duplicating identical code** for different types

### DO NOT use generics

- **When you only need `interface{}`**: if the function does not operate on the type (just passes it through), you don't need generics
- **When there are 1-2 types**: it is clearer to write specific functions
- **For behavioral polymorphism**: use classic interfaces (methods), not generics
- **Prematurely**: do not generalize until you have at least 3 concrete uses

```go
// NO: overcomplicating for nothing
func PrintValue[T any](v T) { fmt.Println(v) }
// Better: func PrintValue(v any) { fmt.Println(v) }

// YES: you really need type safety and to avoid duplication
func Min[T cmp.Ordered](a, b T) T { ... }
```

## Current limitations

1. **No method type parameters**: you cannot have type parameters on individual methods (only on the type)
   ```go
   // Does NOT compile
   func (s *Stack) Map[U any](fn func(T) U) *Stack[U] { ... }
   // Workaround: free function
   func MapStack[T, U any](s *Stack[T], fn func(T) U) *Stack[U] { ... }
   ```

2. **No specialization**: you cannot have different implementations for specific types

3. **No variadic type parameters**: you cannot have `func F[T1, T2, T3... any](...)`

## Common interview questions

1. **Since which Go version do generics exist?**
   Go 1.18 (March 2022). Before that, `interface{}` or code generation was used.

2. **What is a constraint?**
   An interface that defines which types can be used as a type parameter. It can require methods (like normal interfaces) or concrete types with `|`.

3. **What does the `~` operator do?**
   `~T` means "any type whose underlying type is T". Necessary to accept custom types (like `type Celsius float64`).

4. **When would you NOT use generics?**
   When you only work with one concrete type, when `any` suffices (you don't need type safety), or when the abstraction does not simplify the code. The Go proverbs say "a little copying is better than a little dependency".

5. **Difference between `any` constraint and `any` as a type?**
   `any` as a constraint (`func F[T any]`) preserves type safety: the compiler knows the concrete type. `any` as a type (`func F(v any)`) loses type safety: you need type assertions to operate.
