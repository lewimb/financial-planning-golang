# ERD — Financial Planning Backend

Open `erd.excalidraw` at [excalidraw.com](https://excalidraw.com) (File → Open) to view the diagram.

---

## Tables

| Table | PK | FKs | Notes |
|---|---|---|---|
| `users` | `id` | — | Central entity; soft-delete via `deleted_at` |
| `transactions` | `id` | `user_id → users.id` | `type` CHECK: `INCOME\|EXPENSE`; indexed on `(user_id, category, date)` |
| `budgets` | `id` | `user_id → users.id` | UNIQUE on `(user_id, category, period, month, year)` prevents duplicates |
| `goals` | `id` | `user_id → users.id` | `status` DEFAULT `ONGOING`; `description` added in migration 009 |
| `ai_logs` | `id` | `user_id → users.id` | Stores AI Q&A pairs per user; no handler yet (schema only) |
| `reports` | `id` | `user_id → users.id` | No handler yet (schema only); `type` MONTHLY\|YEARLY |
| `settings` | `id` | `user_id → users.id` UNIQUE | UNIQUE on `user_id` enforces **1:1** with `users` |

---

## Relationships

```
users ──────────────────────────── settings      (1 : 1)
  │   UNIQUE FK on settings.user_id
  │
  ├──────────────────────────────── transactions  (1 : N)
  ├──────────────────────────────── budgets       (1 : N)
  ├──────────────────────────────── goals         (1 : N)
  ├──────────────────────────────── ai_logs       (1 : N)
  └──────────────────────────────── reports       (1 : N)
```

No junction tables. No N:N relationships exist in the current schema.

---

## Assumptions & Inferences

- **Soft deletes** — `users`, `transactions`, `budgets`, `goals`, `ai_logs` all have `deleted_at TIMESTAMP`. The repository queries do not explicitly filter `WHERE deleted_at IS NULL` in all cases; this may need auditing.
- **`reports` and `ai_logs`** have no route handlers in the current implementation. The tables exist in migrations but are dead schema — likely planned features.
- **`budgets` ↔ `transactions`** — the `GetBudgetUsage` service joins them in-memory by matching `category`, `user_id`, `month`, and `year`. There is no formal FK between the two tables; the relationship is logical, not structural.
- **`settings` cardinality** — enforced solely by the `UNIQUE` constraint on `settings.user_id`. Application code does not yet expose settings CRUD endpoints.
- **`goals.description`** was added in migration `000009` as an `ALTER TABLE`, meaning it is nullable with no default.
- **Performance indices** (migration 008): `idx_transactions_user_category_date`, `idx_transactions_full`, `idx_budgets_user`.
