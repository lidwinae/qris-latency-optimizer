# QRIS Transaction System - Legacy Version (Baseline)

This legacy app is used for comparing the performance of a standard, unoptimized transaction system against the QRIS Latency Optimizer (the optimized version). It serves as a baseline to demonstrate how a system behaves when every transaction and inquiry request hits the primary database (PostgreSQL) directly without a caching layer.

## 📂 Project Structure

This repository is organized as a monorepo containing the backend, merchant UI, and customer UI:

- **`/backend`**: The Go backend (Gin framework). Handles QR generation and transaction.
- **`/frontend`**: The **Merchant Dashboard** UI (React + Vite). Runs on port 5173.
- **`/customer-app`**: The **Customer Mobile** UI (React + Vite). Runs on port 5174.

## 🚀 How to Run

### Prerequisites
Before starting, ensure that **Docker Desktop** or the **Docker Engine** is already running on your machine. You will also need **Node.js** and **Go** installed locally for development.

### 1. Start the Infrastructure
Navigate to the root project folder and run the following command to start all databases and monitoring tools:
```bash
docker-compose up -d
```

### 2. Start the Backend API
Open a terminal and run:
```bash
cd backend
# Ensure your .env is configured (DB_HOST=localhost)
go run cmd/main.go
```
*(The backend runs on http://localhost:8080)*

### 3. Start the Frontend UI / Merchant App
Open a new terminal and run:
```bash
cd frontend
npm install   # Required only for the first time
npm run dev
```
*(The merchant dashboard runs on http://localhost:5173)*

### 4. Start the Customer App
Open a new terminal and run:
```bash
cd customer-app
npm install   # Required only for the first time
npm run dev
```
*(The customer app runs on http://localhost:5174)*
