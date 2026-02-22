package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/vgartg/goauction/internal/models"
	"github.com/vgartg/goauction/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

const (
	tokenTTL      = 24 * time.Hour
	cookieName    = "goauction_session"
	cookieMaxAge  = int(24 * 60 * 60)
	minPasswordLn = 6
)

type Service struct {
	users     repository.UserRepository
	jwtSecret []byte
}

func NewService(users repository.UserRepository, jwtSecret string) *Service {
	return &Service{
		users:     users,
		jwtSecret: []byte(jwtSecret),
	}
}

type Claims struct {
	UserID   string `json:"uid"`
	Username string `json:"un"`
	jwt.RegisteredClaims
}

func (s *Service) Register(ctx context.Context, username, email, password string) (*models.User, string, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(strings.ToLower(email))
	if username == "" || email == "" {
		return nil, "", errors.New("username and email are required")
	}
	if len(password) < minPasswordLn {
		return nil, "", fmt.Errorf("password must be at least %d characters", minPasswordLn)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := s.users.CreateUser(ctx, user); err != nil {
		return nil, "", err
	}
	token, err := s.issueToken(user)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}
	token, err := s.issueToken(user)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *Service) issueToken(user *models.User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(s.jwtSecret)
}

func (s *Service) ParseToken(raw string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// SetSessionCookie writes the JWT as an HTTP-only cookie for the web UI.
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func extractToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); h != "" {
		if strings.HasPrefix(h, "Bearer ") {
			return strings.TrimPrefix(h, "Bearer ")
		}
	}
	if c, err := r.Cookie(cookieName); err == nil {
		return c.Value
	}
	return ""
}

type ctxKey int

const userCtxKey ctxKey = 0

// Middleware enforces authentication. On failure returns 401 for API or
// redirects to /login for HTML requests.
func (s *Service) Middleware(redirectHTML bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := s.tryAuth(r)
			if !ok {
				if redirectHTML {
					http.Redirect(w, r, "/login?next="+r.URL.Path, http.StatusSeeOther)
					return
				}
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userCtxKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Attach injects claims into context if present but does not require auth.
// Useful for pages that render differently for guests vs logged-in users.
func (s *Service) Attach(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if claims, ok := s.tryAuth(r); ok {
			ctx := context.WithValue(r.Context(), userCtxKey, claims)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Service) tryAuth(r *http.Request) (*Claims, bool) {
	raw := extractToken(r)
	if raw == "" {
		return nil, false
	}
	claims, err := s.ParseToken(raw)
	if err != nil {
		return nil, false
	}
	return claims, true
}

// FromContext returns the authenticated claims, if any.
func FromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(userCtxKey).(*Claims)
	return c, ok
}
