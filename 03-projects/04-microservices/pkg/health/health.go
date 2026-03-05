// Package health proporciona un sistema de health checks para microservicios.
// Permite registrar funciones de verificacion y expone un endpoint HTTP
// que reporta el estado de salud de todos los componentes registrados.
package health

import (
	"encoding/json"
	"net/http"
	"sync"
)

// Status representa el resultado del health check de un componente.
type Status struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

// Response es la respuesta completa del health check.
type Response struct {
	Status   string   `json:"status"`
	Services []Status `json:"services"`
}

// Checker gestiona los health checks de multiples componentes.
type Checker struct {
	mu     sync.RWMutex
	checks map[string]func() error
}

// NewChecker crea un nuevo Checker.
func NewChecker() *Checker {
	return &Checker{
		checks: make(map[string]func() error),
	}
}

// Register registra una funcion de health check para un servicio.
func (c *Checker) Register(name string, check func() error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = check
}

// CheckAll ejecuta todos los health checks registrados y devuelve el resultado.
func (c *Checker) CheckAll() Response {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var services []Status
	allHealthy := true

	for name, check := range c.checks {
		s := Status{Service: name, Status: "healthy"}
		if err := check(); err != nil {
			s.Status = "unhealthy"
			s.Error = err.Error()
			allHealthy = false
		}
		services = append(services, s)
	}

	overall := "healthy"
	if !allHealthy {
		overall = "unhealthy"
	}

	return Response{
		Status:   overall,
		Services: services,
	}
}

// Handler devuelve un http.Handler que responde con el estado de salud en JSON.
// Devuelve 200 si todos los servicios estan sanos, 503 si alguno falla.
func (c *Checker) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := c.CheckAll()

		w.Header().Set("Content-Type", "application/json")
		if resp.Status != "healthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(resp)
	})
}
