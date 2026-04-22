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
	service *taskusecase.Service
}

func NewTaskHandler(service *taskusecase.Service) *TaskHandler {
	return &TaskHandler{service: service}
}

// Create создает новую задачу (шаблон или разовую)
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, errors.New("title is required"))
		return
	}
	if req.Status == "" {
		req.Status = taskdomain.StatusNew
	}
	if !req.Status.Valid() {
		writeError(w, http.StatusBadRequest, errors.New("invalid status"))
		return
	}
	if err := validateRecurrence(req.RecurrenceType, req.RecurrenceConfig); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	task := &taskdomain.Task{
		Title:          req.Title,
		Description:    req.Description,
		Status:         req.Status,
		RecurrenceType: req.RecurrenceType,
	}
	// Исправлено: разыменовываем указатель, если он не nil
	if req.RecurrenceConfig != nil {
		task.RecurrenceConfig = *req.RecurrenceConfig
	}

	created, err := h.service.Create(r.Context(), task)
	if err != nil {
		writeServiceError(w, err)
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
	task, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
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

	if req.Status != "" && !req.Status.Valid() {
		writeError(w, http.StatusBadRequest, errors.New("invalid status"))
		return
	}
	if err := validateRecurrence(req.RecurrenceType, req.RecurrenceConfig); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	task := &taskdomain.Task{
		ID:             id,
		Title:          req.Title,
		Description:    req.Description,
		Status:         req.Status,
		RecurrenceType: req.RecurrenceType,
	}
	if req.RecurrenceConfig != nil {
		task.RecurrenceConfig = *req.RecurrenceConfig
	}

	updated, err := h.service.Update(r.Context(), task)
	if err != nil {
		writeServiceError(w, err)
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
	if err := h.service.Delete(r.Context(), id); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// List возвращает список задач с фильтрацией
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	parentTaskIDStr := r.URL.Query().Get("parent_task_id")
	recurrenceType := r.URL.Query().Get("recurrence_type")

	var tasks []taskdomain.Task
	var err error

	switch {
	case parentTaskIDStr != "" && recurrenceType != "":
		parentID, parseErr := strconv.ParseInt(parentTaskIDStr, 10, 64)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid parent_task_id"))
			return
		}
		tasks, err = h.service.ListByParentAndType(r.Context(), parentID, recurrenceType)

	case parentTaskIDStr != "":
		parentID, parseErr := strconv.ParseInt(parentTaskIDStr, 10, 64)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid parent_task_id"))
			return
		}
		tasks, err = h.service.ListInstances(r.Context(), parentID)

	case recurrenceType != "":
		tasks, err = h.service.ListByRecurrenceType(r.Context(), recurrenceType)

	default:
		tasks, err = h.service.List(r.Context())
	}

	if err != nil {
		writeServiceError(w, err)
		return
	}

	resp := make([]taskDTO, 0, len(tasks))
	for i := range tasks {
		resp = append(resp, newTaskDTO(&tasks[i]))
	}
	writeJSON(w, http.StatusOK, resp)
}

// ---------- helpers ----------
func getIDFromRequest(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	rawID := vars["id"]
	if rawID == "" {
		return 0, errors.New("missing task id")
	}
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid task id")
	}
	return id, nil
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, taskdomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func validateRecurrence(recurrenceType string, config *taskdomain.RecurrenceConfig) error {
	if recurrenceType == "" || recurrenceType == "none" {
		return nil
	}
	if config == nil {
		return errors.New("recurrence_config required for recurring tasks")
	}
	switch recurrenceType {
	case "daily":
		if config.Interval < 1 {
			return errors.New("daily interval must be >= 1")
		}
	case "monthly":
		if len(config.Days) == 0 {
			return errors.New("monthly requires at least one day")
		}
		for _, d := range config.Days {
			if d < 1 || d > 30 {
				return errors.New("day must be between 1 and 30")
			}
		}
	case "specific_dates":
		if len(config.Dates) == 0 {
			return errors.New("specific_dates requires at least one date")
		}
	case "even_odd":
		if config.Parity != "even" && config.Parity != "odd" {
			return errors.New("parity must be 'even' or 'odd'")
		}
	default:
		return errors.New("unknown recurrence_type")
	}
	return nil
}
