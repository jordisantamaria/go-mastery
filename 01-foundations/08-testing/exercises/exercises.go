// Package exercises contiene codigo para que TU escribas los tests.
// Este modulo es diferente: el codigo ya esta implementado, tu tarea es escribir los tests.
// El archivo exercises_test.go tiene las funciones de test vacias — implementalas.
// Ejecuta: go test -v ./01-foundations/08-testing/exercises/...
package exercises

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEmpty    = errors.New("empty input")
	ErrNegative = errors.New("negative number")
	ErrNotFound = errors.New("not found")
)

// Factorial calcula n!.
// Devuelve error si n < 0.
func Factorial(n int) (int, error) {
	if n < 0 {
		return 0, fmt.Errorf("factorial: %w", ErrNegative)
	}
	if n <= 1 {
		return 1, nil
	}
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result, nil
}

// Palindrome verifica si un string es palindromo (case-insensitive).
func Palindrome(s string) bool {
	s = strings.ToLower(s)
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		if runes[i] != runes[j] {
			return false
		}
	}
	return true
}

// TitleCase convierte "hello world" a "Hello World".
func TitleCase(s string) string {
	if s == "" {
		return ""
	}
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

// StringStore es una store simple para demostrar mocking con interfaces.
type StringStore interface {
	Get(key string) (string, error)
	Set(key string, value string) error
}

// CacheService usa un StringStore para caching.
type CacheService struct {
	store StringStore
}

func NewCacheService(store StringStore) *CacheService {
	return &CacheService{store: store}
}

// GetOrDefault devuelve el valor del store, o defaultVal si no existe.
func (c *CacheService) GetOrDefault(key, defaultVal string) string {
	val, err := c.store.Get(key)
	if err != nil {
		return defaultVal
	}
	return val
}

// SetIfNotExists guarda el valor solo si la key no existe.
// Devuelve true si se guardo, false si ya existia.
func (c *CacheService) SetIfNotExists(key, value string) (bool, error) {
	_, err := c.store.Get(key)
	if err == nil {
		return false, nil // ya existe
	}
	if !errors.Is(err, ErrNotFound) {
		return false, err // error real
	}
	if err := c.store.Set(key, value); err != nil {
		return false, err
	}
	return true, nil
}
