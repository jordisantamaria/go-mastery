package exercises

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"
)

// =============================================
// IMPORTANTE: ejecutar SIEMPRE con -race
// go test -race ./01-foundations/06-concurrency/exercises/...
// =============================================

func TestConcurrentSum(t *testing.T) {
	tests := []struct {
		name       string
		nums       []int
		numWorkers int
		want       int
	}{
		{"basic", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 3, 55},
		{"single worker", []int{1, 2, 3, 4, 5}, 1, 15},
		{"more workers than items", []int{1, 2, 3}, 10, 6},
		{"empty", []int{}, 3, 0},
		{"single element", []int{42}, 2, 42},
		{"large", makeRange(1, 1000), 5, 500500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConcurrentSum(tt.nums, tt.numWorkers)
			if got != tt.want {
				t.Errorf("ConcurrentSum() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSafeMap(t *testing.T) {
	m := NewSafeMap()

	// Concurrent writes
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", id)
			m.Set(key, id)
		}(i)
	}
	wg.Wait()

	if m.Len() != 100 {
		t.Errorf("Len() = %d, want 100", m.Len())
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", id)
			val, ok := m.Get(key)
			if !ok {
				t.Errorf("Get(%q) not found", key)
			}
			if val != id {
				t.Errorf("Get(%q) = %d, want %d", key, val, id)
			}
		}(i)
	}
	wg.Wait()

	// Get non-existent
	_, ok := m.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) should return false")
	}

	// Delete
	m.Delete("key-0")
	if m.Len() != 99 {
		t.Errorf("Len after delete = %d, want 99", m.Len())
	}
}

func TestPipeline(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"basic", []int{1, 2, 3, 4, 5}, []int{4, 16}},
		{"all even squares", []int{2, 4, 6}, []int{4, 16, 36}},
		{"no even squares", []int{1, 3, 5}, []int{}},
		{"empty", []int{}, []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RunPipeline(tt.input)
			if got == nil {
				got = []int{}
			}
			sort.Ints(got)
			sort.Ints(tt.want)
			if len(got) != len(tt.want) {
				t.Errorf("RunPipeline() returned %d elements, want %d: got %v", len(got), len(tt.want), got)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("RunPipeline() = %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}

func TestWorkerPool(t *testing.T) {
	double := func(n int) int { return n * 2 }

	tests := []struct {
		name       string
		jobs       []int
		numWorkers int
		want       []int
	}{
		{"basic", []int{1, 2, 3, 4, 5}, 3, []int{2, 4, 6, 8, 10}},
		{"single worker", []int{1, 2, 3}, 1, []int{2, 4, 6}},
		{"empty", []int{}, 3, []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WorkerPool(tt.jobs, tt.numWorkers, double)
			if got == nil {
				got = []int{}
			}
			sort.Ints(got)
			sort.Ints(tt.want)
			if len(got) != len(tt.want) {
				t.Errorf("WorkerPool() returned %d elements, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("WorkerPool() = %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}

func TestFirstResult(t *testing.T) {
	t.Run("returns fastest", func(t *testing.T) {
		fast := func(ctx context.Context) (string, error) {
			time.Sleep(50 * time.Millisecond)
			return "fast", nil
		}
		slow := func(ctx context.Context) (string, error) {
			time.Sleep(500 * time.Millisecond)
			return "slow", nil
		}

		ctx := context.Background()
		got, err := FirstResult(ctx, fast, slow)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if got != "fast" {
			t.Errorf("FirstResult() = %q, want %q", got, "fast")
		}
	})

	t.Run("respects parent context cancellation", func(t *testing.T) {
		slow := func(ctx context.Context) (string, error) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(5 * time.Second):
				return "done", nil
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := FirstResult(ctx, slow)
		if err == nil {
			t.Error("expected context deadline error")
		}
	})
}

func TestProcessWithLimit(t *testing.T) {
	// Track max concurrency
	var (
		mu             sync.Mutex
		current, maxC  int
	)

	process := func(n int) int {
		mu.Lock()
		current++
		if current > maxC {
			maxC = current
		}
		mu.Unlock()

		time.Sleep(50 * time.Millisecond) // simular trabajo

		mu.Lock()
		current--
		mu.Unlock()

		return n * 2
	}

	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	maxConcurrent := 3

	got := ProcessWithLimit(items, maxConcurrent, process)
	sort.Ints(got)

	// Verificar resultados
	want := []int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20}
	if len(got) != len(want) {
		t.Errorf("ProcessWithLimit() returned %d elements, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("ProcessWithLimit() = %v, want %v", got, want)
			break
		}
	}

	// Verificar que la concurrencia no excedio el limite
	if maxC > maxConcurrent {
		t.Errorf("Max concurrency was %d, limit is %d", maxC, maxConcurrent)
	}
}

func TestMerge(t *testing.T) {
	t.Run("merge 3 channels", func(t *testing.T) {
		ch1 := makeChannel(1, 2, 3)
		ch2 := makeChannel(4, 5, 6)
		ch3 := makeChannel(7, 8, 9)

		merged := Merge(ch1, ch2, ch3)

		var got []int
		for v := range merged {
			got = append(got, v)
		}

		sort.Ints(got)
		want := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
		if len(got) != len(want) {
			t.Errorf("Merge() returned %d elements, want %d", len(got), len(want))
			return
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("Merge() = %v, want %v", got, want)
				return
			}
		}
	})

	t.Run("merge empty channels", func(t *testing.T) {
		ch1 := makeChannel()
		ch2 := makeChannel()

		merged := Merge(ch1, ch2)

		var got []int
		for v := range merged {
			got = append(got, v)
		}

		if len(got) != 0 {
			t.Errorf("Merge(empty, empty) returned %d elements, want 0", len(got))
		}
	})

	t.Run("merge single channel", func(t *testing.T) {
		ch := makeChannel(1, 2, 3)
		merged := Merge(ch)

		var got []int
		for v := range merged {
			got = append(got, v)
		}

		sort.Ints(got)
		want := []int{1, 2, 3}
		if len(got) != len(want) {
			t.Errorf("Merge(single) returned %d elements, want %d", len(got), len(want))
		}
	})
}

// --- Helpers ---

func makeRange(from, to int) []int {
	nums := make([]int, 0, to-from+1)
	for i := from; i <= to; i++ {
		nums = append(nums, i)
	}
	return nums
}

func makeChannel(nums ...int) <-chan int {
	ch := make(chan int)
	go func() {
		for _, n := range nums {
			ch <- n
		}
		close(ch)
	}()
	return ch
}
