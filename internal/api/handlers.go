package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/vgartg/goauction/internal/auction"
	"github.com/vgartg/goauction/internal/auth"
	"github.com/vgartg/goauction/internal/repository"
)

type Handlers struct {
	engine *auction.Engine
	auth   *auth.Service
	users  repository.UserRepository
}

func NewHandlers(engine *auction.Engine, authSvc *auth.Service, users repository.UserRepository) *Handlers {
	return &Handlers{engine: engine, auth: authSvc, users: users}
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
	if req.Title == "" || req.StartPrice <= 0 || req.MinStep <= 0 || req.ClosingAt.Before(time.Now()) {
		http.Error(w, "invalid lot parameters", http.StatusBadRequest)
		return
	}
	lot, err := h.engine.CreateLot(r.Context(), req.Title, req.StartPrice, req.MinStep, req.ClosingAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, lot)
}

func (h *Handlers) GetLot(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	lot, err := h.engine.GetLot(r.Context(), id)
	if err != nil {
		http.Error(w, "lot not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, lot)
}

func (h *Handlers) ListLots(w http.ResponseWriter, r *http.Request) {
	lots, err := h.engine.ListLots(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, lots)
}

type PlaceBidRequest struct {
	Amount float64 `json:"amount"`
}

func (h *Handlers) PlaceBid(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	lotID := chi.URLParam(r, "id")
	var req PlaceBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	lot, err := h.engine.PlaceBid(r.Context(), lotID, claims.UserID, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, lot)
}

type authRequest struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	user, token, err := h.auth.Register(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	auth.SetSessionCookie(w, token)
	writeJSON(w, http.StatusCreated, authResponse{Token: token, UserID: user.ID, Username: user.Username})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	user, token, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	auth.SetSessionCookie(w, token)
	writeJSON(w, http.StatusOK, authResponse{Token: token, UserID: user.ID, Username: user.Username})
}

func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := h.users.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handlers) UserStats(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	stats, err := h.users.GetUserStats(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload) // nolint:errcheck
}
