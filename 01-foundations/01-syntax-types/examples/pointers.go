package main

import "fmt"

func main() {
	// =============================================
	// PUNTEROS: basico
	// =============================================

	x := 42
	p := &x // p es *int — un puntero a int

	fmt.Println("Valor de x:", x)
	fmt.Println("Direccion de x:", p)
	fmt.Println("Valor apuntado por p:", *p)

	// Modificar via puntero
	*p = 100
	fmt.Println("x despues de *p = 100:", x) // 100

	// =============================================
	// PUNTEROS EN FUNCIONES
	// =============================================

	// Go pasa todo por valor (copia).
	// Sin puntero, la funcion trabaja con una copia.

	a := 5
	doubleByValue(a)
	fmt.Println("Despues de doubleByValue:", a) // 5 — no cambio

	doubleByPointer(&a)
	fmt.Println("Despues de doubleByPointer:", a) // 10 — si cambio

	// =============================================
	// PUNTEROS: new()
	// =============================================

	// new(T) aloca memoria para T y devuelve un *T con zero value
	n := new(int)
	fmt.Println("new(int):", *n) // 0
	*n = 42
	fmt.Println("Despues:", *n)  // 42

	// =============================================
	// nil POINTER
	// =============================================

	var ptr *int // nil — no apunta a nada
	fmt.Println("nil pointer:", ptr)

	// Desreferenciar nil pointer -> PANIC
	// fmt.Println(*ptr)  // panic: runtime error: invalid memory address

	// Siempre verifica antes de desreferenciar
	if ptr != nil {
		fmt.Println("Valor:", *ptr)
	} else {
		fmt.Println("El puntero es nil")
	}
}

// Recibe una COPIA de n — no modifica el original
func doubleByValue(n int) {
	n *= 2
}

// Recibe un PUNTERO a n — modifica el original
func doubleByPointer(n *int) {
	*n *= 2
}
