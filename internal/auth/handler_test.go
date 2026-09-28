package auth_test

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
	"github.com/go-chi/chi/v5"
)

type mockAuthService struct {
	registerResp *auth.AuthResponse
	registerErr  error
	loginResp    *auth.AuthResponse
	loginErr     error
	refreshResp  *auth.AuthResponse
	refreshErr   error
}

func (m *mockAuthService) Register(ctx context.Context, req auth.RegisterRequest) (*auth.AuthResponse, error) {
	return m.registerResp, m.registerErr
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*auth.AuthResponse, error) {
	return m.loginResp, m.loginErr
}

func (m *mockAuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*auth.AuthResponse, error) {
	return m.refreshResp, m.refreshErr
}

func setupRouter(svc auth.Service) http.Handler {
	r := chi.NewRouter()
	h := auth.NewHandler(svc)
	r.Route("/auth", h.Routes())
	return r
}

func TestHandler_InternalErrorsNeverLeaked(t *testing.T) {
	t.Run("Login internal database error does not leak to client", func(t *testing.T) {
		svc := &mockAuthService{
			loginErr: errors.New("pq: password authentication failed for user 'postgres' host: 10.0.0.1 database 'fluxa_prod'"),
		}
		router := setupRouter(svc)

		reqBody := `{"email":"ops@example.com","password":"mypassword123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}

		body := rec.Body.String()
		// Must not contain database details
		if strings.Contains(body, "postgres") || strings.Contains(body, "fluxa_prod") || strings.Contains(body, "10.0.0.1") {
			t.Fatalf("internal error text was leaked in response body: %q", body)
		}

		if strings.TrimSpace(body) != "internal server error" {
			t.Fatalf("expected 'internal server error', got: %q", body)
		}
	})

	t.Run("Register internal database error does not leak to client", func(t *testing.T) {
		svc := &mockAuthService{
			registerErr: errors.New("pq: relation 'users' does not exist at character 13 in transaction 92834"),
		}
		router := setupRouter(svc)

		reqBody := `{"name":"Alice","email":"alice@example.com","password":"validpassword123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}

		body := rec.Body.String()
		if strings.Contains(body, "relation") || strings.Contains(body, "transaction") || strings.Contains(body, "92834") {
			t.Fatalf("internal error text was leaked in response body: %q", body)
		}

		if strings.TrimSpace(body) != "internal server error" {
			t.Fatalf("expected 'internal server error', got: %q", body)
		}
	})
}

func TestHandler_ExpectedStatusCodes(t *testing.T) {
	t.Run("Login invalid credentials returns 401", func(t *testing.T) {
		svc := &mockAuthService{
			loginErr: domain.ErrInvalidCredentials,
		}
		router := setupRouter(svc)

		reqBody := `{"email":"ops@example.com","password":"wrongpassword"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(reqBody))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("Register user already exists returns 409", func(t *testing.T) {
		svc := &mockAuthService{
			registerErr: domain.ErrUserAlreadyExists,
		}
		router := setupRouter(svc)

		reqBody := `{"name":"Bob","email":"bob@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(reqBody))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d", rec.Code)
		}
	})

	t.Run("Register password too short returns 400", func(t *testing.T) {
		svc := &mockAuthService{
			registerErr: auth.ErrPasswordTooShort,
		}
		router := setupRouter(svc)

		reqBody := `{"name":"Bob","email":"bob@example.com","password":"short"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(reqBody))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("Register password too long returns 400", func(t *testing.T) {
		svc := &mockAuthService{
			registerErr: auth.ErrPasswordTooLong,
		}
		router := setupRouter(svc)

		reqBody := `{"name":"Bob","email":"bob@example.com","password":"` + strings.Repeat("x", 80) + `"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(reqBody))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
	})
}
