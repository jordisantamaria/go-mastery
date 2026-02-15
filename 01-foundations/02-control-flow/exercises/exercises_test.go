package exercises

import "testing"

func TestFizzBuzz(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{1, "1"},
		{3, "Fizz"},
		{5, "Buzz"},
		{15, "FizzBuzz"},
		{30, "FizzBuzz"},
		{9, "Fizz"},
		{10, "Buzz"},
		{7, "7"},
		{0, "FizzBuzz"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FizzBuzz(tt.input)
			if got != tt.want {
				t.Errorf("FizzBuzz(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		temp int
		want string
	}{
		{-10, "freezing"},
		{-1, "freezing"},
		{0, "cold"},
		{15, "cold"},
		{16, "mild"},
		{25, "mild"},
		{26, "warm"},
		{35, "warm"},
		{36, "hot"},
		{50, "hot"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := Classify(tt.temp)
			if got != tt.want {
				t.Errorf("Classify(%d) = %q, want %q", tt.temp, got, tt.want)
			}
		})
	}
}

func TestSumRange(t *testing.T) {
	tests := []struct {
		name     string
		from, to int
		want     int
	}{
		{"1 a 5", 1, 5, 15},
		{"1 a 1", 1, 1, 1},
		{"3 a 3", 3, 3, 3},
		{"1 a 10", 1, 10, 55},
		{"invertido", 5, 1, 0},
		{"negativos", -3, 3, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SumRange(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("SumRange(%d, %d) = %d, want %d", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestFindIndex(t *testing.T) {
	items := []string{"alpha", "beta", "gamma", "delta"}

	tests := []struct {
		target string
		want   int
	}{
		{"alpha", 0},
		{"gamma", 2},
		{"delta", 3},
		{"omega", -1},
		{"", -1},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			got := FindIndex(items, tt.target)
			if got != tt.want {
				t.Errorf("FindIndex(items, %q) = %d, want %d", tt.target, got, tt.want)
			}
		})
	}

	// Test con slice vacio
	t.Run("empty slice", func(t *testing.T) {
		got := FindIndex([]string{}, "anything")
		if got != -1 {
			t.Errorf("FindIndex([], anything) = %d, want -1", got)
		}
	})
}

func TestMatrixSum(t *testing.T) {
	tests := []struct {
		name   string
		matrix [][]int
		want   int
	}{
		{"2x2", [][]int{{1, 2}, {3, 4}}, 10},
		{"3x3", [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, 45},
		{"1x1", [][]int{{42}}, 42},
		{"vacio", [][]int{}, 0},
		{"nil", nil, 0},
		{"jagged", [][]int{{1}, {2, 3}, {4, 5, 6}}, 21},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatrixSum(tt.matrix)
			if got != tt.want {
				t.Errorf("MatrixSum(%v) = %d, want %d", tt.matrix, got, tt.want)
			}
		})
	}
}

func TestCollatz(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want int
	}{
		{"1", 1, 0},
		{"2", 2, 1},
		{"6", 6, 8},
		{"27", 27, 111},
		{"invalido negativo", -1, -1},
		{"invalido cero", 0, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Collatz(tt.n)
			if got != tt.want {
				t.Errorf("Collatz(%d) = %d, want %d", tt.n, got, tt.want)
			}
		})
	}
}
