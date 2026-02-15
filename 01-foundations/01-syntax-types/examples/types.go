package main

import "fmt"

func main() {
	// =============================================
	// TIPOS NUMERICOS
	// =============================================

	var i int = 42         // tamanyo depende de la plataforma (64-bit en tu Mac)
	var f float64 = 3.14   // punto flotante — usa float64 por defecto
	var b byte = 'A'       // alias de uint8 — un byte
	var r rune = '🚀'      // alias de int32 — un code point Unicode

	fmt.Printf("int: %d, float: %.2f, byte: %c (%d), rune: %c (%d)\n",
		i, f, b, b, r, r)

	// =============================================
	// CONVERSIONES EXPLICITAS
	// =============================================

	x := 42
	y := float64(x)  // int -> float64
	z := int(y)       // float64 -> int (trunca, no redondea)

	fmt.Printf("int: %d -> float64: %f -> int: %d\n", x, y, z)

	// Esto NO compila:
	// var n int32 = 10
	// var m int64 = n  // ERROR: cannot use n (int32) as int64

	// =============================================
	// STRINGS: inmutables y UTF-8
	// =============================================

	s := "Hola 🌍"

	fmt.Println("String:", s)
	fmt.Println("len() en bytes:", len(s))  // 10 — porque el emoji ocupa 4 bytes

	// Iterar por bytes (normalmente NO quieres esto)
	fmt.Print("Bytes: ")
	for i := 0; i < len(s); i++ {
		fmt.Printf("%x ", s[i])
	}
	fmt.Println()

	// Iterar por runes (caracteres) — esto SI quieres
	fmt.Print("Runes: ")
	for _, r := range s {
		fmt.Printf("%c ", r)
	}
	fmt.Println()

	// Contar caracteres reales
	runes := []rune(s)
	fmt.Println("Caracteres reales:", len(runes)) // 6

	// =============================================
	// STRINGS: operaciones comunes
	// =============================================

	// Concatenacion
	greeting := "Hello" + " " + "World"
	fmt.Println(greeting)

	// String -> []byte -> String
	bs := []byte("hello")
	bs[0] = 'H'              // podemos modificar el byte slice
	modified := string(bs)
	fmt.Println(modified)     // "Hello"

	// Multi-line strings con backticks (raw strings)
	json := `{
    "name": "Jordi",
    "age": 28
}`
	fmt.Println(json)
}
