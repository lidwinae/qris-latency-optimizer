# QRIS Latency Optimizer Flow

```mermaid
flowchart TD
    A[Backend Start] --> B[Load backend/.env]
    B --> C[Connect Postgres]
    C --> D[AutoMigrate merchants and transactions]
    D --> E[Seed default merchants]
    E --> F[Connect Redis]
    F --> G[Warm merchant cache]
    G --> G1[Connect RabbitMQ]
    G1 --> G2[Start payment consumer worker]
    G2 --> G3[Start HTTP server with latency middleware]

    H[Frontend Merchant Page] --> I[GET /api/merchants]
    I --> J[Query active merchants from Postgres]
    J --> K[Return merchant list]

    L[Frontend Generate QR] --> M[GET /api/qris]
    M --> N[Validate merchant UUID and amount]
    N --> O[Load merchant from Postgres]
    O --> P[Cache merchant in Redis]
    P --> Q[Prefetch related merchants]
    Q --> R[Generate QRIS payload]
    R --> S[Return qris_payload]

    T[Customer Scan QR] --> U[Extract QRID and amount from payload]
    U --> V[POST /api/transactions/scan]
    V --> W[Find merchant by UUID or QRID]
    W --> X[Check Redis merchant cache first]
    X --> Y[Fallback to Postgres if needed]
    Y --> Z[Validate QR CRC, merchant, amount]
    Z --> AA[Create PENDING transaction in Postgres]
    AA --> AB[Cache transaction in Redis]
    AB --> AC[Return transaction_id]

    AD[Customer Check Status] --> AE[GET /api/transactions/:id]
    AE --> AF[Check Redis transaction cache]
    AF -->|Hit| AG[Return cached transaction]
    AF -->|Miss| AH[Query Postgres]
    AH --> AI[Cache transaction in Redis]
    AI --> AJ[Return DB transaction]

    AK[Customer Confirm Payment] --> AL[POST /api/transactions/:id/confirm]
    AL --> AM[Publish transaction_id to RabbitMQ]
    AM --> AN[Return PROCESSING immediately]
    AM --> AO[Payment consumer reads queue]
    AO --> AP[Update status to SUCCESS in Postgres]
    AP --> AQ[Delete Redis transaction cache]

    AR[Baseline Confirm Payment] --> AS[POST /api/transactions/:id/confirm-sync]
    AS --> AT[Update status to SUCCESS in Postgres synchronously]
    AT --> AU[Delete Redis transaction cache]
    AU --> AV[Return SUCCESS transaction]

    AW[Later Status Check] --> AE
```

## Notes

- Postgres is source of truth.
- Redis is cache layer for merchants and transactions.
- QRID like `TEST001` is QR payload merchant identifier.
- Merchant UUID is database primary key.
- Optimized confirm returns `PROCESSING` and finishes through RabbitMQ worker.
- Baseline confirm-sync writes to Postgres before responding.

## Monitoring Flow

```mermaid
flowchart LR
    A[K6 optimized script] -->|POST /api/monitor/k6/data + summary| B[Backend in-memory K6 store]
    C[K6 non-optimized script] -->|POST /api/monitor/k6/data + summary| B
    B -->|GET /api/monitor/k6| D["/latency live dashboard"]

    E[Backend latency middleware] -->|GET /api/monitor/live| D

    A -->|--out influxdb=http://localhost:8086/k6| F[(InfluxDB k6 database)]
    C -->|--out influxdb=http://localhost:8086/k6| F
    F -->|InfluxQL queries| G[Grafana QRIS Performance & Latency dashboard]

    H[Docker Compose] --> I[Provision Grafana datasource]
    H --> J[Provision Grafana dashboard]
    I --> G
    J --> G
```

## Monitoring Notes

- `/latency` is live/in-memory and resets when backend restarts.
- Grafana is persisted through InfluxDB and `grafana_data` volume.
- Grafana datasource name is `QRIS K6 InfluxDB`.
- Grafana dashboard title is `QRIS Performance & Latency`.
- K6 scenario tags are `Event_Driven_Async` and `Synchronous_DB`.
