# Seeder Flow

Derived from `db/seeder/` — the standalone seeder binary using GoSeeder v1.0.5.

---

## 1. Seeder Entry Point

```mermaid
flowchart TD
    A([go run ./db/seeder]) --> B[godotenv.Load\nread .env file]
    B --> C[Parse CLI flags\n-fresh bool\n-only comma-separated names]
    C --> D[config.Connect\nopen pgx connection to PostgreSQL]
    D --> E{-fresh flag set?}
    E -- Yes --> F[config.Truncate\nTRUNCATE with RESTART IDENTITY CASCADE\nOrder: ai_logs · user_financial_goals\nuser_financial_profiles · goals\nbudgets · transactions · users]
    E -- No --> G[Skip truncation]
    F --> H[Register seeders in FK-safe order]
    G --> H

    H --> I1[goseeder.Register SeedUsers]
    I1 --> I2[goseeder.Register SeedFinancialProfiles]
    I2 --> I3[goseeder.Register SeedTransactions]
    I3 --> I4[goseeder.Register SeedBudgets]
    I4 --> I5[goseeder.Register SeedGoals]
    I5 --> I6[goseeder.Register SeedAiLogs]

    I6 --> J{-only flag set?}
    J -- Yes --> K[ForSpecificSeeds option\nfilter by name]
    J -- No --> L[Run all registered seeders]
    K --> M[goseeder.Execute db options]
    L --> M

    M --> N{Execute error?}
    N -- Yes --> O[os.Exit 1]
    N -- No --> P[Seeding complete!]
```

---

## 2. Individual Seeder Logic

### SeedUsers

```mermaid
flowchart TD
    A[SeedUsers] --> B[For each UserSeed in factories.Users]
    B --> C[bcrypt.GenerateFromPassword\nDefaultCost]
    C --> D[INSERT INTO users email·name·password\nON CONFLICT email DO NOTHING]
    D --> E{error?}
    E -- Yes --> F[log.Fatalf]
    E -- No --> G[Next user]
    G --> B
```

### SeedFinancialProfiles

```mermaid
flowchart TD
    A[SeedFinancialProfiles] --> B[config.GetUserIDs\nSELECT id FROM users ORDER BY id]
    B --> C{len userIDs == 0?}
    C -- Yes --> D[log.Fatal: no users found]
    C -- No --> E[For each userID index i]
    E --> F[INSERT INTO user_financial_profiles\nON CONFLICT user_id DO UPDATE all fields]
    F --> G[DELETE FROM user_financial_goals WHERE user_id]
    G --> H[For each goal type in profile.Goals]
    H --> I[INSERT INTO user_financial_goals\nON CONFLICT user_id goal_type DO NOTHING]
    I --> J[Next user]
    J --> E
```

### SeedTransactions

```mermaid
flowchart TD
    A[SeedTransactions] --> B[config.GetUserIDs]
    B --> C[For each userID index i]
    C --> D[generators.Generate i\nDeterministic rand seeded by user index\nProduces 7 months of transactions]
    D --> E[For each TX in generated slice]
    E --> F[INSERT INTO transactions\nuser_id·amount·category·type·date·description]
    F --> G[Next TX]
    G --> E
    E --> H[Next user]
    H --> C
```

### SeedBudgets

```mermaid
flowchart TD
    A[SeedBudgets] --> B[config.GetUserIDs]
    B --> C[For each userID index i]
    C --> D[For each BudgetSeed in factories.Budgets index i]
    D --> E{period == MONTHLY?}
    E -- Yes --> F[INSERT ON CONFLICT\nuser_id·category·period·month·year DO NOTHING]
    E -- No --> G[INSERT WHERE NOT EXISTS\nguard for NULL month YEARLY budgets\nPostgreSQL NULL != NULL in UNIQUE constraint]
    F --> H[Next budget]
    G --> H
    H --> D
    D --> I[Next user]
    I --> C
```

### SeedGoals

```mermaid
flowchart TD
    A[SeedGoals] --> B[config.GetUserIDs]
    B --> C[For each userID index i]
    C --> D[For each GoalSeed in factories.Goals index i]
    D --> E[time.Parse '2006-01-02' deadline string]
    E --> F[INSERT INTO goals\nuser_id·name·target_amount·current_amount\nstatus·deadline·description]
    F --> G[Next goal]
    G --> D
    D --> H[Next user]
    H --> C
```

### SeedAiLogs

```mermaid
flowchart TD
    A[SeedAiLogs] --> B[config.GetUserIDs]
    B --> C[For each userID]
    C --> D[For each AiLogSeed in factories.AiLogs 3 entries]
    D --> E[INSERT INTO ai_logs\nuser_id·question·response]
    E --> F[Next log]
    F --> D
    D --> G[Next user]
    G --> C
```

---

## 3. Transaction Generator Algorithm

```mermaid
flowchart TD
    A[generators.Generate userIndex] --> B[rand.New\nrand.NewSource userIndex*99991+12345\nDeterministic seed]
    B --> C[baseSalary = salaries index\n10M · 7.5M · 6M · 8.5M · 2.5M IDR]
    C --> D[For monthsBack 6 down to 0\nNov 2025 → May 2026]

    D --> E[Compute monthStart and lastDay\nif current month: cap at day 19]

    E --> F[INCOME: Insert salary\non 25th or lastDay if less\nvariation ±150K IDR]

    F --> G{userIndex == 1 Siti or 3 Dewi?}
    G -- Yes --> H[67% chance: Insert freelance income\n500K–3.5M IDR random day]
    G -- No --> I[Skip freelance]
    H --> J{past month and 17% chance?}
    I --> J
    J -- Yes --> K[Insert bonus income 1M–5M]
    J -- No --> L[Skip bonus]

    K --> M[For each day 1..lastDay]
    L --> M

    M --> N[Insert utility bills on days 1–5\nPLN · internet · PDAM · gas\nSkip for Eko student index 4]

    N --> O{rng 85% chance?}
    O -- Yes --> P[Insert 1–2 food transactions\nfrom foodTemplates array]
    O -- No --> Q[Skip food]

    P --> R{Weekday Mon–Fri?}
    Q --> R
    R -- Yes --> S[Insert 1–2 transport transactions]
    R -- No --> T[Skip transport]

    S --> U{rng 25% chance?}
    T --> U
    U -- Yes --> V[Insert random other expense\nEntertainment · Shopping · Health\nEducation · Bills]
    U -- No --> W[Next day]
    V --> W
    W --> M
    M --> X[Next month]
    X --> D
    D --> Y[Return TX slice ~170–220 per user]
```

---

## 4. Data Summary

| Entity | Records per User | Total (5 users) |
|---|---|---|
| users | 1 | 5 |
| financial profiles | 1 | 5 |
| user financial goal tags | 2–3 | ~12 |
| transactions | ~170–220 | ~1,000 |
| budgets | 3–6 | 21 |
| goals | 2–3 | 12 (1 COMPLETED) |
| ai_logs | 3 | 15 |
