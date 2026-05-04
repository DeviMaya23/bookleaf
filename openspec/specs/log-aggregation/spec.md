# log-aggregation Specification

## Purpose
Defines the Loki and Promtail log aggregation infrastructure for the local development stack, including container log collection, structured log parsing, and bidirectional trace-log correlation in Grafana.

## Requirements

### Requirement: Loki Log Aggregation Service

The repository SHALL include a `loki/loki.yml` configuration file and a `loki` Docker Compose service. Loki SHALL run in single-process mode and store data on the container filesystem. It SHALL listen for log pushes on port `3100`.

#### Scenario: Loki receives logs from Promtail

- **WHEN** the local dev stack is running and the app processes a request
- **THEN** Loki contains a log entry for that request queryable via LogQL in Grafana Explore

### Requirement: Promtail Log Collection

The repository SHALL include a `promtail/promtail.yml` configuration file and a `promtail` Docker Compose service. Promtail SHALL:

- Mount the Docker socket (`/var/run/docker.sock`) to discover running containers
- Collect stdout/stderr from all containers in the Compose stack
- Apply a `container` label using the container name
- Parse each log line as JSON and extract `trace_id` as a Loki stream label
- Push collected logs to `http://loki:3100/loki/api/v1/push`

#### Scenario: Structured log fields are queryable in Loki

- **WHEN** the app emits a JSON log line containing `trace_id`
- **THEN** `trace_id` is available as a queryable label in Loki
- **AND** the log line is retrievable via `{container="bookleaf-app-1"} | json | trace_id="<id>"`

### Requirement: Bidirectional Trace-Log Correlation in Grafana

Grafana SHALL be provisioned with a `tempo.yml` datasource file and a `loki.yml` datasource file that configure bidirectional trace-log linking using `trace_id` as the correlation key.

The Tempo datasource SHALL include a `tracesToLogsV2` block pointing at the Loki datasource, filtering by the `trace_id` label.

The Loki datasource SHALL include a `derivedFields` block that matches the `trace_id` JSON field in log lines and creates a link to the Tempo datasource.

#### Scenario: Trace links to correlated logs

- **WHEN** a trace is open in Grafana's Tempo explore view
- **THEN** a "Logs" link is visible that navigates to Loki filtered to the matching `trace_id`

#### Scenario: Log line links to its trace

- **WHEN** a log line is open in Grafana's Loki explore view and contains a `trace_id` field
- **THEN** a link is visible that navigates to the matching Tempo trace
