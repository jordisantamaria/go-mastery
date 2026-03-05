// TU TAREA: implementar todos los tests de este archivo.
// Este modulo es al reves — el codigo existe, tu escribes los tests.
// Cada funcion Test tiene comentarios explicando que debes testear.
// Ejecuta: go test -v ./01-foundations/08-testing/exercises/...
package exercises

import (
	"testing"
)

// --- Ejercicio 1: TestFactorial ---
// Escribe un TABLE-DRIVEN TEST para Factorial.
// Casos que debes cubrir:
// - Factorial(0) = 1
// - Factorial(1) = 1
// - Factorial(5) = 120
// - Factorial(10) = 3628800
// - Factorial(-1) = error (debe wrappear ErrNegative)
// Usa t.Run para cada caso.

func TestFactorial(t *testing.T) {
	// TODO: implementar table-driven test
	t.Skip("TODO: implement this test")
}

// --- Ejercicio 2: TestPalindrome ---
// Escribe un table-driven test para Palindrome.
// Casos: "racecar" (true), "hello" (false), "Madam" (true, case-insensitive),
// "" (true), "a" (true), "ab" (false)

func TestPalindrome(t *testing.T) {
	// TODO: implementar
	t.Skip("TODO: implement this test")
}

// --- Ejercicio 3: TestTitleCase ---
// Escribe un table-driven test para TitleCase.
// Casos: "hello world" -> "Hello World", "" -> "", "HELLO" -> "Hello",
// "  extra   spaces  " -> "Extra Spaces"

func TestTitleCase(t *testing.T) {
	// TODO: implementar
	t.Skip("TODO: implement this test")
}

// --- Ejercicio 4: TestCacheService con Mock ---
// Crea un mockStringStore que implemente StringStore.
// Testea:
// - GetOrDefault devuelve el valor cuando existe
// - GetOrDefault devuelve el default cuando no existe
// - SetIfNotExists guarda cuando la key no existe
// - SetIfNotExists NO guarda cuando la key ya existe
//
// Pista: tu mock necesita un map interno y debe devolver ErrNotFound cuando la key no existe.

// TODO: crear mockStringStore struct que implemente StringStore

func TestCacheServiceGetOrDefault(t *testing.T) {
	// TODO: implementar con mock
	t.Skip("TODO: implement this test")
}

func TestCacheServiceSetIfNotExists(t *testing.T) {
	// TODO: implementar con mock
	t.Skip("TODO: implement this test")
}

// --- Ejercicio 5: Test Helper ---
// Crea una funcion helper assertError(t, err, target) que:
// - Usa t.Helper()
// - Verifica que err no es nil
// - Verifica que errors.Is(err, target) es true
// Luego usala en TestFactorial para los casos de error.

func assertError(t *testing.T, err error, target error) {
	// TODO: implementar
	t.Helper()
	t.Skip("TODO: implement this helper")
}

// --- Ejercicio 6: Benchmark ---
// Escribe un benchmark para Factorial(20).

func BenchmarkFactorial(b *testing.B) {
	// TODO: implementar
	b.Skip("TODO: implement this benchmark")
}

// --- Ejercicio 7: Fuzz ---
// Escribe un fuzz test para Palindrome que verifique:
// Si Palindrome(s) es true, entonces revertir s deberia dar el mismo resultado.

func FuzzPalindrome(f *testing.F) {
	// TODO: implementar
	f.Skip("TODO: implement this fuzz test")
}
