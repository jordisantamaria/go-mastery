package exercises

import (
	"sort"
	"testing"
)

// =============================================
// Estos tests verifican tus soluciones.
// Ejecuta: go test ./01-foundations/01-syntax-types/exercises/...
// Con detalle: go test -v ./01-foundations/01-syntax-types/exercises/...
// =============================================

func TestZeroValues(t *testing.T) {
	i, f, s, b := ZeroValues()

	if i != 0 {
		t.Errorf("int zero value: got %d, want 0", i)
	}
	if f != 0.0 {
		t.Errorf("float64 zero value: got %f, want 0.0", f)
	}
	if s != "" {
		t.Errorf("string zero value: got %q, want \"\"", s)
	}
	if b != false {
		t.Errorf("bool zero value: got %t, want false", b)
	}
}

func TestSwap(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		wantA    int
		wantB    int
	}{
		{"positivos", 1, 2, 2, 1},
		{"negativos", -5, -10, -10, -5},
		{"iguales", 3, 3, 3, 3},
		{"con cero", 0, 42, 42, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b := Swap(tt.a, tt.b)
			if a != tt.wantA || b != tt.wantB {
				t.Errorf("Swap(%d, %d) = (%d, %d), want (%d, %d)",
					tt.a, tt.b, a, b, tt.wantA, tt.wantB)
			}
		})
	}
}

func TestRuneCount(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{"", 0},
		{"Hola 🌍", 6},
		{"日本語", 3},
		{"café", 4},
		{"🏳️‍🌈", 4}, // flag emoji: multiple runes
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := RuneCount(tt.input)
			if got != tt.want {
				t.Errorf("RuneCount(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestSumSlice(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  int
	}{
		{"varios", []int{1, 2, 3, 4, 5}, 15},
		{"uno", []int{42}, 42},
		{"vacio", []int{}, 0},
		{"nil", nil, 0},
		{"negativos", []int{-1, -2, -3}, -6},
		{"mixto", []int{-10, 5, 5}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SumSlice(tt.input)
			if got != tt.want {
				t.Errorf("SumSlice(%v) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestUniqueStrings(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{"con duplicados", []string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{"sin duplicados", []string{"x", "y", "z"}, []string{"x", "y", "z"}},
		{"vacio", []string{}, []string{}},
		{"un elemento", []string{"solo"}, []string{"solo"}},
		{"todos iguales", []string{"a", "a", "a"}, []string{"a"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UniqueStrings(tt.input)
			sort.Strings(got)
			sort.Strings(tt.want)
			if len(got) != len(tt.want) {
				t.Errorf("UniqueStrings(%v) returned %d elements, want %d",
					tt.input, len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("UniqueStrings(%v) = %v, want %v", tt.input, got, tt.want)
					return
				}
			}
		})
	}
}

func TestWordCount(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]int
	}{
		{
			"basico",
			"hello world hello",
			map[string]int{"hello": 2, "world": 1},
		},
		{
			"una palabra",
			"go",
			map[string]int{"go": 1},
		},
		{
			"vacio",
			"",
			map[string]int{},
		},
		{
			"espacios extra",
			"  go   is   great  ",
			map[string]int{"go": 1, "is": 1, "great": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WordCount(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("WordCount(%q) returned %d entries, want %d",
					tt.input, len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("WordCount(%q)[%q] = %d, want %d",
						tt.input, k, got[k], v)
				}
			}
		})
	}
}

func TestReverseSlice(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"varios", []int{1, 2, 3, 4, 5}, []int{5, 4, 3, 2, 1}},
		{"dos", []int{1, 2}, []int{2, 1}},
		{"uno", []int{1}, []int{1}},
		{"vacio", []int{}, []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Guardar copia del original para verificar que no se modifica
			original := make([]int, len(tt.input))
			copy(original, tt.input)

			got := ReverseSlice(tt.input)

			// Verificar resultado
			if len(got) != len(tt.want) {
				t.Errorf("ReverseSlice(%v) returned %d elements, want %d",
					tt.input, len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ReverseSlice(%v) = %v, want %v", original, got, tt.want)
					return
				}
			}

			// Verificar que el original NO fue modificado
			for i := range tt.input {
				if tt.input[i] != original[i] {
					t.Errorf("ReverseSlice modifico el slice original: %v -> %v",
						original, tt.input)
					return
				}
			}
		})
	}
}

func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name string
		a    map[string]int
		b    map[string]int
		want map[string]int
	}{
		{
			"merge basico",
			map[string]int{"a": 1, "b": 2},
			map[string]int{"b": 3, "c": 4},
			map[string]int{"a": 1, "b": 3, "c": 4},
		},
		{
			"b vacio",
			map[string]int{"a": 1},
			map[string]int{},
			map[string]int{"a": 1},
		},
		{
			"a vacio",
			map[string]int{},
			map[string]int{"x": 10},
			map[string]int{"x": 10},
		},
		{
			"ambos vacios",
			map[string]int{},
			map[string]int{},
			map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeMaps(tt.a, tt.b)
			if len(got) != len(tt.want) {
				t.Errorf("MergeMaps returned %d entries, want %d", len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("MergeMaps[%q] = %d, want %d", k, got[k], v)
				}
			}
		})
	}
}
