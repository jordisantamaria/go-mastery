# 02 - Control Flow

## if / else

En Go, **no hay parentesis** alrededor de la condicion. Las llaves son **obligatorias**:

```go
x := 10
if x > 5 {
    fmt.Println("mayor que 5")
} else if x > 0 {
    fmt.Println("positivo")
} else {
    fmt.Println("cero o negativo")
}
```

### if con statement inicial (idiomatico)

Puedes declarar una variable dentro del `if` — su scope se limita al bloque:

```go
if err := doSomething(); err != nil {
    fmt.Println("error:", err)
    return
}
// err no existe aqui fuera
```

> Este patron es **extremadamente comun** en Go. Lo veras en cada funcion que maneje errores.

## switch

Mucho mas potente que en otros lenguajes. **No necesita `break`** (no hay fall-through por defecto):

```go
day := "Monday"
switch day {
case "Monday", "Tuesday", "Wednesday", "Thursday", "Friday":
    fmt.Println("Dia laboral")
case "Saturday", "Sunday":
    fmt.Println("Fin de semana")
default:
    fmt.Println("Dia invalido")
}
```

### switch sin expresion (reemplaza if/else largo)

```go
score := 85
switch {
case score >= 90:
    fmt.Println("A")
case score >= 80:
    fmt.Println("B")
case score >= 70:
    fmt.Println("C")
default:
    fmt.Println("F")
}
```

### Type switch (importante para interfaces)

```go
func describe(i interface{}) string {
    switch v := i.(type) {
    case int:
        return fmt.Sprintf("entero: %d", v)
    case string:
        return fmt.Sprintf("string: %s", v)
    case bool:
        return fmt.Sprintf("bool: %t", v)
    default:
        return fmt.Sprintf("tipo desconocido: %T", v)
    }
}
```

### fallthrough (raro, pero existe)

```go
switch 1 {
case 1:
    fmt.Println("uno")
    fallthrough // fuerza la ejecucion del siguiente case
case 2:
    fmt.Println("dos") // se ejecuta aunque case != 2
}
```

## for (el unico loop en Go)

Go solo tiene `for`. No hay `while` ni `do-while`:

```go
// for clasico
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// "while" loop
count := 0
for count < 5 {
    count++
}

// loop infinito
for {
    // usa break para salir
    break
}
```

### range — iterar sobre colecciones

```go
// Slice
nums := []int{10, 20, 30}
for i, v := range nums {
    fmt.Printf("indice %d: valor %d\n", i, v)
}

// Solo valores (descarta indice)
for _, v := range nums {
    fmt.Println(v)
}

// Solo indices
for i := range nums {
    fmt.Println(i)
}

// Map
m := map[string]int{"a": 1, "b": 2}
for key, value := range m {
    fmt.Printf("%s: %d\n", key, value)
}

// String (itera por runes, no bytes)
for i, r := range "Hola 🌍" {
    fmt.Printf("byte %d: rune %c\n", i, r)
}
```

### break y continue

```go
for i := 0; i < 100; i++ {
    if i%2 == 0 {
        continue // salta a la siguiente iteracion
    }
    if i > 10 {
        break // sale del loop
    }
    fmt.Println(i) // 1, 3, 5, 7, 9
}
```

### Labels (break/continue en loops anidados)

```go
outer:
for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
        if i == 1 && j == 1 {
            break outer // sale de AMBOS loops
        }
        fmt.Printf("(%d, %d) ", i, j)
    }
}
```

## defer

`defer` programa una funcion para ejecutarse **al salir de la funcion actual** (despues del return). Se ejecutan en orden **LIFO** (ultimo en entrar, primero en salir):

```go
func readFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close() // se ejecuta al salir de readFile, pase lo que pase

    // trabajar con el archivo...
    return nil
}
```

### defer LIFO

```go
func main() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
    // Output: 3, 2, 1
}
```

### defer evalua argumentos inmediatamente

```go
x := 10
defer fmt.Println(x) // imprime 10, no 20
x = 20
```

> Los argumentos se evaluan cuando se declara el `defer`, no cuando se ejecuta.

## panic y recover

`panic` detiene la ejecucion normal. `recover` lo captura (solo dentro de `defer`):

```go
func safeDiv(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic recovered: %v", r)
        }
    }()

    return a / b, nil // si b == 0, Go hace panic
}
```

> **Regla**: no uses `panic` para errores normales. Solo para situaciones verdaderamente irrecuperables (bug en el programa, estado corrupto). Usa `error` para todo lo demas.

## Preguntas de entrevista frecuentes

1. **Go tiene while loop?**
   No. `for` cubre todos los casos: `for condition {}` es el equivalente a while.

2. **Que pasa si haces defer dentro de un loop?**
   Los defers se acumulan y se ejecutan todos al salir de la funcion (no al salir del loop). Puede causar memory leaks si el loop es largo. Solucion: extraer el cuerpo del loop a una funcion separada.

3. **En que orden se ejecutan los defers?**
   LIFO — ultimo defer declarado, primero en ejecutarse.

4. **Cuando usar panic vs error?**
   `error` para flujo normal (archivo no encontrado, input invalido, etc). `panic` solo para bugs irrecuperables (indice fuera de rango, nil pointer en lugar imposible).

5. **Por que Go no tiene fall-through por defecto en switch?**
   Porque el fall-through implicito (como en C/Java) es fuente de bugs. Si lo necesitas, usas `fallthrough` explicitamente.
