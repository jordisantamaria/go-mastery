package main

import (
	"cmp"
	"fmt"
	"strings"
)

// =============================================
// CONSTRAINTS
// =============================================

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64
}

// =============================================
// FUNCIONES GENERICAS
// =============================================

func Sum[T Number](nums []T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}

func Min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Contains[T comparable](items []T, target T) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

// =============================================
// MAP, FILTER, REDUCE genericos
// =============================================

func Map[T any, U any](items []T, fn func(T) U) []U {
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = fn(item)
	}
	return result
}

func Filter[T any](items []T, pred func(T) bool) []T {
	var result []T
	for _, item := range items {
		if pred(item) {
			result = append(result, item)
		}
	}
	return result
}

func Reduce[T any, U any](items []T, initial U, fn func(U, T) U) U {
	acc := initial
	for _, item := range items {
		acc = fn(acc, item)
	}
	return acc
}

// =============================================
// TIPO GENERICO: Stack
// =============================================

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, true
}

func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], false
}

func (s *Stack[T]) Len() int {
	return len(s.items)
}

// =============================================
// KEYS / VALUES generico para maps
// =============================================

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// =============================================
// ~ operator: custom types
// =============================================

type Celsius float64

func main() {
	// --- Funciones genericas basicas ---
	fmt.Println("=== Generic Functions ===")
	fmt.Println("Sum ints:", Sum([]int{1, 2, 3, 4, 5}))
	fmt.Println("Sum floats:", Sum([]float64{1.1, 2.2, 3.3}))

	fmt.Println("Min:", Min(10, 3))
	fmt.Println("Max:", Max("apple", "banana")) // strings tambien son Ordered

	fmt.Println("Contains 3:", Contains([]int{1, 2, 3}, 3))
	fmt.Println("Contains z:", Contains([]string{"a", "b"}, "z"))

	// --- Map, Filter, Reduce ---
	fmt.Println("\n=== Map / Filter / Reduce ===")
	names := []string{"Alice", "Bob", "Charlie", "David"}

	lengths := Map(names, func(s string) int { return len(s) })
	fmt.Println("Lengths:", lengths) // [5, 3, 7, 5]

	upper := Map(names, strings.ToUpper)
	fmt.Println("Upper:", upper) // [ALICE, BOB, CHARLIE, DAVID]

	long := Filter(names, func(s string) bool { return len(s) > 4 })
	fmt.Println("Long names:", long) // [Alice, Charlie, David]

	totalLen := Reduce(names, 0, func(acc int, s string) int { return acc + len(s) })
	fmt.Println("Total length:", totalLen) // 20

	// Encadenar: nombres largos en mayusculas
	result := Map(
		Filter(names, func(s string) bool { return len(s) > 4 }),
		strings.ToUpper,
	)
	fmt.Println("Long + Upper:", result) // [ALICE, CHARLIE, DAVID]

	// --- Stack generico ---
	fmt.Println("\n=== Generic Stack ===")
	intStack := &Stack[int]{}
	intStack.Push(10)
	intStack.Push(20)
	intStack.Push(30)
	fmt.Println("Stack len:", intStack.Len())

	val, ok := intStack.Pop()
	fmt.Printf("Pop: %d (ok=%t), len: %d\n", val, ok, intStack.Len())

	strStack := &Stack[string]{}
	strStack.Push("hello")
	strStack.Push("world")
	s, _ := strStack.Pop()
	fmt.Println("String stack pop:", s)

	// --- ~ operator ---
	fmt.Println("\n=== ~ Operator ===")
	temps := []Celsius{36.6, 37.0, 38.5}
	avgTemp := Sum(temps) / Celsius(len(temps))
	fmt.Printf("Average temp: %.1f°C\n", avgTemp)

	// --- Keys generico ---
	fmt.Println("\n=== Generic Keys ===")
	ages := map[string]int{"Alice": 30, "Bob": 25, "Charlie": 35}
	fmt.Println("Keys:", Keys(ages))
}
