package wallet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestFriendbotProvisionerRequestsAddressAndRetries(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Query().Get("addr") != "GABC" {
			t.Errorf("addr query = %q", r.URL.Query().Get("addr"))
		}
		if calls.Load() == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provisioner := NewFriendbotProvisioner(server.URL)
	provisioner.Attempts = 2
	if err := provisioner.EnsureAccount(context.Background(), "GABC"); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}
}

func TestFriendbotProvisionerDoesNotHidePermanentFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()
	if err := NewFriendbotProvisioner(server.URL).EnsureAccount(context.Background(), "GABC"); err == nil {
		t.Fatal("expected permanent Friendbot failure")
	}
}
