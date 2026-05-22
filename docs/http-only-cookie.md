# HttpOnly Cookie Migration — Frontend Guide

The login endpoint now sets an HttpOnly, `SameSite=None; Secure` (on HTTPS) cookie containing the JWT. This means **JavaScript cannot read the token from `document.cookie`** — and doesn't need to.

## What changed

| Attribute | Before | After |
|-----------|--------|-------|
| `HttpOnly` | `false` | `true` |
| `Secure` | `false` | auto: `true` on HTTPS, `false` on HTTP |
| `SameSite` | (none) | `None` on HTTPS, `Lax` on HTTP |
| `Domain` | `"*"` (invalid) | `""` (defaults to request origin) |

## What the frontend must do

### 1. Remove `document.cookie` reads

Delete any code reading the token from cookies:

```ts
// ❌ REMOVE — HttpOnly makes this return empty
const token = document.cookie
  .split("; ")
  .find((row) => row.startsWith("accessToken="))
  ?.split("=")[1];
```

### 2. Get the token from the login response body

The login endpoint still returns the token in `data.token`. Store it in memory (React state / context / zustand) instead of cookies:

```ts
const res = await fetch("/api/v1/login", { method: "POST", body: ... });
const json = await res.json();
const token = json.data.token; // still available
```

### 3. Send the token as `Authorization: Bearer <token>`

Attach it to every authenticated request (the auth middleware expects this header):

```ts
fetch("/api/auth/v1/transactions", {
  headers: { Authorization: `Bearer ${token}` }
});
```

### 4. Add `credentials: "include"` to all fetch calls

This tells the browser to include the HttpOnly cookie on cross-origin requests (needed when frontend and backend are on different ports/domains):

```diff
 fetch(url, {
+  credentials: "include",
   headers: { Authorization: `Bearer ${token}` },
 });
```

For **axios**, set it globally:

```ts
axios.defaults.withCredentials = true;
```

Or per-request:

```diff
-axios.get("/api/auth/v1/transactions");
+axios.get("/api/auth/v1/transactions", { withCredentials: true });
```

### 5. Store the token in memory, not in localStorage/sessionStorage

```ts
// In auth context/state — NOT in document.cookie or localStorage
let _token: string | null = null;

export function setToken(t: string) { _token = t; }
export function getToken(): string | null { return _token; }
```

On page reload the token is lost — the login page will redirect (the server can later implement a `/auth/refresh` endpoint using the HttpOnly cookie).

## Why this matters

| Concern | Without HttpOnly | With HttpOnly |
|---------|-----------------|---------------|
| XSS stealing the JWT | `document.cookie` leaks it | Cookie is invisible to JS |
| CSRF | Needs manual CSRF token | `SameSite=None; Secure` prevents cross-site form posts |
| Token sent automatically | No — manual header | Browser sends it (once middleware is updated) |

## When will the cookie be used server-side?

Currently the auth middleware only reads the `Authorization` header. The cookie is set but **not yet consumed** by the backend. Once the middleware is updated to fall back to the cookie, the `Authorization` header can be removed — the browser will send the cookie automatically with `credentials: "include"`.
