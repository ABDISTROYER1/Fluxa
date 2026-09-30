# Hosted Payment Links and Refunds

## Payment Links

Create a tenant-scoped payment link for a wallet. Amount and currency are fixed; callers cannot change either during checkout. Expiration must be in the future and no more than 90 days away.

```http
POST /v1/payment-links
Authorization: Bearer <api-key>
Content-Type: application/json

{
  "wallet_id": "<wallet-uuid>",
  "amount": "2500.00",
  "currency": "NGN",
  "expires_at": "2026-10-07T12:00:00Z"
}
```

The response contains a checkout URL such as `/pay/<token>`. Share the URL with the customer. Public checkout displays only the fixed amount, currency, status, and expiry; the token is an unguessable UUID. The first checkout request atomically claims the link. Replays, concurrent attempts, cancelled links, and expired links receive a conflict response. A failed payment-provider initiation marks the link failed, so create a new link before trying again.

Authenticated endpoints support listing, retrieving, and cancelling unused links. Link data is isolated by tenant and live/test environment. Successful and failed fiat deposit webhooks update the associated link status.

## Refunds

Only confirmed or settled transfers can be refunded. Create a refund using a unique `Idempotency-Key` and the original transaction ID:

```http
POST /v1/refunds
Authorization: Bearer <api-key>
Idempotency-Key: <unique-key>
Content-Type: application/json

{
  "original_transaction_id": "<transaction-uuid>",
  "amount": "20.0000000",
  "reason": "Order returned"
}
```

Amounts use the original transaction asset and cannot exceed its net credited amount after fees. Concurrent and repeated refunds share a database reservation, so requested, pending, and succeeded refunds together cannot exceed that cap. A refund is executed as a new reverse transfer; its status is `pending` until settlement, then `succeeded` or `failed`. The refund record includes its new transaction ID. Poll `GET /v1/refunds/{id}` or list with `GET /v1/refunds?original_transaction_id=<uuid>` to follow the lifecycle.

Refunds are isolated by tenant and environment. Creation requires an idempotency key; reusing that key with different transaction, amount, or reason returns a conflict. Mutations are recorded in the append-only audit log without customer email or payment-link tokens.