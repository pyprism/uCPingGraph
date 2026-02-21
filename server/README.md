# uCPingGraph Server

## Run

```bash
go run . server
```

Default URL: `http://127.0.0.1:8880`

## CLI

```bash
go run . network add
go run . device add
go run . generate
go run . cleanup
```

## API

### `POST /api/stats`

Headers:

- `Authorization: <device_token>`
- `Content-Type: application/json`

Body:

```json
{
  "latency_ms": 14.2,
  "sent_packets": 5,
  "received_packets": 4,
  "packet_loss_percent": 20,
  "target": "1.1.1.1",
  "platform": "esp8266",
  "rssi": -68
}
```

Notes:

- `latency` (legacy key) is still accepted for backward compatibility.
- If packet-loss is omitted, it is derived from sent/received counters.

### `GET /api/networks`

Returns network list for dashboard selectors.

### `GET /api/networks/:network/devices`

Returns device list for a given network.

### `GET /api/series?network=<name>&device=<name>&minutes=<1-10080>`

Returns chart series and summary metrics.

## Tests

```bash
GOCACHE=$(pwd)/.cache/go-build go test ./...
```
