package userservice

import (
	"fmt"
	"testing"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/04-microservices/pkg/model"
)

// TestCreateUser verifica que se puede crear un usuario correctamente.
func TestCreateUser(t *testing.T) {
	svc := NewService()

	args := model.CreateUserArgs{Name: "Ana Garcia", Email: "ana@example.com"}
	var user model.User

	if err := svc.Create(args, &user); err != nil {
		t.Fatalf("error al crear usuario: %v", err)
	}

	if user.ID == "" {
		t.Error("el ID del usuario no deberia estar vacio")
	}
	if user.Name != "Ana Garcia" {
		t.Errorf("nombre esperado 'Ana Garcia', obtenido '%s'", user.Name)
	}
	if user.Email != "ana@example.com" {
		t.Errorf("email esperado 'ana@example.com', obtenido '%s'", user.Email)
	}
	if user.CreatedAt.IsZero() {
		t.Error("CreatedAt no deberia estar vacio")
	}
}

// TestCreateUserValidation verifica las validaciones al crear un usuario.
func TestCreateUserValidation(t *testing.T) {
	svc := NewService()

	tests := []struct {
		name    string
		args    model.CreateUserArgs
		wantErr error
	}{
		{
			name:    "nombre vacio",
			args:    model.CreateUserArgs{Name: "", Email: "test@example.com"},
			wantErr: ErrInvalidName,
		},
		{
			name:    "email vacio",
			args:    model.CreateUserArgs{Name: "Test", Email: ""},
			wantErr: ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var user model.User
			err := svc.Create(tt.args, &user)
			if err == nil {
				t.Fatal("se esperaba un error")
			}
			if err != tt.wantErr {
				t.Errorf("error esperado '%v', obtenido '%v'", tt.wantErr, err)
			}
		})
	}
}

// TestCreateUserDuplicateEmail verifica que no se puede crear un usuario con email duplicado.
func TestCreateUserDuplicateEmail(t *testing.T) {
	svc := NewService()

	args := model.CreateUserArgs{Name: "Ana", Email: "ana@example.com"}
	var user model.User
	if err := svc.Create(args, &user); err != nil {
		t.Fatalf("error al crear primer usuario: %v", err)
	}

	args2 := model.CreateUserArgs{Name: "Otra Ana", Email: "ana@example.com"}
	var user2 model.User
	err := svc.Create(args2, &user2)
	if err != ErrEmailExists {
		t.Errorf("error esperado '%v', obtenido '%v'", ErrEmailExists, err)
	}
}

// TestGetByID verifica que se puede obtener un usuario por ID.
func TestGetByID(t *testing.T) {
	svc := NewService()

	// Crear usuario primero
	createArgs := model.CreateUserArgs{Name: "Carlos", Email: "carlos@example.com"}
	var created model.User
	if err := svc.Create(createArgs, &created); err != nil {
		t.Fatalf("error al crear usuario: %v", err)
	}

	// Buscar por ID
	var found model.User
	if err := svc.GetByID(model.GetByIDArgs{ID: created.ID}, &found); err != nil {
		t.Fatalf("error al buscar usuario: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("ID esperado '%s', obtenido '%s'", created.ID, found.ID)
	}
}

// TestGetByIDNotFound verifica el error cuando el usuario no existe.
func TestGetByIDNotFound(t *testing.T) {
	svc := NewService()

	var user model.User
	err := svc.GetByID(model.GetByIDArgs{ID: "no-existe"}, &user)
	if err != ErrUserNotFound {
		t.Errorf("error esperado '%v', obtenido '%v'", ErrUserNotFound, err)
	}
}

// TestListUsers verifica que se pueden listar todos los usuarios.
func TestListUsers(t *testing.T) {
	svc := NewService()

	// Crear varios usuarios
	names := []string{"Ana", "Carlos", "Elena"}
	for i, name := range names {
		args := model.CreateUserArgs{
			Name:  name,
			Email: fmt.Sprintf("%s@example.com", name),
		}
		var user model.User
		if err := svc.Create(args, &user); err != nil {
			t.Fatalf("error al crear usuario %d: %v", i, err)
		}
	}

	var users []model.User
	if err := svc.List(model.ListArgs{}, &users); err != nil {
		t.Fatalf("error al listar usuarios: %v", err)
	}

	if len(users) != 3 {
		t.Errorf("se esperaban 3 usuarios, se obtuvieron %d", len(users))
	}
}

// TestUpdateUser verifica que se puede actualizar un usuario.
func TestUpdateUser(t *testing.T) {
	svc := NewService()

	// Crear usuario
	createArgs := model.CreateUserArgs{Name: "Ana", Email: "ana@example.com"}
	var created model.User
	if err := svc.Create(createArgs, &created); err != nil {
		t.Fatalf("error al crear usuario: %v", err)
	}

	// Actualizar
	updateArgs := model.UpdateUserArgs{
		ID:    created.ID,
		Name:  "Ana Garcia",
		Email: "ana.garcia@example.com",
	}
	var updated model.User
	if err := svc.Update(updateArgs, &updated); err != nil {
		t.Fatalf("error al actualizar usuario: %v", err)
	}

	if updated.Name != "Ana Garcia" {
		t.Errorf("nombre esperado 'Ana Garcia', obtenido '%s'", updated.Name)
	}
	if updated.Email != "ana.garcia@example.com" {
		t.Errorf("email esperado 'ana.garcia@example.com', obtenido '%s'", updated.Email)
	}
}

// TestDeleteUser verifica que se puede eliminar un usuario.
func TestDeleteUser(t *testing.T) {
	svc := NewService()

	// Crear usuario
	createArgs := model.CreateUserArgs{Name: "Ana", Email: "ana@example.com"}
	var created model.User
	if err := svc.Create(createArgs, &created); err != nil {
		t.Fatalf("error al crear usuario: %v", err)
	}

	// Eliminar
	var deleted bool
	if err := svc.Delete(model.DeleteArgs{ID: created.ID}, &deleted); err != nil {
		t.Fatalf("error al eliminar usuario: %v", err)
	}
	if !deleted {
		t.Error("se esperaba deleted = true")
	}

	// Verificar que ya no existe
	var user model.User
	err := svc.GetByID(model.GetByIDArgs{ID: created.ID}, &user)
	if err != ErrUserNotFound {
		t.Errorf("error esperado '%v', obtenido '%v'", ErrUserNotFound, err)
	}
}
