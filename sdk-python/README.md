# Fluxa Python SDK

Typed asynchronous client for the Fluxa payment API. The package is imported as `fluxa` and uses one reusable `httpx.AsyncClient` per `FluxaClient`.

## Install

```bash
pip install fluxa-api-sdk
```

To publish a release, configure PyPI trusted publishing for the `pypi` GitHub environment and push a tag such as `sdk-python/v0.1.0`.

Python 3.11 or newer is required.

## Quick start

```python
import asyncio
from fluxa import FluxaClient


async def main() -> None:
    async with FluxaClient("sk_live_...") as client:
        wallet = await client.wallets.create()
        transfer = await client.transfers.create(
            {
                "from_wallet_id": wallet["id"],
                "to_wallet_id": "recipient-wallet-id",
                "asset": "USDC",
                "amount": "10.0000000",
            }
        )
        print(transfer["id"])


asyncio.run(main())
```

## Configuration and retries

```python
client = FluxaClient(
    "sk_live_...",
    base_url="https://api.fluxa.io",
    timeout=30.0,
    max_retries=2,  # retries are disabled by default
    retry_delay=0.5,
)
```

Every method is async and supports `RequestOptions(timeout=...)`; cancellation is provided by normal `asyncio` task cancellation. Financial POST operations receive a generated `Idempotency-Key`. A supplied key is preserved:

```python
from fluxa import RequestOptions

await client.transfers.create(payload, RequestOptions(idempotency_key="payout-2026-09-28"))
```

Retries are opt-in. Reads may be retried when enabled; mutations are retried only when the request carries an idempotency key. `Retry-After` is honored for retryable responses.

## Resources

| Resource | Methods |
| --- | --- |
| `wallets` | `create`, `get_balances`, `create_trustline` |
| `transfers` | `create`, `get`, `list`, `create_batch`, `get_batch`, `export_batch` |
| `fx` | `quote`, `convert`, `get_rates` |
| `schedules` | `create`, `list`, `update`, `delete` |
| `webhooks` | `create`, `list`, `delete`, `get_deliveries` |
| `fees` | `get`, `list_collected` |
| `keys` | `create`, `list`, `delete` |
| `fiat` | `deposit`, `withdraw` |
| `payment_links` | `create`, `list`, `get`, `cancel` |
| `refunds` | `create`, `get`, `list` |

All resource methods return typed dictionaries generated from `docs/openapi.yaml`. Errors are `FluxaError` instances with `code`, `message`, `http_status`, and `details`; common HTTP failures have specialized subclasses.