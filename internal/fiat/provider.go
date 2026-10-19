package fiat

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// Sentinel errors used by HandleWebhookWithHeaders to classify webhook
// failures for the HTTP layer. The handler maps these to 4xx (permanent
// rejection, no retry) vs 5xx (transient failure, provider should retry).

// ErrWebhookSignatureInvalid is returned when the provider HMAC/hash check
// fails. Providers must not retry — the same payload will always fail.
var ErrWebhookSignatureInvalid = errors.New("webhook signature invalid")

// ErrWebhookPayloadInvalid is returned when the payload cannot be decoded or
// fails structural validation. Providers must not retry.
var ErrWebhookPayloadInvalid = errors.New("webhook payload invalid")

// ErrWebhookEventUnknown is returned when the event type in the payload is
// not one this integration handles. Providers must not retry.
var ErrWebhookEventUnknown = errors.New("webhook event type unknown or unsupported")

type (
	QuoteRequest struct {
		Side         string // "deposit" or "withdraw"
		FiatCurrency string
		FiatAmount   decimal.Decimal
		Country      string
	}

	FiatQuote struct {
		Provider     string
		FiatAmount   decimal.Decimal
		FiatCurrency string
		USDCAmount   decimal.Decimal
		Rate         decimal.Decimal
		Fee          decimal.Decimal
		MinLimit     decimal.Decimal
		MaxLimit     decimal.Decimal
		ExpiresAt    time.Time
	}

	DepositInstruction struct {
		ProviderRef  string
		Instructions map[string]string
	}

	WithdrawalRequest struct {
		WalletID      string
		ProviderRef   string
		FiatAmount    decimal.Decimal
		FiatCurrency  string
		Country       string
		AccountBank   string
		AccountNumber string
		CustomerEmail string
		CustomerName  string
	}

	WithdrawalResult struct {
		ProviderRef string
		Status      string
	}

	RailEvent struct {
		Type        string
		ProviderRef string
		// EventID is the provider's own unique identifier for this
		// occurrence (e.g. Flutterwave's numeric transaction id). Unlike
		// ProviderRef (the merchant-supplied reference, stable across a
		// retried/replayed delivery of the same event), this identifies
		// the specific delivery/charge.
		EventID  string
		Status   string
		Amount   decimal.Decimal
		Currency string
	}
)

const (
	EventDepositConfirmed = "deposit.confirmed"
	EventDepositFailed    = "deposit.failed"
	EventWithdrawalSent   = "withdrawal.sent"
	EventWithdrawalFailed = "withdrawal.failed"
)

type Provider interface {
	Name() string
	SupportedCountries() []string
	// SupportedCurrencies lists the ISO 4217 fiat currency codes (upper case)
	// this provider can settle. The fiat service validates requests against it
	// before pricing them.
	SupportedCurrencies() []string
	GetQuote(ctx context.Context, req QuoteRequest) (*FiatQuote, error)
	InitiateDeposit(ctx context.Context, req DepositRequest) (*DepositInstruction, error)
	InitiateWithdrawal(ctx context.Context, req WithdrawalRequest) (*WithdrawalResult, error)
	GetStatus(ctx context.Context, providerRef string) (*RailEvent, error)
	HandleWebhook(ctx context.Context, payload []byte, headers http.Header) (*RailEvent, error)
}
