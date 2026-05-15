# QRIS Latency Optimizer Flow

```mermaid
flowchart TD
    A[Backend Start] --> B[Load backend/.env]
    B --> C[Connect Postgres]
    C --> D[AutoMigrate merchants and transactions]
    D --> E[Seed default merchants]
    E --> F[Connect Redis]
    F --> G[Warm merchant cache]

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
    AL --> AM[Update status to SUCCESS in Postgres]
    AM --> AN[Delete Redis transaction cache]
    AN --> AO[Read updated transaction from Postgres]
    AO --> AP[Return SUCCESS transaction]

    AQ[Later Status Check] --> AE
```

## Notes

- Postgres is source of truth.
- Redis is cache layer for merchants and transactions.
- QRID like `TEST001` is QR payload merchant identifier.
- Merchant UUID is database primary key.
