package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/fluxa/fluxa/internal/api"
	"github.com/fluxa/fluxa/internal/apikey"
	"github.com/fluxa/fluxa/internal/auth"
	"github.com/fluxa/fluxa/internal/domain"
	"github.com/fluxa/fluxa/internal/postgres"
	"github.com/fluxa/fluxa/internal/tenant"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(
			log.Logger.With().
				Str("request_id", id).
				Str("operation", "http_request").
				Logger().
				WithContext(r.Context()),
		))
	})
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		zerolog.Ctx(r.Context()).Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.Status()).
			Dur("duration", time.Since(start)).
			Msg("request")
	})
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rv := recover(); rv != nil {
				zerolog.Ctx(r.Context()).Error().Interface("panic", rv).Msg("panic recovered")
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[strings.TrimSpace(o)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if _, ok := originSet[origin]; ok || (len(originSet) == 1 && originSet["*"] == struct{}{}) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				// Handle prefix matching for things like localhost:*
				for allowed := range originSet {
					if strings.HasSuffix(allowed, "*") {
						prefix := strings.TrimSuffix(allowed, "*")
						if strings.HasPrefix(origin, prefix) {
							w.Header().Set("Access-Control-Allow-Origin", origin)
							break
						}
					}
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID, X-Fluxa-Mode")
			w.Header().Set("Access-Control-Max-Age", "86400")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// MembershipValidator revalidates a user's current membership and role
// against the database on every authenticated request.
type MembershipValidator interface {
	GetMember(ctx context.Context, tenantID, userID string) (*domain.OrgMember, error)
}

// AuthMiddleware validates the JWT or API key and, for JWT auth, revalidates
// the user's membership and role against the database so that demotions,
// removals, and role changes take effect immediately rather than at token
// expiry. It also pins the request to exactly one environment: an API key
// carries its environment in the credential itself, while a user JWT selects
// one explicitly.
func AuthMiddleware(repo *postgres.APIKeyRepo, jwtSecret []byte, validator MembershipValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
				http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
				return
			}

			rawToken := authHeader[7:]

			// Check if token is JWT (contains 2 dots)
			if strings.Count(rawToken, ".") == 2 {
				claims, err := auth.ParseToken(rawToken, jwtSecret)
				if err == nil && claims.TokenType == "access" {
					// Revalidate membership against the database so stale tokens
					// cannot be used after removal or demotion.
					if validator != nil {
						member, mErr := validator.GetMember(r.Context(), claims.TenantID, claims.Sub)
						if mErr != nil || member == nil {
							http.Error(w, "membership not found or revoked", http.StatusForbidden)
							return
						}
						// Use the current role from the database, not the stale JWT claim.
						claims.Role = member.Role
					}
					mode := domain.ModeLive
					if requested := r.Header.Get("X-Fluxa-Mode"); requested != "" {
						parsed, parseErr := domain.ParseMode(requested)
						if parseErr != nil {
							api.Error(w, http.StatusBadRequest, "INVALID_ENVIRONMENT_MODE", parseErr.Error())
							return
						}
						mode = parsed
					}
					ctx := tenant.WithID(r.Context(), claims.TenantID)
					ctx = tenant.WithMode(ctx, mode)
					ctx = tenant.WithUser(ctx, claims.Sub, claims.Role)
					requestLogger := zerolog.Ctx(ctx).With().
						Str("tenant_id", claims.TenantID).
						Logger()
					ctx = requestLogger.WithContext(ctx)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// Fallback to API Key auth
			hash := apikey.Hash(rawToken)
			key, err := repo.GetByHash(r.Context(), hash)
			if err != nil || key == nil {
				http.Error(w, "invalid api key or authentication token", http.StatusUnauthorized)
				return
			}
			if key.RevokedAt != nil {
				http.Error(w, "revoked api key", http.StatusUnauthorized)
				return
			}
			// The environment is a property of the credential, not of the
			// request, so a key minted for one environment can never reach the
			// other even if the raw key collides with a persisted record.
			rawMode, modeErr := apikey.ModeFromRaw(rawToken)
			if modeErr != nil || rawMode != key.Mode {
				http.Error(w, "api key mode does not match persisted authorization", http.StatusUnauthorized)
				return
			}

			_ = repo.UpdateLastUsed(r.Context(), key.ID)

			ctx := tenant.WithID(r.Context(), key.TenantID)
			ctx = tenant.WithMode(ctx, key.Mode)
			ctx = tenant.WithUser(ctx, "", key.Role)
			ctx = tenant.WithScopes(ctx, key.Scopes)
			ctx = tenant.WithAPIKeyID(ctx, key.ID)
			requestLogger := zerolog.Ctx(ctx).With().
				Str("tenant_id", key.TenantID).
				Str("mode", string(key.Mode)).
				Logger()
			ctx = requestLogger.WithContext(ctx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireScope(requiredScope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scopes, hasScopes := tenant.ScopesFromContext(r.Context())
			if hasScopes && !domain.HasScope(scopes, requiredScope) {
				api.Error(w, http.StatusForbidden, "INSUFFICIENT_SCOPE", "API key does not have the required scope: "+requiredScope)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := tenant.RoleFromContext(r.Context())
			if role == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			for _, allowed := range allowedRoles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "insufficient permissions", http.StatusForbidden)
		})
	}
}

// RequireTestMode protects sandbox-only routes. API-key requests always carry
// the persisted key mode, so a client-supplied header cannot bypass it.
func RequireTestMode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mode, ok := tenant.ModeFromContext(r.Context())
		if !ok {
			api.Error(w, http.StatusUnauthorized, "ENVIRONMENT_REQUIRED", "an authenticated environment is required")
			return
		}
		if mode != domain.ModeTest {
			api.Error(w, http.StatusForbidden, "TEST_MODE_REQUIRED", "this endpoint requires an sk_test_ API key")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireNotViewer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := tenant.RoleFromContext(r.Context())
		if role == domain.RoleViewer {
			http.Error(w, "viewer role is read-only", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequirePlatformOperator() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := tenant.IDFromContext(r.Context())
			if tenantID != "platform" && tenantID != "system" && tenantID != "operator" {
				http.Error(w, "unauthorized: platform operator access required", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
