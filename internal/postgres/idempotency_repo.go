package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fluxa/fluxa/internal/server/idempotency"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type IdempotencyRepo struct {
	db DB
}

func NewIdempotencyRepo(db DB) *IdempotencyRepo {
	return &IdempotencyRepo{db: db}
}

func (r *IdempotencyRepo) TryAcquire(ctx context.Context, orgID, key, requestHash string, expiresAt time.Time) (*idempotency.Record, bool, error) {
	acquired, err := r.insertIfAbsent(ctx, orgID, key, requestHash, expiresAt)
	if err != nil {
		return nil, false, err
	}
	if acquired {
		return nil, false, nil
	}

	rec, found, err := r.lookupLive(ctx, orgID, key)
	if err != nil {
		return nil, false, err
	}
	if found {
		return rec, true, nil
	}

	// No live record was found, so whatever is blocking the unique index is
	// expired — clear it and retry the insert once.
	if _, err := r.db.Exec(ctx,
		`DELETE FROM idempotency_records WHERE org_id = $1 AND key = $2 AND expires_at <= NOW()`,
		orgID, key,
	); err != nil {
		return nil, false, fmt.Errorf("clear expired idempotency record: %w", err)
	}

	acquired, err = r.insertIfAbsent(ctx, orgID, key, requestHash, expiresAt)
	if err != nil {
		return nil, false, err
	}
	if acquired {
		return nil, false, nil
	}

	// Lost a race against a concurrent retry of the same expired key —
	// whoever won is now the live record.
	rec, _, err = r.lookupLive(ctx, orgID, key)
	if err != nil {
		return nil, false, err
	}
	return rec, true, nil
}

func (r *IdempotencyRepo) insertIfAbsent(ctx context.Context, orgID, key, requestHash string, expiresAt time.Time) (bool, error) {
	var insertedID string
	err := r.db.QueryRow(ctx,
		`INSERT INTO idempotency_records (id, org_id, key, request_hash, status, created_at, expires_at)
		 VALUES ($1, $2, $3, $4, 'processing', NOW(), $5)
		 ON CONFLICT (org_id, key) DO NOTHING
		 RETURNING id`,
		uuid.New().String(), orgID, key, requestHash, expiresAt,
	).Scan(&insertedID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return false, fmt.Errorf("acquire idempotency record: %w", err)
}

// lookupLive reads the current record for (orgID, key), if any that has not
// expired. It runs the read as SELECT ... FOR UPDATE SKIP LOCKED inside its
// own short transaction, matching the locking discipline used elsewhere in
// this codebase for claim-style reads (see TransactionRepo.GetPendingTxesForReconciliation).
func (r *IdempotencyRepo) lookupLive(ctx context.Context, orgID, key string) (*idempotency.Record, bool, error) {
	dbTx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("begin idempotency lookup: %w", err)
	}
	defer dbTx.Rollback(ctx)

	rec := &idempotency.Record{OrgID: orgID, Key: key}
	var responseStatus *int
	var responseBody []byte
	err = dbTx.QueryRow(ctx,
		`SELECT status, request_hash, response_status, response_body
		 FROM idempotency_records
		 WHERE org_id = $1 AND key = $2 AND expires_at > NOW()
		 FOR UPDATE SKIP LOCKED`,
		orgID, key,
	).Scan(&rec.Status, &rec.RequestHash, &responseStatus, &responseBody)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("lookup idempotency record: %w", err)
	}

	if err := dbTx.Commit(ctx); err != nil {
		return nil, false, fmt.Errorf("commit idempotency lookup: %w", err)
	}

	if responseStatus != nil {
		rec.ResponseStatus = *responseStatus
	}
	rec.ResponseBody = responseBody
	return rec, true, nil
}

func (r *IdempotencyRepo) Complete(ctx context.Context, orgID, key string, responseStatus int, responseBody []byte) error {
	_, err := r.db.Exec(ctx,
		`UPDATE idempotency_records
		 SET status = 'complete', response_status = $3, response_body = $4
		 WHERE org_id = $1 AND key = $2`,
		orgID, key, responseStatus, responseBody,
	)
	if err != nil {
		return fmt.Errorf("complete idempotency record: %w", err)
	}
	return nil
}

// DeleteExpired removes up to batchSize rows whose expires_at is in the past.
// Callers should loop until the returned count is zero to drain the backlog
// without issuing one enormous DELETE that could spike I/O or hold locks.
//
// The per-key opportunistic delete in TryAcquire is kept: it clears the
// unique-index conflict on the hot path (same key reused after expiry) without
// waiting for the next scheduled sweep, which is important for low-traffic orgs
// whose key may not be visited again within the TTL window.
func (r *IdempotencyRepo) DeleteExpired(ctx context.Context, batchSize int) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM idempotency_records
		 WHERE id IN (
		     SELECT id FROM idempotency_records
		     WHERE expires_at <= NOW()
		     LIMIT $1
		 )`,
		batchSize,
	)
	if err != nil {
		return 0, fmt.Errorf("purge expired idempotency records: %w", err)
	}
	return tag.RowsAffected(), nil
}
