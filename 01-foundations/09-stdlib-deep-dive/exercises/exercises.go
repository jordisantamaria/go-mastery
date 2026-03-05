// Package exercises contiene funciones para practicar la standard library de Go.
// Ejecuta los tests con: go test ./01-foundations/09-stdlib-deep-dive/exercises/...
package exercises

import (
	"io"
	"time"
)

// --- Ejercicio 1: CountWords ---
// Cuenta la frecuencia de cada palabra en un Reader.
// Las palabras se separan por espacios/tabs/newlines (usa strings.Fields).
// Las palabras se convierten a minusculas para contar.
// CountWords(strings.NewReader("Go go GO")) -> map[string]int{"go": 3}
func CountWords(r io.Reader) (map[string]int, error) {
	// TODO: implementar
	panic("TODO")
}

// --- Ejercicio 2: LineCount ---
// Cuenta el numero de lineas en un Reader.
// Una linea se define por el caracter '\n'.
// Un Reader vacio devuelve 0.
// "hello\nworld\n" -> 2
// "hello\nworld" -> 2 (la ultima linea sin newline tambien cuenta)
// "" -> 0
func LineCount(r io.Reader) (int, error) {
	// TODO: implementar
	panic("TODO")
}

// --- Ejercicio 3: ToJSON ---
// Convierte cualquier valor a un string JSON con indentacion (pretty print).
// Usa json.MarshalIndent con prefix="" e indent="  " (2 espacios).
// ToJSON(map[string]int{"a": 1}) -> "{\n  \"a\": 1\n}"
func ToJSON(v any) (string, error) {
	// TODO: implementar
	panic("TODO")
}

// --- Ejercicio 4: FromJSON ---
// Deserializa un string JSON a un tipo T (generico).
// FromJSON[map[string]int](`{"a": 1}`) -> map[string]int{"a": 1}, nil
// FromJSON[int](`"invalid"`) -> 0, error
func FromJSON[T any](data string) (T, error) {
	// TODO: implementar
	panic("TODO")
}

// --- Ejercicio 5: CopyFile ---
// Copia un archivo de src a dst usando io.Copy.
// Devuelve el numero de bytes copiados.
// Si src no existe, devuelve error.
// Crea dst si no existe, lo sobreescribe si existe.
func CopyFile(src, dst string) (int64, error) {
	// TODO: implementar
	panic("TODO")
}

// --- Ejercicio 6: FindFiles ---
// Encuentra todos los archivos con la extension dada bajo el directorio root.
// La extension incluye el punto (ej: ".go", ".txt").
// Devuelve paths relativos al root, ordenados alfabeticamente.
// No incluye directorios, solo archivos regulares.
func FindFiles(root string, ext string) ([]string, error) {
	// TODO: implementar
	panic("TODO")
}

// --- Ejercicio 7: HTTPGet ---
// Hace un GET request a la URL dada con el timeout especificado.
// Usa context.WithTimeout para limitar la duracion del request.
// Devuelve el body como []byte.
// Si el status code no es 200, devuelve un error con el formato:
//
//	"unexpected status: <status_code>"
func HTTPGet(url string, timeout time.Duration) ([]byte, error) {
	// TODO: implementar
	panic("TODO")
}

// --- Ejercicio 8: MultiMerge ---
// Combina multiples io.Readers en un solo io.Reader usando io.MultiReader.
// Los readers se leen secuencialmente (primero r1, luego r2, etc.).
// Si no se pasan readers, devuelve un reader vacio.
func MultiMerge(readers ...io.Reader) io.Reader {
	// TODO: implementar
	panic("TODO")
}
