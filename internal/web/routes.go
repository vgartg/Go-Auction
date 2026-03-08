package web

import (
	"github.com/go-chi/chi/v5"

	"github.com/vgartg/goauction/internal/auth"
	"github.com/vgartg/goauction/internal/httpx"
)

func SetupRoutes(
	r chi.Router,
	h *Handlers,
	authSvc *auth.Service,
	bidLimiter *httpx.RateLimiter,
	authLimiter *httpx.RateLimiter,
) {
	r.Group(func(r chi.Router) {
		r.Use(authSvc.Attach)

		r.Get("/", h.Home)
		r.Get("/lots/{id}", h.LotPage)
		r.Get("/users/{id}", h.UserProfile)

		r.Group(func(r chi.Router) {
			r.Use(authLimiter.Middleware())
			r.Get("/login", h.LoginForm)
			r.Post("/login", h.Login)
			r.Get("/register", h.RegisterForm)
			r.Post("/register", h.Register)
		})
		r.Post("/logout", h.Logout)

		r.Group(func(r chi.Router) {
			r.Use(authSvc.Middleware(true))
			r.Get("/lots/new", h.NewLotForm)
			r.Post("/lots", h.CreateLot)
			r.With(bidLimiter.Middleware()).Post("/lots/{id}/bids", h.PlaceBid)
		})
	})
}
