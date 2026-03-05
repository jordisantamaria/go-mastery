// Tests para JSONStore — validacion del almacen de tareas basado en fichero JSON.
// Demuestra: t.TempDir(), table-driven tests, subtests, verificacion de errores.
package task

import (
	"errors"
	"path/filepath"
	"testing"
)

// newTestStore crea un JSONStore temporal para tests.
// Usa t.TempDir() que se limpia automaticamente al finalizar el test.
func newTestStore(t *testing.T) *JSONStore {
	t.Helper()
	dir := t.TempDir()
	return NewJSONStore(filepath.Join(dir, "tasks.json"))
}

func TestAdd(t *testing.T) {
	store := newTestStore(t)

	t.Run("anadir tarea valida", func(t *testing.T) {
		task, err := store.Add("Comprar leche")
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		if task.ID != 1 {
			t.Errorf("ID esperado 1, obtenido %d", task.ID)
		}
		if task.Title != "Comprar leche" {
			t.Errorf("titulo esperado 'Comprar leche', obtenido '%s'", task.Title)
		}
		if task.Status != StatusPending {
			t.Errorf("estado esperado 'pending', obtenido '%s'", task.Status)
		}
		if task.DoneAt != nil {
			t.Error("DoneAt deberia ser nil para tarea nueva")
		}
	})

	t.Run("titulo vacio devuelve error", func(t *testing.T) {
		_, err := store.Add("")
		if !errors.Is(err, ErrEmptyTitle) {
			t.Errorf("esperado ErrEmptyTitle, obtenido: %v", err)
		}
	})

	t.Run("auto-incremento de IDs", func(t *testing.T) {
		s := newTestStore(t)

		t1, _ := s.Add("Tarea 1")
		t2, _ := s.Add("Tarea 2")
		t3, _ := s.Add("Tarea 3")

		if t1.ID != 1 || t2.ID != 2 || t3.ID != 3 {
			t.Errorf("IDs esperados 1,2,3 — obtenidos %d,%d,%d", t1.ID, t2.ID, t3.ID)
		}
	})
}

func TestList(t *testing.T) {
	store := newTestStore(t)

	t.Run("almacen vacio devuelve lista vacia", func(t *testing.T) {
		tasks, err := store.List(false)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if len(tasks) != 0 {
			t.Errorf("esperadas 0 tareas, obtenidas %d", len(tasks))
		}
	})

	// Preparar datos para los siguientes subtests.
	store.Add("Tarea pendiente 1")
	store.Add("Tarea pendiente 2")
	store.Add("Tarea completada")
	store.Complete(3)

	t.Run("listar solo pendientes", func(t *testing.T) {
		tasks, err := store.List(false)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if len(tasks) != 2 {
			t.Errorf("esperadas 2 tareas pendientes, obtenidas %d", len(tasks))
		}
		for _, task := range tasks {
			if task.Status != StatusPending {
				t.Errorf("tarea #%d deberia estar pendiente, estado: %s", task.ID, task.Status)
			}
		}
	})

	t.Run("listar todas incluidas completadas", func(t *testing.T) {
		tasks, err := store.List(true)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if len(tasks) != 3 {
			t.Errorf("esperadas 3 tareas totales, obtenidas %d", len(tasks))
		}
	})
}

func TestComplete(t *testing.T) {
	store := newTestStore(t)

	t.Run("completar tarea existente", func(t *testing.T) {
		store.Add("Tarea a completar")

		err := store.Complete(1)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		tasks, _ := store.List(true)
		if tasks[0].Status != StatusDone {
			t.Error("la tarea deberia estar completada")
		}
		if tasks[0].DoneAt == nil {
			t.Error("DoneAt deberia estar establecido")
		}
	})

	t.Run("completar tarea inexistente", func(t *testing.T) {
		err := store.Complete(999)
		if !errors.Is(err, ErrTaskNotFound) {
			t.Errorf("esperado ErrTaskNotFound, obtenido: %v", err)
		}
	})

	t.Run("completar tarea ya completada", func(t *testing.T) {
		err := store.Complete(1)
		if err == nil {
			t.Error("deberia devolver error al completar tarea ya completada")
		}
	})
}

func TestDelete(t *testing.T) {
	store := newTestStore(t)

	store.Add("Tarea a eliminar")
	store.Add("Tarea que permanece")

	t.Run("eliminar tarea existente", func(t *testing.T) {
		err := store.Delete(1)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		tasks, _ := store.List(true)
		if len(tasks) != 1 {
			t.Errorf("esperada 1 tarea, obtenidas %d", len(tasks))
		}
		if tasks[0].ID != 2 {
			t.Errorf("la tarea restante deberia tener ID 2, obtenido %d", tasks[0].ID)
		}
	})

	t.Run("eliminar tarea inexistente", func(t *testing.T) {
		err := store.Delete(999)
		if !errors.Is(err, ErrTaskNotFound) {
			t.Errorf("esperado ErrTaskNotFound, obtenido: %v", err)
		}
	})

	t.Run("eliminar y verificar persistencia", func(t *testing.T) {
		s := newTestStore(t)

		s.Add("A")
		s.Add("B")
		s.Add("C")

		s.Delete(2)

		tasks, _ := s.List(true)
		if len(tasks) != 2 {
			t.Fatalf("esperadas 2 tareas, obtenidas %d", len(tasks))
		}

		// Verificar que los IDs restantes son correctos.
		ids := make(map[int]bool)
		for _, task := range tasks {
			ids[task.ID] = true
		}
		if !ids[1] || !ids[3] {
			t.Error("deberien quedar las tareas con ID 1 y 3")
		}
	})
}

func TestPersistence(t *testing.T) {
	// Verificar que los datos sobreviven entre instancias del store.
	dir := t.TempDir()
	fp := filepath.Join(dir, "tasks.json")

	// Primera instancia: crear tareas.
	s1 := NewJSONStore(fp)
	s1.Add("Tarea persistente 1")
	s1.Add("Tarea persistente 2")
	s1.Complete(1)

	// Segunda instancia: leer los mismos datos.
	s2 := NewJSONStore(fp)
	tasks, err := s2.List(true)
	if err != nil {
		t.Fatalf("error al leer desde nueva instancia: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("esperadas 2 tareas, obtenidas %d", len(tasks))
	}

	if tasks[0].Status != StatusDone {
		t.Error("la primera tarea deberia estar completada en la segunda instancia")
	}

	// Verificar que el auto-incremento continua correctamente.
	t3, _ := s2.Add("Tarea nueva en instancia 2")
	if t3.ID != 3 {
		t.Errorf("ID esperado 3 (auto-incremento continuo), obtenido %d", t3.ID)
	}
}
