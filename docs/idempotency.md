# Idempotency Keys

## What they are, and why they matter

Any client that calls a state-mutating Fluxa endpoint can lose the response to a
network error or timeout without knowing whether the request actually
succeeded server-side. Retrying blindly risks double-processing — e.g. a
second `POST /v1/transfers` that moves the same funds twice.

An idempotency key breaks that ambiguity. The client generates a unique key
once per _logical_ operation and sends it as the `X-Idempotency-Key` (or `Idempotency-Key`) header.
Fluxa remembers the outcome of the first request under that key for 24 hours
(configurable via `IDEMPOTENCY_TTL_HOURS`); any retry with the same key and the
same request body gets back the exact same response, byte for byte, without the
operation running again. This is the same model used by Stripe, Adyen, and other
payment processors.

## Endpoints that support idempotency keys

| Method | Path                               | Header                                   | Note                                                                            |
| ------ | ---------------------------------- | ---------------------------------------- | ------------------------------------------------------------------------------- |
| `POST` | `/v1/transfers`                    | `X-Idempotency-Key` or `Idempotency-Key` | Optional, prevents duplicate transfers                                          |
| `POST` | `/v1/withdrawals`                  | `X-Idempotency-Key` or `Idempotency-Key` | Optional, prevents duplicate withdrawals                                        |
| `POST` | `/v1/wallets/:id/withdraw`         | `X-Idempotency-Key` or `Idempotency-Key` | Optional                                                                        |
| `POST` | `/v1/transfers/batch`              | `Idempotency-Key` or `X-Idempotency-Key` | Required; retries with the same key and body replay the original batch response |
| `POST` | `/v1/fx/convert`                   | `Idempotency-Key` or `X-Idempotency-Key` | FX conversions                                                                  |
| `POST` | `/v1/wallets`                      | `Idempotency-Key` or `X-Idempotency-Key` | Wallet creation                                                                 |
| `POST` | `/v1/wallets/:id/trustlines`       | `Idempotency-Key` or `X-Idempotency-Key` | Trustline configuration                                                         |
| `POST` | `/v1/claimable-balances`           | `Idempotency-Key` or `X-Idempotency-Key` | Claimable balance creation                                                      |
| `POST` | `/v1/claimable-balances/:id/claim` | `Idempotency-Key` or `X-Idempotency-Key` | Claimable balance claiming                                                      |

Read-only endpoints (e.g. `GET /v1/transfers/:id`) never require or use the header.

## Generating a key

Clients can provide a **UUID v4** or any unique client reference string (e.g. `tx_order_987213`). Non-UUID strings are deterministically transformed into a RFC 4122 compliant UUID using SHA-256:

```bash
# macOS / Linux
uuidgen

# Python
python3 -c "import uuid; print(uuid.uuid4())"

# Node.js
node -e "console.log(require('crypto').randomUUID())"
```

Generate the key **once** when the operation is first attempted, and reuse
that same key for every retry of that operation — never generate a new one
per HTTP attempt.

## Retry strategy

- Use exponential backoff with jitter between retries (e.g. `base * 2^attempt + random_jitter`).
- Cap retries at **5 attempts**.
- **Never rotate the idempotency key on retry.** Rotating it defeats the
  entire mechanism — the server will see it as a brand-new operation and
  execute it again.
- If a retry returns `409 REQUEST_IN_PROGRESS`, back off and retry the same
  key later — the original request is still being processed.

## Example

```bash
curl -X POST https://api.fluxa.example/v1/transfers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: $(uuidgen)" \
  -d '{
    "from_wallet_id": "b3f1...",
    "to_wallet_id": "9ac2...",
    "asset": "USDC",
    "amount": "100.00"
  }'
```

Retrying the exact same request (same key, same body) returns the original
response — no duplicate transfer is created.

## Error codes

| Code                                         | HTTP status | Meaning                                                                                                           |
| -------------------------------------------- | ----------- | ----------------------------------------------------------------------------------------------------------------- |
| `IDEMPOTENCY_KEY_REQUIRED`                   | 400         | The `Idempotency-Key` header was missing on an endpoint that requires it.                                         |
| `REQUEST_IN_PROGRESS`                        | 409         | A request with this key is already being processed; retry later.                                                  |
| `IDEMPOTENCY_KEY_REUSED_WITH_DIFFERENT_BODY` | 409         | This key was already used with a different request body — generate a new key for a genuinely different operation. |

## How it works server-side

Each request is fingerprinted as `SHA-256(method + path + body)`. The first
request for a given `(org, key)` pair claims the key and proceeds to the
handler; the response is cached against the key once the handler completes.
Any later request with the same key:

- while the original is still in flight → `409 REQUEST_IN_PROGRESS`
- after it completed, with the same body → the cached response, replayed exactly
- after it completed, with a different body → `409 IDEMPOTENCY_KEY_REUSED_WITH_DIFFERENT_BODY`

Records are scoped per organization and expire after the configured TTL
(default 24 hours), after which the same key can be reused for a new operation.

## TTL and cleanup

The TTL for idempotency records defaults to **24 hours** and is controlled by
the `IDEMPOTENCY_TTL_HOURS` environment variable (minimum: 1 hour). Setting a
shorter TTL reduces storage but also shrinks the replay window — clients must
complete all retries within the configured window.

### Background cleanup job

The worker process runs a cleanup goroutine every **hour** that removes all
rows whose `expires_at` is in the past. Deletes are issued in batches of 1 000
rows to avoid long-held locks or I/O spikes. On startup the worker immediately
drains any backlog accumulated during downtime.

The `idempotency_records` table has an index on `expires_at`
(`idx_idempotency_records_expires_at`, added in migration
`20260927000000`) so the batch delete uses an index scan rather than a
sequential scan.

### Per-key opportunistic delete

`TryAcquire` also issues a targeted `DELETE … WHERE org_id = $1 AND key = $2
AND expires_at <= NOW()` on the hot path when a key-reuse attempt hits an
expired row blocking the unique index. This is deliberately kept alongside the
background sweep:

- It clears the conflict immediately on the request path, so a client reusing a
  key after its TTL gets a fresh response without waiting for the next hourly
  sweep.
- It is scoped to a single `(org_id, key)` pair and never performs a
  table-wide scan, so it has no impact on vacuum or I/O for unrelated rows.

The background job covers the common case (keys used once, never retried) where
the opportunistic delete never fires.

## Data retention

Every idempotency record stores a `request_hash` (a SHA-256 digest of the
method, path, and request body) and a `response_body` (the serialized HTTP
response). This data:

- Is retained for the duration of the TTL (`IDEMPOTENCY_TTL_HOURS`, default 24h).
- Is deleted by the background cleanup job once `expires_at` has elapsed.
- Should be considered in any organization-level data-retention or right-to-erasure policy,
  since `response_body` may contain transaction IDs and amounts.
