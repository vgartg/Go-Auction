package api

import (
    "encoding/json"
    "net/http"
)

type Handlers struct {
    repo *repository.PostgresRepo
}

func NewHandlers(repo *repository.PostgresRepo) *Handlers {
    return &Handlers{repo: repo}
}

func (h *Handlers) CreateLot(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotImplemented)
    json.NewEncoder(w).Encode(map[string]string{"error": "not implemented"})
}

func (h *Handlers) GetLot(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotImplemented)
    json.NewEncoder(w).Encode(map[string]string{"error": "not implemented"})
}

func (h *Handlers) PlaceBid(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotImplemented)
    json.NewEncoder(w).Encode(map[string]string{"error": "not implemented"})
}