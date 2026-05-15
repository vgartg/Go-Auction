package auth

import (
	"testing"

	"github.com/vgartg/goauction/internal/models"
)

func TestService_TokenRoundTrip(t *testing.T) {
	s := NewService(nil, "test-secret")
	tok, err := s.issueToken(&models.User{ID: "user-1", Username: "alice"})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	claims, err := s.ParseToken(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != "user-1" || claims.Username != "alice" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestService_RejectsTokenSignedWithDifferentSecret(t *testing.T) {
	s1 := NewService(nil, "secret-A")
	s2 := NewService(nil, "secret-B")
	tok, err := s1.issueToken(&models.User{ID: "u", Username: "x"})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if _, err := s2.ParseToken(tok); err == nil {
		t.Fatal("token signed with secret-A must not validate under secret-B")
	}
}
