# 04 - Structs & Interfaces

This is probably **the most important module** for understanding Go. Go does not have classes or inheritance — it uses **structs + interfaces + composition**.

## Structs

A struct groups related fields:

```go
type User struct {
    Name  string
    Email string
    Age   int
}

// Create
u1 := User{Name: "Jordi", Email: "jordi@email.com", Age: 28}
u2 := User{"Jordi", "jordi@email.com", 28} // by position (fragile, avoid)
u3 := User{Name: "Jordi"}                  // Age=0, Email="" (zero values)

// Access
fmt.Println(u1.Name)
u1.Age = 29
```

### Zero value of a struct

An uninitialized struct has **all its fields at zero value**:

```go
var u User
// u.Name == "", u.Email == "", u.Age == 0
```

> This is powerful: many structs in Go are designed to be useful with zero value (e.g., `sync.Mutex{}`).

### Struct literals and pointers

```go
// Create a pointer to a struct
u := &User{Name: "Jordi"}  // u is *User

// Go allows accessing fields without dereferencing
u.Name = "Jordi S."  // equivalent to (*u).Name — Go does it automatically
```

### Anonymous structs (inline)

Useful for tests and one-off JSON:

```go
point := struct {
    X, Y int
}{10, 20}

// Very common in table-driven tests
tests := []struct {
    input    int
    expected int
}{
    {1, 1},
    {2, 4},
    {3, 9},
}
```

### Struct tags

Metadata for serialization/validation:

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age,omitempty"`
}
```

- `json:"name"` — the field is serialized as "name" in JSON
- `omitempty` — omitted if it has zero value
- Tags are read with reflection (`reflect` package)

## Methods

A method is a function associated with a type:

```go
type Rectangle struct {
    Width, Height float64
}

// Method with VALUE receiver
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

// Method with POINTER receiver
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}

rect := Rectangle{Width: 10, Height: 5}
fmt.Println(rect.Area())  // 50
rect.Scale(2)
fmt.Println(rect.Area())  // 200
```

### Value receiver vs Pointer receiver

| | Value receiver `(r Rect)` | Pointer receiver `(r *Rect)` |
|---|---|---|
| **Modifies the original?** | No (works with a copy) | Yes |
| **Copies the data?** | Yes | No (only copies the pointer) |
| **When to use** | Reading, small structs | Mutation, large structs |

> **Practical rule**: if **any** method needs a pointer receiver, use pointer receiver on **all** methods of the type. This maintains consistency and avoids subtle bugs with interfaces.

### Constructor pattern (New...)

Go does not have constructors. By convention, a `New` function is used:

```go
func NewRectangle(w, h float64) *Rectangle {
    return &Rectangle{Width: w, Height: h}
}

// When there is validation
func NewRectangle(w, h float64) (*Rectangle, error) {
    if w <= 0 || h <= 0 {
        return nil, fmt.Errorf("dimensions must be positive")
    }
    return &Rectangle{Width: w, Height: h}, nil
}
```

## Embedding (composition)

Go does not have inheritance. It uses **embedding** for composition:

```go
type Animal struct {
    Name string
}

func (a Animal) Speak() string {
    return a.Name + " makes a sound"
}

type Dog struct {
    Animal  // embedding — Dog "inherits" the fields and methods of Animal
    Breed string
}

d := Dog{
    Animal: Animal{Name: "Rex"},
    Breed:  "Labrador",
}

fmt.Println(d.Name)    // "Rex" — promoted from Animal
fmt.Println(d.Speak()) // "Rex makes a sound" — promoted method
fmt.Println(d.Breed)   // "Labrador"
```

- Fields and methods of the embedded type are "promoted" to the outer type
- **It is not inheritance**: Dog is not "an" Animal, but "has" an Animal
- If Dog defines its own `Speak()`, it takes priority (shadowing)

### Override (shadowing)

```go
func (d Dog) Speak() string {
    return d.Name + " barks!"
}

d.Speak()          // "Rex barks!" — Dog's method
d.Animal.Speak()   // "Rex makes a sound" — Animal's method (direct access)
```

## Interfaces

Interfaces in Go are **implicit** — a type satisfies an interface if it implements all its methods. There is no `implements` keyword:

```go
type Shape interface {
    Area() float64
    Perimeter() float64
}

type Circle struct {
    Radius float64
}

func (c Circle) Area() float64 {
    return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
    return 2 * math.Pi * c.Radius
}

// Circle satisfies Shape automatically — no need to declare it
var s Shape = Circle{Radius: 5}
fmt.Println(s.Area())
```

> **This is revolutionary** compared to Java/C#: you can implement interfaces from third parties without modifying their code, and a type can satisfy multiple interfaces without knowing it.

### Small interfaces (the Go philosophy)

```go
// Interfaces with 1-2 methods are the norm in Go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// Interface composition
type ReadWriter interface {
    Reader
    Writer
}
```

> **"The bigger the interface, the weaker the abstraction."** — Rob Pike. 1-method interfaces are the most powerful.

### Interface satisfied by pointer vs value

```go
type Mutable interface {
    SetName(string)
}

type User struct {
    Name string
}

// Method with pointer receiver
func (u *User) SetName(name string) {
    u.Name = name
}

// var m Mutable = User{}    // ERROR: User does not satisfy Mutable
var m Mutable = &User{}     // OK: *User does satisfy Mutable
```

> If a method has a **pointer receiver**, only `*T` satisfies the interface, not `T`. If all methods are **value receiver**, both `T` and `*T` satisfy it.

### Empty interface (`any`)

```go
// any is an alias for interface{} (Go 1.18+)
func printAnything(v any) {
    fmt.Printf("Type: %T, Value: %v\n", v, v)
}

printAnything(42)
printAnything("hello")
printAnything([]int{1, 2, 3})
```

- `any` accepts ANY type
- Loses type safety — use only when necessary (JSON parsing, logging, etc)

### Type assertions

Extract the concrete type from an interface:

```go
var i any = "hello"

// Type assertion with check (safe)
s, ok := i.(string)
if ok {
    fmt.Println("It's a string:", s)
}

// Type assertion without check (can panic)
s = i.(string) // OK
// n := i.(int)  // PANIC: interface conversion: interface {} is string, not int
```

### Type switch

```go
func describe(i any) string {
    switch v := i.(type) {
    case string:
        return fmt.Sprintf("string of length %d", len(v))
    case int:
        return fmt.Sprintf("integer: %d", v)
    case bool:
        return fmt.Sprintf("bool: %t", v)
    case nil:
        return "nil"
    default:
        return fmt.Sprintf("unknown type: %T", v)
    }
}
```

## Common stdlib interfaces

### Stringer (Go's ToString)

```go
type Stringer interface {
    String() string
}

type User struct {
    Name string
    Age  int
}

func (u User) String() string {
    return fmt.Sprintf("%s (%d years)", u.Name, u.Age)
}

u := User{Name: "Jordi", Age: 28}
fmt.Println(u) // "Jordi (28 years)" — fmt.Println calls String() automatically
```

### error interface

```go
type error interface {
    Error() string
}

// Any type with Error() string is an error
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
```

### io.Reader / io.Writer

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

// Files, HTTP responses, strings, buffers... all implement Reader
// This allows writing functions that work with ANY data source
func countBytes(r io.Reader) (int64, error) {
    // works with files, HTTP, strings, etc.
    return io.Copy(io.Discard, r)
}
```

## Common interview questions

1. **Does Go have inheritance?**
   No. Go uses composition via embedding. A struct can embed other structs and their methods are promoted, but there is no "is-a" relationship, only "has-a".

2. **What does it mean that interfaces in Go are implicit?**
   A type satisfies an interface simply by implementing all its methods, without explicitly declaring it (there is no `implements`). This allows satisfying interfaces from packages you don't know about.

3. **Difference between value receiver and pointer receiver with interfaces?**
   If a method has a pointer receiver, only `*T` satisfies the interface. If it has a value receiver, both `T` and `*T` satisfy it. This is because Go can get `&T` from an addressable `T`, but cannot always safely get `T` from a `*T`.

4. **Why should interfaces in Go be small?**
   Small interfaces (1-2 methods) are easier to implement, more flexible, and promote composition. `io.Reader` (1 method) is implemented by dozens of types. A large interface reduces its usefulness.

5. **What is the "accept interfaces, return structs" pattern?**
   Functions should accept interfaces as parameters (for flexibility) but return concrete types (for clarity). This maximizes reusability and minimizes coupling.

6. **How does the nil interface gotcha work?**
   An interface value is nil only if both its type and its value are nil. If you assign a typed nil pointer to it (`var p *MyError = nil; var err error = p`), `err != nil` is **true** even though the underlying value is nil. This is one of the most common bugs in Go.
