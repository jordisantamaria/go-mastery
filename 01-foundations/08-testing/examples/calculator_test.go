// Este archivo demuestra TODOS los patrones de testing de Go.
// Es el "ejemplo perfecto" que puedes referenciar.
// Ejecutar: go test -v ./01-foundations/08-testing/examples/...
package calculator

import (
	"errors"
	"math"
	"testing"
)

// =============================================
// TABLE-DRIVEN TEST — el patron estandar
// =============================================

func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want float64
	}{
		{"positive", 2, 3, 5},
		{"negative", -1, -2, -3},
		{"zero", 0, 0, 0},
		{"mixed", -1.5, 2.5, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Add(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Add(%g, %g) = %g, want %g", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// =============================================
// TEST con ERROR checking
// =============================================

func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr error
	}{
		{"normal", 10, 2, 5, nil},
		{"decimal", 7, 2, 3.5, nil},
		{"division by zero", 10, 0, 0, ErrDivisionByZero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Divide(tt.a, tt.b)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Divide(%g, %g) error = %v, want %v",
						tt.a, tt.b, err, tt.wantErr)
				}
				return // no verificar el valor si esperamos error
			}

			if err != nil {
				t.Fatalf("Divide(%g, %g) unexpected error: %v", tt.a, tt.b, err)
			}
			if got != tt.want {
				t.Errorf("Divide(%g, %g) = %g, want %g", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// =============================================
// TEST con EPSILON (floating point comparison)
// =============================================

func TestSqrt(t *testing.T) {
	const epsilon = 1e-10

	tests := []struct {
		name    string
		input   float64
		want    float64
		wantErr bool
	}{
		{"sqrt 4", 4, 2, false},
		{"sqrt 9", 9, 3, false},
		{"sqrt 2", 2, math.Sqrt(2), false},
		{"sqrt 0", 0, 0, false},
		{"negative", -1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sqrt(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Sqrt(%g) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("Sqrt(%g) unexpected error: %v", tt.input, err)
			}

			// Floating point: comparar con epsilon
			if diff := math.Abs(got - tt.want); diff > epsilon {
				t.Errorf("Sqrt(%g) = %g, want %g (diff %g)", tt.input, got, tt.want, diff)
			}
		})
	}
}

// =============================================
// HELPER function con t.Helper()
// =============================================

func assertFloat(t *testing.T, got, want float64) {
	t.Helper()
	if got != want {
		t.Errorf("got %g, want %g", got, want)
	}
}

func TestMultiply(t *testing.T) {
	assertFloat(t, Multiply(2, 3), 6)
	assertFloat(t, Multiply(-1, 5), -5)
	assertFloat(t, Multiply(0, 100), 0)
}

// =============================================
// PARALLEL tests
// =============================================

func TestIsPrime(t *testing.T) {
	tests := []struct {
		n    int
		want bool
	}{
		{0, false},
		{1, false},
		{2, true},
		{3, true},
		{4, false},
		{17, true},
		{100, false},
		{997, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel() // cada subtest corre en paralelo
			got := IsPrime(tt.n)
			if got != tt.want {
				t.Errorf("IsPrime(%d) = %t, want %t", tt.n, got, tt.want)
			}
		})
	}
}

// =============================================
// BENCHMARK
// =============================================

func BenchmarkIsPrime(b *testing.B) {
	for b.Loop() {
		IsPrime(997)
	}
}

func BenchmarkIsPrimeLarge(b *testing.B) {
	for b.Loop() {
		IsPrime(999983) // primo grande
	}
}

// =============================================
// FUZZ test
// =============================================

func FuzzReverse(f *testing.F) {
	f.Add("hello")
	f.Add("world")
	f.Add("")
	f.Add("日本語")

	f.Fuzz(func(t *testing.T, s string) {
		result := Reverse(Reverse(s))
		if result != s {
			t.Errorf("Reverse(Reverse(%q)) = %q", s, result)
		}
	})
}
