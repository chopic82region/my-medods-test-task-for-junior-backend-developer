package postgres // Уже правильно

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository { // Правильное имя конструктора
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, created_at, updated_at, 
		                   recurrence_type, recurrence_config, parent_task_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	var recurrenceConfigJSON []byte
	var err error

	if task.RecurrenceConfig.Interval != 0 || len(task.RecurrenceConfig.Days) > 0 ||
		len(task.RecurrenceConfig.Dates) > 0 || task.RecurrenceConfig.Parity != "" {
		recurrenceConfigJSON, err = json.Marshal(task.RecurrenceConfig)
		if err != nil {
			return nil, err
		}
	}

	err = r.pool.QueryRow(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.CreatedAt,
		task.UpdatedAt,
		task.RecurrenceType,
		recurrenceConfigJSON,
		task.ParentTaskID,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return task, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at,
		       recurrence_type, recurrence_config, parent_task_id
		FROM tasks
		WHERE id = $1
	`

	return scanTask(r.pool.QueryRow(ctx, query, id))
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks 
		SET title = $1, description = $2, status = $3, updated_at = $4,
		    recurrence_type = $5, recurrence_config = $6, parent_task_id = $7
		WHERE id = $8
		RETURNING updated_at
	`

	var recurrenceConfigJSON []byte
	var err error

	if task.RecurrenceConfig.Interval != 0 || len(task.RecurrenceConfig.Days) > 0 ||
		len(task.RecurrenceConfig.Dates) > 0 || task.RecurrenceConfig.Parity != "" {
		recurrenceConfigJSON, err = json.Marshal(task.RecurrenceConfig)
		if err != nil {
			return nil, err
		}
	}

	err = r.pool.QueryRow(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.UpdatedAt,
		task.RecurrenceType,
		recurrenceConfigJSON,
		task.ParentTaskID,
		task.ID,
	).Scan(&task.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return task, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at, 
		       recurrence_type, recurrence_config, parent_task_id
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *Repository) ListByRecurrenceType(ctx context.Context, recurrenceType string) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at,
		       recurrence_type, recurrence_config, parent_task_id
		FROM tasks
		WHERE recurrence_type = $1
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query, recurrenceType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	return tasks, nil
}

func (r *Repository) ListByParentAndType(ctx context.Context, parentID int64, recurrenceType string) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at,
		       recurrence_type, recurrence_config, parent_task_id
		FROM tasks
		WHERE parent_task_id = $1 AND recurrence_type = $2
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, parentID, recurrenceType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	return tasks, nil
}

// InstanceExists проверяет, существует ли уже экземпляр задачи на указанную дату
func (r *Repository) InstanceExists(ctx context.Context, templateID int64, date time.Time) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1 FROM tasks 
			WHERE parent_task_id = $1 AND DATE(created_at) = DATE($2)
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, templateID, date).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *Repository) ListInstances(ctx context.Context, parentID int64) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at,
		       recurrence_type, recurrence_config, parent_task_id
		FROM tasks
		WHERE parent_task_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	return tasks, nil
}

func (r *Repository) ListTemplates(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at,
		       recurrence_type, recurrence_config, parent_task_id
		FROM tasks
		WHERE recurrence_type IS NOT NULL AND recurrence_type != 'none' AND parent_task_id IS NULL
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	return tasks, nil
}

type TaskScanner interface {
	Scan(dest ...any) error
}

func scanTask(row pgx.Row) (*taskdomain.Task, error) {
	var task taskdomain.Task
	var recurrenceConfigJSON []byte
	var parentTaskID *int64

	err := row.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.RecurrenceType,
		&recurrenceConfigJSON,
		&parentTaskID,
	)
	if err != nil {
		return nil, err
	}

	// Парсим recurrence_config из JSON
	if len(recurrenceConfigJSON) > 0 {
		if err := json.Unmarshal(recurrenceConfigJSON, &task.RecurrenceConfig); err != nil {
			return nil, err
		}
	}

	task.ParentTaskID = parentTaskID

	return &task, nil
}
