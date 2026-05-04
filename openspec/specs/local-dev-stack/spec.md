# local-dev-stack Specification

## Purpose
Defines the Docker Compose local development stack including Tempo, Loki, Promtail, Prometheus, and Grafana services, along with their configuration and provisioning files.

## Requirements

### Requirement: Docker Compose Local Development Stack

The repository SHALL contain a `docker-compose.yml` at the repo root defining the following services:

- **`app`** — built from `./backend`; depends on `tempo`, `prometheus`, `loki`; env vars: `OTEL_EXPORTER=tempo`, `OTEL_METRICS_EXPORTER=prometheus`, `LOG_FORMAT=json`, `OTEL_TEMPO_ENDPOINT=tempo:4317`, plus all other required app env vars loaded from a `.env` file; exposes port `8080`; `extra_hosts: ["host.docker.internal:host-gateway"]` for Linux compatibility
- **`tempo`** — image `grafana/tempo:latest`; mounts `./tempo/tempo.yml` as its config; exposes `4317` (OTLP gRPC receiver); runs in single-binary mode
- **`prometheus`** — image `prom/prometheus:latest`; mounts `./prometheus/prometheus.yml` as its config; scrapes the app's `/metrics` endpoint; exposes port `9090`
- **`loki`** — image `grafana/loki:latest`; mounts `./loki/loki.yml` as its config; exposes port `3100`
- **`promtail`** — image `grafana/promtail:latest`; mounts `./promtail/promtail.yml` as its config and `/var/run/docker.sock` (read-only) for container log discovery; depends on `loki`
- **`grafana`** — image `grafana/grafana:latest`; exposes port `3000`; provisioned with Tempo, Loki, and Prometheus datasources

The PostgreSQL database SHALL NOT be a Docker service; the `app` container reaches it via `host.docker.internal`.

The `jaeger` service is removed entirely.

#### Scenario: Developer starts the local stack

- **WHEN** `docker compose up` is run from the repo root with a valid `.env` file
- **THEN** the `app`, `tempo`, `prometheus`, `loki`, `promtail`, and `grafana` containers start successfully
- **AND** the app is reachable at `http://localhost:8080`
- **AND** Grafana UI is reachable at `http://localhost:3000`
- **AND** Prometheus UI is reachable at `http://localhost:9090`

#### Scenario: App reaches native PostgreSQL

- **WHEN** `DATABASE_URL` in `.env` uses `host.docker.internal` as the host
- **THEN** the app container connects to the natively-running PostgreSQL instance successfully

### Requirement: Prometheus Scrape Configuration

The repository SHALL include `prometheus/prometheus.yml` that configures Prometheus to scrape the app's `/metrics` endpoint. The scrape target SHALL use `host.docker.internal:8080` so Prometheus running in Docker can reach the app container's exposed port.

#### Scenario: Prometheus scrapes app metrics

- **WHEN** Prometheus starts via Docker Compose
- **THEN** it scrapes `http://app:8080/metrics` (using the Docker Compose service name) at the configured interval
- **AND** the app's HTTP request duration histograms and active request counter are queryable in Prometheus

### Requirement: Grafana Datasource Provisioning

The repository SHALL include Grafana provisioning files under `grafana/provisioning/datasources/`:

- `tempo.yml` — Tempo datasource pointing at `http://tempo:3200`; includes `tracesToLogsV2` config linking to the Loki datasource via `trace_id`
- `loki.yml` — Loki datasource pointing at `http://loki:3100`; includes `derivedFields` config linking `trace_id` to the Tempo datasource
- `prometheus.yml` — Prometheus datasource pointing at `http://prometheus:9090`

`jaeger.yml` is removed. All three datasources SHALL be available in Grafana without any manual configuration after startup.

#### Scenario: Grafana starts with all datasources pre-configured

- **WHEN** Grafana starts via Docker Compose
- **THEN** the Tempo, Loki, and Prometheus datasources are available without manual setup
