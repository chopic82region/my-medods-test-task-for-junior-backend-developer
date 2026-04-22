ALTER TABLE tasks DROP COLUMN IF EXISTS recurrence_type;
ALTER TABLE tasks DROP COLUMN IF EXISTS recurrence_config;
ALTER TABLE tasks DROP COLUMN IF EXISTS parent_task_id;