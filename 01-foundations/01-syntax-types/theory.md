# 01 - Syntax & Types

## First program

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Go!")
}
```

Key concepts:
- **Every Go file belongs to a `package`**. The `main` package is special: it is the entry point.
- **`func main()`** is the function that runs when the program is executed.
- **`fmt`** is the stdlib package for formatting and printing.

## Variables

### Explicit declaration

```go
var name string = "Jordi"
var age int = 28
var active bool = true
```

### Short declaration (`:=`) — the most common

```go
name := "Jordi"    // Go infers the type (string)
age := 28          // int
active := true     // bool
```

> `:=` only works **inside functions**. At the package level, use `var`.

### Multiple declaration

```go
var (
    host = "localhost"
    port = 8080
)
```

### Zero values

In Go, **every variable has a default value** (zero value). There is no `null`/`nil` for basic types:

| Type | Zero value |
|------|-----------|
| `int`, `float64` | `0` |
| `string` | `""` (empty string) |
| `bool` | `false` |
| pointers, slices, maps, interfaces | `nil` |

This eliminates an entire category of bugs. There is no "undefined" or accidental "null pointer" for basic types.

### Constants

```go
const Pi = 3.14159
const (
    StatusOK    = 200
    StatusError = 500
)
```

Constants are evaluated at **compile time**. You cannot assign them the result of a function.

## Basic types

### Numeric

```go
// Signed integers
var a int     // platform-dependent (32 or 64 bits)
var b int8    // -128 to 127
var c int16   // -32768 to 32767
var d int32   // ~-2 billion to ~2 billion
var e int64   // very large

// Unsigned integers
var f uint    // platform-dependent
var g uint8   // 0 to 255 (alias: byte)
var h uint64

// Floating point
var i float32
var j float64 // the one you will use almost always
```

> **Practical rule**: use `int` for integers and `float64` for decimals. Only specify size when you need to optimize memory.

### Strings

Strings in Go are **immutable** and encoded in **UTF-8**:

```go
s := "Hola mundo"
fmt.Println(len(s))    // 10 — bytes, NOT characters
fmt.Println(s[0])      // 72 — the byte 'H', not the character

// To iterate by characters (runes):
for i, r := range s {
    fmt.Printf("index=%d rune=%c\n", i, r)
}
```

### byte vs rune

```go
var b byte = 'A'    // alias for uint8 — one byte
var r rune = 'A'    // alias for int32 — a Unicode code point
```

- **`byte`**: for binary data, ASCII.
- **`rune`**: for Unicode characters. `'ñ'` takes more than 1 byte but is 1 rune.

### Type conversions

Go **does not do implicit conversions**. Always explicit:

```go
x := 42          // int
y := float64(x)  // int -> float64
z := int(y)      // float64 -> int (truncates decimals)

// string <-> []byte
s := "hello"
bs := []byte(s)     // string -> []byte
s2 := string(bs)    // []byte -> string
```

## Pointers

A pointer holds the **memory address** of a variable:

```go
x := 42
p := &x       // p is a *int (pointer to int)
fmt.Println(*p) // 42 — dereference the pointer
*p = 100
fmt.Println(x)  // 100 — x changed because we modified it via the pointer
```

- **`&x`** — "give me the address of x"
- **`*p`** — "give me the value that p points to"

> Unlike C/C++, **there is no pointer arithmetic in Go**. They are safe.

### When to use pointers

1. **Modify a value** inside a function (Go passes everything by value/copy)
2. **Avoid large copies** — if a struct is large, pass a pointer
3. **Indicate absence** — a pointer can be `nil`, a value cannot

```go
func double(n *int) {
    *n *= 2
}

x := 5
double(&x)
fmt.Println(x) // 10
```

## Arrays

Arrays have a **fixed size** defined at compile time:

```go
var numbers [5]int                  // [0, 0, 0, 0, 0]
colors := [3]string{"red", "green", "blue"}
auto := [...]int{1, 2, 3}          // the compiler counts: [3]int
```

> In practice, **you will almost never use arrays directly**. Use slices.

## Slices

Slices are **dynamic views** over an underlying array. They are the most commonly used type:

```go
// Create a slice
nums := []int{1, 2, 3, 4, 5}

// Slice of a slice (does not copy data, shares the array)
sub := nums[1:3]   // [2, 3]

// Append — may create a new array if it doesn't fit
nums = append(nums, 6)

// Make — create with length and capacity
s := make([]int, 5)      // len=5, cap=5
s2 := make([]int, 0, 10) // len=0, cap=10
```

### Slice internals (important for interviews)

A slice is a struct with 3 fields:

```
┌─────────┬─────┬──────────┐
│ pointer │ len │ capacity │
└─────────┴─────┴──────────┘
```

- **pointer**: address of the first element in the underlying array
- **len**: how many elements the slice has
- **cap**: how many fit before needing a new array

```go
s := make([]int, 3, 5)
fmt.Println(len(s)) // 3
fmt.Println(cap(s)) // 5
```

> When `append` exceeds the capacity, Go creates a new array (usually double the size) and copies the data. This is O(n) occasionally but amortized O(1).

## Maps

Hash tables — key-value pairs:

```go
// Create
ages := map[string]int{
    "alice": 30,
    "bob":   25,
}

// Read (with comma-ok pattern)
age, ok := ages["alice"]
if !ok {
    fmt.Println("does not exist")
}

// Write
ages["charlie"] = 35

// Delete
delete(ages, "bob")

// Iterate (order is NOT guaranteed)
for name, age := range ages {
    fmt.Printf("%s: %d\n", name, age)
}
```

> **The comma-ok pattern** (`value, ok := m[key]`) is idiomatic in Go. You will see it in maps, type assertions, and channel receives.

### Maps: nil vs empty

```go
var m map[string]int    // nil — reading OK, writing PANIC
m2 := map[string]int{}  // empty — reading and writing OK
m3 := make(map[string]int) // equivalent to m2
```

> **Never declare a map without initializing it** if you are going to write to it.

## Type definitions

```go
// Create a new type based on another
type Celsius float64
type Fahrenheit float64

// You cannot mix them without explicit conversion
var temp Celsius = 36.6
// var f Fahrenheit = temp  // COMPILE ERROR
var f Fahrenheit = Fahrenheit(temp) // OK
```

This provides **type safety** — the compiler prevents you from mixing units, IDs, etc.

## iota (enums in Go)

Go does not have `enum`, but `iota` serves the same purpose:

```go
type Weekday int

const (
    Sunday Weekday = iota  // 0
    Monday                 // 1
    Tuesday                // 2
    Wednesday              // 3
    Thursday               // 4
    Friday                 // 5
    Saturday               // 6
)
```

`iota` resets to 0 in each `const` block and increments with each line.

## Common interview questions

1. **What is the difference between array and slice?**
   Array: fixed size, the type includes the size (`[5]int != [3]int`). Slice: dynamic, it is a descriptor (pointer + len + cap) over an array.

2. **What happens when you append to a slice that is full?**
   Go allocates a new array with more capacity (usually 2x), copies the elements, and returns a new slice pointing to the new array.

3. **Why doesn't Go have implicit conversions?**
   To avoid subtle bugs. Explicit conversion forces the programmer to be aware of possible precision loss or semantic changes.

4. **What is a zero value and why is it important?**
   It is the default value for each type. It allows variables to always have a valid state without explicit initialization, eliminating an entire class of "undefined" bugs.

5. **Difference between `nil` map and empty map?**
   A nil map allows reading (returns zero value) but **panics on write**. An empty map allows both operations.
