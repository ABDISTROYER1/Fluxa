# Scheduled tenant payouts

Schedules create recurring transfers between two wallets owned by the authenticated tenant. The source and destination wallets are validated at creation; normal transfer fee, balance, asset, and compliance checks run for each payout. Schedule IDs and run history are tenant-scoped.

## Create and manage a schedule

Create a schedule with `POST /v1/schedules`, providing `from_wallet_id`, `to_wallet_id`, `asset`, `amount`, `frequency`, and an RFC3339 `start_date`. An optional `end_date` must be at or after `start_date`. The optional `timezone` is an IANA time zone and defaults to `UTC`. `missed_run_policy` can be `skip` or `run_once`; it defaults to `run_once`. Supply an `Idempotency-Key` and reuse it only when retrying the same request. Reusing the key with a different body returns a conflict.

Schedules can be paused or resumed with `PATCH /v1/schedules/{id}` and permanently cancelled with `DELETE /v1/schedules/{id}`. A failed schedule can be resumed to continue future installments after its cause is corrected; the failed occurrence is not retried automatically. Mutations are recorded in the tenant audit log without recording payout amounts, wallet identifiers, or provider error details.

## Execution and recovery

The worker checks for due schedules every minute. Each occurrence has a durable run record unique to its schedule and expected run time. Concurrent workers claim the schedule atomically, and transfer initiation uses a stable idempotency key so retries do not create a second transfer. A failed initiation marks the run and schedule as failed; the run history exposes a safe summary, not internal provider details. Resume a failed schedule after addressing the underlying wallet, asset, or balance issue; that does not replay the failed occurrence.

Daily, weekly, and monthly occurrences use the schedule time zone. `skip` advances past elapsed occurrences when a paused schedule is resumed; when the worker returns to an active schedule with a backlog, it records a skipped run and advances beyond that backlog. `run_once` retains one overdue occurrence for execution after resume and coalesces an active backlog into a single catch-up payout. The schedule listing returns at most 100 records. Run history defaults to 20 records and is capped at 100; use `limit` and `offset` to page through it.

## Configuration and limits

The worker interval (one minute), due-schedule page size (100), and API listing limits are currently fixed in the application; there are no schedule-specific environment variables. The overall API idempotency retention period is configured by `IDEMPOTENCY_TTL_HOURS` (default 24 hours). The create endpoint requires a UUID v4 idempotency key, and existing migration defaults preserve schedules created before timezone and missed-run policy were added.
