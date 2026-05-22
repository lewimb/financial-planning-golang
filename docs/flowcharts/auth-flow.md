# Authentication Flow

All flows derived from `internal/usecase/user.go`, `internal/delivery/http/middleware/auth.go`, and `utils/`.

---

## 1. User Registration

```mermaid
flowchart TD
    A([Client: POST /api/v1/register]) --> B[Parse JSON body\nemail · password · name]
    B --> C{Bind valid?}
    C -- No --> D[400 Invalid input]
    C -- Yes --> E[UserUseCase.Register]
    E --> F[bcrypt.GenerateFromPassword\ncost = 10]
    F --> G[UserRepository.Create\nINSERT INTO users]
    G --> H{PgError 23505\nduplicate email?}
    H -- Yes --> I[domain.ErrConflict\n→ ErrUserExists]
    I --> J[409 User already exists]
    H -- No --> K{DB error?}
    K -- Yes --> L[500 Registration failed]
    K -- No --> M[200 User created successfully]
```

---

## 2. User Login and JWT Issuance

```mermaid
flowchart TD
    A([Client: POST /api/v1/login]) --> B[Parse JSON body\nemail · password]
    B --> C{Bind valid?}
    C -- No --> D[400 Invalid input]
    C -- Yes --> E[UserUseCase.Login]
    E --> F[UserRepository.FindByEmail\nSELECT id,email,name,password\nWHERE email = $1]
    F --> G{Row found?}
    G -- No --> H[ErrInvalidCredentials]
    H --> I[400 invalid email or password]
    G -- Yes --> J[bcrypt.CompareHashAndPassword\nstored hash vs plain password]
    J --> K{Match?}
    K -- No --> H
    K -- Yes --> L[utils.GenerateJwt\nHS256 · claims: id·name·email\nsigned with SECRET_KEY]
    L --> M[Set-Cookie: accessToken\nMaxAge=3600 · path=/]
    M --> N[200 Login Successfully\nbody: token string]
```

> **JWT structure:** Header `{alg:HS256}` + Claims `{id, name, email, iss:lewimb}` + HMAC-SHA256 signature.  
> No `exp` claim is embedded — session length is enforced solely by the 1-hour cookie TTL.

---

## 3. JWT Middleware — Protected Route Guard

```mermaid
flowchart TD
    A([Incoming /api/auth/v1/* request]) --> B[Read Authorization header]
    B --> C{Header present?}
    C -- No --> D[400 Missing Authorization!]
    C -- Yes --> E[Split on space\n'Bearer TOKEN']
    E --> F{Exactly 2 parts?}
    F -- No --> G[401 format must be Bearer token]
    F -- Yes --> H[Extract token string]
    H --> I[jwt.ParseWithClaims\nverify HS256 signature\nwith SECRET_KEY]
    I --> J{token.Valid == true?}
    J -- No --> K[401 Invalid token\nc.Abort]
    J -- Yes --> L[c.Set 'claims'\nMyCustomClaims stored in Gin context]
    L --> M[c.Next → handler proceeds]
    M --> N[utils.ClaimId reads claims.Id\nfrom context for all protected handlers]
```

---

## 4. Error Propagation Map

```mermaid
flowchart LR
    subgraph Repository
        E1[ErrConflict\npgErr.Code == 23505]
        E2[sql.ErrNoRows]
        E3[DB error]
    end
    subgraph UseCase
        U1[ErrUserExists]
        U2[ErrInvalidCredentials]
    end
    subgraph Handler
        H1[409 Conflict]
        H2[400 Bad credentials]
        H3[500 Internal]
    end
    E1 --> U1 --> H1
    E2 --> U2 --> H2
    E3 --> U2
    E3 --> H3
```
