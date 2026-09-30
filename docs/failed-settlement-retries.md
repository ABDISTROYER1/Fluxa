# Retrying failed settlements

Owners and Admins can request a retry for a failed transfer with:

```http
POST /v1/admin/transfers/{transferID}/force-settle
Authorization: Bearer <tenant API key>
```

The endpoint atomically changes `failed` to `pending`, clears the failed
attempt's transaction hash and safe failure diagnostics, then queues the
normal settlement worker. A `202` response means the task was queued; check
the transfer status to confirm its eventual outcome. The transfer remains
scoped to the authenticated tenant and live or test environment. Missing or
out-of-tenant IDs return `404`; a transfer that is no longer failed returns
`409`, so simultaneous or duplicate requests cannot queue two retries.

Only a definite failed settlement is eligible. A submitted transfer with an
unknown Stellar outcome must be resolved by reconciliation before an operator
can retry it; this avoids paying twice. The normal settlement worker still
performs its configured bounded network submission attempts (three by
default, with a two-second linear backoff). Operators may retry again only
after another definite failure. There are no retry-specific environment
settings or separate UI; this is an audited admin API operation. If queueing
fails after the atomic status change, the transfer remains pending and the
normal stuck-transfer recovery loop can enqueue it again. Audit records
contain the transfer ID and action only, not amounts, wallet data, provider
responses, or signing material.
