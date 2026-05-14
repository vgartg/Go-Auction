package api

import (
    "github.com/go-chi/chi/v5"
)

func SetupRoutes(r chi.Router, h *Handlers, wsManager *WebSocketManager) {
    r.Route("/api/lots", func(r chi.Router) {
        r.Post("/", h.CreateLot)
        r.Get("/{id}", h.GetLot)
        r.Post("/{id}/bids", h.PlaceBid)
    })
    r.Get("/ws/lots/{id}", wsManager.HandleWebSocket)
}