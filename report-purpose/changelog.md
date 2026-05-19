Changelog
=========

Branch Comparison: `testarea` vs `upstream/main`
================================================

Compared against `upstream/main` at `f53d78e`.

Summary
-------

- `testarea` is 2 commits ahead of `upstream/main`.
- Main branch work adds backend clean architecture refactor, RabbitMQ async payment confirmation, Redis cache hardening, QRIS payload validation, monitoring dashboards, and K6 load tests.
- Current working tree additionally wires Grafana to real K6/InfluxDB metrics with provisioning and updated run commands.

Backend Architecture
--------------------

1. Handler and routing refactor
- Replaced the older combined REST setup with explicit handlers for merchants, QRIS, transactions, ping, and monitoring.
- Added a central router setup that registers route groups and middleware.
- Moved CORS from handler package into middleware package.

2. Domain and repository layering
- Moved models into `domain/entity`.
- Added repository interfaces under `domain/repository`.
- Split Postgres implementation into merchant and transaction repositories.
- Removed older database package files in favor of explicit config and repository startup.

3. Configuration and startup
- Added centralized config loading.
- Moved `.env_example` from backend-local path to repo root.
- Added root-level `docker-compose.yml` and removed backend-local compose file.
- Backend startup now connects Postgres, Redis, warms Redis cache, connects RabbitMQ, starts payment consumer, and shuts down gracefully.

Payment Flow
------------

1. QRIS and transaction validation
- QRIS payload generation now uses merchant database data and request amount.
- Added payload parsing, CRC validation, merchant QRID extraction, and amount validation.
- Transaction scan accepts UUID or QRID merchant identifiers and validates merchant/activity before creating transaction.

2. Optimized vs non-optimized confirmation
- Added optimized `/api/transactions/:id/confirm` route that publishes confirmation work to RabbitMQ and returns quickly with `PROCESSING`.
- Kept baseline `/api/transactions/:id/confirm-sync` route for synchronous DB confirmation and latency comparison.
- Added RabbitMQ publisher and payment consumer worker for async transaction status updates.
- Transaction cache is invalidated after status updates.

Caching
-------

1. Redis cache behavior
- Redis connection is now explicit during startup instead of hidden package initialization.
- Added merchant cache warm-up, QRID lookup, cache storage, and related merchant prefetch helpers.
- Transaction status lookup checks Redis first, falls back to Postgres, and removes corrupted cache payloads.

Monitoring and Load Testing
---------------------------

1. Backend monitoring API
- Added `/api/monitor/system`, `/api/monitor/live`, `/api/monitor/k6`, `/api/monitor/k6/data`, and `/api/monitor/k6/summary`.
- Added latency tracking middleware for API request duration and error capture.
- Added in-memory K6 aggregation for live dashboard comparisons.

2. Built-in monitoring pages
- Added `/monitor` for live system/service metrics.
- Added `/latency` for optimized vs non-optimized payment latency comparison.
- Updated `/latency` run instructions to use K6 InfluxDB output so Grafana receives real metrics.

3. K6 suite
- Added polling, optimized confirmation, synchronous baseline, and dashboard-oriented K6 scenarios under `tests_script/`.
- Added branch dashboard export `tests_script/dashboard-1778674676803.json`.
- Current working tree adds `tests_script/run-grafana-tests.sh` helper for running optimized and non-optimized tests with InfluxDB output.

4. Grafana and InfluxDB wiring
- Docker Compose now starts InfluxDB and Grafana from repo root.
- Current working tree provisions Grafana datasource `QRIS K6 InfluxDB`.
- Current working tree provisions dashboard `QRIS Performance & Latency`.
- Grafana now reads persisted K6 metrics from InfluxDB; `/latency` still reads live comparison data from backend monitor API.

Tests and Docs
--------------

1. Tests added
- Added QRIS payload parser/validator tests.
- Added monitoring aggregation tests.

2. Documentation updated
- Updated flow documentation for monitoring and transaction behavior.
- Updated README with repo-root Docker usage, Grafana details, and K6 commands using `--out influxdb=http://localhost:8086/k6`.

Backend
-------

1. QRIS payload generation
- Replaced mostly static QR payload generation with input-based generation.
- Payload now uses:
  - merchant name from database
  - merchant QRID from database
  - amount from request
- Added QRIS CRC generation.
- Default city set to MALANG.
- Added QR payload parsing and validation:
  - parse merchant QRID from tag 26.01
  - parse amount from tag 54
  - validate CRC

2. Merchant identifier cleanup
- Clarified merchant identifiers:
  - `ID` = UUID primary key
  - `QRID` = merchant QR identifier, stored in `qr_id`
- Renamed merchant field from confusing `QRIS` to `QRID`.

3. Transaction scan flow hardening
- Added validation for `merchant_id`.
- Allowed `merchant_id` in scan request to be:
  - UUID
  - QRID like `TEST001`
- Added merchant active/existence validation before transaction create.
- Added QR payload validation in scan flow:
  - merchant in QR must match selected merchant
  - amount in QR must match request amount
- Removed panic-prone `uuid.MustParse` usage in scan flow.

4. Transaction status and confirm flow hardening
- Transaction status now:
  - checks Redis first
  - falls back to Postgres
  - deletes corrupted cached transaction payload
- Confirm payment now:
  - checks rows affected
  - returns not found if transaction does not exist
  - reloads updated transaction safely

5. Redis changes
- Redis startup changed from implicit package `init()` to explicit startup call.
- Added merchant cache helpers:
  - `WarmUpCache()`
  - `PrefetchMerchant()`
  - `PrefetchRelatedMerchants()`
  - `GetMerchant()`
  - `CacheMerchant()`
- Warm-up of merchant cache now runs at backend startup.
- Transaction cache uses shared TTL constant.
- Merchant cache used in customer scan flow for QRID lookup.
- Qris generation flow now caches merchant and triggers related merchant prefetch.

6. Database bootstrap
- Removed SQL init dependency from Docker startup.
- Deleted old `backend/init-db/init.sql`.
- Database creation now handled from Go:
  - `pgcrypto` extension creation
  - `AutoMigrate`
  - default merchant seed
- Added Go seed for:
  - `TEST001` / `Kantin FILKOM UB`
  - `TEST002` / `TESTING STORE`

7. Handler/service organization
- Merged server-side transaction status handler into `backend/usecase/service/qris.go`.
- Deleted old separate `backend/usecase/service/transaction.go`.

8. CORS
- Replaced unsafe wildcard CORS setup.
- Added env-based allowed origins using `CORS_ALLOWED_ORIGINS`.
- Safer headers/method configuration.

9. Environment and Compose
- Repo layout shifted to backend-local setup:
  - `backend/.env`
  - `backend/.env_example`
  - `backend/docker-compose.yml`
- LoadEnv currently expects `.env` in backend working directory.
- Compose now includes:
  - Postgres
  - Redis
  - RedisInsight
  - PgAdmin
  - InfluxDB
  - Grafana

10. Asynchronous payment confirmation with RabbitMQ
- Integrated the `github.com/rabbitmq/amqp091-go` message broker package.
- Added `backend/repository/rabbitmq/rabbitmq.go` to handle connections, channel establishment, retry logic (3 attempts with backoff), and publishing message payloads.
- Implemented a background payment consumer worker (`backend/worker/payment_consumer.go`) that consumes from the `payment_confirmations` queue, asynchronously updates the transaction status to `SUCCESS` in PostgreSQL, and invalidates the cached transaction payload in Redis.
- Updated payment confirmation endpoint routing:
  - Split into optimized asynchronous `/api/transactions/:id/confirm` (publishes message to RabbitMQ queue, returning instantly).
  - Maintained synchronous `/api/transactions/:id/confirm-sync` as a baseline reference for direct load-test comparison.

11. Graceful shutdown and improved startup sequence
- Refactored `backend/cmd/main.go` to implement robust graceful shutdown handling using a signal listener (`SIGINT`, `SIGTERM`).
- The server now closes its RabbitMQ channel/connection cleanly and allows active HTTP requests to complete within a 5-second graceful timeout.
- Added clean connection/startup verification messages (`✓`) in terminal logs.

12. CORS and configuration updates
- Refactored CORS configuration in `backend/delivery/handler/cors.go` with dynamic allowed origins logic to support any development origin under ports `:5173` and `:5174` (allowing local LAN testing/IPs) and the monitoring dashboard served on `:8080`.
- Added fallback default value for `RABBITMQ_URL` env variable in `backend/.env_example`.
- Shifted default PostgreSQL timezone setting from `Asia/Shanghai` to `Asia/Jakarta` in `backend/repository/database/pg.go`.


Customer App / Frontend
-----------------------

1. Customer scan behavior
- Customer app scans QR payload.
- Customer app extracts merchant QRID and amount from scanned QR.
- Customer app sends:
  - `qr_payload`
  - `merchant_id`
  - `amount`


DevOps, Monitoring & Load Testing
---------------------------------

1. Real-time monitoring dashboard
- Built a highly visual and responsive dual-dashboard frontend under `backend/monitoring/`:
  - `/monitor` (`index.html`): Real-time system monitoring (CPU, RAM, Go runtime, Goroutine count, queue status) and K6 load-test visualizer.
  - `/latency` (`latency.html`): Real-time endpoint request duration histograms and moving average graphs comparing optimized vs non-optimized routes.
- Served dashboards directly from Go/Gin static routes.
- Added a custom latency tracker middleware (`backend/usecase/service/latency_tracker.go`) to transparently capture stats for all API requests.
- Added system and metrics collection REST APIs:
  - `/api/monitor/system` (resource usage, service status).
  - `/api/monitor/live` (live endpoint stats, average/p95 latency, error count).
  - `/api/monitor/k6` / `/api/monitor/k6/data` / `/api/monitor/k6/summary` (K6 testing live ingestion endpoint).

2. Load testing suite (K6 integration)
- Added new K6 load test scenarios under `tests_script/`:
  - `03-polling-test.js`: Simulates clients aggressively checking transaction statuses.
  - `04-payment-confirm.js`: Load tests the optimized, asynchronous, RabbitMQ-backed `/api/transactions/:id/confirm` endpoint.
  - `05-payment-confirm-lama.js`: Load tests the unoptimized, synchronous, DB-blocking `/api/transactions/:id/confirm-sync` endpoint.
  - `06-optimized-dashboard.js` & `07-non-optimized-dashboard.js`: Test configurations matching dashboard visualizations.
- Added unit tests for service layer monitoring (`monitor_test.go`) and payload verification logic (`payload_test.go`).
- Added JSON configuration for Grafana / K6 optimization dashboards (`dashboard-1778674676803.json`).

3. Infrastructure components in Docker
- Re-enabled RabbitMQ container service in `backend/docker-compose.yml` (`guest:guest` auth, standard data port `5672`, and management dashboard on `15672`).
- Preconfigured InfluxDB and Grafana services for load test metrics storage and visual charting.


Docs
----

1. Added flow documentation
- `flow.txt`
- `flow-mermaid.md`
