# 01 - Syntax & Types

## Primer programa

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Go!")
}
```

Conceptos clave:
- **Todo archivo Go pertenece a un `package`**. El package `main` es especial: es el punto de entrada.
- **`func main()`** es la funcion que se ejecuta al correr el programa.
- **`fmt`** es el package de la stdlib para formatear e imprimir.

## Variables

### Declaracion explicita

```go
var name string = "Jordi"
var age int = 28
var active bool = true
```

### Declaracion corta (`:=`) — la mas comun

```go
name := "Jordi"    // Go infiere el tipo (string)
age := 28          // int
active := true     // bool
```

> `:=` solo funciona **dentro de funciones**. A nivel de package, usa `var`.

### Declaracion multiple

```go
var (
    host = "localhost"
    port = 8080
)
```

### Zero values

En Go, **toda variable tiene un valor por defecto** (zero value). No hay `null`/`nil` para tipos basicos:

| Tipo | Zero value |
|------|-----------|
| `int`, `float64` | `0` |
| `string` | `""` (string vacio) |
| `bool` | `false` |
| punteros, slices, maps, interfaces | `nil` |

Esto elimina una categoria entera de bugs. No hay "undefined" ni "null pointer" accidentales en tipos basicos.

### Constantes

```go
const Pi = 3.14159
const (
    StatusOK    = 200
    StatusError = 500
)
```

Las constantes se evaluan en **tiempo de compilacion**. No puedes asignarles el resultado de una funcion.

## Tipos basicos

### Numericos

```go
// Enteros con signo
var a int     // depende de la plataforma (32 o 64 bits)
var b int8    // -128 a 127
var c int16   // -32768 a 32767
var d int32   // ~-2 billion a ~2 billion
var e int64   // muy grande

// Enteros sin signo
var f uint    // depende de la plataforma
var g uint8   // 0 a 255 (alias: byte)
var h uint64

// Punto flotante
var i float32
var j float64 // el que usaras casi siempre
```

> **Regla practica**: usa `int` para enteros y `float64` para decimales. Solo especifica tamanyo cuando necesites optimizar memoria.

### Strings

Los strings en Go son **inmutables** y estan codificados en **UTF-8**:

```go
s := "Hola mundo"
fmt.Println(len(s))    // 10 — bytes, NO caracteres
fmt.Println(s[0])      // 72 — el byte 'H', no el caracter

// Para iterar por caracteres (runes):
for i, r := range s {
    fmt.Printf("indice=%d rune=%c\n", i, r)
}
```

### byte vs rune

```go
var b byte = 'A'    // alias de uint8 — un byte
var r rune = 'A'    // alias de int32 — un code point Unicode
```

- **`byte`**: para datos binarios, ASCII.
- **`rune`**: para caracteres Unicode. `'ñ'` ocupa mas de 1 byte pero es 1 rune.

### Conversiones de tipo

Go **no hace conversiones implicitas**. Siempre explicitas:

```go
x := 42          // int
y := float64(x)  // int -> float64
z := int(y)      // float64 -> int (trunca decimales)

// string <-> []byte
s := "hello"
bs := []byte(s)     // string -> []byte
s2 := string(bs)    // []byte -> string
```

## Punteros

Un puntero guarda la **direccion de memoria** de una variable:

```go
x := 42
p := &x       // p es un *int (puntero a int)
fmt.Println(*p) // 42 — desreferenciar el puntero
*p = 100
fmt.Println(x)  // 100 — x cambio porque modificamos via el puntero
```

- **`&x`** — "dame la direccion de x"
- **`*p`** — "dame el valor al que apunta p"

> A diferencia de C/C++, **no hay aritmetica de punteros en Go**. Son seguros.

### Cuando usar punteros

1. **Modificar un valor** dentro de una funcion (Go pasa todo por valor/copia)
2. **Evitar copias grandes** — si un struct es grande, pasa un puntero
3. **Indicar ausencia** — un puntero puede ser `nil`, un valor no

```go
func double(n *int) {
    *n *= 2
}

x := 5
double(&x)
fmt.Println(x) // 10
```

## Arrays

Los arrays tienen **tamanyo fijo** definido en compilacion:

```go
var numbers [5]int                  // [0, 0, 0, 0, 0]
colors := [3]string{"red", "green", "blue"}
auto := [...]int{1, 2, 3}          // el compilador cuenta: [3]int
```

> En la practica, **casi nunca usaras arrays directamente**. Usa slices.

## Slices

Los slices son **vistas dinamicas** sobre un array subyacente. Son el tipo mas usado:

```go
// Crear un slice
nums := []int{1, 2, 3, 4, 5}

// Slice de un slice (no copia datos, comparte el array)
sub := nums[1:3]   // [2, 3]

// Append — puede crear un nuevo array si no cabe
nums = append(nums, 6)

// Make — crear con tamanyo y capacidad
s := make([]int, 5)      // len=5, cap=5
s2 := make([]int, 0, 10) // len=0, cap=10
```

### Internals del slice (importante para entrevista)

Un slice es un struct de 3 campos:

```
┌─────────┬─────┬──────────┐
│ pointer │ len │ capacity │
└─────────┴─────┴──────────┘
```

- **pointer**: direccion del primer elemento en el array subyacente
- **len**: cuantos elementos tiene el slice
- **cap**: cuantos caben antes de necesitar un nuevo array

```go
s := make([]int, 3, 5)
fmt.Println(len(s)) // 3
fmt.Println(cap(s)) // 5
```

> Cuando `append` excede la capacidad, Go crea un nuevo array (normalmente el doble) y copia los datos. Esto es O(n) puntualmente pero amortizado O(1).

## Maps

Tablas hash — pares clave-valor:

```go
// Crear
ages := map[string]int{
    "alice": 30,
    "bob":   25,
}

// Leer (con comma-ok pattern)
age, ok := ages["alice"]
if !ok {
    fmt.Println("no existe")
}

// Escribir
ages["charlie"] = 35

// Borrar
delete(ages, "bob")

// Iterar (orden NO garantizado)
for name, age := range ages {
    fmt.Printf("%s: %d\n", name, age)
}
```

> **El comma-ok pattern** (`value, ok := m[key]`) es idiomatico en Go. Lo veras en maps, type assertions, y channel receives.

### Maps: nil vs empty

```go
var m map[string]int    // nil — leer OK, escribir PANIC
m2 := map[string]int{}  // empty — leer y escribir OK
m3 := make(map[string]int) // equivalente a m2
```

> **Nunca declares un map sin inicializarlo** si vas a escribir en el.

## Type definitions

```go
// Crear un tipo nuevo basado en otro
type Celsius float64
type Fahrenheit float64

// No puedes mezclarlos sin conversion explicita
var temp Celsius = 36.6
// var f Fahrenheit = temp  // ERROR de compilacion
var f Fahrenheit = Fahrenheit(temp) // OK
```

Esto da **type safety** — el compilador evita que mezcles unidades, IDs, etc.

## iota (enums en Go)

Go no tiene `enum`, pero `iota` cumple la misma funcion:

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

`iota` se resetea a 0 en cada bloque `const` y se incrementa con cada linea.

## Preguntas de entrevista frecuentes

1. **Cual es la diferencia entre array y slice?**
   Array: tamanyo fijo, tipo incluye el tamanyo (`[5]int != [3]int`). Slice: dinamico, es un descriptor (pointer + len + cap) sobre un array.

2. **Que pasa cuando haces append a un slice que esta lleno?**
   Go aloca un nuevo array con mas capacidad (generalmente 2x), copia los elementos, y devuelve un nuevo slice apuntando al nuevo array.

3. **Por que Go no tiene conversiones implicitas?**
   Para evitar bugs sutiles. La conversion explicita fuerza al programador a ser consciente de posibles perdidas de precision o cambios de semantica.

4. **Que es un zero value y por que es importante?**
   Es el valor por defecto de cada tipo. Permite que las variables siempre tengan un estado valido sin inicializacion explicita, eliminando una clase entera de bugs tipo "undefined".

5. **Diferencia entre `nil` map y empty map?**
   Un nil map permite lectura (devuelve zero value) pero **panic en escritura**. Un empty map permite ambas operaciones.
