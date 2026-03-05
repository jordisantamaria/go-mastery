package jwt

import (
	"testing"
	"time"
)

// TestSignAndVerify comprueba que un token firmado se puede verificar correctamente.
func TestSignAndVerify(t *testing.T) {
	secret := []byte("test-secret-key-12345")

	claims := Claims{
		UserID: "user-123",
		Email:  "test@example.com",
		Exp:    time.Now().Add(time.Hour).Unix(),
	}

	token, err := Sign(claims, secret)
	if err != nil {
		t.Fatalf("error al firmar token: %v", err)
	}

	if token == "" {
		t.Fatal("el token no debería estar vacío")
	}

	// Verificar el token
	verified, err := Verify(token, secret)
	if err != nil {
		t.Fatalf("error al verificar token: %v", err)
	}

	if verified.UserID != claims.UserID {
		t.Errorf("UserID esperado %q, obtenido %q", claims.UserID, verified.UserID)
	}

	if verified.Email != claims.Email {
		t.Errorf("Email esperado %q, obtenido %q", claims.Email, verified.Email)
	}
}

// TestVerifyExpiredToken comprueba que un token expirado se rechaza.
func TestVerifyExpiredToken(t *testing.T) {
	secret := []byte("test-secret-key-12345")

	claims := Claims{
		UserID: "user-123",
		Email:  "test@example.com",
		Exp:    time.Now().Add(-time.Hour).Unix(), // Expirado hace una hora
	}

	token, err := Sign(claims, secret)
	if err != nil {
		t.Fatalf("error al firmar token: %v", err)
	}

	_, err = Verify(token, secret)
	if err != ErrExpiredToken {
		t.Errorf("se esperaba ErrExpiredToken, obtenido: %v", err)
	}
}

// TestVerifyInvalidSignature comprueba que un token con firma incorrecta se rechaza.
func TestVerifyInvalidSignature(t *testing.T) {
	secret := []byte("test-secret-key-12345")
	wrongSecret := []byte("wrong-secret-key-67890")

	claims := Claims{
		UserID: "user-123",
		Email:  "test@example.com",
		Exp:    time.Now().Add(time.Hour).Unix(),
	}

	token, err := Sign(claims, secret)
	if err != nil {
		t.Fatalf("error al firmar token: %v", err)
	}

	// Verificar con clave incorrecta
	_, err = Verify(token, wrongSecret)
	if err != ErrInvalidToken {
		t.Errorf("se esperaba ErrInvalidToken, obtenido: %v", err)
	}
}

// TestVerifyMalformedToken comprueba que tokens con formato incorrecto se rechazan.
func TestVerifyMalformedToken(t *testing.T) {
	secret := []byte("test-secret-key-12345")

	testCases := []struct {
		name  string
		token string
	}{
		{"vacío", ""},
		{"una parte", "abc"},
		{"dos partes", "abc.def"},
		{"cuatro partes", "abc.def.ghi.jkl"},
		{"base64 inválido", "abc.!!!.def"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Verify(tc.token, secret)
			if err == nil {
				t.Error("se esperaba un error para token malformado")
			}
		})
	}
}

// TestTokenStructure comprueba que el token tiene el formato esperado con tres partes.
func TestTokenStructure(t *testing.T) {
	secret := []byte("test-secret-key-12345")

	claims := Claims{
		UserID: "user-123",
		Email:  "test@example.com",
		Exp:    time.Now().Add(time.Hour).Unix(),
	}

	token, err := Sign(claims, secret)
	if err != nil {
		t.Fatalf("error al firmar token: %v", err)
	}

	// Un JWT debe tener exactamente 3 partes separadas por puntos
	parts := 0
	for _, c := range token {
		if c == '.' {
			parts++
		}
	}
	if parts != 2 {
		t.Errorf("el token debería tener 3 partes (2 puntos), tiene %d puntos", parts)
	}
}
