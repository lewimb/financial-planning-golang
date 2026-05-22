# Notification Flow

Added in v1.3 to support budget threshold alerts and in-app notification bell.

---

## Budget Alert Notification (triggered on transaction mutation)

```mermaid
flowchart TD
    A[Transaction Create / Update / Delete] --> B[Handler calls CheckBudgetAlerts async goroutine]
    B --> C[NotificationUseCase.CheckBudgetAlerts]
    C --> D{Get notification preferences}
    D -- budget_alerts disabled --> Z[Skip]
    D -- budget_alerts enabled --> E[BudgetRepo.GetUsage current month]
    E --> F{For each budget usage}
    F --> G{Status?}
    G -- SAFE --> F
    G -- WARNING --> H[notifType = BUDGET_WARNING]
    G -- EXCEEDED --> I[notifType = BUDGET_EXCEEDED]
    H & I --> J{ExistsRecent in last 24h?}
    J -- yes --> F
    J -- no --> K[NotificationRepo.Create]
    K --> F
```

---

## Frontend Notification Bell Flow

```mermaid
flowchart LR
    FE["Frontend (React)"] -->|GET /notifications?unread_only=true| API
    API --> DB[(notifications table)]
    DB --> API
    API -->|"{ data: [...], unread_count: N }"| FE
    FE --> Bell["Notification Bell\n(badge shows unread_count)"]
    Bell --> List["Notification List\nSorted newest first"]
    List -->|PATCH /notifications/:id/read| API
    List -->|PATCH /notifications/read-all| API
    List -->|DELETE /notifications/:id| API
```

---

## Notification Preference Management

```mermaid
flowchart TD
    User["User opens Settings → Notifications"] -->|GET /notifications/preferences| API
    API --> Prefs[(notification_preferences table)]
    Prefs --> API
    API --> Form["Settings Form\n(budget_alerts, goal_reminders, anomaly_alerts)"]
    Form -->|PUT /notifications/preferences| API
    API -->|UPSERT notification_preferences| Prefs
    API --> User
```
