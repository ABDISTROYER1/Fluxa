package apikey

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fluxa/fluxa/internal/api"
	"github.com/fluxa/fluxa/internal/domain"
	"github.com/fluxa/fluxa/internal/postgres"
	"github.com/fluxa/fluxa/internal/tenant"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	repo *postgres.APIKeyRepo
}

func NewHandler(repo *postgres.APIKeyRepo) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := tenant.IDFromContext(r.Context())
	mode, ok := tenant.ModeFromContext(r.Context())
	if tenantID == "" || !ok {
		api.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "tenant environment is required")
		return
	}

	var req struct {
		Label *string `json:"label"`
		Role  string  `json:"role"`
		Mode  string  `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	if req.Role == "" {
		req.Role = domain.RoleDeveloper
	}
	if req.Role != domain.RoleOwner && req.Role != domain.RoleAdmin && req.Role != domain.RoleDeveloper && req.Role != domain.RoleViewer {
		api.BadRequest(w, "invalid role")
		return
	}

	requestedMode := mode
	if req.Mode != "" {
		parsed, err := domain.ParseMode(req.Mode)
		if err != nil {
			api.BadRequest(w, err.Error())
			return
		}
		requestedMode = parsed
	}
	// API-key callers are environment admins, not platform key managers. A user
	// JWT may create either environment; an API key may only manage its own.
	if tenant.UserIDFromContext(r.Context()) == "" && requestedMode != mode {
		api.Error(w, http.StatusForbidden, "CROSS_MODE_KEY_OPERATION", "an API key cannot create keys for another environment")
		return
	}

	raw, prefix, err := Generate(requestedMode)
	if err != nil {
		log.Error().Err(err).Msg("generate api key")
		api.InternalError(w, err)
		return
	}

	key := &domain.APIKey{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		KeyHash:   Hash(raw),
		Prefix:    prefix,
		Mode:      requestedMode,
		Label:     req.Label,
		Role:      req.Role,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.repo.Create(r.Context(), key); err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Str("mode", string(requestedMode)).Msg("create api key")
		api.InternalError(w, err)
		return
	}

	api.JSON(w, http.StatusCreated, map[string]interface{}{
		"id":         key.ID,
		"key":        raw, // raw key is returned exactly once
		"prefix":     key.Prefix,
		"mode":       key.Mode,
		"label":      key.Label,
		"role":       key.Role,
		"created_at": key.CreatedAt,
	})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := tenant.IDFromContext(r.Context())
	mode, ok := tenant.ModeFromContext(r.Context())
	if tenantID == "" || !ok {
		api.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "tenant environment is required")
		return
	}

	keys, err := h.repo.ListByTenant(r.Context(), tenantID, mode)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Str("mode", string(mode)).Msg("list api keys")
		api.InternalError(w, err)
		return
	}

	// The response shape is explicit so the stored key hash can never leak.
	res := make([]map[string]interface{}, 0, len(keys))
	for _, k := range keys {
		res = append(res, map[string]interface{}{
			"id":           k.ID,
			"prefix":       k.Prefix,
			"mode":         k.Mode,
			"label":        k.Label,
			"role":         k.Role,
			"last_used_at": k.LastUsedAt,
			"revoked_at":   k.RevokedAt,
			"created_at":   k.CreatedAt,
		})
	}
	api.JSON(w, http.StatusOK, res)
}

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	tenantID := tenant.IDFromContext(r.Context())
	mode, ok := tenant.ModeFromContext(r.Context())
	if tenantID == "" || !ok {
		api.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "tenant environment is required")
		return
	}
	id := chi.URLParam(r, "id")

	if err := h.repo.Revoke(r.Context(), id, tenantID, mode); err != nil {
		log.Error().Err(err).Str("key_id", id).Msg("revoke api key")
		api.Error(w, http.StatusNotFound, "API_KEY_NOT_FOUND", "API key not found in this environment")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
