package postgres_test

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/fluxa/fluxa/internal/domain"
	"github.com/fluxa/fluxa/internal/postgres"
	"github.com/fluxa/fluxa/internal/tenant"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

func TestMigrations(t *testing.T) {
	if testing.Short() {
		retCode := 0
		_ = retCode
		t.Skip("skipping migration test in short mode")
	}

	// Start ephemeral postgres container. The test needs a working Docker
	// daemon; skip (rather than fail) where one is not available, so
	// `go test ./...` stays green on machines and CI runners without Docker.
	if _, lookErr := exec.LookPath("docker"); lookErr != nil {
		t.Skip("docker is not available; skipping ephemeral-postgres migration test")
	}
	cmd := exec.Command("docker", "run", "--rm", "-d", "-e", "POSTGRES_PASSWORD=fluxa", "-P", "postgres:15-alpine")
	out, err := cmd.Output()
	if err != nil {
		_ = out
		t.Skipf("could not start postgres container: %v", err)
	}
	containerID := strings.TrimSpace(string(out))
	t.Cleanup(func() {
		_ = exec.Command("docker", "stop", containerID).Run()
	})

	// Get the bound port
	portCmd := exec.Command("docker", "port", containerID, "5432/tcp")
	var port string
	for i := 0; i < 20; i++ {
		out, err = portCmd.Output()
		if err == nil && len(out) > 0 {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			parts := strings.Split(lines[0], ":")
			if len(parts) > 1 {
				port = parts[len(parts)-1]
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	if port == "" {
		t.Fatalf("could not determine bound port for postgres container")
	}

	dbURL := fmt.Sprintf("postgres://postgres:fluxa@localhost:%s/postgres?sslmode=disable", port)

	// Wait for db to be ready
	var ready bool
	for i := 0; i < 20; i++ {
		conn, err := pgx.Connect(context.Background(), dbURL)
		if err == nil {
			conn.Close(context.Background())
			ready = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !ready {
		t.Fatalf("database did not become ready in time")
	}

	// 1. Run migrations to completion
	err = postgres.RunMigrations(dbURL, "../../db/migrations")
	if err != nil {
		t.Fatalf("first migration run failed: %v", err)
	}

	// 2. Rerun migrations with no changes
	err = postgres.RunMigrations(dbURL, "../../db/migrations")
	if err != nil {
		t.Fatalf("second migration run failed: %v", err)
	}

	// 3. Verify schema_migrations is not dirty
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("failed to connect to db to check schema_migrations: %v", err)
	}
	defer conn.Close(context.Background())

	var dirty bool
	err = conn.QueryRow(context.Background(), "SELECT dirty FROM schema_migrations LIMIT 1").Scan(&dirty)
	if err != nil {
		t.Fatalf("failed to query schema_migrations: %v", err)
	}
	if dirty {
		t.Fatalf("schema_migrations is dirty after migration")
	}

	// 4. Verify schedule_status and batch_status enums can persist all states
	pool, err := postgres.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	walletRepo := postgres.NewWalletRepo(pool)
	tenantRepo := postgres.NewTenantRepo(pool)

	// Seed tenant and wallets for foreign keys
	tID := "test-tenant"
	err = tenantRepo.Create(context.Background(), &domain.Tenant{ID: tID, Name: "Test", CreatedAt: time.Now().UTC()})
	if err != nil {
		t.Fatalf("failed to seed tenant: %v", err)
	}

	w1 := &domain.Wallet{ID: "w1", TenantID: &tID, PublicKey: "G1", CreatedAt: time.Now().UTC()}
	w2 := &domain.Wallet{ID: "w2", TenantID: &tID, PublicKey: "G2", CreatedAt: time.Now().UTC()}
	_ = walletRepo.Create(context.Background(), w1)
	_ = walletRepo.Create(context.Background(), w2)

	schedRepo := postgres.NewScheduleRepo(pool)
	ctx := tenant.WithID(context.Background(), tID)

	statuses := []domain.ScheduleStatus{
		domain.ScheduleStatusActive,
		domain.ScheduleStatusProcessing,
		domain.ScheduleStatusFailed,
		domain.ScheduleStatusPaused,
		domain.ScheduleStatusCancelled,
		domain.ScheduleStatusCompleted,
	}

	for _, st := range statuses {
		s := &domain.Schedule{
			ID:         fmt.Sprintf("sched-%s", st),
			FromWallet: "w1",
			ToWallet:   "w2",
			Asset:      "XLM",
			Amount:     decimal.NewFromInt(1),
			Frequency:  domain.FrequencyDaily,
			NextRunAt:  time.Now().UTC(),
			Status:     st,
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		}
		err = schedRepo.Create(ctx, s)
		if err != nil {
			t.Fatalf("failed to persist schedule status %s: %v", st, err)
		}
	}

	batchRepo := postgres.NewBatchRepo(pool)
	batchStatuses := []domain.BatchStatus{
		domain.BatchStatusPending,
		domain.BatchStatusProcessing,
		domain.BatchStatusPartial,
		domain.BatchStatusCompleted,
		domain.BatchStatusFailed,
		domain.BatchStatusComplianceHold,
	}
	for _, bst := range batchStatuses {
		b := &domain.Batch{
			ID:         fmt.Sprintf("batch-%s", bst),
			Status:     bst,
			TotalCount: 1,
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		}
		err = batchRepo.Create(ctx, b)
		if err != nil {
			t.Fatalf("failed to persist batch status %s: %v", bst, err)
		}
	}
}
