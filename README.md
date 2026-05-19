# QRIS Latency Optimizer 🚀

Full-stack QRIS payment simulation with:
- Go backend
- merchant dashboard
- customer scanner app
- Postgres for source of truth
- Redis for cache and prefetch
- RabbitMQ for async payment processing

## Project Structure

- `backend/`
  - Go API with Gin
  - QR generation
  - transaction lifecycle
  - Redis cache and merchant prefetch
  - RabbitMQ async payment confirmation
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
- RabbitMQ
- pgAdmin

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

### RabbitMQ Management

Open:

```text
http://localhost:15672
```

Default credentials: `guest` / `guest`

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
- start HTTP server

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

### Confirm Payment (Optimized - Async)

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

### Confirm Payment (Baseline - Sync)

Endpoint:

```text
POST /api/transactions/:id/confirm-sync
```

Flow:
- validate UUID
- update transaction to `SUCCESS` directly in Postgres
- delete old transaction cache
- return updated transaction

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

## API Routes

```text
GET  /api/ping
GET  /api/merchants
GET  /api/qris?merchant_id=<merchant_uuid>&amount=<amount>
POST /api/transactions/scan
GET  /api/transactions/:id
POST /api/transactions/:id/confirm
POST /api/transactions/:id/confirm-sync
```

## Testing Quick Examples

### Check transaction status

```bash
curl http://localhost:8080/api/transactions/<transaction_id>
```

### Confirm payment (async)

```bash
curl -X POST http://localhost:8080/api/transactions/<transaction_id>/confirm
```

### Confirm payment (sync baseline)

```bash
curl -X POST http://localhost:8080/api/transactions/<transaction_id>/confirm-sync
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
