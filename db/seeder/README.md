# Database Seeder

Seeds the PostgreSQL database with realistic Indonesian financial demo data using [GoSeeder](https://github.com/kristijorgji/goseeder).

## Prerequisites

- Migrations already applied (`migrate up`)
- `.env` populated with DB credentials (same file used by the main app)

## Quick Start

```bash
# Seed fresh (wipe + re-seed) — recommended for clean demo state
go run ./db/seeder -fresh

# Seed without wiping existing data
go run ./db/seeder

# Run only specific seeders
go run ./db/seeder -only SeedUsers,SeedTransactions

# Build a seeder binary
go build -o ./tmp/seeder.exe ./db/seeder
./tmp/seeder.exe -fresh
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-fresh` | false | Truncate all seeded tables (RESTART IDENTITY CASCADE) before seeding |
| `-only` | "" | Comma-separated seeder names; omit to run all |

## Seeder Execution Order

Seeders run in registration order to satisfy FK constraints:

1. **SeedUsers** — 5 demo users
2. **SeedFinancialProfiles** — financial profile + goal tags per user
3. **SeedTransactions** — 7 months of transaction history (Nov 2025 – May 2026)
4. **SeedBudgets** — monthly + yearly budgets for current period
5. **SeedGoals** — financial savings goals
6. **SeedAiLogs** — sample AI chat history

## Seeded Data Summary

### Users (5)

| Email | Name | Role | Password |
|-------|------|------|----------|
| budi.santoso@gmail.com | Budi Santoso | Software Engineer | Password123 |
| siti.rahayu@gmail.com | Siti Rahayu | Pemilik Toko Online | Password123 |
| ahmad.fauzi@gmail.com | Ahmad Fauzi | PNS Golongan III | Password123 |
| dewi.permata@gmail.com | Dewi Permata | UI/UX Freelancer | Password123 |
| eko.prabowo@gmail.com | Eko Prabowo | Mahasiswa Part-time | Password123 |

### Transactions (~1,200 total)

- **Period:** Nov 2025 – May 2026 (7 months)
- **Income categories:** Gaji, Freelance, Bonus
- **Expense categories:** Makanan & Minuman, Transportasi, Utilitas, Hiburan & Rekreasi, Belanja, Kesehatan, Pendidikan, Tagihan
- **Pattern:** Salary on 25th, daily food/transport, monthly utility bills, occasional other expenses
- **Deterministic:** same seed data every run (math/rand seeded by user index)

Monthly salary per user (IDR):
- Budi: ~Rp 10,000,000
- Siti: ~Rp 7,500,000 + freelance income
- Ahmad: ~Rp 6,000,000
- Dewi: ~Rp 8,500,000 + freelance income
- Eko: ~Rp 2,500,000

### Budgets (21 total)

Monthly budgets for May 2026 + yearly budgets for 2026.
Categories: Makanan & Minuman, Transportasi, Hiburan & Rekreasi, Belanja, Utilitas, Kesehatan, Pendidikan, Tagihan, Peralatan Kerja, Investasi.

### Goals (12 total, 1 COMPLETED)

- Budi: Dana Darurat, DP Rumah, Liburan ke Jepang
- Siti: Modal Usaha Tambahan, Beli Mobil Honda Brio
- Ahmad: Biaya Umroh, Dana Pendidikan Anak, Renovasi Dapur (COMPLETED)
- Dewi: Laptop MacBook Pro M3, Liburan Backpacker Eropa
- Eko: Motor Honda Beat, Laptop untuk Kuliah

### Financial Profiles (5)

Each user has: monthly income, fixed expenses, current savings, debt, employment status, spending habit, risk level, and financial goal tags.

### AI Logs (15 total)

3 sample chat Q&A entries per user.

## Common Workflows

```bash
# Full reset + seed (development)
go run ./db/seeder -fresh

# Re-seed only transactions (data exploration)
go run ./db/seeder -fresh -only SeedTransactions

# Add new users without touching other tables
go run ./db/seeder -only SeedUsers
```

## Migration + Seed Workflow

```bash
# 1. Apply all migrations
migrate -database "postgres://user:pass@localhost:5432/financial_planning?sslmode=disable" \
        -path db/migrations up

# 2. Seed
go run ./db/seeder -fresh

# 3. Start server
air
```

## Directory Structure

```
db/seeder/
├── main.go                  Entry point — registers seeders, runs Execute
├── config/
│   └── db.go                DB connect, truncate, GetUserIDs helper
├── factories/
│   └── static.go            Static seed data (users, profiles, budgets, goals, AI logs)
├── generators/
│   └── transactions.go      Programmatic transaction generator (deterministic rand)
├── seeders/
│   ├── users.go             Inserts users with bcrypt-hashed passwords
│   ├── financial_profiles.go Upserts financial profiles + goal tags
│   ├── transactions.go      Inserts generated transaction history
│   ├── budgets.go           Inserts monthly/yearly budgets
│   ├── goals.go             Inserts financial goals
│   └── ai_logs.go           Inserts sample AI chat logs
└── README.md
```

## Adding New Seeders

1. Create `db/seeder/seeders/my_entity.go` with `func SeedMyEntity(s goseeder.Seeder)`
2. Register in `db/seeder/main.go`:
   ```go
   goseeder.Register(seeders.SeedMyEntity)
   ```
3. Add the table to `config.Truncate()` in correct FK order

## Environment Variables

Same as the main application `.env`:

```env
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=financial_planning
DB_HOST=localhost
DB_PORT=5432
```

`SECRET_KEY` and `GEMINI_API_KEY` are not required by the seeder.
