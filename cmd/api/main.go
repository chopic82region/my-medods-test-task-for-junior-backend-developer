package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	taskrepository "example.com/taskservice/internal/repository/postgres"
	transporthttp "example.com/taskservice/internal/transport/http"
	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

func main() {
	// Инициализация логгера
	logger := log.New(os.Stdout, "[TASK-SERVICE] ", log.LstdFlags|log.Lshortfile)

	// Загрузка конфигурации
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable")
	serverAddr := getEnv("SERVER_ADDR", ":8080")

	// Создание контекста с отменой для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Подключение к БД
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Проверка подключения (ИСПРАВЛЕНО)
	if err := dbPool.Ping(ctx); err != nil {
		logger.Fatalf("Unable to ping database: %v", err)
	}
	logger.Println("Successfully connected to database")

	// Инициализация компонентов (ИСПРАВЛЕНО)
	taskRepo := taskrepository.New(dbPool)
	taskService := taskusecase.NewService(taskRepo)         // Исправлено: taskusecase.NewService
	taskHandler := httphandlers.NewTaskHandler(taskService) // Исправлено: передан taskService
	docsHandler := swaggerdocs.NewHandler()

	// Настройка роутера
	router := transporthttp.NewRouter(taskHandler, docsHandler)

	// Настройка HTTP сервера
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск планировщика
	scheduler := taskusecase.NewScheduler(taskService, logger)
	scheduler.Start(ctx)
	defer scheduler.Stop()

	// Запуск HTTP сервера
	go func() {
		logger.Printf("Starting HTTP server on %s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Could not start server: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	<-ctx.Done()
	logger.Println("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	scheduler.Stop()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Printf("Server shutdown error: %v", err)
	}

	logger.Println("Server exited gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
