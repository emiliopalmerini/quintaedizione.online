# ADR-021: Observability Stack (Prometheus + Grafana + Uptime Kuma)

## Status

Accepted

## Context

The app runs on a private VPS managed by Dokploy. Dokploy provides basic container-level metrics (CPU, RAM, disk, I/O) but we lack:

- **Application-level metrics**: request rates, latency distribution, error rates, search volume
- **Uptime alerting**: external notification when the app goes down
- **Dashboards**: historical view of app health and performance

The app already has a basic `GET /health` endpoint returning static JSON, but it doesn't verify actual readiness and can't be used for meaningful health monitoring.

## Decision

### 1. Enhance health endpoint

Upgrade `GET /health` to verify the in-memory store is loaded (check that at least one collection has documents). Return HTTP 200 with `"status": "healthy"` when OK, HTTP 503 with `"status": "unhealthy"` when the store is empty or not ready.

### 2. Prometheus instrumentation

Add `prometheus/client_golang` and instrument the app with:

- **HTTP metrics middleware** (applied to the main mux):
  - `http_requests_total` (counter) — labels: `method`, `path`, `status`
  - `http_request_duration_seconds` (histogram) — labels: `method`, `path`
  - `http_requests_in_flight` (gauge)
- **Search metrics** (instrumented in search handler):
  - `search_queries_total` (counter) — labels: `type` (global, collection, dropdown)
- **Expose `GET /metrics`** — registered on a separate internal mux (not the public mux) so it's only reachable from the Docker network, not from the internet.

**Internal metrics server**: A second `http.Server` listening on port `9090` (configurable via `METRICS_PORT` env var) serves only `/metrics`. Since Dokploy only exposes the app port (8000) to the internet, the metrics port stays internal to the Docker network.

### 3. Monitoring stack (Dokploy services)

Deploy as separate Dokploy services on the same VPS:

- **Prometheus**: scrapes the app's internal `:9090/metrics` endpoint. Configured via `prometheus.yml` with a scrape target pointing to the app container by Docker network hostname.
- **Grafana**: connects to Prometheus as datasource. Pre-configured with a dashboard for HTTP request rate, latency percentiles, error rate, and search volume.
- **Uptime Kuma**: monitors the public `https://quintaedizione.online/health` endpoint. Sends alerts via Telegram bot when the site goes down.

### 4. Path label normalization

To avoid high cardinality in Prometheus labels, normalize URL paths:
- `/srd/{collection}/{source}/{slug}` → `/srd/{collection}/{source}/{slug}` (parameterized)
- `/mappe/{slug}` → `/mappe/{slug}`
- `/generatori/{slug}` → `/generatori/{slug}`
- Static paths (`/`, `/health`, `/srd/search`) remain as-is

## Inputs

- Docker network connectivity between containers (same Dokploy project or shared network)
- `METRICS_PORT` env var (default: `9090`)
- Telegram bot token and chat ID for Uptime Kuma alerts

## Outputs

- `GET /health` — enhanced readiness check (public, port 8000)
- `GET /metrics` — Prometheus metrics (internal only, port 9090)
- Grafana dashboard accessible via Dokploy proxy
- Uptime Kuma monitoring with Telegram alerts

## Edge Cases

- **App startup**: metrics server starts alongside main server; `/health` returns 503 until store is loaded
- **Metrics port conflict**: configurable via env var to avoid clashes with other services
- **Path cardinality explosion**: parameterized path labels prevent unbounded label growth
- **Graceful shutdown**: both HTTP servers (app + metrics) shut down on SIGTERM

## Consequences

- Adds `prometheus/client_golang` dependency
- Adds ~50 lines of middleware code and ~20 lines for the internal metrics server
- Three new Dokploy services to maintain (Prometheus, Grafana, Uptime Kuma)
- Minimal performance overhead from metrics collection (histogram observation per request)
