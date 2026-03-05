// Package discovery implementa un registro de servicios simple en memoria.
// En produccion se usaria Consul, etcd o el DNS de Kubernetes,
// pero este registro demuestra el patron de service discovery.
package discovery

import (
	"errors"
	"sync"
)

// ErrServiceNotFound se devuelve cuando no se encuentra un servicio registrado.
var ErrServiceNotFound = errors.New("servicio no encontrado en el registro")

// Registry es un registro de servicios en memoria.
// Permite registrar, desregistrar y resolver direcciones de servicios.
type Registry struct {
	mu       sync.RWMutex
	services map[string][]string // nombre → []direccion
}

// NewRegistry crea un nuevo Registry vacio.
func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string][]string),
	}
}

// Register registra una direccion para un servicio.
func (r *Registry) Register(name, addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Evitar duplicados
	for _, existing := range r.services[name] {
		if existing == addr {
			return
		}
	}
	r.services[name] = append(r.services[name], addr)
}

// Deregister elimina una direccion de un servicio.
func (r *Registry) Deregister(name, addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	addrs := r.services[name]
	for i, a := range addrs {
		if a == addr {
			r.services[name] = append(addrs[:i], addrs[i+1:]...)
			return
		}
	}
}

// Resolve devuelve una direccion para el servicio solicitado.
// Devuelve la primera direccion disponible (round-robin simple).
func (r *Registry) Resolve(name string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	addrs, ok := r.services[name]
	if !ok || len(addrs) == 0 {
		return "", ErrServiceNotFound
	}

	// Devolver la primera direccion disponible
	return addrs[0], nil
}

// List devuelve todas las direcciones registradas para un servicio.
func (r *Registry) List(name string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	addrs := r.services[name]
	result := make([]string, len(addrs))
	copy(result, addrs)
	return result
}
