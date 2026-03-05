package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthHandler maneja el endpoint de health check.
type HealthHandler struct {
	startTime time.Time
}

// NewHealthHandler crea una nueva instancia del handler de salud.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
	}
}

// healthResponse es la respuesta del endpoint de health check.
type healthResponse struct {
	Status  string `json:"status"`
	Uptime  string `json:"uptime"`
	Version string `json:"version"`
}

// ServeHTTP responde con el estado del servicio, tiempo de actividad y versión.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := healthResponse{
		Status:  "ok",
		Uptime:  time.Since(h.startTime).String(),
		Version: "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
