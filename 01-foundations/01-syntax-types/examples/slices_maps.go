package main

import "fmt"

func main() {
	// =============================================
	// ARRAYS — tamanyo fijo (rara vez se usan directamente)
	// =============================================

	arr := [3]int{10, 20, 30}
	fmt.Println("Array:", arr, "len:", len(arr))

	// El tamanyo es parte del tipo: [3]int != [5]int
	// var other [5]int = arr  // NO compila

	// =============================================
	// SLICES — lo que usaras siempre
	// =============================================

	// Crear con literal
	nums := []int{1, 2, 3, 4, 5}
	fmt.Println("Slice:", nums)

	// Slicing — crea una VISTA, no una copia
	sub := nums[1:3] // [2, 3] — incluye indice 1, excluye 3
	fmt.Println("Sub-slice:", sub)

	// Modificar el sub-slice afecta al original!
	sub[0] = 999
	fmt.Println("Original despues de modificar sub:", nums) // [1, 999, 3, 4, 5]

	// Append
	nums = append(nums, 6, 7)
	fmt.Println("Despues de append:", nums)

	// Make — crear con tamanyo y capacidad
	s := make([]int, 3, 10)
	fmt.Printf("make([]int, 3, 10): len=%d cap=%d %v\n", len(s), cap(s), s)

	// Patron comun: empezar vacio con capacidad conocida
	names := make([]string, 0, 5)
	names = append(names, "Alice", "Bob", "Charlie")
	fmt.Println("Names:", names, "len:", len(names), "cap:", cap(names))

	// =============================================
	// COPY — para hacer una copia independiente
	// =============================================

	original := []int{1, 2, 3}
	copia := make([]int, len(original))
	copy(copia, original)
	copia[0] = 999
	fmt.Println("Original:", original) // [1, 2, 3] — no afectado
	fmt.Println("Copia:", copia)       // [999, 2, 3]

	// =============================================
	// MAPS — tablas hash
	// =============================================

	// Crear con literal
	ages := map[string]int{
		"alice": 30,
		"bob":   25,
	}

	// Leer con comma-ok pattern
	age, exists := ages["alice"]
	fmt.Printf("alice: %d, exists: %t\n", age, exists)

	age, exists = ages["nobody"]
	fmt.Printf("nobody: %d, exists: %t\n", age, exists) // 0, false

	// Escribir
	ages["charlie"] = 35

	// Borrar
	delete(ages, "bob")

	// Iterar (el orden NO esta garantizado)
	fmt.Println("Todos:")
	for name, age := range ages {
		fmt.Printf("  %s: %d\n", name, age)
	}

	// Contar
	fmt.Println("Total:", len(ages))

	// =============================================
	// nil MAP vs empty MAP
	// =============================================

	var nilMap map[string]int         // nil
	emptyMap := map[string]int{}      // empty
	madeMap := make(map[string]int)   // empty (equivalente)

	// Leer de nil map esta OK
	fmt.Println("nil map read:", nilMap["key"]) // 0

	// Escribir en nil map -> PANIC
	// nilMap["key"] = 1  // panic: assignment to entry in nil map

	_ = emptyMap
	_ = madeMap
}
