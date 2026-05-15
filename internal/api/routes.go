package api

import (
	"github.com/go-chi/chi/v5"

	"github.com/vgartg/goauction/internal/auth"
)

func SetupRoutes(r chi.Router, h *Handlers, wsManager *WebSocketManager, authSvc *auth.Service) {
	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
			r.With(authSvc.Middleware(false)).Get("/me", h.Me)
		})

		r.Route("/lots", func(r chi.Router) {
			r.Get("/", h.ListLots)
			r.Get("/{id}", h.GetLot)
			r.With(authSvc.Middleware(false)).Post("/", h.CreateLot)
			r.With(authSvc.Middleware(false)).Post("/{id}/bids", h.PlaceBid)
		})

		r.Get("/users/{id}/stats", h.UserStats)
	})

	r.Get("/ws/lots/{id}", wsManager.HandleWebSocket)
}
