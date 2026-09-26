package webhook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fluxa/fluxa/internal/domain"
)

func TestDispatcher_SendsNonEmptyBodyMatchingPayload(t *testing.T) {
	repo := newFakeRepo()

	var receivedBody string
	var receivedContentType string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		bs, _ := io.ReadAll(r.Body)
		receivedBody = string(bs)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ep := &domain.WebhookEndpoint{
		ID:        "ep-test-body",
		URL:       ts.URL,
		Secret:    "whsec_test",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = repo.CreateEndpoint(context.Background(), ep)

	payloadObj := map[string]string{"event": "transfer.settled", "id": "tx-123"}
	bs, err := json.Marshal(payloadObj)
	if err != nil {
		t.Fatalf("unexpected error marshalling payload: %v", err)
	}
	payloadStr := string(bs)

	deliv := &domain.WebhookDelivery{
		ID:           "del-body-1",
		EndpointID:   ep.ID,
		EventType:    "transfer.settled",
		Method:       http.MethodPost,
		Payload:      payloadStr,
		Status:       "pending",
		AttemptCount: 0,
		MaxAttempts:  3,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	_ = repo.CreateDelivery(context.Background(), deliv)

	d := NewWebhookDispatcher(repo, nil, 3, false)
	err = d.Deliver(context.Background(), deliv.ID)
	if err != nil {
		t.Fatalf("expected deliver success, got %v", err)
	}

	if receivedBody != payloadStr {
		t.Fatalf(
			"expected dispatched request body %q, got %q",
			payloadStr,
			receivedBody,
		)
	}

	if receivedContentType != "application/json" {
		t.Fatalf(
			"expected Content-Type application/json, got %q",
			receivedContentType,
		)
	}
}
