package web

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"

	"github.com/vgartg/goauction/internal/auction"
	"github.com/vgartg/goauction/internal/auth"
	"github.com/vgartg/goauction/internal/models"
	"github.com/vgartg/goauction/internal/repository"
	"github.com/vgartg/goauction/internal/web/views"
)

const homePageSize = 24

type Handlers struct {
	engine *auction.Engine
	auth   *auth.Service
	users  repository.UserRepository
}

func NewHandlers(engine *auction.Engine, authSvc *auth.Service, users repository.UserRepository) *Handlers {
	return &Handlers{engine: engine, auth: authSvc, users: users}
}

func (h *Handlers) layoutData(r *http.Request, title string) views.LayoutData {
	return views.LayoutData{
		Title:       title,
		CurrentUser: currentUserFromContext(r.Context()),
	}
}

func currentUserFromContext(ctx context.Context) *models.User {
	c, ok := auth.FromContext(ctx)
	if !ok {
		return nil
	}
	return &models.User{ID: c.UserID, Username: c.Username}
}

func render(w http.ResponseWriter, r *http.Request, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = c.Render(r.Context(), w)
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	opts := repository.LotListOptions{Limit: homePageSize}
	if s := r.URL.Query().Get("status"); s != "" {
		opts.Status = models.LotStatus(s)
	}
	lots, err := h.engine.ListLots(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render(w, r, views.Home(h.layoutData(r, "Auctions"), lots, string(opts.Status)))
}

func (h *Handlers) LotPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	lot, err := h.engine.GetLot(r.Context(), id)
	if err != nil {
		http.Error(w, "lot not found", http.StatusNotFound)
		return
	}
	bids, err := h.engine.RecentBids(r.Context(), id, 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render(w, r, views.LotDetail(views.LotDetailData{
		Layout: h.layoutData(r, lot.Title),
		Lot:    lot,
		Bids:   bids,
	}))
}

func (h *Handlers) NewLotForm(w http.ResponseWriter, r *http.Request) {
	render(w, r, views.NewLot(h.layoutData(r, "Create lot"), ""))
}

func (h *Handlers) CreateLot(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	startPrice, _ := strconv.ParseFloat(r.FormValue("start_price"), 64)
	minStep, _ := strconv.ParseFloat(r.FormValue("min_step"), 64)
	closingStr := r.FormValue("closing_at")
	tzOffsetMinutes, _ := strconv.Atoi(r.FormValue("tz_offset_minutes"))

	loc := time.FixedZone("client", -tzOffsetMinutes*60)
	closingAt, err := time.ParseInLocation("2006-01-02T15:04", closingStr, loc)
	if err != nil {
		render(w, r, views.NewLot(h.layoutData(r, "Create lot"), "Invalid closing time"))
		return
	}
	closingAt = closingAt.UTC()
	if title == "" || startPrice <= 0 || minStep <= 0 || closingAt.Before(time.Now()) {
		render(w, r, views.NewLot(h.layoutData(r, "Create lot"), "All fields are required and must be valid"))
		return
	}
	lot, err := h.engine.CreateLot(r.Context(), title, startPrice, minStep, closingAt)
	if err != nil {
		render(w, r, views.NewLot(h.layoutData(r, "Create lot"), err.Error()))
		return
	}
	http.Redirect(w, r, "/lots/"+lot.ID, http.StatusSeeOther)
}

func (h *Handlers) PlaceBid(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	lotID := chi.URLParam(r, "id")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "bad amount", http.StatusBadRequest)
		return
	}
	lot, err := h.engine.PlaceBid(r.Context(), lotID, claims.UserID, amount)
	if err != nil {
		current, gerr := h.engine.GetLot(r.Context(), lotID)
		if gerr != nil {
			http.Error(w, gerr.Error(), http.StatusInternalServerError)
			return
		}
		render(w, r, views.BidPanel(current, currentUserFromContext(r.Context()), err.Error()))
		return
	}
	render(w, r, views.BidPanel(lot, currentUserFromContext(r.Context()), ""))
}

func (h *Handlers) LoginForm(w http.ResponseWriter, r *http.Request) {
	render(w, r, views.Login(h.layoutData(r, "Log in"), "", "", r.URL.Query().Get("next")))
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")
	next := r.FormValue("next")
	_, token, err := h.auth.Login(r.Context(), email, password)
	if err != nil {
		render(w, r, views.Login(h.layoutData(r, "Log in"), "Invalid credentials", email, next))
		return
	}
	auth.SetSessionCookie(w, token)
	if next == "" {
		next = "/"
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (h *Handlers) RegisterForm(w http.ResponseWriter, r *http.Request) {
	render(w, r, views.Register(h.layoutData(r, "Sign up"), "", "", ""))
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	_, token, err := h.auth.Register(r.Context(), username, email, password)
	if err != nil {
		msg := err.Error()
		if errors.Is(err, repository.ErrUserExists) {
			msg = "Username or email already taken"
		}
		render(w, r, views.Register(h.layoutData(r, "Sign up"), msg, username, email))
		return
	}
	auth.SetSessionCookie(w, token)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	auth.ClearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) UserProfile(w http.ResponseWriter, r *http.Request) {
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
	render(w, r, views.Profile(views.ProfileData{
		Layout: h.layoutData(r, stats.Username),
		Stats:  stats,
	}))
}
