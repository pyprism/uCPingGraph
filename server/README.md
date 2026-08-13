# uCPingGraph Server

## Security model

The read APIs (`/api/networks`, `/api/networks/:network/devices`, `/api/series`)
and the dashboard are unauthenticated, and CORS allows all origins
(`cors.Default()`). This is intentional for **LAN-only deployments**: any
device on the network can read topology and telemetry. If you expose this
server beyond a trusted LAN, put it behind your own auth and a real CORS
origin allowlist first — do not expose it directly to the internet as-is.

## Run

```bash
go run . server
```

Default URL: `http://127.0.0.1:8080`

## Environment Variables

Copy `.env.example` to `.env` and configure:

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
ansible-playbook deploy.yaml -e @deploy_config.yaml --tags fresh -k
```

### Update systemd deployment (pull + rebuild + restart)

```bash
ansible-playbook deploy.yaml -e @deploy_config.yaml --tags update -k
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

## Tests

```bash
go test ./...
```
