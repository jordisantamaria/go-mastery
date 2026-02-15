package main

import (
	"errors"
	"fmt"
	"strconv"
)

// =============================================
// SENTINEL ERRORS
// =============================================

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrEmpty        = errors.New("empty input")
)

// =============================================
// CUSTOM ERROR TYPE
// =============================================

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

func main() {
	// =============================================
	// BASICO: if err != nil
	// =============================================

	fmt.Println("=== Basic Error Handling ===")
	n, err := strconv.Atoi("42")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Parsed:", n)
	}

	n, err = strconv.Atoi("abc")
	if err != nil {
		fmt.Println("Error:", err)
	}

	// =============================================
	// ERROR WRAPPING
	// =============================================

	fmt.Println("\n=== Error Wrapping ===")
	err = processUser("999")
	if err != nil {
		fmt.Println("Error:", err)
		// Output: "Error: processUser: getUser(999): not found"

		// errors.Is desenvuelve automaticamente
		fmt.Println("Is NotFound?", errors.Is(err, ErrNotFound))       // true
		fmt.Println("Is Unauthorized?", errors.Is(err, ErrUnauthorized)) // false
	}

	// =============================================
	// CUSTOM ERROR con errors.As
	// =============================================

	fmt.Println("\n=== Custom Error Types ===")
	err = validateUser("", -5)
	if err != nil {
		fmt.Println("Error:", err)

		// Extraer el tipo concreto
		var valErr *ValidationError
		if errors.As(err, &valErr) {
			fmt.Printf("  Field: %s, Message: %s\n", valErr.Field, valErr.Message)
		}
	}

	// =============================================
	// MULTIPLES ERRORES EN CADENA
	// =============================================

	fmt.Println("\n=== Error Chain ===")
	err = step3()
	if err != nil {
		fmt.Println("Final error:", err)
		// "step3: step2: step1: original problem"

		fmt.Println("Is original?", errors.Is(err, ErrEmpty))
	}

	// =============================================
	// PATRON: manejar diferentes tipos de error
	// =============================================

	fmt.Println("\n=== Error Dispatch ===")
	for _, id := range []string{"1", "2", "3", "bad"} {
		result, err := fetchResource(id)
		switch {
		case err == nil:
			fmt.Printf("  %s: %s\n", id, result)
		case errors.Is(err, ErrNotFound):
			fmt.Printf("  %s: recurso no encontrado\n", id)
		case errors.Is(err, ErrUnauthorized):
			fmt.Printf("  %s: sin permiso\n", id)
		default:
			fmt.Printf("  %s: error inesperado: %v\n", id, err)
		}
	}
}

// --- Funciones ejemplo ---

func getUser(id string) (string, error) {
	if id == "999" {
		return "", ErrNotFound
	}
	return "User-" + id, nil
}

func processUser(id string) error {
	_, err := getUser(id)
	if err != nil {
		return fmt.Errorf("processUser: getUser(%s): %w", id, err)
	}
	return nil
}

func validateUser(name string, age int) error {
	if name == "" {
		return &ValidationError{Field: "name", Message: "cannot be empty"}
	}
	if age < 0 {
		return &ValidationError{Field: "age", Message: "cannot be negative"}
	}
	return nil
}

func step1() error {
	return fmt.Errorf("step1: %w", ErrEmpty)
}

func step2() error {
	if err := step1(); err != nil {
		return fmt.Errorf("step2: %w", err)
	}
	return nil
}

func step3() error {
	if err := step2(); err != nil {
		return fmt.Errorf("step3: %w", err)
	}
	return nil
}

func fetchResource(id string) (string, error) {
	switch id {
	case "1":
		return "Resource A", nil
	case "2":
		return "", ErrNotFound
	case "3":
		return "", ErrUnauthorized
	default:
		return "", fmt.Errorf("invalid resource id: %s", id)
	}
}
