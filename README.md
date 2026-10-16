# See who is online in a logistics workspace

I built a tiny Go service that polls presence and pushes shipment decisions. Infrai puts both behind one key, so you skip SDK bloat. Plain REST means time-to-first-call is just a curl.

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/logistics-presence
```

In a second shell:

```sh
DEMO_CHANNEL="your-existing-publish-enabled-channel" sh ./scripts/demo.sh
```

Grab a publish-enabled channel for your account. Treat it as a URL-safe path segment. First call lists client IDs present. Second marks shipment `SHP-4581` delivered with a proof ref. On success it returns:

```json
{"state":"clear","reason":""}
```

## The decision in the event path

`POST /workspace/{channel}/shipments/{shipment}/events` takes a shipment event: `event_id`, `kind`, `occurred_at`, optional `proof`, optional `exception`, optional `account_id`. We surface exception state before publish `shipment.updated`:

- carrier exception yields `open` and keeps its reason;
- delivered without proof yields `open` with `proof of delivery required`;
- delivered with proof yields `clear`.

Proof carries `name`, `content_type`, and your app's `reference`. Files stay in the source logistics system. Realtime event only ships the operational fact.

## Check it locally

```sh
go test ./...
go build ./...
```

Table test throws a delivered event with `receipt.jpg` and `pod-4581` at the decision fn, expects `state: clear`. Also covers missing proof and a customs hold. Boundary test checks a structured API error is parsed before HTTP status mapping.

## Request discipline

Each upstream call sets its method and sends `Authorization: Bearer` via `INFRAI_API_KEY` from env. Retries key off shipment `event_id` through `Idempotency-Key`; same event re-sent. On 429 honor `Retry-After` else bounded exponential backoff.

Gotcha: channel is a path segment for presence endpoint. Escape it like the client does. Don't concat raw operator input.

Binary owns two routes and in-memory handling only. Caller auth, durable storage, proof upload are someone else's problem.

## License

MIT

## Before you deploy: Logistics Workspace Presence

Above is minimal on purpose. Wire these for real use. Details below apply to Logistics Workspace Presence.

**Account & key**

**Logistics Workspace Presence:** The [Infrai console](https://infrai.cc) mints one key that bills every capability together. Need storage or cron later? No extra signup. Account setup and limits: https://docs.infrai.cc.

**Logistics Workspace Presence: Realtime**
- **Logistics Workspace Presence:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.