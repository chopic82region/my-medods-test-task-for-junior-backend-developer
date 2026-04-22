package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:            normalized.Title,
		Description:      normalized.Description,
		Status:           normalized.Status,
		RecurrenceType:   normalized.RecurrenceType,
		RecurrenceConfig: *normalized.RecurrenceConfig,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:               id,
		Title:            normalized.Title,
		Description:      normalized.Description,
		Status:           normalized.Status,
		RecurrenceType:   normalized.RecurrenceType,
		RecurrenceConfig: *normalized.RecurrenceConfig,
		UpdatedAt:        s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func (s *Service) ListTemplates(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.ListTemplates(ctx)
}

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

		_, err = s.repo.Create(ctx, instance)
		if err != nil {
			continue
		}
		count++
	}

	return count, nil
}

func (s *Service) generateDates(template *taskdomain.Task, from, to time.Time) ([]time.Time, error) {
	dates := make([]time.Time, 0)

	switch template.RecurrenceType {
	case "daily":
		interval := template.RecurrenceConfig.Interval
		if interval == 0 {
			interval = 1
		}
		for d := from; !d.After(to); d = d.AddDate(0, 0, interval) {
			dates = append(dates, d)
		}

	case "monthly":
		for _, day := range template.RecurrenceConfig.Days {
			current := from
			current = time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, current.Location())
			for !current.After(to) {
				targetDay := day
				lastDayOfMonth := time.Date(current.Year(), current.Month()+1, 0, 0, 0, 0, 0, current.Location()).Day()
				if targetDay > lastDayOfMonth {
					targetDay = lastDayOfMonth
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

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status != "" && !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}
