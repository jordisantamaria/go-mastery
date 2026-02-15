// Package exercises contiene funciones para practicar structs e interfaces en Go.
// Ejecuta los tests con: go test ./01-foundations/04-structs-interfaces/exercises/...
package exercises

import "fmt"

// --- Ejercicio 1: BankAccount ---
// Implementa un BankAccount con los methods indicados.
// - NewBankAccount(owner string, initial float64) crea una cuenta
// - Deposit(amount float64) error — anyadir dinero (error si amount <= 0)
// - Withdraw(amount float64) error — sacar dinero (error si amount <= 0 o saldo insuficiente)
// - Balance() float64 — devuelve el saldo actual
// - Owner() string — devuelve el propietario

type BankAccount struct {
	// TODO: definir campos
}

func NewBankAccount(owner string, initial float64) *BankAccount {
	// TODO: implementar
	panic("not implemented")
}

func (a *BankAccount) Deposit(amount float64) error {
	// TODO: implementar
	panic("not implemented")
}

func (a *BankAccount) Withdraw(amount float64) error {
	// TODO: implementar
	panic("not implemented")
}

func (a *BankAccount) Balance() float64 {
	// TODO: implementar
	panic("not implemented")
}

func (a *BankAccount) Owner() string {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 2: Stringer ---
// Implementa String() para que BankAccount se imprima como:
// "BankAccount{owner: Jordi, balance: 1500.00}"
func (a *BankAccount) String() string {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 3: Shape interface ---
// Implementa las structs Triangle y Square que satisfagan esta interface.

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Triangle struct {
	// TODO: definir campos (Base, Height, SideA, SideB, SideC)
	// SideA, SideB, SideC son los tres lados para el perimetro
}

type Square struct {
	// TODO: definir campo (Side)
}

func (t Triangle) Area() float64 {
	// Area = (base * height) / 2
	// TODO: implementar
	panic("not implemented")
}

func (t Triangle) Perimeter() float64 {
	// TODO: implementar
	panic("not implemented")
}

func (s Square) Area() float64 {
	// TODO: implementar
	panic("not implemented")
}

func (s Square) Perimeter() float64 {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 4: TotalArea ---
// Calcula el area total de un slice de Shapes.
// TotalArea([]Shape{Square{5}, Triangle{...}}) -> suma de areas
func TotalArea(shapes []Shape) float64 {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 5: Embedding + Override ---
// Crea una struct LoggingAccount que embeba BankAccount y override Deposit
// para que imprima un mensaje antes de depositar.
// El mensaje debe ser: "Depositing %.2f to %s's account\n"

type LoggingAccount struct {
	// TODO: embeber BankAccount
}

func NewLoggingAccount(owner string, initial float64) *LoggingAccount {
	// TODO: implementar
	panic("not implemented")
}

// Override Deposit para anyadir logging
func (la *LoggingAccount) Deposit(amount float64) error {
	// TODO: imprimir mensaje y llamar al Deposit original
	// fmt.Printf("Depositing %.2f to %s's account\n", amount, la.Owner())
	panic("not implemented")
}

// --- Ejercicio 6: Describer interface ---
// Implementa una funcion Describe que use type switch para describir valores:
// - int: "integer: <value>"
// - string: "string: <value> (length <len>)"
// - bool: "boolean: <value>"
// - Shape: "shape with area <area>"
// - nil: "nothing"
// - default: "unknown: <type>"
func Describe(v any) string {
	// TODO: implementar con type switch
	_ = fmt.Sprintf // puedes usar fmt.Sprintf
	panic("not implemented")
}
