# uCPingGraph Server

## Run

```bash
go run . server
```

Default URL: `http://127.0.0.1:8080`

## Environment Variables

Copy `.env.example` to `.env` and configure:

| Variable        | Default         | Description                                   |
| --------------- | --------------- | --------------------------------------------- |
| `SERVER_PORT`   | `8080`          | HTTP listen port                              |
| `DEBUG`         | `""`            | Set to `True` for debug mode                  |
| `DB_PATH`       | `./storage/db/uCPingGraph.db` | SQLite database path              |
| `CLEANUP_DAYS`  | `30`            | Days of stats to retain                       |
| `LOG_DIR`       | `./logs`        | Directory for rotated log files               |
| `SENTRY_DSN`    | `""`            | Sentry DSN for error tracking (optional)      |
| `APP_ENV`       | `production`    | Environment label for Sentry                  |

## Logging

The server writes structured JSON logs to `logs/server.log` (rotated at 50 MB, 5 backups, 30-day retention, compressed) and human-readable logs to stdout.

## Sentry

Set `SENTRY_DSN` in your `.env` to enable error tracking. Errors are automatically captured and reported.

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

## Deployment

All deployment is managed via a single Ansible playbook and a single config file in `deploy/`. No separate inventory/hosts file is needed — the host is built dynamically from `deploy_config.yaml`.

### Setup

```bash
cd deploy
cp deploy_config.example.yaml deploy_config.yaml
# Edit deploy_config.yaml with your server details (ssh_host, ssh_user, deploy_dir, etc.)
```

### Fresh systemd deployment (clone + build + install service)

```bash
ansible-playbook deploy.yaml -e @deploy_config.yaml --tags fresh
```

### Update systemd deployment (pull + rebuild + restart)

```bash
ansible-playbook deploy.yaml -e @deploy_config.yaml --tags update
```

### Fresh Docker deployment (clone + docker compose up)

```bash
ansible-playbook deploy.yaml -e @deploy_config.yaml --tags docker-fresh
```

### Update Docker deployment (pull + rebuild containers)

```bash
ansible-playbook deploy.yaml -e @deploy_config.yaml --tags docker-update
```

### Docker prune

```bash
ansible-playbook deploy.yaml -e @deploy_config.yaml --tags docker-prune
```

### Private repositories

Set `git_token` in `deploy_config.yaml` to a GitHub PAT or deploy token. The playbook automatically constructs the authenticated URL.

## Tests

```bash
go test ./...
```
