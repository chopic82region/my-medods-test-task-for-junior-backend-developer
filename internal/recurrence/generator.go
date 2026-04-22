package recurrence

import (
	"time"

	"example.com/taskservice/internal/domain/task"
)

type Generator interface {
	GeneratorDates(template task.Task, from, to time.Time) ([]time.Time, error)
}
