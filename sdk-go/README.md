# Fluxa Go SDK

Typed Go client for the Fluxa API. Request and response models are generated from `docs/openapi.yaml`; edit the contract and regenerate rather than editing `models.gen.go`.

## Install

```bash
go get github.com/fluxa/fluxa/sdk-go@v0.1.0
```

Go's module proxy serves tagged Go modules automatically. Release this nested module with a repository tag such as `sdk-go/v0.1.0`.

## Quick start

```go
client, err := fluxa.NewClient(fluxa.Config{APIKey: "sk_live_..."})
if err != nil {
	log.Fatal(err)
}

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

wallet, err := client.Wallets.Create(ctx)
if err != nil {
	log.Fatal(err)
}
transfer, err := client.Transfers.Create(ctx, fluxa.CreateTransferRequest{
	FromWalletId: wallet.Id,
	ToWalletId:   "recipient-wallet-id",
	Asset:        "USDC",
	Amount:       "10.0000000",
})
```

The client uses `https://api.fluxa.io` and a 30-second HTTP timeout by default. Override `BaseURL`, `Timeout`, or `HTTPClient` in `Config`. Every method accepts a `context.Context`, and each also accepts optional `RequestOptions` for a per-request timeout, headers, or idempotency key.

## Idempotency and retries

Financial POST requests receive a generated UUID `Idempotency-Key`; a caller-supplied key is kept stable across attempts. Retries are disabled by default. When enabled with `MaxRetries`, reads can be retried; mutations are replayed only when they have an idempotency key. `Retry-After` is honored for retryable responses.

```go
transfer, err := client.Transfers.Create(ctx, request, fluxa.RequestOptions{
	IdempotencyKey: "payout-2026-09-28",
	Timeout:        5 * time.Second,
})
```

Errors are `*fluxa.FluxaError` values with `Code`, `Message`, `HTTPStatus`, and `Details` fields. Use `errors.As` to inspect them.

## Resources

| Resource | Methods |
| --- | --- |
| `Wallets` | `Create`, `GetBalances`, `CreateTrustline` |
| `Transfers` | `Create`, `Get`, `List`, `CreateBatch`, `GetBatch`, `ExportBatch` |
| `FX` | `Quote`, `Convert`, `GetRates` |
| `Schedules` | `Create`, `List`, `Update`, `Delete` |
| `Webhooks` | `Create`, `List`, `Delete`, `GetDeliveries` |
| `Fees` | `Get`, `ListCollected` |
| `Keys` | `Create`, `List`, `Delete` |
| `Fiat` | `Deposit`, `Withdraw` |

## Regenerate models

```bash
python -m pip install -r tools/sdkgen-requirements.txt
python tools/sdkgen.py
python tools/sdkgen.py --check
```