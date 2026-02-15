# 05 - Error Handling

En Go, **los errores son valores**, no excepciones. No hay `try/catch`. Esto es una decision de diseno deliberada que fuerza al programador a manejar errores explicitamente.

## La interface error

```go
type error interface {
    Error() string
}
```

Cualquier tipo que tenga un method `Error() string` es un error. Asi de simple.

## Creando errores

### errors.New — errores simples

```go
import "errors"

func validate(age int) error {
    if age < 0 {
        return errors.New("age cannot be negative")
    }
    return nil  // nil = sin error
}
```

### fmt.Errorf — errores con formato

```go
func validate(name string) error {
    if len(name) < 2 {
        return fmt.Errorf("name %q is too short (min 2 chars)", name)
    }
    return nil
}
```

## El patron basico: if err != nil

```go
result, err := doSomething()
if err != nil {
    return err  // propagar el error
}
// usar result...
```

> Veras `if err != nil` **cientos de veces** en cualquier proyecto Go. Es la forma idiomatica de manejar errores. No intentes evitarlo — abrazalo.

### Return early (el patron)

```go
// MAL — anidacion innecesaria
func process(id int) error {
    user, err := getUser(id)
    if err == nil {
        orders, err := getOrders(user.ID)
        if err == nil {
            // procesar...
            return nil
        }
        return err
    }
    return err
}

// BIEN — return early
func process(id int) error {
    user, err := getUser(id)
    if err != nil {
        return err
    }

    orders, err := getOrders(user.ID)
    if err != nil {
        return err
    }

    // procesar...
    return nil
}
```

## Error wrapping

Desde Go 1.13, puedes **envolver** errores para anyadir contexto sin perder el error original:

```go
func getUser(id int) (*User, error) {
    row := db.QueryRow("SELECT ...")
    if err := row.Scan(&user); err != nil {
        return nil, fmt.Errorf("getUser(%d): %w", id, err)
        //                                    ^^ %w envuelve el error
    }
    return &user, nil
}
```

`%w` (wrap) es diferente de `%v`:
- **`%w`**: envuelve el error — se puede desenvolver con `errors.Is`/`errors.As`
- **`%v`**: solo formatea el mensaje — el error original se pierde

## Sentinel errors

Errores predefinidos que se comparan por identidad:

```go
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrInvalidInput = errors.New("invalid input")
)

func getUser(id int) (*User, error) {
    // ...
    if user == nil {
        return nil, ErrNotFound
    }
    return user, nil
}

// Verificar
err := getUser(42)
if errors.Is(err, ErrNotFound) {
    // manejar "no encontrado"
}
```

> Convencion: los sentinel errors empiezan con `Err` (ej: `ErrNotFound`, `io.EOF`).

## errors.Is y errors.As

### errors.Is — comparar con sentinel error (a traves de wrapping)

```go
// Funciona incluso si el error esta envuelto
err := fmt.Errorf("database query: %w", ErrNotFound)

errors.Is(err, ErrNotFound)  // true! — desenvuelve automaticamente
err == ErrNotFound           // false — no es el mismo objeto
```

### errors.As — extraer un tipo de error especifico

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s - %s", e.Field, e.Message)
}

// Verificar y extraer
var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Println("Campo invalido:", valErr.Field)
    fmt.Println("Mensaje:", valErr.Message)
}
```

- `errors.Is`: **"es este error?"** — compara identidad
- `errors.As`: **"es de este tipo?"** — extrae tipo concreto

## Custom error types

Para errores con informacion extra:

```go
type HTTPError struct {
    StatusCode int
    Message    string
    URL        string
}

func (e *HTTPError) Error() string {
    return fmt.Sprintf("HTTP %d: %s (url: %s)", e.StatusCode, e.Message, e.URL)
}

func fetchData(url string) error {
    resp, err := http.Get(url)
    if err != nil {
        return fmt.Errorf("fetch %s: %w", url, err)
    }
    if resp.StatusCode != 200 {
        return &HTTPError{
            StatusCode: resp.StatusCode,
            Message:    resp.Status,
            URL:        url,
        }
    }
    return nil
}

// Uso
err := fetchData("https://api.example.com/data")
var httpErr *HTTPError
if errors.As(err, &httpErr) {
    if httpErr.StatusCode == 404 {
        // manejar not found
    }
}
```

## Patron: error handling en multiples operaciones

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("open: %w", err)
    }
    defer f.Close()

    data, err := io.ReadAll(f)
    if err != nil {
        return fmt.Errorf("read: %w", err)
    }

    if err := validate(data); err != nil {
        return fmt.Errorf("validate: %w", err)
    }

    if err := save(data); err != nil {
        return fmt.Errorf("save: %w", err)
    }

    return nil
}
```

Cada paso anyadir contexto con `%w` crea un "stack trace" de errores:
`"save: validate: invalid format"` — sabes exactamente donde fallo.

## panic vs error

| | `error` | `panic` |
|---|---|---|
| **Uso** | Situaciones esperadas | Bugs irrecuperables |
| **Ejemplos** | Archivo no existe, input invalido | Indice fuera de rango, nil pointer |
| **Se puede recuperar?** | Si, con if err != nil | Si, con recover() (pero no se recomienda) |
| **Flow** | Se propaga via return | Desenrolla todo el call stack |

> **Regla**: si puedes anticipar el error, usa `error`. `panic` es para bugs del programador o estado corrupto imposible.

### panic en la practica

```go
// Aceptable: inicializacion que DEBE funcionar
func mustCompile(pattern string) *regexp.Regexp {
    re, err := regexp.Compile(pattern)
    if err != nil {
        panic(fmt.Sprintf("invalid regex %q: %v", pattern, err))
    }
    return re
}

// Aceptable: funciones Must* en la stdlib (template.Must, regexp.MustCompile)
var emailRegex = regexp.MustCompile(`^[a-z]+@[a-z]+\.[a-z]+$`)
```

## Preguntas de entrevista frecuentes

1. **Por que Go no tiene excepciones?**
   Para forzar el manejo explicito de errores. Las excepciones crean flujos de control invisibles y hacen dificil saber que puede fallar. En Go, cada error es visible en la firma de la funcion.

2. **Diferencia entre errors.Is y errors.As?**
   `errors.Is` compara identidad (sentinel errors). `errors.As` extrae un tipo concreto. Ambos funcionan a traves de error wrapping.

3. **Que es error wrapping y por que importa?**
   `fmt.Errorf("context: %w", err)` envuelve un error anyadiendo contexto. Permite saber DONDE fallo (via el contexto) y QUE fallo (via errors.Is/As en el error original).

4. **Cuando usarias panic?**
   Solo para bugs del programador (estado imposible). Nunca para errores de I/O, validacion, o input de usuario. Excepciones: funciones `Must*` que se ejecutan en init.

5. **Cual es el problema del error handling verboso en Go?**
   `if err != nil` se repite mucho, lo que puede parecer excesivo. Pero esta verbosidad es deliberada: hace explicito que puede fallar y fuerza a decidir como manejarlo. El tradeoff es claridad sobre brevedad.

6. **Como harias un error que contiene informacion adicional?**
   Creando un custom error type (struct con Error() method). Puedes incluir campos como StatusCode, Field, Timestamp, etc. Se extrae con errors.As.
