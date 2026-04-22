package transporthttp

import (
	"net/http"

	"github.com/gorilla/mux"

	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
)

func NewRouter(taskHandler *httphandlers.TaskHandler, docsHandler *swaggerdocs.Handler) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// Swagger документация
	router.HandleFunc("/swagger/openapi.json", docsHandler.ServeSpec).Methods(http.MethodGet)
	router.HandleFunc("/swagger/", docsHandler.ServeUI).Methods(http.MethodGet)
	router.HandleFunc("/swagger", docsHandler.RedirectToUI).Methods(http.MethodGet)

	// API v1
	api := router.PathPrefix("/api/v1").Subrouter()

	// Базовые CRUD операции
	api.HandleFunc("/tasks", taskHandler.Create).Methods(http.MethodPost)               // Создание задачи
	api.HandleFunc("/tasks", taskHandler.List).Methods(http.MethodGet)                  // Получение списка задач (с поддержкой фильтрации через query параметры)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.GetByID).Methods(http.MethodGet)   // Получение задачи по ID
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.Update).Methods(http.MethodPut)    // Обновление задачи
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.Delete).Methods(http.MethodDelete) // Удаление задачи

	return router
}
