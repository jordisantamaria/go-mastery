package challenges

import (
	"reflect"
	"sort"
	"testing"
)

// --- Helpers ---

// buildList construye una linked list a partir de un slice.
func buildList(vals []int) *ListNode {
	if len(vals) == 0 {
		return nil
	}
	head := &ListNode{Val: vals[0]}
	current := head
	for _, v := range vals[1:] {
		current.Next = &ListNode{Val: v}
		current = current.Next
	}
	return head
}

// listToSlice convierte una linked list a un slice.
func listToSlice(head *ListNode) []int {
	var result []int
	for head != nil {
		result = append(result, head.Val)
		head = head.Next
	}
	return result
}

// buildTree construye un arbol binario a partir de una representacion por niveles.
// Usa -1 como marcador de nodo nulo (asumimos que -1 no es un valor valido en los tests).
func buildTree(vals []int) *TreeNode {
	if len(vals) == 0 {
		return nil
	}
	root := &TreeNode{Val: vals[0]}
	queue := []*TreeNode{root}
	i := 1
	for len(queue) > 0 && i < len(vals) {
		node := queue[0]
		queue = queue[1:]

		if i < len(vals) && vals[i] != -1 {
			node.Left = &TreeNode{Val: vals[i]}
			queue = append(queue, node.Left)
		}
		i++

		if i < len(vals) && vals[i] != -1 {
			node.Right = &TreeNode{Val: vals[i]}
			queue = append(queue, node.Right)
		}
		i++
	}
	return root
}

// --- Tests ---

func TestTwoSum(t *testing.T) {
	tests := []struct {
		name   string
		nums   []int
		target int
		want   [2]int
	}{
		{
			name:   "ejemplo basico",
			nums:   []int{2, 7, 11, 15},
			target: 9,
			want:   [2]int{0, 1},
		},
		{
			name:   "elementos no consecutivos",
			nums:   []int{3, 2, 4},
			target: 6,
			want:   [2]int{1, 2},
		},
		{
			name:   "mismo valor duplicado",
			nums:   []int{3, 3},
			target: 6,
			want:   [2]int{0, 1},
		},
		{
			name:   "numeros negativos",
			nums:   []int{-1, -2, -3, -4, -5},
			target: -8,
			want:   [2]int{2, 4},
		},
		{
			name:   "target cero",
			nums:   []int{-3, 4, 3, 90},
			target: 0,
			want:   [2]int{0, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TwoSum(tt.nums, tt.target)
			// Ordenar para comparar (el orden de los indices no importa)
			a, b := got[0], got[1]
			if a > b {
				a, b = b, a
			}
			wantA, wantB := tt.want[0], tt.want[1]
			if wantA > wantB {
				wantA, wantB = wantB, wantA
			}
			if a != wantA || b != wantB {
				t.Errorf("TwoSum(%v, %d) = %v, quiero %v", tt.nums, tt.target, got, tt.want)
			}
			// Verificar que los valores suman el target
			if tt.nums[got[0]]+tt.nums[got[1]] != tt.target {
				t.Errorf("los valores en indices %v no suman %d", got, tt.target)
			}
		})
	}
}

func TestValidParentheses(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"parentesis simples", "()", true},
		{"multiples tipos", "()[]{}", true},
		{"anidados", "{[()]}", true},
		{"no cerrado", "(]", false},
		{"orden incorrecto", "([)]", false},
		{"string vacio", "", true},
		{"solo apertura", "(", false},
		{"solo cierre", ")", false},
		{"complejo valido", "({[]}())", true},
		{"complejo invalido", "({[}])", false},
		{"muchos anidados", "((((((()))))))", true},
		{"impar", "(()", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidParentheses(tt.input)
			if got != tt.want {
				t.Errorf("ValidParentheses(%q) = %v, quiero %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMergeSortedArrays(t *testing.T) {
	tests := []struct {
		name string
		a    []int
		b    []int
		want []int
	}{
		{
			name: "ambos con elementos",
			a:    []int{1, 3, 5},
			b:    []int{2, 4, 6},
			want: []int{1, 2, 3, 4, 5, 6},
		},
		{
			name: "a vacio",
			a:    []int{},
			b:    []int{1, 2, 3},
			want: []int{1, 2, 3},
		},
		{
			name: "b vacio",
			a:    []int{1, 2, 3},
			b:    []int{},
			want: []int{1, 2, 3},
		},
		{
			name: "ambos vacios",
			a:    []int{},
			b:    []int{},
			want: []int{},
		},
		{
			name: "a completamente menor",
			a:    []int{1, 2, 3},
			b:    []int{4, 5, 6},
			want: []int{1, 2, 3, 4, 5, 6},
		},
		{
			name: "con duplicados",
			a:    []int{1, 3, 3, 5},
			b:    []int{2, 3, 4},
			want: []int{1, 2, 3, 3, 3, 4, 5},
		},
		{
			name: "con negativos",
			a:    []int{-5, -1, 3},
			b:    []int{-3, 0, 2},
			want: []int{-5, -3, -1, 0, 2, 3},
		},
		{
			name: "un solo elemento cada uno",
			a:    []int{1},
			b:    []int{2},
			want: []int{1, 2},
		},
		{
			name: "a nil",
			a:    nil,
			b:    []int{1, 2},
			want: []int{1, 2},
		},
		{
			name: "b nil",
			a:    []int{1, 2},
			b:    nil,
			want: []int{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeSortedArrays(tt.a, tt.b)
			if len(got) == 0 && len(tt.want) == 0 {
				return // ambos vacios, OK
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeSortedArrays(%v, %v) = %v, quiero %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestReverseLinkedList(t *testing.T) {
	tests := []struct {
		name string
		vals []int
		want []int
	}{
		{"lista de 5", []int{1, 2, 3, 4, 5}, []int{5, 4, 3, 2, 1}},
		{"lista de 1", []int{1}, []int{1}},
		{"lista de 2", []int{1, 2}, []int{2, 1}},
		{"lista vacia", []int{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			head := buildList(tt.vals)
			reversed := ReverseLinkedList(head)
			got := listToSlice(reversed)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReverseLinkedList(%v) = %v, quiero %v", tt.vals, got, tt.want)
			}
		})
	}
}

func TestLRUCache(t *testing.T) {
	t.Run("operaciones basicas", func(t *testing.T) {
		cache := NewLRUCache(2)

		cache.Put(1, 1)
		cache.Put(2, 2)

		val, ok := cache.Get(1)
		if !ok || val != 1 {
			t.Errorf("Get(1) = (%d, %v), quiero (1, true)", val, ok)
		}

		// Insertar 3 deberia desalojar key 2 (la menos reciente)
		cache.Put(3, 3)

		_, ok = cache.Get(2)
		if ok {
			t.Error("Get(2) deberia retornar false despues de desalojo")
		}

		val, ok = cache.Get(3)
		if !ok || val != 3 {
			t.Errorf("Get(3) = (%d, %v), quiero (3, true)", val, ok)
		}
	})

	t.Run("actualizar existente", func(t *testing.T) {
		cache := NewLRUCache(2)

		cache.Put(1, 1)
		cache.Put(2, 2)
		cache.Put(1, 10) // actualizar key 1

		val, ok := cache.Get(1)
		if !ok || val != 10 {
			t.Errorf("Get(1) despues de update = (%d, %v), quiero (10, true)", val, ok)
		}

		// key 2 deberia seguir existiendo (key 1 fue actualizada, no es la LRU)
		cache.Put(3, 3) // deberia desalojar key 2

		_, ok = cache.Get(2)
		if ok {
			t.Error("Get(2) deberia retornar false despues de desalojo")
		}
	})

	t.Run("get actualiza recencia", func(t *testing.T) {
		cache := NewLRUCache(2)

		cache.Put(1, 1)
		cache.Put(2, 2)
		cache.Get(1) // hace que key 1 sea la mas reciente

		cache.Put(3, 3) // deberia desalojar key 2 (no key 1)

		val, ok := cache.Get(1)
		if !ok || val != 1 {
			t.Errorf("Get(1) = (%d, %v), quiero (1, true) - no deberia haber sido desalojada", val, ok)
		}

		_, ok = cache.Get(2)
		if ok {
			t.Error("Get(2) deberia retornar false")
		}
	})

	t.Run("capacidad 1", func(t *testing.T) {
		cache := NewLRUCache(1)

		cache.Put(1, 1)
		cache.Put(2, 2) // desaloja key 1

		_, ok := cache.Get(1)
		if ok {
			t.Error("Get(1) deberia retornar false con capacidad 1")
		}

		val, ok := cache.Get(2)
		if !ok || val != 2 {
			t.Errorf("Get(2) = (%d, %v), quiero (2, true)", val, ok)
		}
	})

	t.Run("key inexistente", func(t *testing.T) {
		cache := NewLRUCache(2)

		_, ok := cache.Get(999)
		if ok {
			t.Error("Get de key inexistente deberia retornar false")
		}
	})
}

func TestMaxSubarraySum(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		want int
	}{
		{"ejemplo clasico", []int{-2, 1, -3, 4, -1, 2, 1, -5, 4}, 6},
		{"un solo elemento positivo", []int{1}, 1},
		{"un solo elemento negativo", []int{-1}, -1},
		{"todos positivos", []int{1, 2, 3, 4}, 10},
		{"todos negativos", []int{-1, -2, -3}, -1},
		{"mezcla", []int{5, -3, 5}, 7},
		{"array grande positivo", []int{1, 2, 3, -2, 5}, 9},
		{"maximo al final", []int{-1, -2, 10}, 10},
		{"maximo al inicio", []int{10, -1, -2}, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaxSubarraySum(tt.nums)
			if got != tt.want {
				t.Errorf("MaxSubarraySum(%v) = %d, quiero %d", tt.nums, got, tt.want)
			}
		})
	}
}

func TestBinarySearch(t *testing.T) {
	tests := []struct {
		name   string
		nums   []int
		target int
		want   int
	}{
		{"encontrado al medio", []int{1, 3, 5, 7, 9}, 5, 2},
		{"encontrado al inicio", []int{1, 3, 5, 7, 9}, 1, 0},
		{"encontrado al final", []int{1, 3, 5, 7, 9}, 9, 4},
		{"no encontrado", []int{1, 3, 5, 7, 9}, 4, -1},
		{"array vacio", []int{}, 1, -1},
		{"un elemento encontrado", []int{5}, 5, 0},
		{"un elemento no encontrado", []int{5}, 3, -1},
		{"dos elementos encontrado primero", []int{1, 3}, 1, 0},
		{"dos elementos encontrado segundo", []int{1, 3}, 3, 1},
		{"numeros negativos", []int{-10, -5, 0, 5, 10}, -5, 1},
		{"target menor que todos", []int{10, 20, 30}, 5, -1},
		{"target mayor que todos", []int{10, 20, 30}, 35, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BinarySearch(tt.nums, tt.target)
			if got != tt.want {
				t.Errorf("BinarySearch(%v, %d) = %d, quiero %d", tt.nums, tt.target, got, tt.want)
			}
		})
	}
}

func TestLevelOrderTraversal(t *testing.T) {
	tests := []struct {
		name string
		vals []int
		want [][]int
	}{
		{
			name: "arbol completo",
			vals: []int{3, 9, 20, -1, -1, 15, 7},
			want: [][]int{{3}, {9, 20}, {15, 7}},
		},
		{
			name: "un solo nodo",
			vals: []int{1},
			want: [][]int{{1}},
		},
		{
			name: "arbol vacio",
			vals: []int{},
			want: nil,
		},
		{
			name: "arbol lineal izquierdo",
			vals: []int{1, 2, -1, 3, -1},
			want: [][]int{{1}, {2}, {3}},
		},
		{
			name: "arbol completo de 3 niveles",
			vals: []int{1, 2, 3, 4, 5, 6, 7},
			want: [][]int{{1}, {2, 3}, {4, 5, 6, 7}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := buildTree(tt.vals)
			got := LevelOrderTraversal(root)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LevelOrderTraversal(%v) = %v, quiero %v", tt.vals, got, tt.want)
			}
		})
	}
}

func TestIsAnagram(t *testing.T) {
	tests := []struct {
		name string
		s    string
		t2   string
		want bool
	}{
		{"anagrama basico", "anagram", "nagaram", true},
		{"no anagrama", "rat", "car", false},
		{"strings vacios", "", "", true},
		{"un vacio", "a", "", false},
		{"mismas letras diferente freq", "aab", "abb", false},
		{"un caracter igual", "a", "a", true},
		{"un caracter diferente", "a", "b", false},
		{"con espacios", "abc", "cba", true},
		{"case sensitive", "Hello", "hello", false},
		{"unicode", "cafe", "face", true},
		{"largo", "abcdefghijklmnopqrstuvwxyz", "zyxwvutsrqponmlkjihgfedcba", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAnagram(tt.s, tt.t2)
			if got != tt.want {
				t.Errorf("IsAnagram(%q, %q) = %v, quiero %v", tt.s, tt.t2, got, tt.want)
			}
		})
	}
}

func TestTopKFrequent(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		k    int
		want []int
	}{
		{
			name: "ejemplo basico",
			nums: []int{1, 1, 1, 2, 2, 3},
			k:    2,
			want: []int{1, 2},
		},
		{
			name: "k=1",
			nums: []int{1},
			k:    1,
			want: []int{1},
		},
		{
			name: "todos iguales",
			nums: []int{5, 5, 5, 5},
			k:    1,
			want: []int{5},
		},
		{
			name: "k igual al numero de elementos unicos",
			nums: []int{1, 2, 3},
			k:    3,
			want: []int{1, 2, 3},
		},
		{
			name: "numeros negativos",
			nums: []int{-1, -1, -2, -2, -2, 3},
			k:    2,
			want: []int{-2, -1},
		},
		{
			name: "muchos elementos",
			nums: []int{4, 1, -1, -1, -1, 2, 3, 3, 3, 3},
			k:    2,
			want: []int{3, -1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TopKFrequent(tt.nums, tt.k)

			if len(got) != tt.k {
				t.Fatalf("TopKFrequent(%v, %d) retorno %d elementos, quiero %d", tt.nums, tt.k, len(got), tt.k)
			}

			// Ordenar ambos para comparar (el orden no importa)
			sortedGot := make([]int, len(got))
			copy(sortedGot, got)
			sort.Ints(sortedGot)

			sortedWant := make([]int, len(tt.want))
			copy(sortedWant, tt.want)
			sort.Ints(sortedWant)

			if !reflect.DeepEqual(sortedGot, sortedWant) {
				t.Errorf("TopKFrequent(%v, %d) = %v, quiero %v", tt.nums, tt.k, got, tt.want)
			}
		})
	}
}
