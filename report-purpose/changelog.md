Changelog
=========

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


Customer App / Frontend
-----------------------

1. Customer scan behavior
- Customer app scans QR payload.
- Customer app extracts merchant QRID and amount from scanned QR.
- Customer app sends:
  - `qr_payload`
  - `merchant_id`
  - `amount`

2. HTTPS experiment
- Temporary HTTPS dev-server setup for `customer-app` was added, then reverted.
- Current customer app Vite config is back to normal HTTP dev mode.


Docs
----

1. Added flow documentation
- `flow.txt`
- `flow-mermaid.md`
