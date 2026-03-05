// Package challenges contiene problemas clasicos de algoritmos y estructuras de datos
// para preparacion de entrevistas tecnicas en Go.
package challenges

// ListNode representa un nodo de una singly linked list.
type ListNode struct {
	Val  int
	Next *ListNode
}

// TreeNode representa un nodo de un arbol binario.
type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

// LRUCache implementa un cache Least Recently Used con capacidad fija.
// Get(key) retorna el valor y true si existe, o 0 y false si no.
// Put(key, value) inserta o actualiza un par key-value. Si el cache esta lleno,
// elimina el elemento menos recientemente usado antes de insertar.
type LRUCache struct {
	capacity int
	items    map[int]*lruNode
	head     *lruNode // dummy head (MRU end)
	tail     *lruNode // dummy tail (LRU end)
}

type lruNode struct {
	key, value int
	prev, next *lruNode
}

// NewLRUCache crea un nuevo LRUCache con la capacidad dada.
func NewLRUCache(capacity int) *LRUCache {
	panic("TODO")
}

// Get retorna el valor asociado a la key y true, o (0, false) si no existe.
func (c *LRUCache) Get(key int) (int, bool) {
	panic("TODO")
}

// Put inserta o actualiza un par key-value en el cache.
func (c *LRUCache) Put(key, value int) {
	panic("TODO")
}

// TwoSum encuentra dos indices en nums cuyos valores sumen target.
// Retorna un array de dos indices. Se garantiza que existe exactamente una solucion.
// No se puede usar el mismo elemento dos veces.
func TwoSum(nums []int, target int) [2]int {
	panic("TODO")
}

// ValidParentheses verifica si un string de brackets esta balanceado.
// Soporta '(', ')', '{', '}', '[', ']'.
func ValidParentheses(s string) bool {
	panic("TODO")
}

// MergeSortedArrays fusiona dos slices ordenados en uno solo ordenado.
func MergeSortedArrays(a, b []int) []int {
	panic("TODO")
}

// ReverseLinkedList invierte una singly linked list y retorna la nueva head.
func ReverseLinkedList(head *ListNode) *ListNode {
	panic("TODO")
}

// MaxSubarraySum encuentra la suma maxima de un subarray contiguo (algoritmo de Kadane).
// El slice tiene al menos un elemento.
func MaxSubarraySum(nums []int) int {
	panic("TODO")
}

// BinarySearch busca target en un slice ordenado ascendentemente.
// Retorna el indice si lo encuentra, o -1 si no existe.
func BinarySearch(nums []int, target int) int {
	panic("TODO")
}

// LevelOrderTraversal realiza un recorrido por niveles (BFS) de un arbol binario.
// Retorna un slice de slices, donde cada sub-slice contiene los valores de un nivel.
func LevelOrderTraversal(root *TreeNode) [][]int {
	panic("TODO")
}

// IsAnagram verifica si dos strings son anagramas entre si.
// Dos strings son anagramas si contienen exactamente los mismos caracteres
// con la misma frecuencia (case-sensitive).
func IsAnagram(s, t string) bool {
	panic("TODO")
}

// TopKFrequent retorna los k elementos mas frecuentes de nums.
// Se garantiza que la respuesta es unica. El orden del resultado no importa.
func TopKFrequent(nums []int, k int) []int {
	panic("TODO")
}
