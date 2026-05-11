# QRIS Latency Optimizer 🚀

This project is a full-stack QRIS payment system designed to handle extremely low-latency API responses. It implements a cache-aside architecture using Redis to optimize transaction status polling and a complete monitoring stack to analyze system performance.

## 📂 Project Structure

This repository is organized as a monorepo containing the backend, merchant UI, and customer UI:

- **`/backend`**: The Go backend (Gin framework). Handles QR generation, transaction lifecycle, and caching logic.
- **`/frontend`**: The **Merchant Dashboard** UI (React + Vite). Runs on port 5173.
- **`/customer-app`**: The **Customer Mobile** UI (React + Vite). Runs on port 5174.

## 🛠️ Tech Stack & Infrastructure

The system runs several core services via Docker:

- **PostgreSQL**: Primary persistent storage for transaction data.
- **Redis**: Caching layer for high-speed transaction status checks (Key to P95 latency optimization).
- **RabbitMQ**: Message broker for handling asynchronous tasks and background processing.
- **pgAdmin**: Web interface for managing the PostgreSQL database (Port 5050).
- **Monitoring Stack**: 
  - **InfluxDB**: Time-series database to store load test metrics from k6.
  - **Grafana**: Dashboard for real-time visualization of system metrics and latency (Port 3000).

---

## 🚀 How to Run

### 1. Prerequisites
Before starting, ensure that **Docker Desktop** or the **Docker Engine** is already running on your machine. You will also need **Node.js** and **Go** installed locally for development.

### 2. Start the Infrastructure
Navigate to the root project folder and run the following command to start all databases and monitoring tools:
```bash
docker-compose up -d
```

**2. Start the Backend API**
Open a terminal and run:
```bash
cd backend
# Ensure your .env is configured (DB_HOST=localhost)
go run cmd/main.go
```
*(The backend runs on http://localhost:8080)*

**3. Start the Frontend UI / Merchant App**
Open a new terminal and run:
```bash
cd frontend
npm install   # Required only for the first time
npm run dev
```
*(The merchant dashboard runs on http://localhost:5173)*

**4. Start the Customer App**
Open a new terminal and run:
```bash
cd customer-app
npm install   # Required only for the first time
npm run dev
```
*(The customer app runs on http://localhost:5174)*

## 📚 Architectural Details (Clean Architecture)

The backend follows Clean Architecture principles:
- **`usecase/customer`**: Contains endpoints mimicking customer actions (e.g., scanning the QR code, simulating payment confirmation).
- **`usecase/service`**: Contains endpoints for the merchant backend (e.g., generating the dynamic QR code string, checking transaction status).
- **Latency Optimization**: The `GetTransactionStatus` API queries Redis first. If there's a cache hit, it returns immediately. On a cache miss, it fetches from PostgreSQL and re-populates Redis. When a payment is confirmed, the Redis cache is instantly invalidated to prevent stale data.
