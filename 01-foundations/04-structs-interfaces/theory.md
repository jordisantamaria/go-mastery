# 04 - Structs & Interfaces

Este es probablemente **el modulo mas importante** para entender Go. Go no tiene clases ni herencia — usa **structs + interfaces + composicion**.

## Structs

Un struct agrupa campos relacionados:

```go
type User struct {
    Name  string
    Email string
    Age   int
}

// Crear
u1 := User{Name: "Jordi", Email: "jordi@email.com", Age: 28}
u2 := User{"Jordi", "jordi@email.com", 28} // por posicion (fragil, evitar)
u3 := User{Name: "Jordi"}                  // Age=0, Email="" (zero values)

// Acceder
fmt.Println(u1.Name)
u1.Age = 29
```

### Zero value de un struct

Un struct sin inicializar tiene **todos sus campos en zero value**:

```go
var u User
// u.Name == "", u.Email == "", u.Age == 0
```

> Esto es poderoso: muchos structs en Go estan disenados para ser utiles con zero value (ej: `sync.Mutex{}`).

### Struct literals y punteros

```go
// Crear un puntero a struct
u := &User{Name: "Jordi"}  // u es *User

// Go permite acceder a campos sin desreferenciar
u.Name = "Jordi S."  // equivalente a (*u).Name — Go lo hace automatico
```

### Structs anonimos (inline)

Utiles para tests y JSON one-off:

```go
point := struct {
    X, Y int
}{10, 20}

// Muy comun en table-driven tests
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

Metadatos para serialization/validation:

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age,omitempty"`
}
```

- `json:"name"` — el campo se serializa como "name" en JSON
- `omitempty` — se omite si tiene zero value
- Las tags se leen con reflection (`reflect` package)

## Methods

Un method es una funcion asociada a un tipo:

```go
type Rectangle struct {
    Width, Height float64
}

// Method con VALUE receiver
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

// Method con POINTER receiver
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
| **Modifica el original?** | No (trabaja con copia) | Si |
| **Copia los datos?** | Si | No (solo copia el puntero) |
| **Cuando usar** | Lectura, structs pequenyos | Mutacion, structs grandes |

> **Regla practica**: si **algun** method necesita pointer receiver, usa pointer receiver en **todos** los methods del tipo. Esto mantiene la consistencia y evita bugs sutiles con interfaces.

### Constructor pattern (New...)

Go no tiene constructores. Por convencion, se usa una funcion `New`:

```go
func NewRectangle(w, h float64) *Rectangle {
    return &Rectangle{Width: w, Height: h}
}

// Cuando hay validacion
func NewRectangle(w, h float64) (*Rectangle, error) {
    if w <= 0 || h <= 0 {
        return nil, fmt.Errorf("dimensions must be positive")
    }
    return &Rectangle{Width: w, Height: h}, nil
}
```

## Embedding (composicion)

Go no tiene herencia. Usa **embedding** para composicion:

```go
type Animal struct {
    Name string
}

func (a Animal) Speak() string {
    return a.Name + " makes a sound"
}

type Dog struct {
    Animal  // embedding — Dog "hereda" los campos y methods de Animal
    Breed string
}

d := Dog{
    Animal: Animal{Name: "Rex"},
    Breed:  "Labrador",
}

fmt.Println(d.Name)    // "Rex" — promovido desde Animal
fmt.Println(d.Speak()) // "Rex makes a sound" — method promovido
fmt.Println(d.Breed)   // "Labrador"
```

- Los campos y methods del tipo embebido se "promueven" al tipo exterior
- **No es herencia**: Dog no "es" un Animal, sino que "tiene" un Animal
- Si Dog define su propio `Speak()`, este tiene prioridad (shadowing)

### Override (shadowing)

```go
func (d Dog) Speak() string {
    return d.Name + " barks!"
}

d.Speak()          // "Rex barks!" — method de Dog
d.Animal.Speak()   // "Rex makes a sound" — method de Animal (acceso directo)
```

## Interfaces

Las interfaces en Go son **implicitas** — un tipo satisface una interface si implementa todos sus methods. No hay `implements` keyword:

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

// Circle satisface Shape automaticamente — no necesita declararlo
var s Shape = Circle{Radius: 5}
fmt.Println(s.Area())
```

> **Esto es revolucionario** comparado con Java/C#: puedes implementar interfaces de terceros sin modificar su codigo, y un tipo puede satisfacer multiples interfaces sin saberlo.

### Interfaces pequenyas (la filosofia Go)

```go
// Interfaces de 1-2 methods son la norma en Go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// Composicion de interfaces
type ReadWriter interface {
    Reader
    Writer
}
```

> **"The bigger the interface, the weaker the abstraction."** — Rob Pike. Interfaces de 1 method son las mas poderosas.

### Interface satisfecha por pointer vs value

```go
type Mutable interface {
    SetName(string)
}

type User struct {
    Name string
}

// Method con pointer receiver
func (u *User) SetName(name string) {
    u.Name = name
}

// var m Mutable = User{}    // ERROR: User no satisface Mutable
var m Mutable = &User{}     // OK: *User si satisface Mutable
```

> Si un method tiene **pointer receiver**, solo `*T` satisface la interface, no `T`. Si todos los methods son **value receiver**, tanto `T` como `*T` la satisfacen.

### Empty interface (`any`)

```go
// any es un alias de interface{} (Go 1.18+)
func printAnything(v any) {
    fmt.Printf("Type: %T, Value: %v\n", v, v)
}

printAnything(42)
printAnything("hello")
printAnything([]int{1, 2, 3})
```

- `any` acepta CUALQUIER tipo
- Pierde type safety — usa solo cuando sea necesario (JSON parsing, logging, etc)

### Type assertions

Extraer el tipo concreto de una interface:

```go
var i any = "hello"

// Type assertion con check (segura)
s, ok := i.(string)
if ok {
    fmt.Println("Es string:", s)
}

// Type assertion sin check (puede panic)
s = i.(string) // OK
// n := i.(int)  // PANIC: interface conversion: interface {} is string, not int
```

### Type switch

```go
func describe(i any) string {
    switch v := i.(type) {
    case string:
        return fmt.Sprintf("string de largo %d", len(v))
    case int:
        return fmt.Sprintf("entero: %d", v)
    case bool:
        return fmt.Sprintf("bool: %t", v)
    case nil:
        return "nil"
    default:
        return fmt.Sprintf("tipo desconocido: %T", v)
    }
}
```

## Interfaces comunes de la stdlib

### Stringer (el ToString de Go)

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
fmt.Println(u) // "Jordi (28 years)" — fmt.Println llama String() automaticamente
```

### error interface

```go
type error interface {
    Error() string
}

// Cualquier tipo con Error() string es un error
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

// Archivos, HTTP responses, strings, buffers... todos implementan Reader
// Esto permite escribir funciones que trabajan con CUALQUIER fuente de datos
func countBytes(r io.Reader) (int64, error) {
    // funciona con archivos, HTTP, strings, etc.
    return io.Copy(io.Discard, r)
}
```

## Preguntas de entrevista frecuentes

1. **Go tiene herencia?**
   No. Go usa composicion via embedding. Un struct puede embeber otros structs y sus methods se promueven, pero no hay relacion "is-a", solo "has-a".

2. **Que significa que las interfaces en Go son implicitas?**
   Un tipo satisface una interface simplemente implementando todos sus methods, sin declararlo explicitamente (no hay `implements`). Esto permite satisfacer interfaces de packages que no conoces.

3. **Diferencia entre value receiver y pointer receiver con interfaces?**
   Si un method tiene pointer receiver, solo `*T` satisface la interface. Si tiene value receiver, tanto `T` como `*T` la satisfacen. Esto es porque Go puede obtener `&T` de un `T` addressable, pero no siempre puede obtener `T` de un `*T` de forma segura.

4. **Por que las interfaces en Go deben ser pequenyas?**
   Interfaces pequenyas (1-2 methods) son mas faciles de implementar, mas flexibles, y promueven la composicion. `io.Reader` (1 method) es implementada por docenas de tipos. Una interface grande reduce su utilidad.

5. **Que es el "accept interfaces, return structs" pattern?**
   Las funciones deben aceptar interfaces como parametros (para flexibilidad) pero devolver tipos concretos (para claridad). Esto maximiza la reutilizacion y minimiza el acoplamiento.

6. **Como funciona el nil interface gotcha?**
   Un interface value es nil solo si tanto su tipo como su valor son nil. Si le asignas un puntero nil tipado (`var p *MyError = nil; var err error = p`), `err != nil` es **true** aunque el valor subyacente sea nil. Este es uno de los bugs mas comunes en Go.
