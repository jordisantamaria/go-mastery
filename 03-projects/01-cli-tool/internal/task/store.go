// Archivo store.go — Interfaz de almacenamiento y su implementacion con fichero JSON.
// Demuestra: interfaces, manejo de ficheros, codificacion JSON, mutex para concurrencia.
package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

// Errores comunes del almacen de tareas.
var (
	ErrTaskNotFound = errors.New("tarea no encontrada")
	ErrEmptyTitle   = errors.New("el titulo de la tarea no puede estar vacio")
)

// Store define la interfaz para cualquier almacen de tareas.
// Permite intercambiar la implementacion (JSON, SQLite, etc.) sin cambiar la logica.
type Store interface {
	// Add crea una nueva tarea con el titulo dado y la devuelve.
	Add(title string) (Task, error)
	// List devuelve las tareas. Si includeCompleted es false, solo las pendientes.
	List(includeCompleted bool) ([]Task, error)
	// Complete marca la tarea con el ID dado como completada.
	Complete(id int) error
	// Delete elimina la tarea con el ID dado.
	Delete(id int) error
}

// jsonData es la estructura interna que se serializa al fichero JSON.
// Mantiene un contador para auto-incrementar los IDs.
type jsonData struct {
	NextID int    `json:"next_id"`
	Tasks  []Task `json:"tasks"`
}

// JSONStore implementa Store usando un fichero JSON como persistencia.
// Es seguro para uso concurrente gracias al mutex.
type JSONStore struct {
	filepath string
	mu       sync.Mutex
}

// NewJSONStore crea un nuevo almacen que persiste en el fichero indicado.
// Si el fichero no existe, se creara automaticamente al anadir la primera tarea.
func NewJSONStore(filepath string) *JSONStore {
	return &JSONStore{filepath: filepath}
}

// Add crea una nueva tarea pendiente y la persiste en el fichero JSON.
func (s *JSONStore) Add(title string) (Task, error) {
	if title == "" {
		return Task{}, ErrEmptyTitle
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return Task{}, fmt.Errorf("error al cargar tareas: %w", err)
	}

	task := Task{
		ID:        data.NextID,
		Title:     title,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}

	data.Tasks = append(data.Tasks, task)
	data.NextID++

	if err := s.save(data); err != nil {
		return Task{}, fmt.Errorf("error al guardar tarea: %w", err)
	}

	return task, nil
}

// List devuelve las tareas almacenadas.
// Si includeCompleted es false, filtra las tareas completadas.
func (s *JSONStore) List(includeCompleted bool) ([]Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return nil, fmt.Errorf("error al cargar tareas: %w", err)
	}

	if includeCompleted {
		return data.Tasks, nil
	}

	// Filtrar solo las tareas pendientes.
	var pending []Task
	for _, t := range data.Tasks {
		if t.Status == StatusPending {
			pending = append(pending, t)
		}
	}
	return pending, nil
}

// Complete marca como completada la tarea con el ID dado.
// Devuelve ErrTaskNotFound si no existe una tarea con ese ID.
func (s *JSONStore) Complete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return fmt.Errorf("error al cargar tareas: %w", err)
	}

	for i, t := range data.Tasks {
		if t.ID == id {
			if t.Status == StatusDone {
				return fmt.Errorf("la tarea #%d ya esta completada", id)
			}
			now := time.Now()
			data.Tasks[i].Status = StatusDone
			data.Tasks[i].DoneAt = &now
			return s.save(data)
		}
	}

	return fmt.Errorf("%w: #%d", ErrTaskNotFound, id)
}

// Delete elimina la tarea con el ID dado del almacen.
// Devuelve ErrTaskNotFound si no existe una tarea con ese ID.
func (s *JSONStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return fmt.Errorf("error al cargar tareas: %w", err)
	}

	for i, t := range data.Tasks {
		if t.ID == id {
			// Eliminar el elemento manteniendo el orden.
			data.Tasks = append(data.Tasks[:i], data.Tasks[i+1:]...)
			return s.save(data)
		}
	}

	return fmt.Errorf("%w: #%d", ErrTaskNotFound, id)
}

// load lee el fichero JSON y devuelve los datos parseados.
// Si el fichero no existe, devuelve datos vacios con NextID = 1.
func (s *JSONStore) load() (jsonData, error) {
	data := jsonData{NextID: 1}

	content, err := os.ReadFile(s.filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data, nil
		}
		return data, err
	}

	// Fichero vacio — devolver datos por defecto.
	if len(content) == 0 {
		return data, nil
	}

	if err := json.Unmarshal(content, &data); err != nil {
		return jsonData{NextID: 1}, fmt.Errorf("error al parsear JSON: %w", err)
	}

	return data, nil
}

// save escribe los datos al fichero JSON con formato legible.
func (s *JSONStore) save(data jsonData) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error al serializar JSON: %w", err)
	}

	return os.WriteFile(s.filepath, content, 0644)
}
