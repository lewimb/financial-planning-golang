ALTER TABLE notification_preferences
  DROP COLUMN IF EXISTS weekly_summary,
  DROP COLUMN IF EXISTS push_enabled;
