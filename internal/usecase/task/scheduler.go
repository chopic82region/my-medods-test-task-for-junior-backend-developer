package task

import (
	"context"
	"log"
	"time"
)

type Scheduler struct {
	service   *Service
	logger    *log.Logger
	stopChan  chan struct{}
	isRunning bool
}

func NewScheduler(service *Service, logger *log.Logger) *Scheduler {
	return &Scheduler{
		service:  service,
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// Start запускает планировщик
func (s *Scheduler) Start(ctx context.Context) {
	if s.isRunning {
		s.logger.Println("Scheduler is already running")
		return
	}

	s.isRunning = true
	s.logger.Println("Starting task scheduler...")

	go s.run(ctx)
}

// Stop останавливает планировщик
func (s *Scheduler) Stop() {
	if !s.isRunning {
		return
	}
	s.isRunning = false
	close(s.stopChan)
	s.logger.Println("Task scheduler stopped")
}

// run основной цикл планировщика
func (s *Scheduler) run(ctx context.Context) {
	// Запускаем генерацию сразу при старте
	s.generateInstances(ctx)

	// Настраиваем интервал проверки (каждый час)
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.generateInstances(ctx)
		case <-s.stopChan:
			s.logger.Println("Scheduler loop stopped")
			return
		case <-ctx.Done():
			s.logger.Println("Scheduler context cancelled")
			return
		}
	}
}

// generateInstances генерирует экземпляры для всех активных шаблонов
func (s *Scheduler) generateInstances(ctx context.Context) {
	s.logger.Println("Running scheduled task instance generation...")

	// Получаем все шаблоны периодических задач
	templates, err := s.service.ListTemplates(ctx)
	if err != nil {
		s.logger.Printf("Failed to get templates: %v", err)
		return
	}

	now := time.Now()
	// Генерируем на 30 дней вперед
	toDate := now.AddDate(0, 0, 30)

	totalGenerated := 0

	for _, template := range templates {
		// Пропускаем шаблоны без настроек периодичности
		if template.RecurrenceType == "" || template.RecurrenceType == "none" {
			continue
		}

		count, err := s.service.GenerateInstances(ctx, template.ID, &now, &toDate)
		if err != nil {
			s.logger.Printf("Failed to generate instances for template %d: %v", template.ID, err)
			continue
		}

		if count > 0 {
			s.logger.Printf("Generated %d instances for template %d (%s)", count, template.ID, template.Title)
			totalGenerated += count
		}
	}

	if totalGenerated > 0 {
		s.logger.Printf("Total generated instances: %d", totalGenerated)
	} else {
		s.logger.Println("No new instances generated")
	}
}
