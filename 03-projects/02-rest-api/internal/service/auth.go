package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/model"
	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/repository"
	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/pkg/jwt"
)

var (
	ErrInvalidCredentials = errors.New("credenciales inválidas")
	ErrEmailAlreadyExists = errors.New("el email ya está registrado")
)

// AuthService contiene la lógica de negocio para autenticación.
type AuthService struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

// NewAuthService crea una nueva instancia de AuthService con inyección de dependencias.
func NewAuthService(userRepo repository.UserRepository, jwtSecret []byte, tokenTTL time.Duration) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		tokenTTL:  tokenTTL,
	}
}

// Register crea un nuevo usuario con contraseña hasheada.
// Genera un salt aleatorio y aplica SHA-256 para almacenar la contraseña de forma segura.
func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, error) {
	// Validar campos
	if errs := req.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("validación fallida: %v", errs)
	}

	// Hashear contraseña con salt
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("error al hashear contraseña: %w", err)
	}

	user := &model.User{
		ID:           generateID(),
		Email:        req.Email,
		PasswordHash: passwordHash,
		Name:         req.Name,
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("error al crear usuario: %w", err)
	}

	return user, nil
}

// Login verifica las credenciales y devuelve un token JWT.
func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (string, *model.User, error) {
	// Validar campos
	if errs := req.Validate(); len(errs) > 0 {
		return "", nil, fmt.Errorf("validación fallida: %v", errs)
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, fmt.Errorf("error al buscar usuario: %w", err)
	}

	// Verificar contraseña
	if !verifyPassword(req.Password, user.PasswordHash) {
		return "", nil, ErrInvalidCredentials
	}

	// Generar JWT
	claims := jwt.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Exp:    time.Now().Add(s.tokenTTL).Unix(),
	}

	token, err := jwt.Sign(claims, s.jwtSecret)
	if err != nil {
		return "", nil, fmt.Errorf("error al generar token: %w", err)
	}

	return token, user, nil
}

// hashPassword genera un hash SHA-256 con salt aleatorio.
// Formato almacenado: "salt:hash" donde ambos están en hexadecimal.
func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	saltHex := hex.EncodeToString(salt)
	hash := sha256.Sum256([]byte(saltHex + password))
	hashHex := hex.EncodeToString(hash[:])

	return saltHex + ":" + hashHex, nil
}

// verifyPassword comprueba si una contraseña coincide con su hash almacenado.
func verifyPassword(password, stored string) bool {
	// Separar salt y hash
	parts := splitOnce(stored, ':')
	if len(parts) != 2 {
		return false
	}

	saltHex := parts[0]
	expectedHash := parts[1]

	hash := sha256.Sum256([]byte(saltHex + password))
	hashHex := hex.EncodeToString(hash[:])

	return hashHex == expectedHash
}

// splitOnce divide una cadena en dos partes por el primer separador encontrado.
func splitOnce(s string, sep byte) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

// generateID genera un ID único usando bytes aleatorios en hexadecimal.
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
