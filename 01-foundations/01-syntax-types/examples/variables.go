package main

import "fmt"

func main() {
	// =============================================
	// VARIABLES: las 3 formas de declarar
	// =============================================

	// 1. Declaracion explicita con tipo
	var name string = "Jordi"

	// 2. Declaracion con inferencia (var)
	var age = 28

	// 3. Declaracion corta — la mas comun dentro de funciones
	city := "Barcelona"

	fmt.Println(name, age, city)

	// =============================================
	// ZERO VALUES: todo tiene un valor por defecto
	// =============================================

	var (
		defaultInt    int
		defaultFloat  float64
		defaultString string
		defaultBool   bool
	)

	fmt.Printf("int: %d, float: %f, string: %q, bool: %t\n",
		defaultInt, defaultFloat, defaultString, defaultBool)
	// int: 0, float: 0.000000, string: "", bool: false

	// =============================================
	// CONSTANTES
	// =============================================

	const Pi = 3.14159
	const (
		StatusOK    = 200
		StatusError = 500
	)

	fmt.Println("Pi:", Pi, "Status:", StatusOK)

	// =============================================
	// MULTIPLES ASIGNACIONES
	// =============================================

	x, y := 10, 20
	fmt.Println("Antes:", x, y)

	// Swap sin variable temporal — idiomatico en Go
	x, y = y, x
	fmt.Println("Despues:", x, y)

	// =============================================
	// BLANK IDENTIFIER (_) — descartar valores
	// =============================================

	// Cuando una funcion devuelve multiples valores y no necesitas uno
	result, _ := divide(10, 3)
	fmt.Println("10/3 =", result)
}

func divide(a, b int) (int, int) {
	return a / b, a % b
}
