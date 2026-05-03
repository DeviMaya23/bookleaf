# local-dev-stack Specification

## Purpose
Defines the Docker Compose local development stack including Jaeger, Prometheus, and Grafana services, along with their configuration and provisioning files.

## Requirements

### Requirement: Docker Compose Local Development Stack

The repository SHALL contain a `docker-compose.yml` at the repo root defining the following services:

- **`app`** — built from `./backend`; depends on `jaeger` and `prometheus`; env vars: `OTEL_EXPORTER=jaeger`, `OTEL_METRICS_EXPORTER=prometheus`, `LOG_FORMAT=console`, `OTEL_JAEGER_ENDPOINT=jaeger:4317`, plus all other required app env vars loaded from a `.env` file; exposes port `8080`; `extra_hosts: ["host.docker.internal:host-gateway"]` for Linux compatibility
- **`jaeger`** — image `jaegertracing/all-in-one:latest`; exposes `16686` (UI), `4317` (OTLP gRPC)
- **`prometheus`** — image `prom/prometheus:latest`; mounts `./prometheus/prometheus.yml` as its config; scrapes the app's `/metrics` endpoint; exposes port `9090`
- **`grafana`** — image `grafana/grafana:latest`; exposes port `3000`; provisioned with a Jaeger datasource and a Prometheus datasource

The PostgreSQL database SHALL NOT be a Docker service; the `app` container reaches it via `host.docker.internal`.

#### Scenario: Developer starts the local stack

- **WHEN** `docker compose up` is run from the repo root with a valid `.env` file
- **THEN** the `app`, `jaeger`, `prometheus`, and `grafana` containers start successfully
- **AND** the app is reachable at `http://localhost:8080`
- **AND** Jaeger UI is reachable at `http://localhost:16686`
- **AND** Prometheus UI is reachable at `http://localhost:9090`
- **AND** Grafana UI is reachable at `http://localhost:3000`

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

- `jaeger.yml` — Jaeger datasource pointing at `http://jaeger:16686`
- `prometheus.yml` — Prometheus datasource pointing at `http://prometheus:9090`

Both datasources SHALL be available in Grafana without any manual configuration after startup.

#### Scenario: Grafana starts with all datasources pre-configured

- **WHEN** Grafana starts via Docker Compose
- **THEN** both the Jaeger and Prometheus datasources are available without manual setup
