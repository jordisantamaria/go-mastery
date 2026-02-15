// Package exercises contiene funciones para practicar tipos y sintaxis basica de Go.
// Cada funcion tiene un test correspondiente. Implementa el cuerpo de cada funcion
// y ejecuta los tests con: go test ./01-foundations/01-syntax-types/exercises/...
package exercises

// --- Ejercicio 1: Zero Values ---
// Devuelve los zero values de los tipos basicos de Go.
// Pista: simplemente declara variables sin inicializar.
func ZeroValues() (int, float64, string, bool) {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 2: Swap ---
// Devuelve los dos valores intercambiados.
// Ejemplo: Swap(1, 2) -> (2, 1)
func Swap(a, b int) (int, int) {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 3: RuneCount ---
// Cuenta el numero REAL de caracteres (runes) en un string.
// len("Hola 🌍") devuelve 10 (bytes), pero tiene 6 caracteres.
// Pista: convierte a []rune.
func RuneCount(s string) int {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 4: SumSlice ---
// Calcula la suma de todos los elementos de un slice de enteros.
// SumSlice([]int{1, 2, 3}) -> 6
// SumSlice(nil) -> 0
func SumSlice(nums []int) int {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 5: UniqueStrings ---
// Dado un slice de strings, devuelve un nuevo slice con solo los valores unicos.
// El orden no importa.
// UniqueStrings([]string{"a", "b", "a", "c", "b"}) -> ["a", "b", "c"]
// Pista: usa un map como set.
func UniqueStrings(items []string) []string {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 6: WordCount ---
// Cuenta cuantas veces aparece cada palabra en un string.
// Devuelve un map[string]int.
// WordCount("hello world hello") -> {"hello": 2, "world": 1}
// Pista: usa strings.Fields() para separar por espacios.
func WordCount(s string) map[string]int {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 7: ReverseSlice ---
// Devuelve una nueva copia del slice con los elementos en orden inverso.
// NO modifica el slice original.
// ReverseSlice([]int{1, 2, 3}) -> [3, 2, 1]
func ReverseSlice(nums []int) []int {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 8: MergeMaps ---
// Fusiona dos maps. Si una clave existe en ambos, el valor de b tiene prioridad.
// MergeMaps({"a":1, "b":2}, {"b":3, "c":4}) -> {"a":1, "b":3, "c":4}
func MergeMaps(a, b map[string]int) map[string]int {
	// TODO: implementar
	panic("not implemented")
}
