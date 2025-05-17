package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/domain/fact"
	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/service"
)

type FactHandler struct {
	Service     *service.FactService
	AdminSecret string
}

func NewFactHandler(service *service.FactService, secret string) *FactHandler {
	return &FactHandler{Service: service, AdminSecret: secret}
}

func (h *FactHandler) CreateFact(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Admin-Key") != h.AdminSecret {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var fact fact.ListenFact
	if err := json.NewDecoder(r.Body).Decode(&fact); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.Service.Store(&fact); err != nil {
		http.Error(w, "Failed to store listen fact", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fact)
}
