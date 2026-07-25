Create an ERD for a financial planning app with the following tables and relationships:

**users**: id (PK), email (unique), name, password, first_name, last_name, phone, created_at, deleted_at

**transactions**: id (PK), user_id (FK → users), amount, category, type, date, description, is_recurring, recurrence_interval, created_at, updated_at, deleted_at

**budgets**: id (PK), user_id (FK → users), category, period, month, year, limit_amount, alert_threshold, created_at, updated_at, deleted_at

**goals**: id (PK), user_id (FK → users), name, target_amount, current_amount, deadline, status, description, created_at, updated_at

**ai_logs**: id (PK), user_id (FK → users), question, response, created_at, deleted_at

**user_financial_profiles**: id (PK), user_id (FK → users, unique), monthly_income, fixed_expenses, current_savings, debt, employment_status, spending_habit, risk_level, created_at, updated_at

**user_financial_goals**: id (PK), user_id (FK → users), goal_type

**notifications**: id (PK), user_id (FK → users), type, title, message, entity_type, entity_id, is_read, created_at

**notification_preferences**: id (PK), user_id (FK → users, unique), budget_alerts, goal_reminders, anomaly_alerts, weekly_summary, push_enabled, updated_at

**activity_logs**: id (PK), user_id (FK → users), action, entity_type, entity_id, description, created_at

Relationships:
- users has many transactions, budgets, goals, ai_logs, notifications, activity_logs, user_financial_goals
- users has one user_financial_profiles, one notification_preferences
