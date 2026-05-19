Changelog
=========

Branch Comparison: `testarea` vs `upstream/main`
================================================

Compared against `upstream/main` at `f53d78e`.

Summary
-------

- `testarea` is 2 commits ahead of `upstream/main`.
- Main branch work adds backend clean architecture refactor, RabbitMQ async payment confirmation, Redis cache hardening, and QRIS payload validation.

Backend Architecture
--------------------

1. Handler and routing refactor
- Replaced the older combined REST setup with explicit handlers for merchants, QRIS, transactions, and ping.
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
- Kept baseline `/api/transactions/:id/confirm-sync` route for synchronous DB confirmation.
- Added RabbitMQ publisher and payment consumer worker for async transaction status updates.
- Transaction cache is invalidated after status updates.

Caching
-------

1. Redis cache behavior
- Redis connection is now explicit during startup instead of hidden package initialization.
- Added merchant cache warm-up, QRID lookup, cache storage, and related merchant prefetch helpers.
- Transaction status lookup checks Redis first, falls back to Postgres, and removes corrupted cache payloads.

Tests and Docs
--------------

1. Tests added
- Added QRIS payload parser/validator tests.

2. Documentation updated
- Updated flow documentation for transaction behavior.
- Updated README with repo-root Docker usage.

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
  - RabbitMQ

10. Asynchronous payment confirmation with RabbitMQ
- Integrated the `github.com/rabbitmq/amqp091-go` message broker package.
- Added `backend/repository/rabbitmq/rabbitmq.go` to handle connections, channel establishment, retry logic (3 attempts with backoff), and publishing message payloads.
- Implemented a background payment consumer worker (`backend/worker/payment_consumer.go`) that consumes from the `payment_confirmations` queue, asynchronously updates the transaction status to `SUCCESS` in PostgreSQL, and invalidates the cached transaction payload in Redis.
- Updated payment confirmation endpoint routing:
  - Split into optimized asynchronous `/api/transactions/:id/confirm` (publishes message to RabbitMQ queue, returning instantly).
  - Maintained synchronous `/api/transactions/:id/confirm-sync` as a baseline reference.

11. Graceful shutdown and improved startup sequence
- Refactored `backend/cmd/main.go` to implement robust graceful shutdown handling using a signal listener (`SIGINT`, `SIGTERM`).
- The server now closes its RabbitMQ channel/connection cleanly and allows active HTTP requests to complete within a 5-second graceful timeout.
- Added clean connection/startup verification messages (`✓`) in terminal logs.

12. CORS and configuration updates
- Refactored CORS configuration in `backend/delivery/handler/cors.go` with dynamic allowed origins logic to support any development origin under ports `:5173` and `:5174` (allowing local LAN testing/IPs).
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


Docs
----

1. Added flow documentation
- `flow.txt`
- `flow-mermaid.md`
