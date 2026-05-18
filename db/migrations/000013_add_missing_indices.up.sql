-- goals: GetAll, CountActive, GetSavingsTotal, GetUpcomingMilestones all filter on user_id;
-- CountActive and GetUpcomingMilestones also filter on deadline.
CREATE INDEX idx_goals_user          ON goals(user_id);
CREATE INDEX idx_goals_user_deadline ON goals(user_id, deadline);

-- budgets: existing idx_budgets_user covers user_id scans, but GetUsage also
-- filters on year and period — the composite makes that query index-only.
CREATE INDEX idx_budgets_user_year_period ON budgets(user_id, year, period);

-- ai_logs: user_id index for future chat-history queries.
CREATE INDEX idx_ailogs_user ON ai_logs(user_id);

-- reports: user_id index for future report-listing queries.
CREATE INDEX idx_reports_user ON reports(user_id);
