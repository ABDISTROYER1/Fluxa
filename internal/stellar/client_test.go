package stellar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

const (
	srcAccount = "GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H"
	dstAccount = "GC2BKLYOOYPDEFJKLKY6FNNRQMGFLVHJKQRGNSSRRGSMPGF32LHCQVGF"
	usdcIssuer = "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"
)

func jsonResponse(w http.ResponseWriter, payload string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(payload))
}

func newTestHorizon(t *testing.T, handler http.HandlerFunc) *horizonClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, ok := NewClient(srv.URL, "testnet", 0).(*horizonClient)
	if !ok {
		t.Fatal("NewClient did not return *horizonClient")
	}
	return client
}

// TestFindPathsStrict_PopulatesSourceAndDestination is the regression test for
// the bug where the paying account was written into the destination slot and
// no source account was sent at all.
func TestFindPathsStrict_PopulatesSourceAndDestination(t *testing.T) {
	var gotQuery string
	client := newTestHorizon(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		jsonResponse(w, `{"_links":{"self":{"href":""}},"_embedded":{"records":[{"source_account":"`+srcAccount+`"}]}}`)
	})

	paths, err := client.FindPathsStrict(srcAccount, dstAccount, "USDC", usdcIssuer, "10.0000000")
	if err != nil {
		t.Fatalf("FindPathsStrict() error: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(paths))
	}

	values, err := url.ParseQuery(gotQuery)
	if err != nil {
		t.Fatalf("parse query: %v", err)
	}
	if got := values.Get("source_account"); got != srcAccount {
		t.Errorf("source_account = %q, want %q", got, srcAccount)
	}
	if got := values.Get("destination_account"); got != dstAccount {
		t.Errorf("destination_account = %q, want %q", got, dstAccount)
	}
	if got := values.Get("destination_asset_code"); got != "USDC" {
		t.Errorf("destination_asset_code = %q, want USDC", got)
	}
	if got := values.Get("destination_asset_issuer"); got != usdcIssuer {
		t.Errorf("destination_asset_issuer = %q, want %q", got, usdcIssuer)
	}
	if got := values.Get("destination_amount"); got != "10.0000000" {
		t.Errorf("destination_amount = %q, want 10.0000000", got)
	}
}

func TestFindPathsStrict_HorizonErrorIsWrapped(t *testing.T) {
	client := newTestHorizon(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"type":   "https://stellar.org/horizon-errors/bad_request",
			"title":  "Bad Request",
			"status": 400,
		})
	})

	_, err := client.FindPathsStrict(srcAccount, dstAccount, "USDC", usdcIssuer, "1")
	if err == nil {
		t.Fatal("expected an error for a 400 response")
	}
	if !strings.Contains(err.Error(), "find paths") {
		t.Errorf("error should be wrapped with context, got %v", err)
	}
}

func TestPayments_ReturnsRecords(t *testing.T) {
	var gotQuery string
	client := newTestHorizon(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		jsonResponse(w, `{"_embedded":{"records":[{"id":"1","type":"payment","paging_token":"1"}]}}`)
	})

	records, err := client.Payments(dstAccount, "cursor-1", 25)
	if err != nil {
		t.Fatalf("Payments() error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 payment, got %d", len(records))
	}

	values, _ := url.ParseQuery(gotQuery)
	if got := values.Get("cursor"); got != "cursor-1" {
		t.Errorf("cursor = %q, want cursor-1", got)
	}
	if got := values.Get("limit"); got != "25" {
		t.Errorf("limit = %q, want 25", got)
	}
	if got := values.Get("order"); got != "asc" {
		t.Errorf("order = %q, want asc", got)
	}
}

func TestLoadAccount_Success(t *testing.T) {
	client := newTestHorizon(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/accounts/"+dstAccount {
			http.NotFound(w, r)
			return
		}
		jsonResponse(w, `{"account_id":"`+dstAccount+`","sequence":"42"}`)
	})

	acct, err := client.LoadAccount(dstAccount)
	if err != nil {
		t.Fatalf("LoadAccount() error: %v", err)
	}
	if acct.AccountID != dstAccount {
		t.Errorf("AccountID = %q, want %q", acct.AccountID, dstAccount)
	}
	if acct.Sequence != 42 {
		t.Errorf("Sequence = %d, want 42", acct.Sequence)
	}
}

// TestLoadAccount_NotFoundIsMapped verifies the 404 that callers rely on to
// distinguish "account does not exist yet" from a transport failure.
func TestLoadAccount_NotFoundIsMapped(t *testing.T) {
	client := newTestHorizon(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"type":   "https://stellar.org/horizon-errors/not_found",
			"title":  "Resource Missing",
			"status": 404,
		})
	})

	_, err := client.LoadAccount(dstAccount)
	if err == nil {
		t.Fatal("expected an error for a 404 response")
	}
	if !strings.Contains(err.Error(), "load account") {
		t.Errorf("error should be wrapped with context, got %v", err)
	}
}

func TestNewClient_ConfiguresHTTPTimeout(t *testing.T) {
	def, ok := NewClient("https://horizon-testnet.stellar.org", "testnet", 0).(*horizonClient)
	if !ok {
		t.Fatal("NewClient did not return *horizonClient")
	}
	defHTTP, ok := def.inner.HTTP.(*http.Client)
	if !ok || defHTTP == nil {
		t.Fatal("Horizon client has no *http.Client: requests would be unbounded")
	}
	if defHTTP.Timeout != defaultHorizonTimeout {
		t.Errorf("default timeout = %s, want %s", defHTTP.Timeout, defaultHorizonTimeout)
	}

	custom, ok := NewClient("https://horizon-testnet.stellar.org", "testnet", 3*time.Second).(*horizonClient)
	if !ok {
		t.Fatal("NewClient did not return *horizonClient")
	}
	customHTTP, ok := custom.inner.HTTP.(*http.Client)
	if !ok || customHTTP == nil {
		t.Fatal("Horizon client has no *http.Client")
	}
	if customHTTP.Timeout != 3*time.Second {
		t.Errorf("configured timeout = %s, want 3s", customHTTP.Timeout)
	}
}

func TestBoundCtx_AddsDeadlineWhenAbsent(t *testing.T) {
	client := &horizonClient{timeout: 5 * time.Second}

	bounded, cancel := client.boundCtx(context.Background())
	defer cancel()
	deadline, ok := bounded.Deadline()
	if !ok {
		t.Fatal("boundCtx must add a deadline when the caller has none")
	}
	if remaining := time.Until(deadline); remaining <= 0 || remaining > 5*time.Second {
		t.Errorf("deadline %s is not within the configured timeout", remaining)
	}

	// A caller-supplied deadline must be preserved, not shortened.
	parent, parentCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer parentCancel()
	want, _ := parent.Deadline()
	kept, keepCancel := client.boundCtx(parent)
	defer keepCancel()
	got, _ := kept.Deadline()
	if !got.Equal(want) {
		t.Errorf("boundCtx changed an existing deadline: got %s want %s", got, want)
	}
}
