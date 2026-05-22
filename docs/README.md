# Documentation Index

AI Financial Planning Platform — Go Backend  
Last audited: 2026-05-19

---

## Rendering Mermaid Diagrams

All diagrams use [Mermaid](https://mermaid.js.org/) syntax.

| Tool | How to render |
|---|---|
| **GitHub** | Renders automatically in `.md` files |
| **VS Code** | Install [Markdown Preview Mermaid Support](https://marketplace.visualstudio.com/items?itemName=bierner.markdown-mermaid) |
| **Mermaid Live Editor** | Paste diagram code at [mermaid.live](https://mermaid.live) |
| **Obsidian** | Native Mermaid support |

Diagram types used:
- `flowchart TD` — top-down decision trees and data flows
- `sequenceDiagram` — request/response lifetimes across components
- `erDiagram` — entity-relationship schema
- `graph TB` — architecture overview

---

## Architecture Overview

```
Client (React/Vite :5173)
    │
    ▼ HTTP JSON + Authorization: Bearer token
Go Backend (Gin :8080)
    │
    ├── /api/v1/*          Public routes (register, login)
    └── /api/auth/v1/*     JWT-protected routes
           │
           ├── Delivery Layer    (Gin handlers + middleware)
           ├── Use Case Layer    (business logic + validation)
           ├── Repository Layer  (raw SQL via pgx/v5)
           └── Domain Layer      (interfaces + DTOs + errors)
                   │
                   ▼ raw SQL
              PostgreSQL :5432 (9 tables)
                   
    ├── ML Service (FastAPI :8000)  ← spending analysis, anomaly, forecast
    └── Google Gemini API           ← AI chat responses
    
db/seeder/   — standalone seed binary (GoSeeder v1.0.5)
```

---

## File Index

### Core Architecture

| File | Description |
|---|---|
| [system-architecture.md](system-architecture.md) | High-level diagram, component wiring, clean architecture dependency rule, ML integration, known issues |
| [erd.md](erd.md) | Full entity-relationship diagram, all 9 tables with column types/constraints, design notes |
| [service-logic-overview.md](service-logic-overview.md) | All use case business logic: validation rules, error contracts, cross-domain dependencies |

### Sequence Diagrams

| File | Sequences covered |
|---|---|
| [sequence-diagram.md](sequence-diagram.md) | Registration, login, JWT middleware, create transaction, goal contribution, dashboard, AI chat, financial profile upsert, ML forecast, budget usage |

### Flowcharts

| File | Covers |
|---|---|
| [flowcharts/auth-flow.md](flowcharts/auth-flow.md) | Registration, login, JWT middleware, error propagation |
| [flowcharts/transaction-flow.md](flowcharts/transaction-flow.md) | CRUD, soft delete, monthly aggregations, net savings |
| [flowcharts/budgeting-flow.md](flowcharts/budgeting-flow.md) | Budget CRUD, usage calculation (SAFE/WARNING/EXCEEDED), status state machine |
| [flowcharts/analytics-flow.md](flowcharts/analytics-flow.md) | Dashboard aggregation, ML analysis/anomaly, data transformation |
| [flowcharts/forecasting-flow.md](flowcharts/forecasting-flow.md) | Full ML forecast pipeline, analysis, anomaly, timeout strategy, error handling |
| [flowcharts/seeder-flow.md](flowcharts/seeder-flow.md) | Seeder entry point, per-entity logic, transaction generator algorithm, data summary |
| [flowcharts/api-request-flow.md](flowcharts/api-request-flow.md) | Generic request lifecycle, route map, handler↔usecase↔repo dependency map, error→HTTP mapping |
| [flowcharts/database-sync-flow.md](flowcharts/database-sync-flow.md) | Atomic financial profile upsert, goal contribution cross-domain, budget JOIN pattern, soft vs hard delete, duplicate prevention, index map |

---

## Key Design Decisions

| Decision | Rationale |
|---|---|
| Raw SQL (no ORM) | Full control over query structure; pgx/v5 provides type-safe scanning |
| Clean architecture | Enforces dependency direction inward; domain layer has zero framework imports |
| GoalUseCase injects TransactionRepository | Cross-domain net-savings check without circular use-case dependency |
| FinancialProfile upsert uses DB transaction | Atomically replaces profile + goal tags — no partial state |
| JWT has no `exp` claim | Session length governed only by 1-hour cookie TTL |
| Budget↔Transaction join is logical | Case-insensitive string match on category — no FK enforces consistency |
| Goals use hard delete | `deleted_at` was dropped in migration 012; `DELETE FROM goals` is the only delete path |
| ML service has no auth | Designed for internal use only; exposed on port 8000 without token validation |

---

## Known Issues / Improvement Opportunities

| Issue | Severity | Location |
|---|---|---|
| `UserResponse.Password` included in `GET /users` response | **High** | `UserRepository.GetAll` |
| Sequential DB calls in Dashboard (6 queries) | Medium | `DashboardUseCase.Get` |
| Budget↔Transaction category join has no FK | Medium | `BudgetRepository.GetUsage` |
| JWT has no expiry claim | Medium | `utils/jwt.go` |
| CORS origin hardcoded to `localhost:5173` | Low | `main.go` |
| `reports` and `settings` tables exist but have no routes | Low | migrations 006/007 |
| No pagination on Goals or Budgets `GetAll` | Low | `GoalRepository`, `BudgetRepository` |
| ML service has no authentication | Low | `internal/ml/client.go` |
