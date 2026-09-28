package batch

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/fluxa/fluxa/internal/domain"
	"github.com/fluxa/fluxa/internal/server/idempotency"
	"github.com/fluxa/fluxa/internal/tenant"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// fakeService implements Service, counting how many times CreateBatch
// actually runs so tests can tell a replayed idempotent response apart from
// a reprocessed one.
type fakeService struct {
	mu         sync.Mutex
	createHits int
}

func (f *fakeService) CreateBatch(_ context.Context, fromWalletID string, items []Item) (*Result, error) {
	f.mu.Lock()
	f.createHits++
	f.mu.Unlock()

	now := time.Now().UTC()
	return &Result{
		Batch: &domain.Batch{
			ID:         "batch-1",
			Status:     domain.BatchStatusPending,
			TotalCount: len(items),
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}, nil
}

func (f *fakeService) GetBatch(_ context.Context, id string) (*Result, error) {
	return nil, domain.ErrBatchNotFound
}

func (f *fakeService) ExportCSV(_ context.Context, id string) (string, error) {
	return "", domain.ErrBatchNotFound
}

func (f *fakeService) hits() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.createHits
}

// mockIdemRepo is a minimal in-memory idempotency.Repository, matching the
// semantics exercised in internal/server/idempotency's own middleware
// tests: the first TryAcquire for a (org, key) pair wins and starts
// "processing"; later ones see that record until Complete overwrites it.
type mockIdemRepo struct {
	mu      sync.Mutex
	records map[string]*idempotency.Record
}

func newMockIdemRepo() *mockIdemRepo {
	return &mockIdemRepo{records: map[string]*idempotency.Record{}}
}

func (m *mockIdemRepo) TryAcquire(_ context.Context, orgID, key, requestHash string, _ time.Time) (*idempotency.Record, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := orgID + ":" + key
	if rec, ok := m.records[k]; ok {
		cp := *rec
		return &cp, true, nil
	}
	m.records[k] = &idempotency.Record{OrgID: orgID, Key: key, RequestHash: requestHash, Status: idempotency.StatusProcessing}
	return nil, false, nil
}

func (m *mockIdemRepo) Complete(_ context.Context, orgID, key string, responseStatus int, responseBody []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := orgID + ":" + key
	rec, ok := m.records[k]
	if !ok {
		return nil
	}
	rec.Status = idempotency.StatusComplete
	rec.ResponseStatus = responseStatus
	rec.ResponseBody = responseBody
	return nil
}

func newBatchRouter(svc Service, repo idempotency.Repository) http.Handler {
	h := NewHandler(svc).WithIdempotency(idempotency.RequiredMiddleware(repo))
	r := chi.NewRouter()
	r.Route("/", h.Routes())
	return r
}

func newBatchRequest(t *testing.T, key string) *http.Request {
	t.Helper()
	body := `{"from_wallet_id":"11111111-1111-4111-8111-111111111111","transfers":[{"to_wallet_id":"22222222-2222-4222-8222-222222222222","asset":"USDC","amount":"10"}]}`
	ctx := tenant.WithID(context.Background(), "org-1")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body)).WithContext(ctx)
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	return req
}

func TestCreateBatchWithoutIdempotencyKeyIsRejected(t *testing.T) {
	svc := &fakeService{}
	router := newBatchRouter(svc, newMockIdemRepo())

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, newBatchRequest(t, ""))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if svc.hits() != 0 {
		t.Fatalf("expected CreateBatch not to run, ran %d times", svc.hits())
	}
}

func TestCreateBatchDuplicateKeyReplaysCachedResponseWithoutReprocessing(t *testing.T) {
	svc := &fakeService{}
	router := newBatchRouter(svc, newMockIdemRepo())
	key := uuid.New().String()

	first := httptest.NewRecorder()
	router.ServeHTTP(first, newBatchRequest(t, key))
	second := httptest.NewRecorder()
	router.ServeHTTP(second, newBatchRequest(t, key))

	if svc.hits() != 1 {
		t.Fatalf("expected CreateBatch to run exactly once, ran %d times", svc.hits())
	}
	if first.Code != second.Code || first.Body.String() != second.Body.String() {
		t.Fatalf("expected identical replayed response, got %d %q vs %d %q",
			first.Code, first.Body.String(), second.Code, second.Body.String())
	}
}

func TestCreateBatchAcceptsNonUUIDIdempotencyKey(t *testing.T) {
	svc := &fakeService{}
	router := newBatchRouter(svc, newMockIdemRepo())

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, newBatchRequest(t, "not-a-uuid"))

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for a non-UUID key, got %d: %s", rec.Code, rec.Body.String())
	}
	if svc.hits() != 1 {
		t.Fatalf("expected CreateBatch to run once, ran %d times", svc.hits())
	}
}

func TestToBatchResponseIncludesFailureDiagnostics(t *testing.T) {
	response := toBatchResponse(&Result{
		Batch: &domain.Batch{CreatedAt: time.Now().UTC()},
		Transactions: []*domain.Transaction{{
			Status:         domain.StatusFailed,
			FailureReason:  "transfer_initiation_failed",
			FailureMessage: "destination account does not exist",
		}},
	})

	if len(response.Transfers) != 1 {
		t.Fatalf("got %d transfers, want 1", len(response.Transfers))
	}
	if response.Transfers[0].FailureReason != "transfer_initiation_failed" {
		t.Fatalf("failure reason = %q, want transfer_initiation_failed", response.Transfers[0].FailureReason)
	}
	if response.Transfers[0].FailureMessage != "destination account does not exist" {
		t.Fatalf("failure message = %q, want underlying error", response.Transfers[0].FailureMessage)
	}
}
