// Package task define el modelo de tarea y sus estados posibles.
// Representa el dominio central de la aplicacion CLI de gestion de tareas.
package task

import "time"

// Status representa el estado actual de una tarea.
type Status string

const (
	// StatusPending indica que la tarea esta pendiente de completar.
	StatusPending Status = "pending"
	// StatusDone indica que la tarea ha sido completada.
	StatusDone Status = "done"
)

// Task representa una tarea individual con su metadatos.
type Task struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	Status    Status     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	DoneAt    *time.Time `json:"done_at,omitempty"`
}

// StatusLabel devuelve la etiqueta en espanol del estado de la tarea.
func (t Task) StatusLabel() string {
	switch t.Status {
	case StatusDone:
		return "completada"
	default:
		return "pendiente"
	}
}
