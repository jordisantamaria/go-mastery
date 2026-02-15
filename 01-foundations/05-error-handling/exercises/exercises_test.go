package exercises

import (
	"errors"
	"strings"
	"testing"
)

func TestSafeDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr error
	}{
		{"normal", 10, 2, 5, nil},
		{"decimal", 7, 2, 3.5, nil},
		{"zero dividend", 0, 5, 0, nil},
		{"division by zero", 10, 0, 0, ErrDivisionByZero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeDivide(tt.a, tt.b)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("SafeDivide(%v, %v) error = %v, want %v", tt.a, tt.b, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("SafeDivide(%v, %v) unexpected error: %v", tt.a, tt.b, err)
			}
			if got != tt.want {
				t.Errorf("SafeDivide(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSqrt(t *testing.T) {
	tests := []struct {
		name    string
		n       float64
		want    float64
		wantErr error
		epsilon float64
	}{
		{"sqrt 4", 4, 2, nil, 0.001},
		{"sqrt 9", 9, 3, nil, 0.001},
		{"sqrt 2", 2, 1.4142, nil, 0.001},
		{"sqrt 0", 0, 0, nil, 0.001},
		{"negative", -4, 0, ErrNegativeNumber, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sqrt(tt.n)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Sqrt(%v) error = %v, want %v", tt.n, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Sqrt(%v) unexpected error: %v", tt.n, err)
			}
			diff := got - tt.want
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.epsilon {
				t.Errorf("Sqrt(%v) = %v, want ~%v (diff %v > epsilon %v)", tt.n, got, tt.want, diff, tt.epsilon)
			}
		})
	}
}

func TestParseAge(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr error
	}{
		{"valid", "25", 25, nil},
		{"zero", "0", 0, nil},
		{"max", "150", 150, nil},
		{"empty", "", 0, ErrEmpty},
		{"not number", "abc", 0, nil},      // wraps strconv error, checked separately
		{"negative", "-5", 0, ErrNegativeNumber},
		{"too old", "200", 0, ErrOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAge(tt.input)

			if tt.name == "not number" {
				// Solo verificar que hay error
				if err == nil {
					t.Error("ParseAge(\"abc\") should return error")
				}
				return
			}

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ParseAge(%q) expected error, got nil", tt.input)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseAge(%q) error = %v, want error wrapping %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseAge(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseAge(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestSafeGet(t *testing.T) {
	items := []string{"alpha", "beta", "gamma"}

	tests := []struct {
		name    string
		items   []string
		index   int
		want    string
		wantErr error
	}{
		{"first", items, 0, "alpha", nil},
		{"last", items, 2, "gamma", nil},
		{"negative index", items, -1, "", ErrOutOfRange},
		{"too large", items, 5, "", ErrOutOfRange},
		{"empty slice", []string{}, 0, "", ErrEmpty},
		{"nil slice", nil, 0, "", ErrEmpty},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeGet(tt.items, tt.index)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("SafeGet() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("SafeGet() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("SafeGet() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFindUser(t *testing.T) {
	users := []User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}

	// Found
	u, err := FindUser(users, "Alice")
	if err != nil {
		t.Errorf("FindUser(Alice) unexpected error: %v", err)
	}
	if u.Email != "alice@example.com" {
		t.Errorf("FindUser(Alice).Email = %q, want alice@example.com", u.Email)
	}

	// Not found — should wrap ErrNotFound
	_, err = FindUser(users, "Charlie")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("FindUser(Charlie) should wrap ErrNotFound, got: %v", err)
	}
	if !strings.Contains(err.Error(), "FindUser") {
		t.Errorf("FindUser error should contain context 'FindUser', got: %v", err)
	}

	// Empty name — should wrap ErrEmpty
	_, err = FindUser(users, "")
	if !errors.Is(err, ErrEmpty) {
		t.Errorf("FindUser(\"\") should wrap ErrEmpty, got: %v", err)
	}
}

func TestFieldError(t *testing.T) {
	err := &FieldError{Field: "email", Message: "is required"}
	want := "field email: is required"
	if err.Error() != want {
		t.Errorf("FieldError.Error() = %q, want %q", err.Error(), want)
	}
}

func TestValidateUser(t *testing.T) {
	// Valid
	err := ValidateUser("Jordi", "jordi@email.com", 28)
	if err != nil {
		t.Errorf("ValidateUser(valid) unexpected error: %v", err)
	}

	// Empty name
	err = ValidateUser("", "jordi@email.com", 28)
	var fe *FieldError
	if !errors.As(err, &fe) {
		t.Errorf("ValidateUser(empty name) should return *FieldError, got: %T", err)
	} else if fe.Field != "name" {
		t.Errorf("FieldError.Field = %q, want \"name\"", fe.Field)
	}

	// Invalid email
	err = ValidateUser("Jordi", "no-at-sign", 28)
	if !errors.As(err, &fe) {
		t.Errorf("ValidateUser(bad email) should return *FieldError, got: %T", err)
	} else if fe.Field != "email" {
		t.Errorf("FieldError.Field = %q, want \"email\"", fe.Field)
	}

	// Invalid age
	err = ValidateUser("Jordi", "jordi@email.com", -1)
	if !errors.As(err, &fe) {
		t.Errorf("ValidateUser(bad age) should return *FieldError, got: %T", err)
	} else if fe.Field != "age" {
		t.Errorf("FieldError.Field = %q, want \"age\"", fe.Field)
	}

	err = ValidateUser("Jordi", "jordi@email.com", 200)
	if !errors.As(err, &fe) {
		t.Errorf("ValidateUser(age 200) should return *FieldError")
	}
}

func TestValidateAll(t *testing.T) {
	// All valid
	err := ValidateAll("Jordi", "jordi@email.com", 28)
	if err != nil {
		t.Errorf("ValidateAll(valid) unexpected error: %v", err)
	}

	// Multiple errors
	err = ValidateAll("", "bad-email", -1)
	if err == nil {
		t.Error("ValidateAll(all invalid) should return error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "name") {
		t.Errorf("error should mention name, got: %q", msg)
	}
	if !strings.Contains(msg, "email") {
		t.Errorf("error should mention email, got: %q", msg)
	}
	if !strings.Contains(msg, "age") {
		t.Errorf("error should mention age, got: %q", msg)
	}

	// Single error
	err = ValidateAll("Jordi", "bad-email", 28)
	if err == nil {
		t.Error("ValidateAll(bad email) should return error")
	}
	if !strings.Contains(err.Error(), "email") {
		t.Errorf("error should mention email, got: %q", err.Error())
	}
}
