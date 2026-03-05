package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/middleware"
	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/model"
	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/service"
)

// TransactionHandler maneja los endpoints CRUD de transacciones.
type TransactionHandler struct {
	txService *service.TransactionService
}

// NewTransactionHandler crea una nueva instancia con inyección de dependencias.
func NewTransactionHandler(txService *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		txService: txService,
	}
}

// createTransactionRequest es el cuerpo esperado para crear una transacción.
type createTransactionRequest struct {
	Type        model.TransactionType `json:"type"`
	Amount      float64               `json:"amount"`
	Category    string                `json:"category"`
	Description string                `json:"description"`
	Date        string                `json:"date"` // Formato: "2006-01-02"
}

// listResponse es la respuesta paginada para listar transacciones.
type listResponse struct {
	Data       []model.Transaction `json:"data"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalPages int                 `json:"total_pages"`
}

// Create maneja POST /api/transactions.
// Crea una nueva transacción asociada al usuario autenticado.
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "usuario no autenticado")
		return
	}

	var req createTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
		return
	}

	// Parsear la fecha
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, "formato de fecha inválido, use YYYY-MM-DD")
		return
	}

	tx := &model.Transaction{
		Type:        req.Type,
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
		Date:        date,
	}

	if err := h.txService.Create(r.Context(), userID, tx); err != nil {
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "error al crear transacción")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tx)
}

// List maneja GET /api/transactions.
// Devuelve las transacciones del usuario autenticado con filtros y paginación.
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "usuario no autenticado")
		return
	}

	// Parsear query params para filtros
	filter := model.TransactionFilter{}

	if t := r.URL.Query().Get("type"); t != "" {
		filter.Type = model.TransactionType(t)
	}

	if c := r.URL.Query().Get("category"); c != "" {
		filter.Category = c
	}

	if from := r.URL.Query().Get("from"); from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			filter.From = t
		}
	}

	if to := r.URL.Query().Get("to"); to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			// Incluir todo el día final
			filter.To = t.Add(24*time.Hour - time.Nanosecond)
		}
	}

	if p := r.URL.Query().Get("page"); p != "" {
		if page, err := strconv.Atoi(p); err == nil {
			filter.Page = page
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if limit, err := strconv.Atoi(l); err == nil {
			filter.Limit = limit
		}
	}

	// Valores por defecto
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}

	results, total, err := h.txService.List(r.Context(), userID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error al listar transacciones")
		return
	}

	// Calcular total de páginas
	totalPages := total / filter.Limit
	if total%filter.Limit != 0 {
		totalPages++
	}

	resp := listResponse{
		Data:       results,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// GetByID maneja GET /api/transactions/{id}.
// Devuelve una transacción específica del usuario autenticado.
func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "usuario no autenticado")
		return
	}

	txID := r.PathValue("id")
	if txID == "" {
		writeError(w, http.StatusBadRequest, "ID de transacción requerido")
		return
	}

	tx, err := h.txService.GetByID(r.Context(), userID, txID)
	if err != nil {
		if errors.Is(err, service.ErrTransactionNotFound) {
			writeError(w, http.StatusNotFound, "transacción no encontrada")
			return
		}
		if errors.Is(err, service.ErrUnauthorized) {
			writeError(w, http.StatusForbidden, "no tienes permiso para acceder a esta transacción")
			return
		}
		writeError(w, http.StatusInternalServerError, "error al obtener transacción")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tx)
}

// Update maneja PUT /api/transactions/{id}.
// Actualiza una transacción existente del usuario autenticado.
func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "usuario no autenticado")
		return
	}

	txID := r.PathValue("id")
	if txID == "" {
		writeError(w, http.StatusBadRequest, "ID de transacción requerido")
		return
	}

	var req createTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
		return
	}

	var date time.Time
	if req.Date != "" {
		var err error
		date, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			writeError(w, http.StatusBadRequest, "formato de fecha inválido, use YYYY-MM-DD")
			return
		}
	}

	updates := &model.Transaction{
		Type:        req.Type,
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
		Date:        date,
	}

	updated, err := h.txService.Update(r.Context(), userID, txID, updates)
	if err != nil {
		if errors.Is(err, service.ErrTransactionNotFound) {
			writeError(w, http.StatusNotFound, "transacción no encontrada")
			return
		}
		if errors.Is(err, service.ErrUnauthorized) {
			writeError(w, http.StatusForbidden, "no tienes permiso para modificar esta transacción")
			return
		}
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "error al actualizar transacción")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

// Delete maneja DELETE /api/transactions/{id}.
// Elimina una transacción del usuario autenticado.
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "usuario no autenticado")
		return
	}

	txID := r.PathValue("id")
	if txID == "" {
		writeError(w, http.StatusBadRequest, "ID de transacción requerido")
		return
	}

	if err := h.txService.Delete(r.Context(), userID, txID); err != nil {
		if errors.Is(err, service.ErrTransactionNotFound) {
			writeError(w, http.StatusNotFound, "transacción no encontrada")
			return
		}
		if errors.Is(err, service.ErrUnauthorized) {
			writeError(w, http.StatusForbidden, "no tienes permiso para eliminar esta transacción")
			return
		}
		writeError(w, http.StatusInternalServerError, "error al eliminar transacción")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
