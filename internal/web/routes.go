package web

import (
	"github.com/go-chi/chi/v5"

	"github.com/vgartg/goauction/internal/auth"
)

func SetupRoutes(r chi.Router, h *Handlers, authSvc *auth.Service) {
	// All web routes get Attach so the header can show login state.
	r.Group(func(r chi.Router) {
		r.Use(authSvc.Attach)

		r.Get("/", h.Home)
		r.Get("/lots/{id}", h.LotPage)
		r.Get("/users/{id}", h.UserProfile)

		r.Get("/login", h.LoginForm)
		r.Post("/login", h.Login)
		r.Get("/register", h.RegisterForm)
		r.Post("/register", h.Register)
		r.Post("/logout", h.Logout)

		// Auth-required pages: middleware redirects to /login if missing.
		r.Group(func(r chi.Router) {
			r.Use(authSvc.Middleware(true))
			r.Get("/lots/new", h.NewLotForm)
			r.Post("/lots", h.CreateLot)
			r.Post("/lots/{id}/bids", h.PlaceBid)
		})
	})
}
