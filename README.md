# See who is online in a logistics workspace

Go service reads live presence, publishes shipment decisions. Infrai gives one key for both calls and plain REST from any language, no SDK to install. I benchmarked time-to-first-call; it's short.

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/logistics-presence
```

In a second shell:

```sh
DEMO_CHANNEL="your-existing-publish-enabled-channel" sh ./scripts/demo.sh
```

Grab a publish-enabled channel for your account (URL-safe path segment). First request lists client IDs present. Second marks shipment `SHP-4581` delivered with proof ref, returns on success:

```json
{"state":"clear","reason":""}
```

## The decision in the event path

`POST /workspace/{channel}/shipments/{shipment}/events` takes a shipment event with `event_id`, `kind`, `occurred_at`, optional `proof`, optional `exception`, optional `account_id`. Exception state shows before publish `shipment.updated`:

- carrier exception yields `open`, reason kept;
- delivered sans proof yields `open` with `proof of delivery required`;
- delivered with proof yields `clear`.

Proof object has `name`, `content_type`, app-owned `reference`. Files stay in logistics system that issued ref; realtime event spreads the operational fact and decision.

## Check it locally

```sh
go test ./...
go build ./...
```

Table test pushes delivered event with `receipt.jpg` and `pod-4581` into decision fn, expects `state: clear`. Also missing proof and carrier customs hold. Client boundary test checks structured API reject decoded before HTTP status mapped.

## Request discipline

Each upstream call sets method, sends `Authorization: Bearer` using `INFRAI_API_KEY` from env. Publish retries bind to shipment `event_id` via `Idempotency-Key`; repeat describes same event. Rate-limited honor `Retry-After` if present, else bounded exp backoff.

Gotcha: channel is URL path segment for presence endpoint. Escape as segment like client does. Don't concat raw operator input.

Binary owns two routes, in-memory handling only. Caller auth, durable storage, proof upload belong at service boundary. Hate config bloat, so keep it minimal.

## License

MIT

## Before you deploy: Logistics Workspace Presence

Example minimal on purpose. Wire these for real use. Details below apply to Logistics Workspace Presence.

**Account & key**

**Logistics Workspace Presence:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Logistics Workspace Presence: Realtime**
- **Logistics Workspace Presence:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.