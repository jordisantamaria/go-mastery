package main

import "fmt"

func main() {
	// =============================================
	// IF con statement inicial
	// =============================================

	// Patron tipico: asignar y chequear en la misma linea
	if num := 42; num%2 == 0 {
		fmt.Println(num, "es par")
	}
	// num no existe aqui — su scope es solo el if/else

	// =============================================
	// SWITCH sin expresion (reemplaza if/else chains)
	// =============================================

	temperature := 25
	switch {
	case temperature < 0:
		fmt.Println("Helando")
	case temperature < 15:
		fmt.Println("Frio")
	case temperature < 25:
		fmt.Println("Agradable")
	default:
		fmt.Println("Calor")
	}

	// =============================================
	// FOR: todas las variantes
	// =============================================

	// Clasico (estilo C)
	fmt.Print("Clasico: ")
	for i := 0; i < 5; i++ {
		fmt.Print(i, " ")
	}
	fmt.Println()

	// While-style
	fmt.Print("While: ")
	n := 1
	for n < 32 {
		fmt.Print(n, " ")
		n *= 2
	}
	fmt.Println()

	// Range sobre slice
	fruits := []string{"manzana", "banana", "cereza"}
	fmt.Print("Range: ")
	for i, fruit := range fruits {
		fmt.Printf("[%d]%s ", i, fruit)
	}
	fmt.Println()

	// Range sobre map
	capitals := map[string]string{
		"Spain":  "Madrid",
		"France": "Paris",
		"Italy":  "Rome",
	}
	for country, capital := range capitals {
		fmt.Printf("%s -> %s\n", country, capital)
	}

	// =============================================
	// DEFER: LIFO y evaluacion inmediata
	// =============================================

	fmt.Println("--- Defer demo ---")
	fmt.Println("Start")
	defer fmt.Println("Deferred 1")
	defer fmt.Println("Deferred 2")
	defer fmt.Println("Deferred 3")
	fmt.Println("End")
	// Output: Start, End, Deferred 3, Deferred 2, Deferred 1

	// =============================================
	// LABELS: break en loops anidados
	// =============================================

	fmt.Println("--- Labels demo ---")
	matrix := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	target := 5
search:
	for i, row := range matrix {
		for j, val := range row {
			if val == target {
				fmt.Printf("Encontrado %d en [%d][%d]\n", target, i, j)
				break search
			}
		}
	}
}
