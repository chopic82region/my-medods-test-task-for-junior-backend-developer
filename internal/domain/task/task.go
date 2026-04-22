package task

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// параметры переодтчности
	RecurrenceType   string           `json:"recurrence_type" db:"recurrence_type"`
	RecurrenceConfig RecurrenceConfig `json:"recurrence_config" db:"recurrence_config"`
	ParentTaskID     *int64           `json:"parent_task_id,omitempty" db:"parent_task_id"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

// Scan для RecurrenceConfig (чтобы работать с JSONB в PostgreSQL)
func (rc *RecurrenceConfig) Scan(value interface{}) error {
	if value == nil {
		*rc = RecurrenceConfig{}
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return nil
	}

	return json.Unmarshal(data, rc)
}

// Value для RecurrenceConfig (чтобы сохранять в БД)
func (rc RecurrenceConfig) Value() (driver.Value, error) {
	if rc.Interval == 0 && len(rc.Days) == 0 && len(rc.Dates) == 0 && rc.Parity == "" {
		return nil, nil
	}
	return json.Marshal(rc)
}
