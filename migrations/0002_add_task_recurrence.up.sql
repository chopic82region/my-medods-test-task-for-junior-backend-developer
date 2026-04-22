-- ИСПРАВЛЕНО: было recuerrenc_type, стало recurrence_type
ALTER TABLE tasks ADD COLUMN recurrence_type VARCHAR(20) NOT NULL DEFAULT 'none';
ALTER TABLE tasks ADD COLUMN recurrence_config JSONB;
ALTER TABLE tasks ADD COLUMN parent_task_id BIGINT REFERENCES tasks(id) ON DELETE CASCADE; 

COMMENT ON COLUMN tasks.recurrence_type IS 'Тип периодичности: none, daily, monthly, specific_dates, even_odd';
COMMENT ON COLUMN tasks.recurrence_config IS 'JSON с параметрами периодичности (например, {"interval": 3} для daily, {"days": [5,15,25]} для monthly)';
COMMENT ON COLUMN tasks.parent_task_id IS 'ID шаблона-родителя для периодических задач';