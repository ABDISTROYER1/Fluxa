package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/fluxa/fluxa/internal/domain"
	"github.com/jackc/pgx/v5"
)

type APIKeyRepo struct {
	db DB
}

func NewAPIKeyRepo(db DB) *APIKeyRepo {
	return &APIKeyRepo{db: db}
}

func (r *APIKeyRepo) Create(ctx context.Context, key *domain.APIKey) error {
	if !key.Mode.Valid() {
		return errors.New("api key mode must be live or test")
	}
	db := TxFromContext(ctx, r.db)
	_, err := db.Exec(ctx,
		`INSERT INTO api_keys (id, tenant_id, key_hash, prefix, mode, label, role, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		key.ID, key.TenantID, key.KeyHash, key.Prefix, key.Mode, key.Label, key.Role, key.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert api_key: %w", err)
	}
	return nil
}

func (r *APIKeyRepo) GetByHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	k := &domain.APIKey{}
	err := r.db.QueryRow(ctx,
		`SELECT id, tenant_id, key_hash, prefix, mode, label, role, last_used_at, revoked_at, created_at FROM api_keys WHERE key_hash = $1`,
		hash,
	).Scan(&k.ID, &k.TenantID, &k.KeyHash, &k.Prefix, &k.Mode, &k.Label, &k.Role, &k.LastUsedAt, &k.RevokedAt, &k.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("api key not found")
		}
		return nil, fmt.Errorf("get api_key by hash: %w", err)
	}
	return k, nil
}

// ListByTenant returns the keys for one environment. Keys are environment
// scoped: a live key must never be listed as if it could authenticate against
// testnet.
func (r *APIKeyRepo) ListByTenant(ctx context.Context, tenantID string, mode domain.Mode) ([]*domain.APIKey, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, tenant_id, key_hash, prefix, mode, label, role, last_used_at, revoked_at, created_at
		 FROM api_keys WHERE tenant_id = $1 AND mode = $2 ORDER BY created_at DESC`,
		tenantID, mode,
	)
	if err != nil {
		return nil, fmt.Errorf("list api_keys: %w", err)
	}
	defer rows.Close()

	var keys []*domain.APIKey
	for rows.Next() {
		k := &domain.APIKey{}
		if err := rows.Scan(&k.ID, &k.TenantID, &k.KeyHash, &k.Prefix, &k.Mode, &k.Label, &k.Role, &k.LastUsedAt, &k.RevokedAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func (r *APIKeyRepo) Revoke(ctx context.Context, id string, tenantID string, mode domain.Mode) error {
	res, err := r.db.Exec(ctx,
		`UPDATE api_keys SET revoked_at = NOW() WHERE id = $1 AND tenant_id = $2 AND mode = $3`,
		id, tenantID, mode,
	)
	if err != nil {
		return fmt.Errorf("revoke api_key: %w", err)
	}
	if res.RowsAffected() == 0 {
		return errors.New("api key not found or not owned by tenant")
	}
	return nil
}

func (r *APIKeyRepo) UpdateLastUsed(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("update api_key last_used: %w", err)
	}
	return nil
}
