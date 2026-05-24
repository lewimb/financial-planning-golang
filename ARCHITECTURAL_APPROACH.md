# Architectural Approach

## Classification: Hexagonal Architecture (Ports & Adapters)

Perancangan aplikasi *financial planner* berbasis web pada repositori ini menggunakan pendekatan **Hexagonal Architecture** (Ports & Adapters) dengan memanfaatkan antarmuka (interface) sebagai alat pemodelan sistem. Komponen arsitektur yang digunakan meliputi *domain entities*, *repository interfaces* (ports), *use case layer*, *HTTP handlers* (adapters), dan *PostgreSQL repository implementations* (adapters). Pendekatan Hexagonal Architecture digunakan karena aplikasi yang dikembangkan memiliki beberapa objek dan entitas yang saling berinteraksi di dalam sistem, seperti *user*, *transaction*, *budget*, *goal*, serta *AI recommendation*.

Hexagonal Architecture merupakan metode pengembangan perangkat lunak yang berfokus pada pemisahan antara **logika bisnis inti** (domain) dengan **infrastruktur eksternal** (database, framework, API). Pendekatan ini membantu proses analisis dan perancangan sistem menjadi lebih mudah dipahami karena setiap bagian sistem direpresentasikan dalam bentuk *port* (antarmuka) dan *adapter* (implementasi) yang memiliki tanggung jawab masing-masing. Selain itu, pendekatan ini juga mendukung pengembangan aplikasi yang lebih terstruktur, mudah dikembangkan, serta mempermudah proses *maintenance* sistem di masa mendatang.

Dalam penerapannya, pendekatan Hexagonal Architecture dibangun berdasarkan beberapa konsep utama, yaitu **dependency inversion**, **separation of concerns**, **encapsulation**, dan **interface-based abstraction**. Konsep-konsep tersebut digunakan untuk membantu proses perancangan sistem agar lebih modular dan efisien dalam pengembangan perangkat lunak berbasis Go.

---

## Ports & Adapters Mapping

```
┌─────────────────────────────────────────────────────┐
│                    ADAPTERS (Driving)                │
│  ┌──────────────────────────────────────────────┐   │
│  │         HTTP Handler (Gin framework)          │   │
│  │  ─  menerima request dari client             │   │
│  │  ─  memanggil use case melalui port          │   │
│  └──────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────┤
│                     PORTS (Core)                     │
│  ┌──────────────────────────────────────────────┐   │
│  │              Use Case Layer                    │   │
│  │  ─  logika bisnis aplikasi                   │   │
│  │  ─  bergantung pada interface repository     │   │
│  └──────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────┐   │
│  │              Domain Layer                      │   │
│  │  ─  entities (struct)                         │   │
│  │  ─  repository interfaces (ports)             │   │
│  │  ─  sentinel errors                           │   │
│  └──────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────┤
│                    ADAPTERS (Driven)                │
│  ┌──────────────────────────────────────────────┐   │
│  │         PostgreSQL Repository                 │   │
│  │  ─  mengimplementasikan port repository      │   │
│  │  ─  berisi SQL queries                       │   │
│  └──────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────┐   │
│  │   Gemini Client / ML Service Client           │   │
│  │  ─  adapter untuk AI eksternal               │   │
│  └──────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

### Alur Dependensi

```
Driving Adapter (HTTP/Gin)
       │  memanggil use case via constructor injection
       ▼
Port (UseCase struct — business logic)
       │  bergantung pada interface (port), bukan implementasi konkret
       ▼
Port (Repository Interface — defined in domain)
       ▲
       │  diimplementasikan oleh
Driven Adapter (PostgreSQL Repository)
```

- **Domain layer** tidak mengimpor paket apa pun selain stdlib.
- **Use case layer** hanya mengimpor `domain/` dan stdlib.
- **HTTP handlers** mengimpor `usecase/` dan `domain/` — tidak mengimpor repository.
- **PostgreSQL repository** mengimpor `domain/` dan `database/sql` — tidak mengimpor use case atau handler.

---

## Penerapan Konsep Hexagonal Architecture

### 1. Dependency Inversion (Ports)

Repository *interface* (port) didefinisikan di dalam **domain layer** dan diimplementasikan di **outer layer** (adapter):

```go
// domain/transaction.go — PORT (milik domain)
type TransactionRepository interface {
    GetByUserID(userID, limit, offset int, year, month string) ([]TransactionResponse, int, error)
    Create(userID int, req TransactionRequest) error
}

// repository/postgres/transaction.go — ADAPTER (implementasi konkret)
type transactionRepository struct { db *sql.DB }
func NewTransactionRepository(db *sql.DB) domain.TransactionRepository {
    return &transactionRepository{db: db}
}
```

Use case bergantung pada *port* (interface), bukan pada adapter:

```go
type TransactionUseCase struct {
    repo domain.TransactionRepository  // ← port
}
```

### 2. Constructor Injection (Composition Root)

Seluruh dependensi diinjeksi melalui constructor dan dirakit di satu titik (`main.go`):

```go
// main.go — composition root
txRepo  := postgres.NewTransactionRepository(db)     // adapter
txUC    := usecase.NewTransactionUseCase(txRepo)       // port → terima adapter
handler := handler.NewTransactionHandler(txUC)         // driving adapter → terima use case
```

### 3. Framework Independence

- Tidak ada Gin, `database/sql`, atau konsep HTTP di dalam `usecase/` maupun `domain/`.
- Use case dapat diuji dengan *mock implementation* dari domain interface tanpa perlu database atau server sungguhan.

### 4. Separation of Concerns

| Layer | Tanggung Jawab | Boleh Mengimpor |
|-------|---------------|-----------------|
| `internal/domain/` | Entities, port interfaces, sentinel errors | stdlib |
| `internal/usecase/` | Business logic, validasi, orkestrasi | `domain/`, stdlib |
| `internal/delivery/http/` | HTTP handler, routing, middleware | `domain/`, `usecase/`, Gin |
| `internal/repository/postgres/` | Implementasi SQL | `domain/`, `database/sql` |

### 5. Encapsulation via Private Structs

Detail implementasi disembunyikan dengan membuat struct repository sebagai **private** (`transactionRepository`) dan hanya mengekspos konstruktor publik:

```go
// repository/postgres/transaction.go
type transactionRepository struct { db *sql.DB }  // private

func NewTransactionRepository(db *sql.DB) domain.TransactionRepository {
    return &transactionRepository{db: db}  // publik, kembalikan interface
}
```

### 6. Repository Pattern

Setiap entitas domain memiliki *port repository* yang sesuai. Detail persistensi (SQL mentah, soft-delete via `deleted_at`, fitur PostgreSQL seperti `EXTRACT()` / `NULLIF`) dienkapsulasi di dalam `repository/postgres/` dan tidak terlihat oleh pemanggil.

---

## Perbandingan dengan Pendekatan Lain

| Pendekatan | Mengapa Bukan |
|------------|---------------|
| **OOAD murni** | Tidak menggunakan inheritance atau polymorphism berbasis class. Go tidak mendukung inheritance klasik. Domain bersifat anemic (DTO), tanpa perilaku terenkapsulasi. |
| **Prosedural** | Menggunakan struct methods, interfaces, dan dependency injection secara ekstensif — bukan fungsi datar yang memutasi state global. |
| **Full DDD** | Tidak memiliki domain events, value objects, aggregates, atau ubiquitous language. Domain layer sengaja dijaga tetap tipis. |
| **MVC tradisional** | Tidak ada model yang mengandung logika bisnis dan database sekaligus. Pemisahan lebih ketat dengan dependency inversion. |

---

## Summary

Repositori ini mengimplementasikan **Hexagonal Architecture (Ports & Adapters)** dalam idiom Go yang pragmatis — sepenuhnya menerapkan *dependency inversion* dan *separation of concerns* namun tidak menggunakan konsep DDD yang berat demi kesederhanaan dan kemudahan *maintenance*. Project `ARCHITECTURE.md` mendeskripsikannya sebagai: *"Clean Architecture · Repository Pattern · Dependency Injection"* — yang dalam konteks akademis dapat diklasifikasikan sebagai varian dari Hexagonal Architecture.
