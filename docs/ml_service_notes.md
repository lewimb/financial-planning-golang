# ML Service Notes — Gap Analysis Follow-ups

> Generated: 2026-05-25
> Context: Items from `frontend_backend_gap_analysis.md` that require ML service changes or coordination.

---

## 1. Wire `/ml/insights` into AI Coach panels (frontend task, but ML service must return stable schema)

**Gap:** `FinancialKeyInsights.tsx` shows hardcoded strings. The `/auth/v1/ml/insights` endpoint exists and returns real data.

**ML service requirement:** Ensure the response includes an `insights` array with `type`, `title`, `description`, `status` fields so the frontend can map them directly. If the current schema differs, align it with:

```json
{
  "insights": [
    { "type": "string", "title": "string", "description": "string", "status": "success|warning|info" }
  ]
}
```

---

## 2. `FinancialRecommendation.tsx` — derive from ML insights or new endpoint

**Gap:** Component shows 3 hardcoded strings. Two options:

**Option A:** The ML insights response (`/auth/v1/ml/insights`) already returns `recommendations` — expose them in the response schema and have the frontend read them.

**Option B:** Use the new rule-based `GET /auth/v1/recommendations` endpoint (implemented in this PR). Sufficient for MVP.

**Recommendation:** Use the new `/recommendations` endpoint for MVP. If the ML service produces higher-quality recommendations (using model outputs), add a `source: "ml"` field to the response and switch over.

---

## 3. ML Forecast — Async Job Pattern (BP-12)

**Gap:** `GET /auth/v1/ml/forecast` is synchronous and can hang for 60+ seconds. Frontend has no timeout/abort.

**Required backend changes (ML service):**

```
POST /auth/v1/ml/forecast/start     → 202 Accepted, { "job_id": "uuid" }
GET  /auth/v1/ml/forecast/status/{job_id} → { "status": "pending|running|complete|failed", "result": {...} }
```

**Suggested DB table (Go backend side):**
```sql
CREATE TABLE forecast_jobs (
  id         VARCHAR(36) PRIMARY KEY,
  user_id    BIGINT NOT NULL REFERENCES users(id),
  status     VARCHAR(20) NOT NULL DEFAULT 'pending',
  result     JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**ML service work needed:**
- Accept a `job_id` and run the forecast asynchronously
- Write result back to `forecast_jobs` table on completion
- The Go backend polls or the ML service calls a callback

**Priority:** Medium. Fixes a hard UX hang.

---

## 4. Financial Health Score — ML Enhancement (optional)

**Current implementation:** Rule-based score in `GET /auth/v1/financial-health` (implemented in this PR).

**ML enhancement path:** Replace or augment the rule-based formula with a Gemini call that evaluates the user's financial data and returns a structured score + trend. Example prompt injection:

```
Given the user's data:
- Savings rate: {savings_rate}%
- Budget adherence: {budget_adherence}%
- Goal progress: {goal_progress}%
- Monthly income: {income}, expense: {expense}

Return JSON: { "score": 0-100, "rating": "Poor|Fair|Good|Excellent", "trend": "improving|stable|declining", "summary": "..." }
```

**Consideration:** This adds Gemini latency to a frequently-called endpoint. Cache the result for 24h by user_id (optional `financial_health_cache` table from gap analysis §5.4).

---

## 5. Chat Streaming (BP from gap analysis §7.5)

**Gap:** Chat sends full response after completion. Standard UX expects streaming tokens.

**ML service / Gemini change:** Switch from full response to streaming:
- Use Gemini streaming API
- Stream chunks via Server-Sent Events: `Content-Type: text/event-stream`
- New endpoint: `POST /auth/v1/chat/stream`
- Keep `POST /auth/v1/chat` for backwards compatibility

**Priority:** Medium. Improves perceived responsiveness of the AI Coach.

---

## 6. AI Coach Context Consistency

**Gap (§9.4):** The 4 AI Coach panels (Chat, Health, Insights, Recommendations) can show contradictory data if fetched independently at different times.

**Suggestion:** Add a composite endpoint `GET /auth/v1/ai-coach/context` that returns all four in one response:

```json
{
  "health": { ... },
  "insights": { ... },
  "recommendations": { ... },
  "chat_context": "injected into system prompt"
}
```

This ensures the chatbot's system prompt and the displayed panels use identical data. The Go backend already injects financial context into Gemini via `BuildFinancialProfileContext` — align that context with what the panels show.

---

## Summary Priority Table

| # | Item | Service | Priority |
|---|------|---------|----------|
| 1 | Wire `/ml/insights` schema to frontend | ML + FE | High |
| 2 | Recommendations from ML vs rule-based | ML or BE | Low (rule-based done) |
| 3 | Forecast async job pattern | ML + BE | Medium |
| 4 | Financial health ML enhancement | ML | Low |
| 5 | Chat streaming | ML + BE | Medium |
| 6 | AI Coach context composite endpoint | BE | Medium |
