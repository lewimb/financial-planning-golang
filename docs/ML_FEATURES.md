# ML Features — Product Context

This document explains what each ML feature does, why it matters, and how it improves the financial planning app.

---

## 1. Spending Forecast (`/forecast`)

### What it does

Uses **Facebook Prophet** (a time-series forecasting model) to predict how much the user will spend in the coming days, based on their historical transaction patterns.

Prophet detects weekly patterns (e.g. higher spending on weekends) and extrapolates them forward. The result is a day-by-day prediction for the next N days (default: 30).

### Example output

```json
{
  "predicted_monthly_spending": 1820000,
  "daily_forecast": [
    {"date": "2026-02-01", "predicted_amount": 58000},
    {"date": "2026-02-02", "predicted_amount": 62000}
  ]
}
```

### Why it matters to users

- **Budget awareness:** Users can see if they're on track to exceed their monthly budget before it happens.
- **Planning ahead:** Before a big purchase, the user can check if their predicted remaining spend leaves room.
- **Behavioral insight:** Seeing future projections based on current habits nudges users toward better decisions.

### How it improves the AI chatbot

The AI assistant can reference forecast data when answering questions like:

> "Am I going to overspend this month?"
> "Can I afford a ₦50,000 purchase this week?"

Without forecast data, the AI can only describe the past. With it, the AI can reason about the future.

---

## 2. Anomaly Detection (`/anomaly`)

### What it does

Uses **IsolationForest** (a scikit-learn unsupervised ML algorithm) to flag spending days that are statistically unusual compared to the user's normal pattern.

IsolationForest works by isolating data points — outliers are easier to isolate than normal points. It requires no labeled training data.

### Example output

```json
{
  "anomalies": [
    {"date": "2026-01-15", "amount": 875000}
  ],
  "summary": "You spent unusually high on 1 day(s)"
}
```

### Why it matters to users

- **Fraud awareness:** Unusually large transactions might indicate unauthorized use.
- **Impulse spending detection:** A spike in spending on a specific day prompts reflection.
- **Alerts:** The app can proactively notify users when anomalies are detected, rather than waiting for them to check.

### How it improves the AI chatbot

The AI can surface anomalies proactively:

> "I noticed you spent 18× your average on January 15th. Was that intentional?"

This turns passive data into an active, conversational financial coach.

---

## 3. Spending Analysis (`/analysis`)

### What it does

Aggregates expense transactions using **Pandas** to compute:

- **Total expense:** sum of all spending in the period
- **Average daily expense:** how much the user spends per day on average
- **Top category:** which category accounts for the most spend
- **Spending distribution:** breakdown of spend by category

### Example output

```json
{
  "total_expense": 450000,
  "avg_daily": 64285.71,
  "top_category": "food",
  "spending_distribution": {
    "food": 180000,
    "transport": 150000,
    "utilities": 120000
  }
}
```

### Why it matters to users

- **Habit visibility:** Most people don't know their actual spending patterns until they see the numbers.
- **Category focus:** Knowing that food accounts for 40% of spending helps users decide where to cut back.
- **Simple summary:** A single API call gives the AI everything it needs to describe the user's financial period.

### How it improves the AI chatbot

Analysis data is the foundation of nearly every financial question:

> "How much did I spend last month?"
> "What am I spending the most on?"
> "Is my daily average higher than usual?"

Without this data precomputed, the AI would have to reason over raw transaction lists. With structured analysis results, responses are faster, more accurate, and more specific.

---

## 4. Combined Value to the App

These three features work together to give users — and the AI — a complete picture of their finances:

| Question | Feature that answers it |
|----------|------------------------|
| "What did I spend?" | Analysis |
| "Was anything unusual?" | Anomaly detection |
| "What will I spend?" | Forecast |

### Financial awareness

Users move from reactive (checking balances after the fact) to proactive (understanding patterns and anticipating future spend). The ML layer makes data that existed in raw form actually useful.

### Decision-making support

A user asking "should I buy this?" gets a better answer when the AI has:
- Their current total spend (analysis)
- Whether they've had unusual spending recently (anomaly)
- What their projected spend looks like (forecast)

### AI chatbot quality

The AI assistant becomes significantly more useful when it can ground responses in quantitative data rather than general financial advice. The ML service is the pipeline that turns raw transactions into structured signals the AI can reason over.

---

## Technical Constraints Worth Knowing

| Constraint | Detail |
|-----------|--------|
| Minimum data for anomaly | 5 unique expense days |
| Minimum data for forecast | 1 day (fallback to daily-avg projection) |
| Forecast accuracy | Improves significantly with 30+ days of history |
| Currency | Amounts treated as raw numbers — no conversion |
| INCOME | Silently ignored by all three endpoints |
