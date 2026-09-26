package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// Dispatcher handles webhook dispatches.
type Dispatcher interface {
	Dispatch(ctx context.Context, eventType string, payload interface{}) error
}

type HttpDispatcher struct {
	client *http.Client
	store  DeliveryStore
}

type DeliveryStore interface {
	CreateDelivery(ctx context.Context, ep Endpoint, payload string) error
}

type Endpoint struct {
	URL string
}

func NewHttpDispatcher(client *http.Client, store DeliveryStore) *HttpDispatcher {
	return &HttpDispatcher{
		client: client,
		store:  store,
	}
}

func (d *HttpDispatcher) Dispatch(ctx context.Context, ep Endpoint, eventType string, eventData interface{}) error {
	data, err := json.Marshal(eventData)
	if err != nil {
		return err
	}
	payload := string(data)

	err = d.store.CreateDelivery(ctx, ep, payload)
	if err != nil {
		return err
	}

	var body *bytes.Buffer
	if payload != "" {
		body = bytes.NewBufferString(payload)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ep.URL, body)
	if err != nil {
		return err
	}
	if payload != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
