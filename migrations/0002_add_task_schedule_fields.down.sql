ALTER TABLE tasks
DROP COLUMN IF EXISTS recurrence_config,
DROP COLUMN IF EXISTS recurrence_type,
DROP COLUMN IF EXISTS scheduled_at;