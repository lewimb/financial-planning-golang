# ML Service — Setup & Running Guide

## Overview

This service exposes three HTTP endpoints that analyze financial transaction data:

| Endpoint | What it does |
|----------|-------------|
| `POST /analysis` | Totals, averages, top spending category |
| `POST /anomaly` | Flags unusual spending days |
| `POST /forecast` | Predicts future daily spending |

The service is stateless — it has no database. Every request is self-contained: you send transactions, it returns results.

---

## Requirements

- **Python 3.11+**
- pip (comes with Python)

---

## Installation

### 1. Create a virtual environment (recommended)

```bash
python -m venv venv
venv\Scripts\activate        # Windows
# source venv/bin/activate   # macOS/Linux
```

### 2. Install dependencies

```bash
pip install -r requirements.txt
```

This installs:

| Package | Version | Purpose |
|---------|---------|---------|
| fastapi | 0.115.0 | HTTP framework |
| uvicorn | 0.30.1 | ASGI server |
| pandas | 2.2.2 | Data aggregation |
| scikit-learn | 1.5.1 | Anomaly detection |
| prophet | 1.1.5 | Time-series forecasting |
| pydantic | 2.7.4 | Request validation |

> **Note:** Prophet downloads and compiles Stan models on first install. This can take 2–5 minutes and requires an internet connection.

---

## Running the Service

```bash
uvicorn main:app --reload
```

The service starts at `http://localhost:8000`.

- `--reload` restarts automatically when you edit code (development only)
- Remove `--reload` in production

### Verify it's running

Open `http://localhost:8000/docs` in a browser — Swagger UI loads automatically.

---

## Running Tests

```bash
pytest -v
```

Expected: **36 tests pass**

> The forecast tests call Prophet and take ~15–20 seconds on first run.

---

## Example API Calls

### /analysis

```bash
curl -X POST http://localhost:8000/analysis \
  -H "Content-Type: application/json" \
  -d '[
    {"date":"2026-01-01","amount":50000,"type":"EXPENSE","category":"food"},
    {"date":"2026-01-02","amount":75000,"type":"EXPENSE","category":"transport"},
    {"date":"2026-01-01","amount":500000,"type":"INCOME","category":"salary"}
  ]'
```

Response:
```json
{
  "total_expense": 125000.0,
  "avg_daily": 62500.0,
  "top_category": "transport",
  "spending_distribution": {
    "food": 50000.0,
    "transport": 75000.0
  }
}
```

### /anomaly

```bash
curl -X POST http://localhost:8000/anomaly \
  -H "Content-Type: application/json" \
  -d '[
    {"date":"2026-01-01","amount":50000,"type":"EXPENSE","category":"food"},
    {"date":"2026-01-02","amount":55000,"type":"EXPENSE","category":"food"},
    {"date":"2026-01-03","amount":48000,"type":"EXPENSE","category":"food"},
    {"date":"2026-01-04","amount":52000,"type":"EXPENSE","category":"food"},
    {"date":"2026-01-05","amount":51000,"type":"EXPENSE","category":"food"},
    {"date":"2026-01-06","amount":900000,"type":"EXPENSE","category":"other"}
  ]'
```

Response:
```json
{
  "anomalies": [{"date": "2026-01-06", "amount": 900000.0}],
  "summary": "You spent unusually high on 1 day(s)"
}
```

### /forecast

```bash
curl -X POST "http://localhost:8000/forecast?periods=7" \
  -H "Content-Type: application/json" \
  -d '[
    {"date":"2026-01-01","amount":50000,"type":"EXPENSE","category":"food"},
    {"date":"2026-01-02","amount":75000,"type":"EXPENSE","category":"transport"}
  ]'
```

Response:
```json
{
  "predicted_monthly_spending": 437500.0,
  "daily_forecast": [
    {"date": "2026-01-03", "predicted_amount": 62500.0},
    ...
  ]
}
```

---

## Edge Cases — All Handled Automatically

| Scenario | Behavior |
|----------|----------|
| Empty transaction array | Returns zero/empty results with 200 OK |
| Only INCOME transactions | Treated as no data |
| Fewer than 5 expense days | `/anomaly` skips detection, returns empty list |
| 1 expense day | `/forecast` uses daily-average fallback |
