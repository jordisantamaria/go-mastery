// Package jwt implementa firma y verificación de tokens JWT con HS256.
// No usa dependencias externas — solo stdlib de Go.
package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("token inválido")
	ErrExpiredToken = errors.New("token expirado")
)

// Header representa la cabecera del JWT.
type Header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// Claims contiene los datos del payload del JWT.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Exp    int64  `json:"exp"`
}

// IsExpired comprueba si el token ha expirado.
func (c *Claims) IsExpired() bool {
	return time.Now().Unix() > c.Exp
}

// Sign genera un token JWT firmado con HS256.
// Codifica header y payload en Base64URL, luego firma con HMAC-SHA256.
func Sign(claims Claims, secret []byte) (string, error) {
	header := Header{
		Alg: "HS256",
		Typ: "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("error al serializar header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("error al serializar claims: %w", err)
	}

	headerEncoded := base64URLEncode(headerJSON)
	claimsEncoded := base64URLEncode(claimsJSON)

	signingInput := headerEncoded + "." + claimsEncoded
	signature := sign([]byte(signingInput), secret)
	signatureEncoded := base64URLEncode(signature)

	return signingInput + "." + signatureEncoded, nil
}

// Verify valida un token JWT y devuelve los claims si es válido.
// Comprueba la firma HS256 y la expiración.
func Verify(token string, secret []byte) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	signature, err := base64URLDecode(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}

	expectedSignature := sign([]byte(signingInput), secret)
	if !hmac.Equal(signature, expectedSignature) {
		return nil, ErrInvalidToken
	}

	claimsJSON, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if claims.IsExpired() {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

// sign genera la firma HMAC-SHA256.
func sign(data, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(data)
	return mac.Sum(nil)
}

// base64URLEncode codifica bytes a Base64URL sin padding (estándar JWT).
func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// base64URLDecode decodifica una cadena Base64URL sin padding.
func base64URLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
