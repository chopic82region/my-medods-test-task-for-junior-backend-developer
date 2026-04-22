package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type TaskHandler struct {
	usecase taskusecase.Usecase
}

func NewTaskHandler(usecase taskusecase.Usecase) *TaskHandler {
	return &TaskHandler{usecase: usecase}
}

// Create создает новую задачу (шаблон или разовую)
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	// Валидация периодичности
	if err := validateRecurrence(req.RecurrenceType, req.RecurrenceConfig); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), taskusecase.CreateInput{
		Title:            req.Title,
		Description:      req.Description,
		Status:           req.Status,
		RecurrenceType:   req.RecurrenceType,
		RecurrenceConfig: req.RecurrenceConfig,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newTaskDTO(created))
}

// GetByID возвращает задачу по ID
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskDTO(task))
}

// Update обновляет существующую задачу
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	// Валидация периодичности при обновлении
	if err := validateRecurrence(req.RecurrenceType, req.RecurrenceConfig); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, taskusecase.UpdateInput{
		Title:            req.Title,
		Description:      req.Description,
		Status:           req.Status,
		RecurrenceType:   req.RecurrenceType,
		RecurrenceConfig: req.RecurrenceConfig,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskDTO(updated))
}

// Delete удаляет задачу по ID
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// List возвращает список задач с поддержкой фильтрации
// Поддерживаемые query параметры:
//   - parent_task_id: фильтр по ID родительского шаблона (получить экземпляры)
//   - recurrence_type: фильтр по типу периодичности (daily, monthly, specific_dates, even_odd)
//
// List возвращает список задач с поддержкой фильтрации
// Поддерживаемые query параметры:
//   - parent_task_id: фильтр по ID родительского шаблона (получить экземпляры)
//   - recurrence_type: фильтр по типу периодичности (daily, monthly, specific_dates, even_odd)
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры фильтрации из query string
	parentTaskIDStr := r.URL.Query().Get("parent_task_id")
	recurrenceType := r.URL.Query().Get("recurrence_type")

	var tasks []taskdomain.Task
	var err error

	// Применяем фильтрацию в зависимости от переданных параметров
	switch {
	case parentTaskIDStr != "" && recurrenceType != "":
		// Если переданы оба параметра - фильтруем по обоим
		parentID, parseErr := strconv.ParseInt(parentTaskIDStr, 10, 64)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid parent_task_id format"))
			return
		}
		tasks, err = h.usecase.ListByParentAndType(r.Context(), parentID, recurrenceType)

	case parentTaskIDStr != "":
		// Фильтр по родительскому шаблону (получить экземпляры)
		parentID, parseErr := strconv.ParseInt(parentTaskIDStr, 10, 64)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid parent_task_id format"))
			return
		}
		tasks, err = h.usecase.ListInstances(r.Context(), parentID)

	case recurrenceType != "":
		// Фильтр по типу периодичности
		tasks, err = h.usecase.ListByRecurrenceType(r.Context(), recurrenceType)

	default:
		// Без фильтров - возвращаем все задачи
		tasks, err = h.usecase.List(r.Context())
	}

	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	// Конвертируем задачи в DTO
	response := make([]taskDTO, 0, len(tasks))
	for i := range tasks {
		response = append(response, newTaskDTO(&tasks[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

// ========== Вспомогательные функции ==========

// getIDFromRequest извлекает ID задачи из URL параметров
func getIDFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing task id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, errors.New("invalid task id")
	}

	if id <= 0 {
		return 0, errors.New("invalid task id")
	}

	return id, nil
}

// decodeJSON декодирует JSON из тела запроса
func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

// writeUsecaseError обрабатывает ошибки usecase и отправляет соответствующий HTTP статус
func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, taskdomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

// writeError отправляет ошибку в JSON формате
func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}

// writeJSON отправляет JSON ответ
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}

// validateRecurrence проверяет корректность конфигурации периодичности
func validateRecurrence(recurrenceType string, config *taskdomain.RecurrenceConfig) error {
	// Если тип не указан или "none", то периодичности нет
	if recurrenceType == "" || recurrenceType == "none" {
		return nil
	}

	// Для периодических задач конфиг обязателен
	if config == nil {
		return errors.New("recurrence_config is required for recurring tasks")
	}

	// Валидация в зависимости от типа
	switch recurrenceType {
	case "daily":
		if config.Interval < 1 {
			return errors.New("daily recurrence requires interval >= 1")
		}

	case "monthly":
		if len(config.Days) == 0 {
			return errors.New("monthly recurrence requires at least one day")
		}
		for _, day := range config.Days {
			if day < 1 || day > 30 {
				return errors.New("days must be between 1 and 30")
			}
		}

	case "specific_dates":
		if len(config.Dates) == 0 {
			return errors.New("specific_dates recurrence requires at least one date")
		}

	case "even_odd":
		if config.Parity != "even" && config.Parity != "odd" {
			return errors.New("parity must be 'even' or 'odd'")
		}

	default:
		return errors.New("invalid recurrence_type: must be one of 'daily', 'monthly', 'specific_dates', 'even_odd'")
	}

	return nil
}
