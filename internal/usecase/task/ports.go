package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	ListTemplates(ctx context.Context) ([]taskdomain.Task, error)
	ListInstances(ctx context.Context, parentID int64) ([]taskdomain.Task, error)
	ListByRecurrenceType(ctx context.Context, recurrenceType string) ([]taskdomain.Task, error)
	ListByParentAndType(ctx context.Context, parentID int64, recurrenceType string) ([]taskdomain.Task, error)
	InstanceExists(ctx context.Context, templateID int64, date time.Time) (bool, error)
}

type Usecase interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) // Изменено
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) // Изменено
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	ListTemplates(ctx context.Context) ([]taskdomain.Task, error)
	ListInstances(ctx context.Context, parentID int64) ([]taskdomain.Task, error)
	ListByRecurrenceType(ctx context.Context, recurrenceType string) ([]taskdomain.Task, error)
	ListByParentAndType(ctx context.Context, parentID int64, recurrenceType string) ([]taskdomain.Task, error)
	GenerateInstances(ctx context.Context, templateID int64, fromDate, toDate *time.Time) (int, error)
}

// Эти структуры больше не нужны для Usecase, но могут быть полезны для внутреннего использования
type CreateInput struct {
	Title            string
	Description      string
	Status           taskdomain.Status
	RecurrenceType   string
	RecurrenceConfig *taskdomain.RecurrenceConfig
}

type UpdateInput struct {
	Title            string
	Description      string
	Status           taskdomain.Status
	RecurrenceType   string
	RecurrenceConfig *taskdomain.RecurrenceConfig
}
