# QRIS Latency Optimizer 🚀

Full-stack QRIS payment simulation with:
- Go backend
- merchant dashboard
- customer scanner app
- Postgres for source of truth
- Redis for cache and prefetch
- monitoring tools for load testing

## Project Structure

- `backend/`
  - Go API with Gin
  - QR generation
  - transaction lifecycle
  - Redis cache and merchant prefetch
- `frontend/`
  - merchant dashboard
  - React + Vite
  - default port `5173`
- `customer-app/`
  - customer QR scanner app
  - React + Vite
  - default port `5174`

## Stack

- Go + Gin
- PostgreSQL
- Redis
- RedisInsight
- pgAdmin
- InfluxDB
- Grafana

## Current Architecture Notes

- Postgres is source of truth.
- Redis is optional acceleration layer.
- RabbitMQ is async processing layer for optimized payment confirmation.
- Merchant data is seeded from Go startup, not SQL init file.
- Backend auto-creates DB schema with GORM `AutoMigrate`.
- Merchant cache is warmed into Redis on backend startup.
- Transaction status uses cache-aside pattern:
  - Redis first
  - Postgres fallback
  - cache repopulated after DB read

## How To Run

### Prerequisites

Need:
- Docker / Docker Desktop running
- Go installed
- Node.js installed
- k6 installed for load testing

## 1. Start Infrastructure

Run Docker services from repo root:

```bash
docker compose up -d
```

This starts:
- Postgres
- Redis
- RedisInsight
- pgAdmin
- RabbitMQ
- InfluxDB
- Grafana

## 2. Backend Setup

Start backend:

```bash
cd backend
go run ./cmd
```

Backend runs on:

```text
http://localhost:8080
```

## 3. Merchant Dashboard

```bash
cd frontend
npm install
npm run dev
```

Frontend runs on:

```text
http://localhost:5173
```

## 4. Customer App

```bash
cd customer-app
npm install
npm run dev
```

Customer app runs on:

```text
http://localhost:5174
```

## Docker Web Tools

### pgAdmin

Open:

```text
http://localhost:5050
```


### RedisInsight

Open:

```text
http://localhost:5540
```

Connect to Redis with:

```text
Host: redis
Port: 6379
```

### Grafana

Open:

```text
http://localhost:3000
```

Ready after startup:
- datasource: `QRIS K6 InfluxDB`
- dashboard folder: `QRIS`
- dashboard: `QRIS Performance & Latency`

### InfluxDB

Open:

```text
http://localhost:8086
```

## Load Test Monitoring

Start backend first:

```bash
cd backend
go run cmd/main.go
```

Run k6 with InfluxDB output so Grafana gets native k6 metrics:

```bash
k6 run --out influxdb=http://localhost:8086/k6 tests_script/06-optimized-dashboard.js
k6 run --out influxdb=http://localhost:8086/k6 tests_script/07-non-optimized-dashboard.js
```

What each dashboard uses:
- Grafana reads persisted k6 metrics from InfluxDB
- `http://localhost:8080/latency` reads comparison/live summary from backend monitor API

If Grafana shows no data:
- confirm backend is running on `http://localhost:8080`
- confirm k6 command includes `--out influxdb=http://localhost:8086/k6`
- confirm test finished at least one request successfully

## Main Backend Flow

### Startup

Backend startup does:
- load `backend/.env`
- connect Postgres
- create `pgcrypto` extension if needed
- auto-migrate tables
- seed default merchants
- connect Redis
- warm merchant cache
- connect RabbitMQ
- start payment consumer worker
- start HTTP server with latency middleware

### Merchant Flow

Endpoint:

```text
GET /api/merchants
```

Returns active merchants from Postgres.

### Generate QRIS

Endpoint:

```text
GET /api/qris?merchant_id=<merchant_uuid>&amount=<amount>
```

Flow:
- validate amount
- load merchant by UUID
- generate QRIS payload from merchant QRID, merchant name, and amount
- cache merchant in Redis
- prefetch related merchants

### Customer Scan

Endpoint:

```text
POST /api/transactions/scan
```

Request body:

```json
{
  "qr_payload": "<qris_payload>",
  "merchant_id": "TEST001",
  "amount": 1000
}
```

Flow:
- customer app scans QR
- extracts merchant QRID and amount from payload
- sends payload to backend
- backend accepts merchant ID as UUID or QRID like `TEST001`
- backend validates:
  - merchant exists and active
  - QR CRC valid
  - QR merchant matches request merchant
  - QR amount matches request amount
- backend creates `PENDING` transaction
- backend caches transaction in Redis

### Check Transaction Status

Endpoint:

```text
GET /api/transactions/:id
```

Flow:
- validate UUID
- check Redis key `transaction:<id>`
- if hit, return cached data
- if miss, query Postgres
- cache fresh transaction result

### Confirm Payment

Endpoint:

```text
POST /api/transactions/:id/confirm
```

Flow:
- validate UUID
- publish confirmation event to RabbitMQ
- return transaction with `PROCESSING` status
- worker updates transaction to `SUCCESS`
- worker deletes old transaction cache

### Baseline Confirm Payment

Endpoint:

```text
POST /api/transactions/:id/confirm-sync
```

Flow:
- validate UUID
- update transaction to `SUCCESS` directly in Postgres
- delete old transaction cache
- return updated transaction

## Monitoring and Dashboards

### Live backend dashboards

- `http://localhost:8080/monitor`
  - system metrics
  - service status
  - load test overview
- `http://localhost:8080/latency`
  - live endpoint latency charts
  - optimized vs non-optimized comparison
  - K6 summary data from backend monitor API

### Monitoring APIs

```text
GET    /api/monitor/system
GET    /api/monitor/live
GET    /api/monitor/k6
POST   /api/monitor/k6/data
POST   /api/monitor/k6/summary
DELETE /api/monitor/k6
```

### Grafana + InfluxDB

- Grafana stores dashboards in `grafana_data` Docker volume.
- InfluxDB stores persisted K6 metrics.
- Grafana datasource is provisioned automatically as `QRIS K6 InfluxDB`.
- Grafana dashboard is provisioned automatically as `QRIS Performance & Latency`.
- `/latency` is live in-memory data from backend; Grafana is persisted K6 history from InfluxDB.

## Redis Usage

### Transaction Cache

Used for:
- repeated transaction status polling
- lower DB load
- faster response

Redis key format:

```text
transaction:<transaction_id>
```

### Merchant Cache

Used for:
- QRID-based merchant lookup
- startup warm cache
- speculative related-merchant prefetch

Redis key format:

```text
merchant:<qr_id>
```

If Redis is down:
- backend still works
- cache reads miss
- cache writes are skipped
- Postgres remains source of truth

## Important Identifiers

Merchant has two identifiers:

- `ID`
  - UUID primary key
  - used internally in backend routes
- `QRID`
  - QR merchant code like `TEST001`
  - stored in `qr_id`
  - placed into QRIS payload tag `26.01`

## Main API Routes

```text
GET  /api/ping
GET  /api/merchants
GET  /api/qris?merchant_id=<merchant_uuid>&amount=<amount>
POST /api/transactions/scan
GET  /api/transactions/:id
POST /api/transactions/:id/confirm
POST /api/transactions/:id/confirm-sync
GET  /monitor
GET  /latency
GET  /api/monitor/system
GET  /api/monitor/live
GET  /api/monitor/k6
POST /api/monitor/k6/data
POST /api/monitor/k6/summary
DELETE /api/monitor/k6
```

## Testing Quick Examples

### Check transaction status

```bash
curl http://localhost:8080/api/transactions/<transaction_id>
```

### Confirm payment

```bash
curl -X POST http://localhost:8080/api/transactions/<transaction_id>/confirm
```

### Confirm payment baseline

```bash
curl -X POST http://localhost:8080/api/transactions/<transaction_id>/confirm-sync
```

### Run Grafana-ready K6 comparison

```bash
./tests_script/run-grafana-tests.sh
```

## Extra Docs

- `report-purpose/flow.txt`
- `report-purpose/flow-mermaid.md`
- `report-purpose/changelog.md`

## Notes For Phone Testing

Customer app camera on phone may fail on plain LAN HTTP because browser camera access often requires secure origin.

If camera does not open:
- check browser permission
- try Chrome/Edge on Android or Safari on iPhone
- if testing from phone over LAN, browser security may block camera on plain `http://<ip>:5174`
