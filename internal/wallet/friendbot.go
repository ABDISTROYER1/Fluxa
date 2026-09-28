package wallet

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// FriendbotProvisioner funds Stellar testnet accounts through the public
// Friendbot faucet. It is only invoked for wallets whose authenticated mode is
// test; live wallets never reach this type.
type FriendbotProvisioner struct {
	Endpoint string
	Client   *http.Client
	Attempts int
}

func NewFriendbotProvisioner(endpoint string) *FriendbotProvisioner {
	return &FriendbotProvisioner{
		Endpoint: endpoint,
		Client:   &http.Client{Timeout: 10 * time.Second},
		Attempts: 3,
	}
}

func (p *FriendbotProvisioner) EnsureAccount(ctx context.Context, publicKey string) error {
	if p == nil || p.Endpoint == "" {
		return fmt.Errorf("friendbot endpoint is not configured")
	}
	if _, err := url.ParseRequestURI(p.Endpoint); err != nil {
		return fmt.Errorf("invalid friendbot endpoint: %w", err)
	}
	attempts := p.Attempts
	if attempts < 1 {
		attempts = 1
	}
	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			delay := time.Duration(1<<(attempt-1)) * 250 * time.Millisecond
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
		endpoint, err := url.Parse(p.Endpoint)
		if err != nil {
			return err
		}
		query := endpoint.Query()
		query.Set("addr", publicKey)
		endpoint.RawQuery = query.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
		if err != nil {
			return fmt.Errorf("create friendbot request: %w", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		lastErr = fmt.Errorf("friendbot returned HTTP %s", resp.Status)
		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500 {
			return lastErr
		}
	}
	return fmt.Errorf("friendbot funding failed after %d attempts: %w", attempts, lastErr)
}
