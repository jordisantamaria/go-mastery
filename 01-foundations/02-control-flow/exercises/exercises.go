// Package exercises contiene funciones para practicar control flow en Go.
// Ejecuta los tests con: go test ./01-foundations/02-control-flow/exercises/...
package exercises

// --- Ejercicio 1: FizzBuzz ---
// Dado un numero, devuelve:
// - "FizzBuzz" si es divisible por 3 Y por 5
// - "Fizz" si es divisible por 3
// - "Buzz" si es divisible por 5
// - El numero como string en otro caso
// FizzBuzz(15) -> "FizzBuzz"
// FizzBuzz(9) -> "Fizz"
// FizzBuzz(10) -> "Buzz"
// FizzBuzz(7) -> "7"
func FizzBuzz(n int) string {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 2: Classify ---
// Clasifica una temperatura (Celsius) en una categoria:
// < 0: "freezing"
// 0-15: "cold"
// 16-25: "mild"
// 26-35: "warm"
// > 35: "hot"
// Pista: usa switch sin expresion.
func Classify(temp int) string {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 3: SumRange ---
// Calcula la suma de todos los enteros desde `from` hasta `to` (inclusive).
// SumRange(1, 5) -> 15  (1+2+3+4+5)
// SumRange(3, 3) -> 3
// Si from > to, devuelve 0.
func SumRange(from, to int) int {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 4: FindIndex ---
// Busca un elemento en un slice y devuelve su indice.
// Si no existe, devuelve -1.
// FindIndex([]string{"a", "b", "c"}, "b") -> 1
// FindIndex([]string{"a", "b", "c"}, "z") -> -1
func FindIndex(items []string, target string) int {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 5: MatrixSum ---
// Calcula la suma de todos los elementos en una matriz (slice de slices).
// MatrixSum([][]int{{1,2},{3,4}}) -> 10
// MatrixSum(nil) -> 0
func MatrixSum(matrix [][]int) int {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 6: Collatz ---
// Calcula cuantos pasos tarda la secuencia de Collatz en llegar a 1.
// Reglas: si n es par, n = n/2. Si n es impar, n = 3n+1. Repetir hasta n == 1.
// Collatz(1) -> 0
// Collatz(2) -> 1  (2 -> 1)
// Collatz(6) -> 8  (6 -> 3 -> 10 -> 5 -> 16 -> 8 -> 4 -> 2 -> 1)
// Si n <= 0, devuelve -1 (input invalido).
func Collatz(n int) int {
	// TODO: implementar
	panic("not implemented")
}
