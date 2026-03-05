package exercises

import (
	"sort"
	"testing"
)

func TestMin(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		if got := Min(3, 5); got != 3 {
			t.Errorf("Min(3, 5) = %d, want 3", got)
		}
		if got := Min(5, 3); got != 3 {
			t.Errorf("Min(5, 3) = %d, want 3", got)
		}
		if got := Min(3, 3); got != 3 {
			t.Errorf("Min(3, 3) = %d, want 3", got)
		}
	})
	t.Run("string", func(t *testing.T) {
		if got := Min("apple", "banana"); got != "apple" {
			t.Errorf("Min(apple, banana) = %q, want apple", got)
		}
	})
	t.Run("float", func(t *testing.T) {
		if got := Min(3.14, 2.71); got != 2.71 {
			t.Errorf("Min(3.14, 2.71) = %f, want 2.71", got)
		}
	})
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name       string
		val, lo, hi int
		want       int
	}{
		{"in range", 5, 1, 10, 5},
		{"below", -5, 0, 100, 0},
		{"above", 200, 0, 100, 100},
		{"at min", 0, 0, 100, 0},
		{"at max", 100, 0, 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Clamp(tt.val, tt.lo, tt.hi)
			if got != tt.want {
				t.Errorf("Clamp(%d, %d, %d) = %d, want %d",
					tt.val, tt.lo, tt.hi, got, tt.want)
			}
		})
	}
}

func TestMap(t *testing.T) {
	t.Run("int to int", func(t *testing.T) {
		got := Map([]int{1, 2, 3}, func(n int) int { return n * 2 })
		want := []int{2, 4, 6}
		assertSliceEqual(t, got, want)
	})

	t.Run("string to int", func(t *testing.T) {
		got := Map([]string{"a", "bb", "ccc"}, func(s string) int { return len(s) })
		want := []int{1, 2, 3}
		assertSliceEqual(t, got, want)
	})

	t.Run("empty", func(t *testing.T) {
		got := Map([]int{}, func(n int) int { return n })
		if len(got) != 0 {
			t.Errorf("Map(empty) returned %d elements", len(got))
		}
	})
}

func TestFilter(t *testing.T) {
	t.Run("even numbers", func(t *testing.T) {
		got := Filter([]int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 })
		want := []int{2, 4, 6}
		assertSliceEqual(t, got, want)
	})

	t.Run("no match", func(t *testing.T) {
		got := Filter([]int{1, 3, 5}, func(n int) bool { return n%2 == 0 })
		if got == nil {
			got = []int{}
		}
		if len(got) != 0 {
			t.Errorf("Filter(no match) returned %d elements", len(got))
		}
	})

	t.Run("strings", func(t *testing.T) {
		got := Filter([]string{"alice", "bob", "charlie"},
			func(s string) bool { return len(s) > 3 })
		want := []string{"alice", "charlie"}
		assertStringSliceEqual(t, got, want)
	})
}

func TestReduce(t *testing.T) {
	t.Run("sum", func(t *testing.T) {
		got := Reduce([]int{1, 2, 3, 4}, 0, func(acc, n int) int { return acc + n })
		if got != 10 {
			t.Errorf("Reduce(sum) = %d, want 10", got)
		}
	})

	t.Run("concat", func(t *testing.T) {
		got := Reduce([]string{"a", "b", "c"}, "", func(acc, s string) string { return acc + s })
		if got != "abc" {
			t.Errorf("Reduce(concat) = %q, want abc", got)
		}
	})

	t.Run("empty with initial", func(t *testing.T) {
		got := Reduce([]int{}, 42, func(acc, n int) int { return acc + n })
		if got != 42 {
			t.Errorf("Reduce(empty, 42) = %d, want 42", got)
		}
	})
}

func TestContains(t *testing.T) {
	if !Contains([]int{1, 2, 3}, 2) {
		t.Error("Contains(1,2,3; 2) should be true")
	}
	if Contains([]int{1, 2, 3}, 5) {
		t.Error("Contains(1,2,3; 5) should be false")
	}
	if !Contains([]string{"go", "rust", "python"}, "go") {
		t.Error("Contains strings should find go")
	}
	if Contains([]int{}, 1) {
		t.Error("Contains empty should be false")
	}
}

func TestUniq(t *testing.T) {
	t.Run("ints", func(t *testing.T) {
		got := Uniq([]int{1, 2, 2, 3, 1, 4, 3})
		want := []int{1, 2, 3, 4}
		assertSliceEqual(t, got, want)
	})

	t.Run("strings", func(t *testing.T) {
		got := Uniq([]string{"a", "b", "a", "c", "b"})
		want := []string{"a", "b", "c"}
		assertStringSliceEqual(t, got, want)
	})

	t.Run("no duplicates", func(t *testing.T) {
		got := Uniq([]int{1, 2, 3})
		want := []int{1, 2, 3}
		assertSliceEqual(t, got, want)
	})

	t.Run("empty", func(t *testing.T) {
		got := Uniq([]int{})
		if got == nil {
			got = []int{}
		}
		if len(got) != 0 {
			t.Errorf("Uniq(empty) returned %d elements", len(got))
		}
	})
}

func TestGroupBy(t *testing.T) {
	t.Run("by first letter", func(t *testing.T) {
		items := []string{"apple", "avocado", "banana", "blueberry", "cherry"}
		got := GroupBy(items, func(s string) byte { return s[0] })

		if len(got) != 3 {
			t.Fatalf("GroupBy returned %d groups, want 3", len(got))
		}

		aGroup := got['a']
		sort.Strings(aGroup)
		if len(aGroup) != 2 || aGroup[0] != "apple" || aGroup[1] != "avocado" {
			t.Errorf("group 'a' = %v, want [apple, avocado]", aGroup)
		}

		bGroup := got['b']
		sort.Strings(bGroup)
		if len(bGroup) != 2 || bGroup[0] != "banana" || bGroup[1] != "blueberry" {
			t.Errorf("group 'b' = %v, want [banana, blueberry]", bGroup)
		}
	})

	t.Run("by length", func(t *testing.T) {
		items := []string{"a", "bb", "c", "dd", "eee"}
		got := GroupBy(items, func(s string) int { return len(s) })

		if len(got[1]) != 2 {
			t.Errorf("group len=1 has %d items, want 2", len(got[1]))
		}
		if len(got[2]) != 2 {
			t.Errorf("group len=2 has %d items, want 2", len(got[2]))
		}
	})
}

func TestStack(t *testing.T) {
	t.Run("int stack", func(t *testing.T) {
		s := &Stack[int]{}

		if !s.IsEmpty() {
			t.Error("new stack should be empty")
		}

		s.Push(1)
		s.Push(2)
		s.Push(3)

		if s.Len() != 3 {
			t.Errorf("Len() = %d, want 3", s.Len())
		}

		val, ok := s.Peek()
		if !ok || val != 3 {
			t.Errorf("Peek() = (%d, %t), want (3, true)", val, ok)
		}
		if s.Len() != 3 {
			t.Error("Peek should not remove element")
		}

		val, ok = s.Pop()
		if !ok || val != 3 {
			t.Errorf("Pop() = (%d, %t), want (3, true)", val, ok)
		}
		if s.Len() != 2 {
			t.Errorf("Len after pop = %d, want 2", s.Len())
		}

		s.Pop()
		s.Pop()

		_, ok = s.Pop()
		if ok {
			t.Error("Pop from empty stack should return false")
		}
	})

	t.Run("string stack", func(t *testing.T) {
		s := &Stack[string]{}
		s.Push("hello")
		s.Push("world")

		val, ok := s.Pop()
		if !ok || val != "world" {
			t.Errorf("Pop() = (%q, %t), want (world, true)", val, ok)
		}
	})
}

func TestMapKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	got := MapKeys(m)
	sort.Strings(got)
	want := []string{"a", "b", "c"}
	assertStringSliceEqual(t, got, want)
}

func TestMapValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	got := MapValues(m)
	sort.Ints(got)
	want := []int{1, 2, 3}
	assertSliceEqual(t, got, want)
}

// --- Helpers ---

func assertSliceEqual(t *testing.T, got, want []int) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("got %v (len %d), want %v (len %d)", got, len(got), want, len(want))
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got %v, want %v", got, want)
			return
		}
	}
}

func assertStringSliceEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("got %v (len %d), want %v (len %d)", got, len(got), want, len(want))
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got %v, want %v", got, want)
			return
		}
	}
}
