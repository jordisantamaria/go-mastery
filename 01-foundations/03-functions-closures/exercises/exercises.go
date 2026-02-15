// Package exercises contiene funciones para practicar functions y closures en Go.
// Ejecuta los tests con: go test ./01-foundations/03-functions-closures/exercises/...
package exercises

// --- Ejercicio 1: Apply ---
// Aplica una funcion transformadora a cada elemento de un slice de ints.
// Devuelve un nuevo slice con los resultados.
// Apply([]int{1, 2, 3}, func(n int) int { return n * 2 }) -> [2, 4, 6]
func Apply(nums []int, fn func(int) int) []int {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 2: Filter ---
// Filtra un slice usando una funcion predicado.
// Devuelve un nuevo slice con solo los elementos que cumplen el predicado.
// Filter([]int{1, 2, 3, 4, 5}, func(n int) bool { return n%2 == 0 }) -> [2, 4]
func Filter(nums []int, predicate func(int) bool) []int {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 3: Reduce ---
// Reduce un slice a un solo valor usando una funcion acumuladora.
// Reduce([]int{1, 2, 3, 4}, 0, func(acc, n int) int { return acc + n }) -> 10
func Reduce(nums []int, initial int, fn func(int, int) int) int {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 4: MakeMultiplier ---
// Devuelve una closure que multiplica su argumento por `factor`.
// double := MakeMultiplier(2)
// double(5) -> 10
// triple := MakeMultiplier(3)
// triple(5) -> 15
func MakeMultiplier(factor int) func(int) int {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 5: MakeCounter ---
// Devuelve dos closures: increment y getCount.
// increment() suma 1 al contador interno.
// getCount() devuelve el valor actual.
// inc, get := MakeCounter()
// inc(); inc(); inc()
// get() -> 3
func MakeCounter() (increment func(), getCount func() int) {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 6: Compose ---
// Compone dos funciones: primero aplica f, luego g al resultado.
// Compose(double, addOne)(3) -> 7  (primero 3*2=6, luego 6+1=7)
func Compose(f, g func(int) int) func(int) int {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 7: Memoize ---
// Devuelve una version memoizada de una funcion.
// La primera vez que se llama con un argumento, ejecuta la funcion y guarda el resultado.
// Las siguientes veces con el mismo argumento, devuelve el resultado cacheado.
// Pista: usa un map[int]int como cache dentro de la closure.
// slow := func(n int) int { return n * n }
// fast := Memoize(slow)
// fast(5) -> 25 (calcula)
// fast(5) -> 25 (cache)
func Memoize(fn func(int) int) func(int) int {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 8: Pipeline ---
// Recibe un slice de funciones y devuelve una funcion que las aplica en cadena.
// Pipeline(double, addOne, triple)(2) -> 15
// Explicacion: 2 -> double -> 4 -> addOne -> 5 -> triple -> 15
func Pipeline(fns ...func(int) int) func(int) int {
	// TODO: implementar
	panic("not implemented")
}
