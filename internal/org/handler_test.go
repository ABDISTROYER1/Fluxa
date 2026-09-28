package org_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fluxa/fluxa/internal/auth"
	"github.com/fluxa/fluxa/internal/domain"
	"github.com/fluxa/fluxa/internal/org"
	"github.com/go-chi/chi/v5"
)

type mockService struct {
	acceptInviteResp *auth.AuthResponse
	acceptInviteErr  error
}

func (m *mockService) InviteMember(ctx context.Context, email, role string) (*domain.OrgInvite, error) {
	return nil, nil
}

func (m *mockService) AcceptInvite(ctx context.Context, req org.AcceptInviteRequest) (*auth.AuthResponse, error) {
	return m.acceptInviteResp, m.acceptInviteErr
}

func (m *mockService) ListMembers(ctx context.Context) ([]*domain.OrgMember, error) {
	return nil, nil
}

func (m *mockService) UpdateRole(ctx context.Context, memberID, role string) error {
	return nil
}

func (m *mockService) RemoveMember(ctx context.Context, memberID string) error {
	return nil
}

func setupOrgRouter(svc org.Service) http.Handler {
	r := chi.NewRouter()
	h := org.NewHandler(svc)
	r.Route("/org", h.Routes())
	return r
}

func TestAcceptInvite_ErrorHandling(t *testing.T) {
	t.Run("Internal database error does not leak to client", func(t *testing.T) {
		svc := &mockService{
			acceptInviteErr: errors.New("pq: deadlock detected on relation 'org_members' tx: 88123"),
		}
		router := setupOrgRouter(svc)

		body := `{"token":"inv-token-1","name":"New User","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/org/invites/accept", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
		respBody := rec.Body.String()
		if strings.Contains(respBody, "deadlock") || strings.Contains(respBody, "org_members") || strings.Contains(respBody, "88123") {
			t.Fatalf("internal error text was leaked: %q", respBody)
		}
		if strings.TrimSpace(respBody) != "internal server error" {
			t.Fatalf("expected 'internal server error', got: %q", respBody)
		}
	})

	t.Run("Password too short returns 400", func(t *testing.T) {
		svc := &mockService{
			acceptInviteErr: auth.ErrPasswordTooShort,
		}
		router := setupOrgRouter(svc)

		body := `{"token":"inv-token-1","name":"New User","password":"short"}`
		req := httptest.NewRequest(http.MethodPost, "/org/invites/accept", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("Password too long returns 400", func(t *testing.T) {
		svc := &mockService{
			acceptInviteErr: auth.ErrPasswordTooLong,
		}
		router := setupOrgRouter(svc)

		body := `{"token":"inv-token-1","name":"New User","password":"` + strings.Repeat("x", 80) + `"}`
		req := httptest.NewRequest(http.MethodPost, "/org/invites/accept", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("Invite not found returns 404", func(t *testing.T) {
		svc := &mockService{
			acceptInviteErr: domain.ErrInviteNotFound,
		}
		router := setupOrgRouter(svc)

		body := `{"token":"missing-token"}`
		req := httptest.NewRequest(http.MethodPost, "/org/invites/accept", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})
}
