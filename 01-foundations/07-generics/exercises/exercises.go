// Package exercises contiene funciones para practicar generics en Go.
// Ejecuta los tests con: go test ./01-foundations/07-generics/exercises/...
package exercises

import "cmp"

// --- Ejercicio 1: Min ---
// Devuelve el menor de dos valores ordenables.
// Min(3, 5) -> 3
// Min("apple", "banana") -> "apple"
func Min[T cmp.Ordered](a, b T) T {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 2: Clamp ---
// Limita un valor entre min y max.
// Clamp(5, 1, 10) -> 5
// Clamp(-5, 0, 100) -> 0
// Clamp(200, 0, 100) -> 100
func Clamp[T cmp.Ordered](value, lo, hi T) T {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 3: Map ---
// Aplica una funcion a cada elemento de un slice y devuelve un nuevo slice.
// Map([]int{1,2,3}, func(n int) int { return n*2 }) -> [2,4,6]
// Map([]string{"a","b"}, func(s string) int { return len(s) }) -> [1,1]
func Map[T any, U any](items []T, fn func(T) U) []U {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 4: Filter ---
// Filtra un slice usando un predicado.
// Filter([]int{1,2,3,4,5}, func(n int) bool { return n%2==0 }) -> [2,4]
func Filter[T any](items []T, pred func(T) bool) []T {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 5: Reduce ---
// Reduce un slice a un valor usando una funcion acumuladora.
// Reduce([]int{1,2,3,4}, 0, func(acc,n int) int { return acc+n }) -> 10
func Reduce[T any, U any](items []T, initial U, fn func(U, T) U) U {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 6: Contains ---
// Verifica si un slice contiene un elemento.
// Requiere comparable constraint.
// Contains([]int{1,2,3}, 2) -> true
func Contains[T comparable](items []T, target T) bool {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 7: Uniq ---
// Elimina duplicados de un slice, manteniendo el orden de primera aparicion.
// Uniq([]int{1,2,2,3,1,4}) -> [1,2,3,4]
// Uniq([]string{"a","b","a"}) -> ["a","b"]
func Uniq[T comparable](items []T) []T {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 8: GroupBy ---
// Agrupa elementos de un slice por una clave derivada de cada elemento.
// GroupBy([]string{"apple","avocado","banana","blueberry"},
//   func(s string) byte { return s[0] })
// -> map[byte][]string{'a': ["apple","avocado"], 'b': ["banana","blueberry"]}
func GroupBy[T any, K comparable](items []T, keyFn func(T) K) map[K][]T {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 9: Stack ---
// Implementa un stack generico con Push, Pop, Peek, Len, y IsEmpty.

type Stack[T any] struct {
	// TODO: definir campos
}

func (s *Stack[T]) Push(item T) {
	panic("not implemented")
}

func (s *Stack[T]) Pop() (T, bool) {
	panic("not implemented")
}

func (s *Stack[T]) Peek() (T, bool) {
	panic("not implemented")
}

func (s *Stack[T]) Len() int {
	panic("not implemented")
}

func (s *Stack[T]) IsEmpty() bool {
	panic("not implemented")
}

// --- Ejercicio 10: MapKeys / MapValues ---
// Extrae las claves y valores de un map generico.

func MapKeys[K comparable, V any](m map[K]V) []K {
	// TODO: implementar
	panic("not implemented")
}

func MapValues[K comparable, V any](m map[K]V) []V {
	// TODO: implementar
	panic("not implemented")
}
