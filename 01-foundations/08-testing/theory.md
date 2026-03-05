# 08 - Testing

Go tiene testing como ciudadano de primera clase. No necesitas frameworks externos — el package `testing` de la stdlib es potente y la comunidad lo usa universalmente.

## Basicos

### Convenciones

- Archivos de test: `xxx_test.go` (Go los excluye automaticamente del build)
- Funciones de test: `TestXxx(t *testing.T)` (empieza con `Test` + mayuscula)
- Mismo package que el codigo testeado (o `package xxx_test` para black-box testing)

```go
// math.go
package math

func Add(a, b int) int {
    return a + b
}

// math_test.go
package math

import "testing"

func TestAdd(t *testing.T) {
    got := Add(2, 3)
    if got != 5 {
        t.Errorf("Add(2, 3) = %d, want 5", got)
    }
}
```

### Ejecutar tests

```bash
go test ./...                    # todos los tests del proyecto
go test ./pkg/...                # tests de un package y sub-packages
go test -v ./...                 # verbose — muestra cada test
go test -run TestAdd ./...       # solo tests que matchean el regex
go test -race ./...              # con race detector
go test -count=1 ./...           # sin cache
go test -short ./...             # salta tests marcados como largos
go test -cover ./...             # muestra coverage
go test -coverprofile=cover.out  # genera archivo de coverage
go tool cover -html=cover.out    # abre coverage en el navegador
```

## Table-driven tests (EL patron de Go)

El patron mas importante. **Toda entrevista espera que lo conozcas**:

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 2, 3, 5},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
        {"mixed", -1, 5, 4},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.expected {
                t.Errorf("Add(%d, %d) = %d, want %d",
                    tt.a, tt.b, got, tt.expected)
            }
        })
    }
}
```

### Por que table-driven?

- **Facil anyadir casos**: solo una linea nueva en la tabla
- **Nombrado**: cada caso tiene un nombre descriptivo
- **Subtests**: `t.Run` permite ejecutar casos individuales (`-run TestAdd/positive`)
- **DRY**: la logica de asercion se escribe una sola vez
- **Estandar**: todo el ecosistema Go lo usa

## Subtests con t.Run

```go
func TestUser(t *testing.T) {
    t.Run("creation", func(t *testing.T) {
        u := NewUser("Jordi")
        if u.Name != "Jordi" {
            t.Error("wrong name")
        }
    })

    t.Run("validation", func(t *testing.T) {
        t.Run("empty name", func(t *testing.T) {
            err := ValidateUser("")
            if err == nil {
                t.Error("expected error for empty name")
            }
        })

        t.Run("valid name", func(t *testing.T) {
            err := ValidateUser("Jordi")
            if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    })
}
```

Ejecutar un subtest especifico:
```bash
go test -run TestUser/validation/empty_name ./...
```

## t.Helper()

Marca una funcion como helper — los errores reportan la linea del caller, no del helper:

```go
func assertEqual(t *testing.T, got, want int) {
    t.Helper() // sin esto, el error apunta a esta linea en lugar del test
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}

func TestSomething(t *testing.T) {
    assertEqual(t, Add(1, 2), 3) // el error apuntara AQUI, no dentro de assertEqual
}
```

## t.Parallel()

Ejecutar tests en paralelo:

```go
func TestSlowA(t *testing.T) {
    t.Parallel() // marca como paralelizable
    time.Sleep(time.Second)
    // ...
}

func TestSlowB(t *testing.T) {
    t.Parallel()
    time.Sleep(time.Second)
    // ...
}
// Ambos corren a la vez — total ~1s en vez de ~2s
```

> **Cuidado**: tests paralelos no deben compartir estado mutable. Cada uno debe ser independiente.

### Parallel en table-driven tests

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"case1", 1, 2, 3},
        {"case2", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // cada subtest corre en paralelo
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("got %d, want %d", got, tt.want)
            }
        })
    }
}
```

## t.Cleanup()

Registrar funciones de cleanup que se ejecutan al final del test:

```go
func TestWithDB(t *testing.T) {
    db := setupTestDB(t)
    t.Cleanup(func() {
        db.Close()  // se ejecuta al final, pase lo que pase
    })

    // usar db...
}
```

## TestMain — setup/teardown global

```go
func TestMain(m *testing.M) {
    // Setup global (antes de todos los tests)
    db := setupDatabase()

    code := m.Run() // ejecuta todos los tests

    // Teardown global (despues de todos los tests)
    db.Close()

    os.Exit(code)
}
```

## t.Skip — saltar tests condicionalmente

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    // test largo...
}
```

```bash
go test -short ./...  # salta los tests marcados con t.Skip en short mode
```

## Mocking con interfaces (sin frameworks)

En Go, el mocking se hace con **interfaces**, no con frameworks magicos:

```go
// El codigo real depende de una interface
type UserRepository interface {
    GetByID(id int) (*User, error)
    Save(user *User) error
}

type UserService struct {
    repo UserRepository
}

func (s *UserService) GetUser(id int) (*User, error) {
    return s.repo.GetByID(id)
}

// En tests, creas un mock que implementa la interface
type mockUserRepo struct {
    users map[int]*User
    err   error
}

func (m *mockUserRepo) GetByID(id int) (*User, error) {
    if m.err != nil {
        return nil, m.err
    }
    user, ok := m.users[id]
    if !ok {
        return nil, ErrNotFound
    }
    return user, nil
}

func (m *mockUserRepo) Save(user *User) error {
    return m.err
}

// Test
func TestGetUser(t *testing.T) {
    mock := &mockUserRepo{
        users: map[int]*User{
            1: {ID: 1, Name: "Jordi"},
        },
    }

    service := &UserService{repo: mock}

    user, err := service.GetUser(1)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.Name != "Jordi" {
        t.Errorf("got name %q, want Jordi", user.Name)
    }

    // Test error case
    _, err = service.GetUser(999)
    if !errors.Is(err, ErrNotFound) {
        t.Errorf("expected ErrNotFound, got %v", err)
    }
}
```

> Este patron es **fundamental**. Dependency injection via interfaces + mock structs = testeable sin frameworks.

## Benchmarks

Medir el rendimiento de funciones:

```go
func BenchmarkAdd(b *testing.B) {
    for b.Loop() {  // Go 1.24+ — el runtime controla las iteraciones
        Add(2, 3)
    }
}

// Benchmark con setup
func BenchmarkSort(b *testing.B) {
    data := generateRandomSlice(10000)
    b.ResetTimer() // no contar el setup

    for b.Loop() {
        sorted := make([]int, len(data))
        copy(sorted, data)
        sort.Ints(sorted)
    }
}
```

```bash
go test -bench=. ./...                 # ejecutar benchmarks
go test -bench=BenchmarkAdd ./...      # benchmark especifico
go test -bench=. -benchmem ./...       # incluir allocations
go test -bench=. -count=5 ./...        # 5 iteraciones para estabilidad
```

Output:
```
BenchmarkAdd-8    1000000000    0.3 ns/op    0 B/op    0 allocs/op
```

## Fuzzing (Go 1.18+)

Encontrar bugs con inputs aleatorios:

```go
func FuzzReverse(f *testing.F) {
    // Seed corpus — ejemplos iniciales
    f.Add("hello")
    f.Add("world")
    f.Add("")

    f.Fuzz(func(t *testing.T, s string) {
        reversed := Reverse(s)
        doubleReversed := Reverse(reversed)
        if s != doubleReversed {
            t.Errorf("Reverse(Reverse(%q)) = %q, want %q", s, doubleReversed, s)
        }
    })
}
```

```bash
go test -fuzz=FuzzReverse ./...          # ejecutar fuzzing
go test -fuzz=FuzzReverse -fuzztime=30s  # limitar tiempo
```

> Fuzzing es poderoso para encontrar edge cases que no se te ocurririan: strings Unicode raros, numeros extremos, inputs vacios.

## Build tags para integration tests

```go
//go:build integration

package mypackage

func TestDatabaseIntegration(t *testing.T) {
    // solo se ejecuta con: go test -tags=integration ./...
}
```

```bash
go test ./...                    # NO ejecuta integration tests
go test -tags=integration ./...  # SI los ejecuta
```

## Golden files (snapshot testing)

```go
func TestRenderHTML(t *testing.T) {
    got := RenderHTML(data)

    golden := filepath.Join("testdata", t.Name()+".golden")

    if *update {  // flag -update para regenerar
        os.WriteFile(golden, []byte(got), 0644)
    }

    want, _ := os.ReadFile(golden)
    if got != string(want) {
        t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
    }
}
```

## Errores comunes en testing

### t.Fatal vs t.Error

```go
t.Error("esto falla pero el test SIGUE")   // reporta error, continua
t.Fatal("esto falla y el test PARA")       // reporta error, para inmediatamente
t.Errorf("con formato: got %d", got)       // como Error pero con formato
t.Fatalf("con formato: %v", err)           // como Fatal pero con formato
```

- Usa `t.Fatal` cuando el resto del test no tiene sentido sin ese valor
- Usa `t.Error` cuando quieres reportar multiples fallos

## Preguntas de entrevista frecuentes

1. **Que es un table-driven test y por que es el estandar en Go?**
   Un test que define una tabla de casos (struct slice), itera sobre ellos, y ejecuta cada uno como subtest con t.Run. Es el estandar porque es DRY, facil de ampliar, y cada caso es aislado y nombrado.

2. **Como mockeas dependencias en Go?**
   Con interfaces. El codigo depende de una interface, y en tests pasas una implementacion mock (un struct que la satisface). No se necesitan frameworks.

3. **Que es t.Helper() y cuando lo usas?**
   Marca una funcion como test helper para que los errores reporten la linea del caller, no del helper. Se usa en funciones de asercion reutilizables.

4. **Diferencia entre t.Error y t.Fatal?**
   `t.Error` reporta el fallo y continua. `t.Fatal` reporta y detiene el test inmediatamente. Usa Fatal cuando un fallo hace que el resto del test no tenga sentido.

5. **Como ejecutas benchmarks en Go?**
   `go test -bench=. ./...`. Las funciones de benchmark usan `testing.B` y el metodo `b.Loop()` (Go 1.24+) o `for i := 0; i < b.N; i++`. Con `-benchmem` muestra allocations.

6. **Que es fuzzing y cuando es util?**
   Testing con inputs aleatorios generados automaticamente. Encuentra edge cases (Unicode, numeros extremos, strings largos) que un programador no pensaria. Disponible desde Go 1.18.

7. **Como separas unit tests de integration tests?**
   Con build tags (`//go:build integration`) y ejecutando con `go test -tags=integration`. O con `testing.Short()` y el flag `-short`.
