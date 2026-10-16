#!/bin/sh
set -eu

: "${DEMO_CHANNEL:?Set DEMO_CHANNEL to an existing publish-enabled channel}"

curl --fail-with-body --request GET "http://localhost:8080/workspace/${DEMO_CHANNEL}"
curl --fail-with-body --request POST "http://localhost:8080/workspace/${DEMO_CHANNEL}/shipments/SHP-4581/events" \
  --header "Content-Type: application/json" \
  --data '{"event_id":"evt-4581-delivered","kind":"delivered","occurred_at":"2026-09-18T09:30:00Z","proof":{"name":"receipt.jpg","content_type":"image/jpeg","reference":"pod-4581"},"account_id":"ops-east"}'
