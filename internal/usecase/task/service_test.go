package task

import (
	"context"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

// MockRepository для тестирования
type MockRepository struct {
	tasks  map[int64]*taskdomain.Task
	nextID int64
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		tasks:  make(map[int64]*taskdomain.Task),
		nextID: 1,
	}
}

func (m *MockRepository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	task.ID = m.nextID
	m.nextID++
	m.tasks[task.ID] = task
	return task, nil
}

func (m *MockRepository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	task, ok := m.tasks[id]
	if !ok {
		return nil, taskdomain.ErrNotFound
	}
	return task, nil
}

func (m *MockRepository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	if _, ok := m.tasks[task.ID]; !ok {
		return nil, taskdomain.ErrNotFound
	}
	m.tasks[task.ID] = task
	return task, nil
}

func (m *MockRepository) Delete(ctx context.Context, id int64) error {
	if _, ok := m.tasks[id]; !ok {
		return taskdomain.ErrNotFound
	}
	delete(m.tasks, id)
	return nil
}

func (m *MockRepository) List(ctx context.Context) ([]taskdomain.Task, error) {
	tasks := make([]taskdomain.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, *task)
	}
	return tasks, nil
}

func (m *MockRepository) ListTemplates(ctx context.Context) ([]taskdomain.Task, error) {
	tasks := make([]taskdomain.Task, 0)
	for _, task := range m.tasks {
		if task.RecurrenceType != "" && task.RecurrenceType != "none" && task.ParentTaskID == nil {
			tasks = append(tasks, *task)
		}
	}
	return tasks, nil
}

func (m *MockRepository) ListInstances(ctx context.Context, parentID int64) ([]taskdomain.Task, error) {
	tasks := make([]taskdomain.Task, 0)
	for _, task := range m.tasks {
		if task.ParentTaskID != nil && *task.ParentTaskID == parentID {
			tasks = append(tasks, *task)
		}
	}
	return tasks, nil
}

func (m *MockRepository) ListByRecurrenceType(ctx context.Context, recurrenceType string) ([]taskdomain.Task, error) {
	tasks := make([]taskdomain.Task, 0)
	for _, task := range m.tasks {
		if task.RecurrenceType == recurrenceType {
			tasks = append(tasks, *task)
		}
	}
	return tasks, nil
}

func (m *MockRepository) ListByParentAndType(ctx context.Context, parentID int64, recurrenceType string) ([]taskdomain.Task, error) {
	tasks := make([]taskdomain.Task, 0)
	for _, task := range m.tasks {
		if task.ParentTaskID != nil && *task.ParentTaskID == parentID && task.RecurrenceType == recurrenceType {
			tasks = append(tasks, *task)
		}
	}
	return tasks, nil
}

func (m *MockRepository) InstanceExists(ctx context.Context, templateID int64, date time.Time) (bool, error) {
	for _, task := range m.tasks {
		if task.ParentTaskID != nil && *task.ParentTaskID == templateID {
			if task.CreatedAt.Year() == date.Year() &&
				task.CreatedAt.Month() == date.Month() &&
				task.CreatedAt.Day() == date.Day() {
				return true, nil
			}
		}
	}
	return false, nil
}

func TestService_Create(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	ctx := context.Background()

	task := &taskdomain.Task{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      taskdomain.StatusNew,
	}

	created, err := service.Create(ctx, task)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.ID == 0 {
		t.Error("Create() ID should be set")
	}
	if created.Title != "Test Task" {
		t.Errorf("Create() Title = %v, want %v", created.Title, "Test Task")
	}
}

func TestService_CreateWithRecurrence(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	ctx := context.Background()

	task := &taskdomain.Task{
		Title:          "Daily Task",
		Description:    "Daily recurring task",
		Status:         taskdomain.StatusNew,
		RecurrenceType: "daily",
		RecurrenceConfig: taskdomain.RecurrenceConfig{
			Interval: 1,
		},
	}

	created, err := service.Create(ctx, task)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.RecurrenceType != "daily" {
		t.Errorf("RecurrenceType = %v, want daily", created.RecurrenceType)
	}
	if created.RecurrenceConfig.Interval != 1 {
		t.Errorf("Interval = %v, want 1", created.RecurrenceConfig.Interval)
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	ctx := context.Background()

	_, err := service.GetByID(ctx, 999)
	if err == nil {
		t.Error("GetByID() should return error for non-existent task")
	}
}
