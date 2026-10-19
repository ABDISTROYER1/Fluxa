package tenantdata

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

type rowResult struct {
	data []byte
	err  error
}

func (r rowResult) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*(dest[0].(*[]byte)) = r.data
	return nil
}

type fakeDB struct {
	queries []string
	args    [][]any
	failAt  int
}

func (f *fakeDB) QueryRow(_ context.Context, query string, args ...interface{}) pgx.Row {
	f.queries = append(f.queries, query)
	f.args = append(f.args, args)
	if f.failAt == len(f.queries) {
		return rowResult{err: errors.New("database unavailable")}
	}
	return rowResult{data: []byte("[]")}
}

func TestExportIsTenantScopedAndOmitsCredentialAndWebhookPayloadFields(t *testing.T) {
	db := &fakeDB{}
	result, err := NewService(db).Build(context.Background(), "tenant-123")
	if err != nil {
		t.Fatal(err)
	}
	if result.TenantID != "tenant-123" || len(result.Data) != 11 {
		t.Fatalf("unexpected export: %+v", result)
	}
	for i, query := range db.queries {
		if !strings.Contains(query, "$1") || db.args[i][0] != "tenant-123" {
			t.Errorf("query %d is not tenant-scoped: %s", i, query)
		}
	}
	for _, fragment := range []string{"encrypted_secret", "'secret'", "'payload'", "'response_body'", "'error_message'"} {
		found := false
		for _, query := range db.queries {
			found = found || strings.Contains(query, fragment)
		}
		if !found {
			t.Errorf("expected export query to exclude sensitive field %s", fragment)
		}
	}
}

func TestExportRejectsMissingTenantBeforeQuery(t *testing.T) {
	db := &fakeDB{}
	if _, err := NewService(db).Build(context.Background(), ""); err == nil {
		t.Fatal("expected empty tenant id to be rejected")
	}
	if len(db.queries) != 0 {
		t.Fatal("queried database without tenant scope")
	}
}
