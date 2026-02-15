package main

import (
	"fmt"
	"sort"
	"strings"
)

func main() {
	// =============================================
	// MULTIPLE RETURN VALUES
	// =============================================

	result, remainder := divmod(17, 5)
	fmt.Printf("17 / 5 = %d remainder %d\n", result, remainder)

	// Con error handling
	val, err := safeDivide(10, 0)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Resultado:", val)
	}

	// =============================================
	// VARIADIC FUNCTIONS
	// =============================================

	fmt.Println("Sum:", sum(1, 2, 3, 4, 5))
	fmt.Println("Sum:", sum()) // 0 — funciona con 0 args

	// Pasar un slice
	numbers := []int{10, 20, 30}
	fmt.Println("Sum slice:", sum(numbers...))

	// =============================================
	// FUNCIONES COMO VALORES
	// =============================================

	// Asignar funcion a variable
	double := func(n int) int { return n * 2 }
	triple := func(n int) int { return n * 3 }

	fmt.Println("double(5):", double(5))
	fmt.Println("triple(5):", triple(5))

	// Pasar funcion como argumento
	fmt.Println("apply double:", applyToSlice([]int{1, 2, 3}, double))
	fmt.Println("apply triple:", applyToSlice([]int{1, 2, 3}, triple))

	// =============================================
	// CLOSURES
	// =============================================

	// Counter con closure
	next := counter()
	fmt.Println("Counter:", next(), next(), next()) // 1 2 3

	// Cada closure tiene su propio estado
	next2 := counter()
	fmt.Println("Counter2:", next2()) // 1

	// Closure que acumula
	acc := accumulator(100)
	fmt.Println("Acumulador:", acc(10)) // 110
	fmt.Println("Acumulador:", acc(20)) // 130
	fmt.Println("Acumulador:", acc(-5)) // 125

	// =============================================
	// SORT con funcion anonima
	// =============================================

	people := []struct {
		Name string
		Age  int
	}{
		{"Alice", 30},
		{"Bob", 25},
		{"Charlie", 35},
	}

	// Ordenar por edad usando closure
	sort.Slice(people, func(i, j int) bool {
		return people[i].Age < people[j].Age
	})
	fmt.Println("Sorted by age:", people)

	// =============================================
	// MIDDLEWARE PATTERN
	// =============================================

	// Componer funciones
	shout := compose(strings.ToUpper, addExclamation)
	fmt.Println(shout("hello")) // HELLO!
}

// --- Funciones auxiliares ---

func divmod(a, b int) (int, int) {
	return a / b, a % b
}

func safeDivide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("cannot divide %f by zero", a)
	}
	return a / b, nil
}

func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Recibe una funcion como argumento — higher-order function
func applyToSlice(nums []int, fn func(int) int) []int {
	result := make([]int, len(nums))
	for i, n := range nums {
		result[i] = fn(n)
	}
	return result
}

// Devuelve una closure
func counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

// Closure con estado inicial
func accumulator(initial int) func(int) int {
	total := initial
	return func(n int) int {
		total += n
		return total
	}
}

func addExclamation(s string) string {
	return s + "!"
}

// Composicion de funciones string -> string
func compose(f, g func(string) string) func(string) string {
	return func(s string) string {
		return g(f(s))
	}
}
