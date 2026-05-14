package api

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"

    "github.com/vgartg/goauction/internal/auction"
)

type Handlers struct {
    engine *auction.Engine
}

func NewHandlers(engine *auction.Engine) *Handlers {
    return &Handlers{engine: engine}
}

type CreateLotRequest struct {
    Title      string    `json:"title"`
    StartPrice float64   `json:"start_price"`
    MinStep    float64   `json:"min_step"`
    ClosingAt  time.Time `json:"closing_at"`
}

func (h *Handlers) CreateLot(w http.ResponseWriter, r *http.Request) {
    var req CreateLotRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }
    lot, err := h.engine.CreateLot(r.Context(), req.Title, req.StartPrice, req.MinStep, req.ClosingAt)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(lot)
}

func (h *Handlers) GetLot(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    lot, err := h.engine.GetLot(r.Context(), id)
    if err != nil {
        http.Error(w, "lot not found", http.StatusNotFound)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(lot)
}

type PlaceBidRequest struct {
    UserID string  `json:"user_id"`
    Amount float64 `json:"amount"`
}

func (h *Handlers) PlaceBid(w http.ResponseWriter, r *http.Request) {
    lotID := chi.URLParam(r, "id")
    var req PlaceBidRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }
    lot, err := h.engine.PlaceBid(r.Context(), lotID, req.UserID, req.Amount)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(lot)
}