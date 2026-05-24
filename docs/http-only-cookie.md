# Auth Token Migration — Bearer Token Guide

The login endpoint now returns the JWT in the response body. Cookies are no longer used for authentication.

## What changed

| Attribute | Before | After |
|-----------|--------|-------|
| Token delivery | `Set-Cookie: accessToken` (HttpOnly) | JSON body `accessToken` field |
| Auth method | Cookie sent automatically by browser | `Authorization: Bearer <token>` header |
| `credentials: "include"` required | Yes | No |
| `withCredentials: true` required | Yes | No |
| Token readable by JS | No (HttpOnly) | Yes (response body) |

## Login response

```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

## What the frontend must do

### 1. Store the token from the login response

```ts
const res = await fetch("/api/v1/login", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ email, password }),
});
const { accessToken, user } = await res.json();
// store accessToken in memory (React state / Zustand / context)
// do NOT store in localStorage — XSS can steal it
```

### 2. Attach `Authorization: Bearer` on every authenticated request

```ts
fetch("/api/auth/v1/transactions", {
  headers: { Authorization: `Bearer ${accessToken}` },
});
```

For axios — set globally on the instance:

```ts
const api = axios.create({ baseURL: "https://your-api.railway.app" });

api.interceptors.request.use((config) => {
  const token = getToken(); // from your auth store
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});
```

### 3. Remove any cookie-based auth code

```ts
// ❌ REMOVE — no longer works
document.cookie.split("; ").find(row => row.startsWith("accessToken="));

// ❌ REMOVE — no longer needed
credentials: "include"
withCredentials: true
```

### 4. Handle token expiry (1 hour)

The JWT expires in 1 hour (`exp` claim). When a request returns 401, redirect to login:

```ts
api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      clearToken();
      window.location.href = "/login";
    }
    return Promise.reject(err);
  }
);
```

## Security notes

- Store the token in memory (React context / Zustand) — NOT `localStorage` or `sessionStorage`
- On page reload the token is lost — redirect to login (expected behavior without refresh tokens)
- Token expiry is enforced server-side via the `exp` JWT claim (1 hour)
