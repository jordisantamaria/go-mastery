package exercises

import "testing"

func TestApply(t *testing.T) {
	double := func(n int) int { return n * 2 }
	negate := func(n int) int { return -n }

	tests := []struct {
		name string
		nums []int
		fn   func(int) int
		want []int
	}{
		{"double", []int{1, 2, 3}, double, []int{2, 4, 6}},
		{"negate", []int{1, -2, 3}, negate, []int{-1, 2, -3}},
		{"empty", []int{}, double, []int{}},
		{"nil", nil, double, []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Apply(tt.nums, tt.fn)
			if got == nil {
				got = []int{}
			}
			if len(got) != len(tt.want) {
				t.Errorf("Apply() returned %d elements, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Apply()[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFilter(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }
	isPositive := func(n int) bool { return n > 0 }

	tests := []struct {
		name      string
		nums      []int
		predicate func(int) bool
		want      []int
	}{
		{"evens", []int{1, 2, 3, 4, 5}, isEven, []int{2, 4}},
		{"positives", []int{-2, -1, 0, 1, 2}, isPositive, []int{1, 2}},
		{"no match", []int{1, 3, 5}, isEven, []int{}},
		{"all match", []int{2, 4, 6}, isEven, []int{2, 4, 6}},
		{"empty", []int{}, isEven, []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Filter(tt.nums, tt.predicate)
			if got == nil {
				got = []int{}
			}
			if len(got) != len(tt.want) {
				t.Errorf("Filter() returned %d elements, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Filter()[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestReduce(t *testing.T) {
	add := func(acc, n int) int { return acc + n }
	multiply := func(acc, n int) int { return acc * n }

	tests := []struct {
		name    string
		nums    []int
		initial int
		fn      func(int, int) int
		want    int
	}{
		{"sum", []int{1, 2, 3, 4}, 0, add, 10},
		{"product", []int{1, 2, 3, 4}, 1, multiply, 24},
		{"sum with initial", []int{1, 2, 3}, 10, add, 16},
		{"empty", []int{}, 42, add, 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Reduce(tt.nums, tt.initial, tt.fn)
			if got != tt.want {
				t.Errorf("Reduce() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMakeMultiplier(t *testing.T) {
	double := MakeMultiplier(2)
	triple := MakeMultiplier(3)
	identity := MakeMultiplier(1)

	tests := []struct {
		name   string
		fn     func(int) int
		input  int
		want   int
	}{
		{"double 5", double, 5, 10},
		{"double 0", double, 0, 0},
		{"triple 4", triple, 4, 12},
		{"triple -3", triple, -3, -9},
		{"identity 7", identity, 7, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn(tt.input)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMakeCounter(t *testing.T) {
	inc, get := MakeCounter()

	if got := get(); got != 0 {
		t.Errorf("initial count = %d, want 0", got)
	}

	inc()
	inc()
	inc()

	if got := get(); got != 3 {
		t.Errorf("after 3 increments = %d, want 3", got)
	}

	// Verificar que un segundo counter es independiente
	inc2, get2 := MakeCounter()
	inc2()

	if got := get(); got != 3 {
		t.Errorf("first counter changed: %d, want 3", got)
	}
	if got := get2(); got != 1 {
		t.Errorf("second counter = %d, want 1", got)
	}
}

func TestCompose(t *testing.T) {
	double := func(n int) int { return n * 2 }
	addOne := func(n int) int { return n + 1 }
	square := func(n int) int { return n * n }

	tests := []struct {
		name  string
		f, g  func(int) int
		input int
		want  int
	}{
		{"double then addOne", double, addOne, 3, 7},   // 3*2=6, 6+1=7
		{"addOne then double", addOne, double, 3, 8},   // 3+1=4, 4*2=8
		{"double then square", double, square, 3, 36},  // 3*2=6, 6*6=36
		{"square then double", square, double, 3, 18},  // 3*3=9, 9*2=18
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			composed := Compose(tt.f, tt.g)
			got := composed(tt.input)
			if got != tt.want {
				t.Errorf("Compose()(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestMemoize(t *testing.T) {
	callCount := 0
	expensive := func(n int) int {
		callCount++
		return n * n
	}

	memo := Memoize(expensive)

	// Primera llamada — debe ejecutar la funcion
	got := memo(5)
	if got != 25 {
		t.Errorf("memo(5) = %d, want 25", got)
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (should compute)", callCount)
	}

	// Segunda llamada con mismo argumento — debe usar cache
	got = memo(5)
	if got != 25 {
		t.Errorf("memo(5) cached = %d, want 25", got)
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (should use cache)", callCount)
	}

	// Llamada con argumento diferente — debe ejecutar
	got = memo(3)
	if got != 9 {
		t.Errorf("memo(3) = %d, want 9", got)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}

	// Verificar que 3 esta cacheado
	memo(3)
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2 (3 should be cached)", callCount)
	}
}

func TestPipeline(t *testing.T) {
	double := func(n int) int { return n * 2 }
	addOne := func(n int) int { return n + 1 }
	square := func(n int) int { return n * n }

	tests := []struct {
		name  string
		fns   []func(int) int
		input int
		want  int
	}{
		{
			"double -> addOne -> square",
			[]func(int) int{double, addOne, square},
			2,
			25, // 2*2=4, 4+1=5, 5*5=25
		},
		{
			"single function",
			[]func(int) int{double},
			5,
			10,
		},
		{
			"no functions",
			[]func(int) int{},
			42,
			42, // sin transformacion
		},
		{
			"addOne -> double",
			[]func(int) int{addOne, double},
			3,
			8, // 3+1=4, 4*2=8
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := Pipeline(tt.fns...)
			got := pipeline(tt.input)
			if got != tt.want {
				t.Errorf("Pipeline()(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
