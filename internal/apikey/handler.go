package apikey

import (
	"encoding/json"
	"net/http"
	"time"

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
	if tenantID == "" {
		http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"tenant not found"}}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Label *string `json:"label"`
		Role  string  `json:"role"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.Role == "" {
		req.Role = domain.RoleDeveloper
	}
	if req.Role != domain.RoleOwner && req.Role != domain.RoleAdmin && req.Role != domain.RoleDeveloper && req.Role != domain.RoleViewer {
		http.Error(w, `{"error":{"code":"INVALID_REQUEST","message":"invalid role"}}`, http.StatusBadRequest)
		return
	}

	raw, prefix, err := Generate()
	if err != nil {
		log.Error().Err(err).Msg("generate api key")
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"failed to generate key"}}`, http.StatusInternalServerError)
		return
	}

	key := &domain.APIKey{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		KeyHash:   Hash(raw),
		Prefix:    prefix,
		Label:     req.Label,
		Role:      req.Role,
		CreatedAt: time.Now(),
	}

	if err := h.repo.Create(r.Context(), key); err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("create api key")
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"failed to create key"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         key.ID,
		"key":        raw, // raw key exactly once
		"prefix":     key.Prefix,
		"label":      key.Label,
		"role":       key.Role,
		"created_at": key.CreatedAt,
	})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := tenant.IDFromContext(r.Context())
	if tenantID == "" {
		http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"tenant not found"}}`, http.StatusUnauthorized)
		return
	}

	keys, err := h.repo.ListByTenant(r.Context(), tenantID)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("list api keys")
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"failed to list keys"}}`, http.StatusInternalServerError)
		return
	}

	res := make([]map[string]interface{}, 0, len(keys))
	for _, k := range keys {
		res = append(res, map[string]interface{}{
			"id":           k.ID,
			"prefix":       k.Prefix,
			"label":        k.Label,
			"role":         k.Role,
			"last_used_at": k.LastUsedAt,
			"revoked_at":   k.RevokedAt,
			"created_at":   k.CreatedAt,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	tenantID := tenant.IDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.repo.Revoke(r.Context(), id, tenantID); err != nil {
		log.Error().Err(err).Str("key_id", id).Msg("revoke api key")
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"failed to revoke key"}}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
