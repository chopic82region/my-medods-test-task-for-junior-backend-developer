package task

import (
	"context"
	"fmt"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

// NewService создаёт сервис задач
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

// Create создаёт новую задачу (шаблон или разовую)
func (s *Service) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	// Валидация
	if task.Title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if task.Status == "" {
		task.Status = taskdomain.StatusNew
	}
	if !task.Status.Valid() {
		return nil, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}
	// Валидация периодичности (опционально)
	if task.RecurrenceType != "" && task.RecurrenceType != "none" {
		rc := &task.RecurrenceConfig
		if err := rc.Validate(taskdomain.RecurrenceType(task.RecurrenceType)); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
	}

	now := s.now()
	task.CreatedAt = now
	task.UpdatedAt = now

	return s.repo.Create(ctx, task)
}

// GetByID возвращает задачу по ID
func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.GetByID(ctx, id)
}

// Update обновляет существующую задачу
func (s *Service) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	if task.ID <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	// Получаем существующую задачу
	existing, err := s.repo.GetByID(ctx, task.ID)
	if err != nil {
		return nil, err
	}

	// Обновляем только переданные поля
	if task.Title != "" {
		existing.Title = task.Title
	}
	if task.Description != "" {
		existing.Description = task.Description
	}
	if task.Status != "" && task.Status.Valid() {
		existing.Status = task.Status
	}
	if task.RecurrenceType != "" {
		existing.RecurrenceType = task.RecurrenceType
	}
	// Обновляем конфигурацию, если она не пустая
	if task.RecurrenceConfig.Interval != 0 || len(task.RecurrenceConfig.Days) > 0 ||
		len(task.RecurrenceConfig.Dates) > 0 || task.RecurrenceConfig.Parity != "" {
		existing.RecurrenceConfig = task.RecurrenceConfig
	}

	existing.UpdatedAt = s.now()

	return s.repo.Update(ctx, existing)
}

// Delete удаляет задачу по ID
func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.Delete(ctx, id)
}

// List возвращает все задачи
func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

// ListTemplates возвращает все шаблоны периодических задач
func (s *Service) ListTemplates(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.ListTemplates(ctx)
}

// ListInstances возвращает экземпляры шаблона
func (s *Service) ListInstances(ctx context.Context, parentID int64) ([]taskdomain.Task, error) {
	if parentID <= 0 {
		return nil, fmt.Errorf("%w: parent_id must be positive", ErrInvalidInput)
	}
	return s.repo.ListInstances(ctx, parentID)
}

// ListByRecurrenceType возвращает задачи по типу периодичности
func (s *Service) ListByRecurrenceType(ctx context.Context, recurrenceType string) ([]taskdomain.Task, error) {
	if recurrenceType == "" {
		return nil, fmt.Errorf("%w: recurrence_type is required", ErrInvalidInput)
	}
	return s.repo.ListByRecurrenceType(ctx, recurrenceType)
}

// ListByParentAndType возвращает экземпляры шаблона с фильтрацией по типу
func (s *Service) ListByParentAndType(ctx context.Context, parentID int64, recurrenceType string) ([]taskdomain.Task, error) {
	if parentID <= 0 {
		return nil, fmt.Errorf("%w: parent_id must be positive", ErrInvalidInput)
	}
	if recurrenceType == "" {
		return nil, fmt.Errorf("%w: recurrence_type is required", ErrInvalidInput)
	}
	return s.repo.ListByParentAndType(ctx, parentID, recurrenceType)
}

// GenerateInstances генерирует экземпляры задачи на период
func (s *Service) GenerateInstances(ctx context.Context, templateID int64, fromDate, toDate *time.Time) (int, error) {
	if templateID <= 0 {
		return 0, fmt.Errorf("%w: template_id must be positive", ErrInvalidInput)
	}

	template, err := s.repo.GetByID(ctx, templateID)
	if err != nil {
		return 0, err
	}

	if template.RecurrenceType == "" || template.RecurrenceType == "none" {
		return 0, fmt.Errorf("%w: task is not a recurring template", ErrInvalidInput)
	}

	now := s.now()
	from := now
	if fromDate != nil {
		from = *fromDate
	}
	to := now.AddDate(0, 1, 0)
	if toDate != nil {
		to = *toDate
	}

	dates, err := s.generateDates(template, from, to)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, date := range dates {
		exists, err := s.repo.InstanceExists(ctx, templateID, date)
		if err != nil {
			continue
		}
		if exists {
			continue
		}

		instance := &taskdomain.Task{
			Title:          template.Title,
			Description:    template.Description,
			Status:         taskdomain.StatusNew,
			CreatedAt:      date,
			UpdatedAt:      date,
			RecurrenceType: "none",
			ParentTaskID:   &templateID,
		}
		if _, err := s.repo.Create(ctx, instance); err != nil {
			continue
		}
		count++
	}
	return count, nil
}

// generateDates возвращает даты для шаблона в заданном диапазоне
func (s *Service) generateDates(template *taskdomain.Task, from, to time.Time) ([]time.Time, error) {
	var dates []time.Time
	switch template.RecurrenceType {
	case "daily":
		interval := template.RecurrenceConfig.Interval
		if interval < 1 {
			interval = 1
		}
		for d := from; !d.After(to); d = d.AddDate(0, 0, interval) {
			dates = append(dates, d)
		}
	case "monthly":
		for _, day := range template.RecurrenceConfig.Days {
			current := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
			for !current.After(to) {
				lastDay := time.Date(current.Year(), current.Month()+1, 0, 0, 0, 0, 0, current.Location()).Day()
				targetDay := day
				if targetDay > lastDay {
					targetDay = lastDay
				}
				targetDate := time.Date(current.Year(), current.Month(), targetDay, 0, 0, 0, 0, current.Location())
				if !targetDate.Before(from) && !targetDate.After(to) {
					dates = append(dates, targetDate)
				}
				current = current.AddDate(0, 1, 0)
			}
		}
	case "specific_dates":
		for _, dateStr := range template.RecurrenceConfig.Dates {
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				continue
			}
			if !date.Before(from) && !date.After(to) {
				dates = append(dates, date)
			}
		}
	case "even_odd":
		for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
			day := d.Day()
			isEven := day%2 == 0
			if (template.RecurrenceConfig.Parity == "even" && isEven) ||
				(template.RecurrenceConfig.Parity == "odd" && !isEven) {
				dates = append(dates, d)
			}
		}
	}
	return dates, nil
}
