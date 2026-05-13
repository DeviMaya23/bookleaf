### Requirement: HTTP Request Count Metric

The system SHALL record an `http.server.request.count` Int64Counter instrument in `MetricsMiddleware` for every completed HTTP request. The counter SHALL be incremented once per request inside the existing `defer` block, after the status code is resolved. Attributes SHALL match those on `http.server.request.duration`: `http.request.method`, `http.route`, and `http.response.status_code`.

#### Scenario: Request count incremented on success

- **WHEN** an HTTP request completes with a `2xx` status code
- **THEN** `http.server.request.count` is incremented by 1 with the correct method, route, and status code attributes

#### Scenario: Request count incremented on error

- **WHEN** an HTTP request completes with a `4xx` or `5xx` status code
- **THEN** `http.server.request.count` is incremented by 1 with the actual status code as an attribute

### Requirement: HTTP Error Rate Metrics

The system SHALL record an `http.server.request.errors` Int64Counter instrument in `MetricsMiddleware`. The counter SHALL be incremented only when the resolved status code is `4xx` or `5xx`. The counter SHALL carry a `http.status_class` attribute with value `"4xx"` for status codes 400–499 and `"5xx"` for status codes 500–599, in addition to `http.request.method` and `http.route`.

#### Scenario: 4xx error increments error counter with correct class

- **WHEN** a handler returns a `4xx` status code
- **THEN** `http.server.request.errors` is incremented by 1
- **AND** the `http.status_class` attribute is `"4xx"`

#### Scenario: 5xx error increments error counter with correct class

- **WHEN** a handler returns a `5xx` status code
- **THEN** `http.server.request.errors` is incremented by 1
- **AND** the `http.status_class` attribute is `"5xx"`

#### Scenario: Successful request does not increment error counter

- **WHEN** a handler returns a `2xx` status code
- **THEN** `http.server.request.errors` is NOT incremented
