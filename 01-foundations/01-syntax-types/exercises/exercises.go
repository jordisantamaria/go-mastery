// SOLUCIONES — NO mires hasta haber intentado resolver los ejercicios!
// Renombra este archivo a solutions.go para compilar y testear.

package exercises

import "strings"

func ZeroValues() (int, float64, string, bool) {
	var i int
	var f float64
	var s string
	var b bool
	return i, f, s, b
}

func Swap(a, b int) (int, int) {
	return b, a
}

func RuneCount(s string) int {
	return len([]rune(s))
}

func SumSlice(nums []int) int {
	sum := 0
	for _, n := range nums {
		sum += n
	}
	return sum
}

func UniqueStrings(items []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

func WordCount(s string) map[string]int {
	counts := make(map[string]int)
	for _, word := range strings.Fields(s) {
		counts[word]++
	}
	return counts
}

func ReverseSlice(nums []int) []int {
	result := make([]int, len(nums))
	for i, v := range nums {
		result[len(nums)-1-i] = v
	}
	return result
}

func MergeMaps(a, b map[string]int) map[string]int {
	result := make(map[string]int)
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}
