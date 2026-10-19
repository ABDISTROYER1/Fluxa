# API key scopes

A tenant admin can issue API keys that only reach the resources and actions an
integration actually needs, instead of handing out full tenant access.

## Scope format

Scopes are `<resource>:<action>`, where the action is `read` or `write`.

| Resource | Read | Write | Routes |
|---|---|---|---|
| Wallets | `wallets:read` | `wallets:write` | `/v1/wallets/**` |
| Transfers | `transfers:read` | `transfers:write` | `/v1/transfers/**`, `/v1/transactions`, `/v1/schedules/**`, `/v1/refunds/**` |
| Batches | `batches:read` | `batches:write` | `/v1/transfers/batch/**` |
| Webhooks | `webhooks:read` | `webhooks:write` | `/v1/webhooks/**` |
| Reports | `reports:read` | `reports:write` | `/v1/usage` (`reports:write` is reserved for report generation and currently gates nothing) |
| Fiat | `fiat:read` | `fiat:write` | `/v1/fiat/**`, `/v1/wallets/{id}/deposit`, `/v1/wallets/{id}/withdraw`, `/v1/payment-links/**` |
| Beneficiaries | `beneficiaries:read` | `beneficiaries:write` | `/v1/beneficiaries/**` |
| API keys | `keys:read` | `keys:write` | `/v1/keys/**` |

`audit:read`, `fees:read`, `fx:read`/`fx:write` and `compliance:read`/`compliance:write`
also exist. `<resource>:*` grants both actions on one resource and `*` grants
everything.

## How enforcement works

- On the scoped groups above, `GET`, `HEAD` and `OPTIONS` need the `read` scope
  and every other method needs `write`. The check is applied to the whole route
  group, so a mutating endpoint added later can't be reached with a read scope.
- A request without the required scope gets `403` with the error code
  `INSUFFICIENT_SCOPE`. The code is stable; the message names the missing scope.
- Each denial is written to the tenant audit log as `api_key.scope_denied`, with
  the key id, the required scope, the key's granted scopes, and the method and
  path. The raw key, its hash and the `Authorization` header are never recorded.
- The SDKs call the same HTTP API, so the same rules apply. The TypeScript SDK
  raises a `PermissionError` whose `requiredScope` is the missing scope.
- Dashboard sessions (JWT) have no scopes; they are governed by member roles.

## Creating a scoped key

```http
POST /v1/keys
{
  "label": "Payouts integration",
  "scopes": ["wallets:read", "transfers:read", "batches:write"]
}
```

Unknown scopes are rejected with `400`. A request made with a scoped key that
holds `keys:write` can only grant scopes that key holds itself, and cannot create
a full-access key. `GET /v1/keys` returns each key's `scopes` alongside its
public `prefix`; the secret is only ever returned once, at creation.

## Compatibility

- **Existing keys keep full access.** The `scopes` column was added with a
  default of `'{}'`, and an empty scope list means unrestricted. Every key that
  existed before scopes, and any key created without `scopes`, keeps working
  exactly as before.
- **Batches.** Batch payouts used to require `transfers:write`. Keys holding
  `transfers:read`/`transfers:write` (or `transfers:*`) still get
  `batches:read`/`batches:write`. New keys can be limited to batches alone.
- **Write routes that only checked read.** `POST /v1/wallets`, trustlines,
  `POST /v1/transfers`, transfer cancellation and beneficiary changes used to
  accept a key with only the matching `read` scope, and deposits/withdrawals
  under `/v1/wallets/{id}` had no scope check. A scoped key now needs the
  `write` scope for these. Unscoped keys are unaffected.
- **Usage.** `GET /v1/usage` now needs `reports:read` for scoped keys.
