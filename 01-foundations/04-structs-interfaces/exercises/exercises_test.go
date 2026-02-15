package exercises

import (
	"strings"
	"testing"
)

// =============================================
// BankAccount tests
// =============================================

func TestNewBankAccount(t *testing.T) {
	acc := NewBankAccount("Jordi", 1000)
	if acc.Owner() != "Jordi" {
		t.Errorf("Owner() = %q, want %q", acc.Owner(), "Jordi")
	}
	if acc.Balance() != 1000 {
		t.Errorf("Balance() = %.2f, want 1000.00", acc.Balance())
	}
}

func TestDeposit(t *testing.T) {
	acc := NewBankAccount("Jordi", 100)

	err := acc.Deposit(50)
	if err != nil {
		t.Errorf("Deposit(50) unexpected error: %v", err)
	}
	if acc.Balance() != 150 {
		t.Errorf("Balance after deposit = %.2f, want 150.00", acc.Balance())
	}

	// Deposito negativo
	err = acc.Deposit(-10)
	if err == nil {
		t.Error("Deposit(-10) should return error")
	}

	// Deposito cero
	err = acc.Deposit(0)
	if err == nil {
		t.Error("Deposit(0) should return error")
	}
}

func TestWithdraw(t *testing.T) {
	acc := NewBankAccount("Jordi", 100)

	err := acc.Withdraw(30)
	if err != nil {
		t.Errorf("Withdraw(30) unexpected error: %v", err)
	}
	if acc.Balance() != 70 {
		t.Errorf("Balance after withdraw = %.2f, want 70.00", acc.Balance())
	}

	// Saldo insuficiente
	err = acc.Withdraw(100)
	if err == nil {
		t.Error("Withdraw(100) should return error (insufficient funds)")
	}
	if acc.Balance() != 70 {
		t.Errorf("Balance should not change after failed withdraw: %.2f", acc.Balance())
	}

	// Retiro negativo
	err = acc.Withdraw(-10)
	if err == nil {
		t.Error("Withdraw(-10) should return error")
	}
}

func TestBankAccountString(t *testing.T) {
	acc := NewBankAccount("Jordi", 1500)
	got := acc.String()
	want := "BankAccount{owner: Jordi, balance: 1500.00}"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

// =============================================
// Shape tests
// =============================================

func TestTriangle(t *testing.T) {
	tri := Triangle{Base: 10, Height: 5, SideA: 10, SideB: 7, SideC: 7}

	if got := tri.Area(); got != 25 {
		t.Errorf("Triangle.Area() = %.2f, want 25.00", got)
	}
	if got := tri.Perimeter(); got != 24 {
		t.Errorf("Triangle.Perimeter() = %.2f, want 24.00", got)
	}

	// Verificar que satisface Shape
	var _ Shape = tri
}

func TestSquare(t *testing.T) {
	sq := Square{Side: 4}

	if got := sq.Area(); got != 16 {
		t.Errorf("Square.Area() = %.2f, want 16.00", got)
	}
	if got := sq.Perimeter(); got != 16 {
		t.Errorf("Square.Perimeter() = %.2f, want 16.00", got)
	}

	// Verificar que satisface Shape
	var _ Shape = sq
}

func TestTotalArea(t *testing.T) {
	shapes := []Shape{
		Square{Side: 5},                                              // 25
		Triangle{Base: 10, Height: 4, SideA: 10, SideB: 5, SideC: 5}, // 20
		Square{Side: 3},                                              // 9
	}

	got := TotalArea(shapes)
	want := 54.0
	if got != want {
		t.Errorf("TotalArea() = %.2f, want %.2f", got, want)
	}

	// Slice vacio
	if got := TotalArea(nil); got != 0 {
		t.Errorf("TotalArea(nil) = %.2f, want 0", got)
	}
}

// =============================================
// Embedding + Override tests
// =============================================

func TestLoggingAccount(t *testing.T) {
	la := NewLoggingAccount("Alice", 500)

	// Debe funcionar como BankAccount
	if la.Balance() != 500 {
		t.Errorf("Balance() = %.2f, want 500.00", la.Balance())
	}
	if la.Owner() != "Alice" {
		t.Errorf("Owner() = %q, want %q", la.Owner(), "Alice")
	}

	// Deposit con logging
	err := la.Deposit(100)
	if err != nil {
		t.Errorf("Deposit(100) unexpected error: %v", err)
	}
	if la.Balance() != 600 {
		t.Errorf("Balance after deposit = %.2f, want 600.00", la.Balance())
	}

	// Withdraw sigue funcionando (promovido)
	err = la.Withdraw(50)
	if err != nil {
		t.Errorf("Withdraw(50) unexpected error: %v", err)
	}
	if la.Balance() != 550 {
		t.Errorf("Balance after withdraw = %.2f, want 550.00", la.Balance())
	}
}

// =============================================
// Describe (type switch) tests
// =============================================

func TestDescribe(t *testing.T) {
	tests := []struct {
		input any
		want  string
	}{
		{42, "integer: 42"},
		{0, "integer: 0"},
		{"hello", "string: hello (length 5)"},
		{"", "string:  (length 0)"},
		{true, "boolean: true"},
		{false, "boolean: false"},
		{Square{Side: 5}, "shape with area 25.00"},
		{nil, "nothing"},
		{3.14, "unknown: float64"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := Describe(tt.input)
			if !strings.Contains(got, tt.want) && got != tt.want {
				t.Errorf("Describe(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
