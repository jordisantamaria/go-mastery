package main

import (
	"encoding/json"
	"fmt"
	"math"
)

// =============================================
// STRUCTS + METHODS
// =============================================

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

// Stringer — controla como se imprime
func (c Circle) String() string {
	return fmt.Sprintf("Circle(r=%.2f)", c.Radius)
}

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

func (r *Rectangle) Scale(factor float64) {
	r.Width *= factor
	r.Height *= factor
}

func (r Rectangle) String() string {
	return fmt.Sprintf("Rect(%.0fx%.0f)", r.Width, r.Height)
}

// =============================================
// INTERFACES (implicitas)
// =============================================

type Shape interface {
	Area() float64
	Perimeter() float64
}

// printShape acepta cualquier Shape — polimorfismo
func printShape(s Shape) {
	fmt.Printf("  %v -> Area=%.2f Perimeter=%.2f\n", s, s.Area(), s.Perimeter())
}

// =============================================
// EMBEDDING (composicion)
// =============================================

type Address struct {
	Street string
	City   string
}

type Person struct {
	Name string
	Age  int
}

func (p Person) Greet() string {
	return fmt.Sprintf("Hi, I'm %s", p.Name)
}

type Employee struct {
	Person  // embedding
	Address // embedding
	Role    string
}

// =============================================
// STRUCT TAGS + JSON
// =============================================

type APIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

func main() {
	// --- Structs y methods ---
	c := Circle{Radius: 5}
	r := Rectangle{Width: 10, Height: 5}

	fmt.Println("=== Shapes ===")
	fmt.Println(c, "area:", c.Area())

	// Pointer receiver — modifica el original
	fmt.Println("Antes scale:", r)
	r.Scale(2)
	fmt.Println("Despues scale:", r)

	// --- Polimorfismo via interface ---
	fmt.Println("\n=== Polimorfismo ===")
	shapes := []Shape{
		Circle{Radius: 3},
		Rectangle{Width: 4, Height: 6},
		Circle{Radius: 1},
	}
	for _, s := range shapes {
		printShape(s)
	}

	// --- Embedding ---
	fmt.Println("\n=== Embedding ===")
	emp := Employee{
		Person:  Person{Name: "Jordi", Age: 28},
		Address: Address{Street: "Calle Mayor 1", City: "Barcelona"},
		Role:    "Backend Developer",
	}

	// Campos promovidos — acceso directo
	fmt.Println("Name:", emp.Name)       // emp.Person.Name
	fmt.Println("City:", emp.City)       // emp.Address.City
	fmt.Println("Role:", emp.Role)
	fmt.Println("Greet:", emp.Greet())   // method promovido

	// --- JSON serialization ---
	fmt.Println("\n=== JSON ===")
	resp := APIResponse{Status: "ok", Code: 200}
	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(data))
	// "message" se omite por omitempty (zero value "")

	resp2 := APIResponse{Status: "error", Message: "not found", Code: 404}
	data2, _ := json.MarshalIndent(resp2, "", "  ")
	fmt.Println(string(data2))

	// --- Type assertion y type switch ---
	fmt.Println("\n=== Type Switch ===")
	values := []any{42, "hello", true, Circle{Radius: 1}, nil}
	for _, v := range values {
		fmt.Printf("  %v -> %s\n", v, describe(v))
	}

	// --- Nil interface gotcha ---
	fmt.Println("\n=== Nil Interface Gotcha ===")
	var err error = nil
	fmt.Println("nil error:", err == nil) // true

	var p *Person = nil
	// err = p  // si hicieramos esto: err != nil seria TRUE aunque p sea nil
	_ = p
}

func describe(i any) string {
	switch v := i.(type) {
	case int:
		return fmt.Sprintf("int(%d)", v)
	case string:
		return fmt.Sprintf("string(%q)", v)
	case bool:
		return fmt.Sprintf("bool(%t)", v)
	case Shape:
		return fmt.Sprintf("Shape(area=%.2f)", v.Area())
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("unknown(%T)", v)
	}
}
