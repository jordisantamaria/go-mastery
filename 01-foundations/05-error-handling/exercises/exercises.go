// Package exercises contiene funciones para practicar error handling en Go.
// Ejecuta los tests con: go test ./01-foundations/05-error-handling/exercises/...
package exercises

import "errors"

// Sentinel errors para los ejercicios
var (
	ErrDivisionByZero = errors.New("division by zero")
	ErrNegativeNumber = errors.New("negative number")
	ErrEmpty          = errors.New("empty input")
	ErrNotFound       = errors.New("not found")
	ErrOutOfRange     = errors.New("out of range")
)

// --- Ejercicio 1: SafeDivide ---
// Divide dos numeros con manejo de errores.
// Si b == 0, devuelve (0, ErrDivisionByZero).
func SafeDivide(a, b float64) (float64, error) {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 2: Sqrt ---
// Calcula la raiz cuadrada de un numero.
// Si n < 0, devuelve (0, ErrNegativeNumber).
// Usa el metodo de Newton: empezar con guess = n/2, iterar guess = (guess + n/guess) / 2
// Iterar 10 veces es suficiente precision.
func Sqrt(n float64) (float64, error) {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 3: ParseAge ---
// Convierte un string a edad (int).
// Errores posibles:
// - String vacio: devuelve error wrapping ErrEmpty
// - No es numero: devuelve error wrapping el error de strconv
// - Numero negativo: devuelve error wrapping ErrNegativeNumber
// - Mayor que 150: devuelve error wrapping ErrOutOfRange
// Usa fmt.Errorf con %w para wrapping.
func ParseAge(s string) (int, error) {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 4: SafeGet ---
// Devuelve el elemento en el indice dado de un slice.
// Si el indice esta fuera de rango, devuelve ("", ErrOutOfRange).
// Si el slice esta vacio, devuelve ("", ErrEmpty).
func SafeGet(items []string, index int) (string, error) {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 5: FindUser ---
// Busca un usuario por nombre en un slice de structs.
// Si el nombre esta vacio, devuelve error wrapping ErrEmpty.
// Si no se encuentra, devuelve error wrapping ErrNotFound.
// El error debe incluir contexto: "FindUser: <error>"

type User struct {
	Name  string
	Email string
}

func FindUser(users []User, name string) (*User, error) {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 6: Custom Error Type ---
// Crea un FieldError type con campos Field y Message.
// Implementa Error() string que devuelva "field <Field>: <Message>".
// Crea ValidateUser que valide:
// - Name no vacio (FieldError con Field="name")
// - Email contiene "@" (FieldError con Field="email")
// - Age entre 0 y 150 (FieldError con Field="age")
// Devuelve el PRIMER error encontrado.

type FieldError struct {
	// TODO: definir campos
}

func (e *FieldError) Error() string {
	// TODO: implementar
	panic("not implemented")
}

func ValidateUser(name, email string, age int) error {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 7: MultiError ---
// Implementa una funcion que ejecuta multiples validaciones y acumula TODOS los errores.
// Devuelve nil si no hay errores.
// Si hay errores, devuelve un error cuyo mensaje sea todos los errores separados por "; ".
// Ejemplo: "name: required; email: must contain @"

func ValidateAll(name, email string, age int) error {
	// TODO: implementar
	// Pista: acumula los errores en un slice y luego une con strings.Join
	panic("not implemented")
}
