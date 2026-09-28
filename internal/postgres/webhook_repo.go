package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fluxa/fluxa/internal/domain"
	"github.com/fluxa/fluxa/internal/tenant"
	"github.com/jackc/pgx/v5"
)

type WebhookRepository struct {
	db DB
}

// NewWebhookRepo is retained for worker code compiled against the older
// constructor name.
func NewWebhookRepo(db DB) *WebhookRepository {
	return NewWebhookRepository(db)
}

func NewWebhookRepository(db DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

func webhookMode(ctx context.Context) domain.Mode {
	return tenant.ModeOrDefault(ctx, domain.ModeLive)
}

func (r *WebhookRepository) CreateEndpoint(ctx context.Context, ep *domain.WebhookEndpoint) error {
	// An endpoint belongs to exactly one environment; test-mode deliveries must
	// never reach an endpoint registered by the live environment.
	if !ep.Mode.Valid() {
		ep.Mode = webhookMode(ctx)
	}
	query := `
		INSERT INTO webhook_endpoints (id, tenant_id, mode, url, secret, events, active, success_count, failure_count, last_delivered_at, notified_failing, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	db := TxFromContext(ctx, r.db)
	_, err := db.Exec(ctx, query, ep.ID, ep.TenantID, ep.Mode, ep.URL, ep.Secret, ep.Events, ep.Active, ep.SuccessCount, ep.FailureCount, ep.LastDeliveredAt, ep.NotifiedFailing, ep.CreatedAt, ep.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create webhook endpoint: %w", err)
	}
	return nil
}

func (r *WebhookRepository) GetEndpoint(ctx context.Context, id string) (*domain.WebhookEndpoint, error) {
	ep := &domain.WebhookEndpoint{}
	mode := webhookMode(ctx)
	query := `SELECT id, tenant_id, mode, url, secret, events, active, success_count, failure_count, last_delivered_at, notified_failing, created_at, updated_at
	          FROM webhook_endpoints WHERE id = $1 AND mode = $2`
	args := []interface{}{id, mode}
	if tenantID := tenant.IDFromContext(ctx); tenantID != "" {
		query += ` AND tenant_id = $3`
		args = append(args, tenantID)
	}
	err := r.db.QueryRow(ctx, query, args...).Scan(&ep.ID, &ep.TenantID, &ep.Mode, &ep.URL, &ep.Secret, &ep.Events, &ep.Active, &ep.SuccessCount, &ep.FailureCount, &ep.LastDeliveredAt, &ep.NotifiedFailing, &ep.CreatedAt, &ep.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrWebhookNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get webhook endpoint: %w", err)
	}
	return ep, nil
}

func (r *WebhookRepository) ListEndpoints(ctx context.Context, tenantID *string) ([]*domain.WebhookEndpoint, error) {
	mode := webhookMode(ctx)
	query := `SELECT id, tenant_id, mode, url, secret, events, active, success_count, failure_count, last_delivered_at, notified_failing, created_at, updated_at
	          FROM webhook_endpoints WHERE mode = $1`
	args := []interface{}{mode}
	if tenantID != nil && *tenantID != "" {
		query += ` AND tenant_id = $2`
		args = append(args, *tenantID)
	}
	query += ` ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list webhook endpoints: %w", err)
	}
	defer rows.Close()
	endpoints := make([]*domain.WebhookEndpoint, 0)
	for rows.Next() {
		ep := &domain.WebhookEndpoint{}
		if err := rows.Scan(&ep.ID, &ep.TenantID, &ep.Mode, &ep.URL, &ep.Secret, &ep.Events, &ep.Active, &ep.SuccessCount, &ep.FailureCount, &ep.LastDeliveredAt, &ep.NotifiedFailing, &ep.CreatedAt, &ep.UpdatedAt); err != nil {
			return nil, err
		}
		endpoints = append(endpoints, ep)
	}
	return endpoints, rows.Err()
}

func (r *WebhookRepository) UpdateEndpoint(ctx context.Context, ep *domain.WebhookEndpoint) error {
	_, err := r.db.Exec(ctx,
		`UPDATE webhook_endpoints SET url=$2, secret=$3, events=$4, active=$5, success_count=$6, failure_count=$7,
		 last_delivered_at=$8, notified_failing=$9, updated_at=$10 WHERE id=$1 AND mode=$11`,
		ep.ID, ep.URL, ep.Secret, ep.Events, ep.Active, ep.SuccessCount, ep.FailureCount,
		ep.LastDeliveredAt, ep.NotifiedFailing, ep.UpdatedAt, webhookMode(ctx))
	if err != nil {
		return fmt.Errorf("update webhook endpoint: %w", err)
	}
	return nil
}

func (r *WebhookRepository) DeleteEndpoint(ctx context.Context, id string) error {
	query := `DELETE FROM webhook_endpoints WHERE id=$1 AND mode=$2`
	args := []interface{}{id, webhookMode(ctx)}
	if tenantID := tenant.IDFromContext(ctx); tenantID != "" {
		query += ` AND tenant_id=$3`
		args = append(args, tenantID)
	}
	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete webhook endpoint: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrWebhookNotFound
	}
	return nil
}

func (r *WebhookRepository) CreateSubscription(ctx context.Context, sub *domain.WebhookSubscription) error {
	if tID := tenant.IDFromContext(ctx); tID != "" {
		sub.TenantID = &tID
	}
	sub.Mode = webhookMode(ctx)
	_, err := r.db.Exec(ctx,
		`INSERT INTO webhook_subscriptions (id, tenant_id, mode, event_type, webhook_url, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		sub.ID, nullableUUID(sub.TenantID), sub.Mode, sub.EventType, sub.WebhookURL, sub.CreatedAt)
	if err != nil {
		return fmt.Errorf("create webhook subscription: %w", err)
	}
	return nil
}

func (r *WebhookRepository) DeleteSubscription(ctx context.Context, id string) error {
	query := `DELETE FROM webhook_subscriptions WHERE id=$1 AND mode=$2`
	args := []interface{}{id, webhookMode(ctx)}
	if tenantID := tenant.IDFromContext(ctx); tenantID != "" {
		query += ` AND tenant_id=$3`
		args = append(args, tenantID)
	}
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *WebhookRepository) ListSubscriptions(ctx context.Context, tenantID *string) ([]*domain.WebhookSubscription, error) {
	query := `SELECT id, tenant_id, mode, event_type, webhook_url, created_at FROM webhook_subscriptions WHERE mode=$1`
	args := []interface{}{webhookMode(ctx)}
	if tenantID != nil && *tenantID != "" {
		query += ` AND tenant_id=$2`
		args = append(args, *tenantID)
	}
	query += ` ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	subs := make([]*domain.WebhookSubscription, 0)
	for rows.Next() {
		sub := &domain.WebhookSubscription{}
		if err := rows.Scan(&sub.ID, &sub.TenantID, &sub.Mode, &sub.EventType, &sub.WebhookURL, &sub.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}

func (r *WebhookRepository) GetSubscriptionsForEvent(ctx context.Context, tenantID *string, eventType string) ([]*domain.WebhookSubscription, error) {
	query := `SELECT id, tenant_id, mode, event_type, webhook_url, created_at FROM webhook_subscriptions WHERE mode=$1 AND event_type=$2`
	args := []interface{}{webhookMode(ctx), eventType}
	if tenantID != nil && *tenantID != "" {
		query += ` AND tenant_id=$3`
		args = append(args, *tenantID)
	}
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	subs := make([]*domain.WebhookSubscription, 0)
	for rows.Next() {
		sub := &domain.WebhookSubscription{}
		if err := rows.Scan(&sub.ID, &sub.TenantID, &sub.Mode, &sub.EventType, &sub.WebhookURL, &sub.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}

func (r *WebhookRepository) CreateDelivery(ctx context.Context, d *domain.WebhookDelivery) error {
	d.Mode = webhookMode(ctx)
	_, err := r.db.Exec(ctx,
		`INSERT INTO webhook_deliveries (id, endpoint_id, tenant_id, mode, event_type, method, payload, status, response_code, response_body, error_message, attempt_count, max_attempts, next_attempt_at, last_attempt, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		d.ID, d.EndpointID, nullableUUID(d.TenantID), d.Mode, d.EventType, d.Method, d.Payload, d.Status,
		d.ResponseCode, d.ResponseBody, d.ErrorMessage, d.AttemptCount, d.MaxAttempts, d.NextAttemptAt, d.LastAttempt, d.CreatedAt, d.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create webhook delivery: %w", err)
	}
	return nil
}

func (r *WebhookRepository) GetDelivery(ctx context.Context, id string) (*domain.WebhookDelivery, error) {
	d := &domain.WebhookDelivery{}
	err := r.db.QueryRow(ctx,
		`SELECT id, endpoint_id, tenant_id, mode, event_type, method, payload, status, response_code, response_body, error_message, attempt_count, max_attempts, next_attempt_at, last_attempt, created_at, updated_at
		 FROM webhook_deliveries WHERE id=$1 AND mode=$2`, id, webhookMode(ctx)).Scan(
		&d.ID, &d.EndpointID, &d.TenantID, &d.Mode, &d.EventType, &d.Method, &d.Payload, &d.Status, &d.ResponseCode, &d.ResponseBody, &d.ErrorMessage, &d.AttemptCount, &d.MaxAttempts, &d.NextAttemptAt, &d.LastAttempt, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrWebhookDeliveryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get webhook delivery: %w", err)
	}
	return d, nil
}

func (r *WebhookRepository) UpdateDelivery(ctx context.Context, d *domain.WebhookDelivery) error {
	_, err := r.db.Exec(ctx,
		`UPDATE webhook_deliveries SET status=$2, response_code=$3, response_body=$4, error_message=$5, attempt_count=$6, max_attempts=$7, next_attempt_at=$8, last_attempt=$9, updated_at=$10 WHERE id=$1 AND mode=$11`,
		d.ID, d.Status, d.ResponseCode, d.ResponseBody, d.ErrorMessage, d.AttemptCount, d.MaxAttempts, d.NextAttemptAt, d.LastAttempt, d.UpdatedAt, webhookMode(ctx))
	return err
}

func (r *WebhookRepository) ListDeliveries(ctx context.Context, endpointID string, limit, offset int) ([]*domain.WebhookDelivery, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, endpoint_id, tenant_id, mode, event_type, method, payload, status, response_code, response_body, error_message, attempt_count, max_attempts, next_attempt_at, last_attempt, created_at, updated_at
		 FROM webhook_deliveries WHERE endpoint_id=$1 AND mode=$2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`, endpointID, webhookMode(ctx), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	deliveries := make([]*domain.WebhookDelivery, 0)
	for rows.Next() {
		d := &domain.WebhookDelivery{}
		if err := rows.Scan(&d.ID, &d.EndpointID, &d.TenantID, &d.Mode, &d.EventType, &d.Method, &d.Payload, &d.Status, &d.ResponseCode, &d.ResponseBody, &d.ErrorMessage, &d.AttemptCount, &d.MaxAttempts, &d.NextAttemptAt, &d.LastAttempt, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}

func (r *WebhookRepository) CreateDeadLetter(ctx context.Context, dl *domain.WebhookDeadLetter) error {
	dl.Mode = webhookMode(ctx)
	_, err := r.db.Exec(ctx,
		`INSERT INTO webhook_dead_letters (id, endpoint_id, tenant_id, mode, delivery_id, payload, error_message, attempt_count, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		dl.ID, dl.EndpointID, nullableUUID(dl.TenantID), dl.Mode, dl.DeliveryID, dl.Payload, dl.ErrorMessage, dl.AttemptCount, dl.CreatedAt)
	return err
}

func (r *WebhookRepository) GetDeadLetter(ctx context.Context, id string) (*domain.WebhookDeadLetter, error) {
	dl := &domain.WebhookDeadLetter{}
	err := r.db.QueryRow(ctx,
		`SELECT id, endpoint_id, tenant_id, mode, delivery_id, payload, error_message, attempt_count, created_at FROM webhook_dead_letters WHERE id=$1 AND mode=$2`, id, webhookMode(ctx)).Scan(
		&dl.ID, &dl.EndpointID, &dl.TenantID, &dl.Mode, &dl.DeliveryID, &dl.Payload, &dl.ErrorMessage, &dl.AttemptCount, &dl.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("dead letter not found")
	}
	return dl, err
}

func (r *WebhookRepository) ListDeadLetters(ctx context.Context, tenantID *string, limit, offset int) ([]*domain.WebhookDeadLetter, error) {
	query := `SELECT id, endpoint_id, tenant_id, mode, delivery_id, payload, error_message, attempt_count, created_at FROM webhook_dead_letters WHERE mode=$1`
	args := []interface{}{webhookMode(ctx)}
	if tenantID != nil && *tenantID != "" {
		query += ` AND tenant_id=$2`
		args = append(args, *tenantID)
	}
	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)
	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	deadLetters := make([]*domain.WebhookDeadLetter, 0)
	for rows.Next() {
		dl := &domain.WebhookDeadLetter{}
		if err := rows.Scan(&dl.ID, &dl.EndpointID, &dl.TenantID, &dl.Mode, &dl.DeliveryID, &dl.Payload, &dl.ErrorMessage, &dl.AttemptCount, &dl.CreatedAt); err != nil {
			return nil, err
		}
		deadLetters = append(deadLetters, dl)
	}
	return deadLetters, rows.Err()
}

func (r *WebhookRepository) GetConfig(ctx context.Context, tenantID string) (*domain.TenantWebhookConfig, error) {
	config := &domain.TenantWebhookConfig{}
	err := r.db.QueryRow(ctx,
		`SELECT tenant_id, enabled, url, secret, signing_algorithm, events, paused, resume_at, last_delivered_at, created_at, updated_at
		 FROM tenant_webhook_configs WHERE tenant_id = $1`,
		tenantID,
	).Scan(
		&config.TenantID, &config.Enabled, &config.URL, &config.Secret, &config.SigningAlgorithm, &config.Events,
		&config.Paused, &config.ResumeAt, &config.LastDeliveredAt, &config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWebhookConfigNotFound
		}
		return nil, fmt.Errorf("get tenant webhook config: %w", err)
	}
	return config, nil
}

func (r *WebhookRepository) UpsertConfig(ctx context.Context, config *domain.TenantWebhookConfig) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO tenant_webhook_configs
		 (tenant_id, enabled, url, secret, signing_algorithm, events, paused, resume_at, last_delivered_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (tenant_id) DO UPDATE SET
		 enabled = EXCLUDED.enabled,
		 url = EXCLUDED.url,
		 secret = EXCLUDED.secret,
		 signing_algorithm = EXCLUDED.signing_algorithm,
		 events = EXCLUDED.events,
		 paused = EXCLUDED.paused,
		 resume_at = EXCLUDED.resume_at,
		 last_delivered_at = EXCLUDED.last_delivered_at,
		 updated_at = EXCLUDED.updated_at`,
		config.TenantID, config.Enabled, config.URL, config.Secret, config.SigningAlgorithm, config.Events,
		config.Paused, config.ResumeAt, config.LastDeliveredAt, config.CreatedAt, config.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert tenant webhook config: %w", err)
	}
	return nil
}

func (r *WebhookRepository) ListEnabledConfigs(ctx context.Context) ([]*domain.TenantWebhookConfig, error) {
	rows, err := r.db.Query(ctx,
		`SELECT tenant_id, enabled, url, secret, signing_algorithm, events, paused, resume_at, last_delivered_at, created_at, updated_at
		 FROM tenant_webhook_configs WHERE enabled = TRUE ORDER BY created_at`,
	)
	if err != nil {
		return nil, fmt.Errorf("list tenant webhook configs: %w", err)
	}
	defer rows.Close()

	var configs []*domain.TenantWebhookConfig
	for rows.Next() {
		config := &domain.TenantWebhookConfig{}
		if err := rows.Scan(
			&config.TenantID, &config.Enabled, &config.URL, &config.Secret, &config.SigningAlgorithm, &config.Events,
			&config.Paused, &config.ResumeAt, &config.LastDeliveredAt, &config.CreatedAt, &config.UpdatedAt,
		); err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, rows.Err()
}

func (r *WebhookRepository) CreateConfigDelivery(ctx context.Context, delivery *domain.TenantWebhookDelivery) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO tenant_webhook_deliveries
		 (id, tenant_id, event_type, payload, status, response_code, attempt_count, last_attempt, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		delivery.ID, delivery.TenantID, string(delivery.EventType), delivery.Payload, string(delivery.Status),
		delivery.ResponseCode, delivery.AttemptCount, delivery.LastAttempt, delivery.CreatedAt, delivery.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert tenant webhook delivery: %w", err)
	}
	return nil
}

func (r *WebhookRepository) UpdateConfigDelivery(ctx context.Context, delivery *domain.TenantWebhookDelivery) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tenant_webhook_deliveries
		 SET status = $1, response_code = $2, attempt_count = $3, last_attempt = $4, updated_at = $5
		 WHERE id = $6 AND tenant_id = $7`,
		string(delivery.Status), delivery.ResponseCode, delivery.AttemptCount, delivery.LastAttempt,
		delivery.UpdatedAt, delivery.ID, delivery.TenantID,
	)
	if err != nil {
		return fmt.Errorf("update tenant webhook delivery: %w", err)
	}
	return nil
}

func (r *WebhookRepository) GetConfigDelivery(ctx context.Context, id, tenantID string) (*domain.TenantWebhookDelivery, error) {
	delivery := &domain.TenantWebhookDelivery{}
	var eventType, status string
	err := r.db.QueryRow(ctx,
		`SELECT id, tenant_id, event_type, payload, status, response_code, attempt_count, last_attempt, created_at, updated_at
		 FROM tenant_webhook_deliveries WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	).Scan(
		&delivery.ID, &delivery.TenantID, &eventType, &delivery.Payload, &status, &delivery.ResponseCode,
		&delivery.AttemptCount, &delivery.LastAttempt, &delivery.CreatedAt, &delivery.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWebhookDeliveryNotFound
		}
		return nil, fmt.Errorf("get tenant webhook delivery: %w", err)
	}
	delivery.EventType = domain.EventType(eventType)
	delivery.Status = domain.DeliveryStatus(status)
	return delivery, nil
}

func (r *WebhookRepository) ListConfigDeliveries(ctx context.Context, tenantID string, limit, offset int) ([]*domain.TenantWebhookDelivery, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, tenant_id, event_type, payload, status, response_code, attempt_count, last_attempt, created_at, updated_at
		 FROM tenant_webhook_deliveries WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		tenantID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list tenant webhook deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*domain.TenantWebhookDelivery
	for rows.Next() {
		delivery := &domain.TenantWebhookDelivery{}
		var eventType, status string
		if err := rows.Scan(
			&delivery.ID, &delivery.TenantID, &eventType, &delivery.Payload, &status, &delivery.ResponseCode,
			&delivery.AttemptCount, &delivery.LastAttempt, &delivery.CreatedAt, &delivery.UpdatedAt,
		); err != nil {
			return nil, err
		}
		delivery.EventType = domain.EventType(eventType)
		delivery.Status = domain.DeliveryStatus(status)
		deliveries = append(deliveries, delivery)
	}
	return deliveries, rows.Err()
}

func (r *WebhookRepository) UpdateConfigLastDelivered(ctx context.Context, tenantID string, deliveredAt time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tenant_webhook_configs SET last_delivered_at = $1, updated_at = $1 WHERE tenant_id = $2`,
		deliveredAt, tenantID,
	)
	if err != nil {
		return fmt.Errorf("update tenant webhook last delivered: %w", err)
	}
	return nil
}
