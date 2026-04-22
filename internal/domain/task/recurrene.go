package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// RecurrenceType - тип периодичности задачи
type RecurrenceType string

const (
	// RecurrenceNone - разовая задача (без периодичности)
	RecurrenceNone RecurrenceType = "none"

	// RecurrenceDaily - ежедневные задачи (каждый n-й день)
	RecurrenceDaily RecurrenceType = "daily"

	// RecurrenceMonthly - ежемесячные задачи (на определенные числа)
	RecurrenceMonthly RecurrenceType = "monthly"

	// RecurrenceSpecificDates - задачи на конкретные даты
	RecurrenceSpecificDates RecurrenceType = "specific_dates"

	// RecurrenceEvenOdd - задачи на четные/нечетные дни
	RecurrenceEvenOdd RecurrenceType = "even_odd"
)

// String возвращает строковое представление типа
func (rt RecurrenceType) String() string {
	return string(rt)
}

// IsValid проверяет, является ли тип корректным
func (rt RecurrenceType) IsValid() bool {
	switch rt {
	case RecurrenceNone, RecurrenceDaily, RecurrenceMonthly,
		RecurrenceSpecificDates, RecurrenceEvenOdd:
		return true
	}
	return false
}

// DailyConfig - конфигурация для ежедневных задач
type DailyConfig struct {
	Interval int `json:"interval"` // каждый N-й день (1, 2, 3...)
}

// MonthlyConfig - конфигурация для ежемесячных задач
type MonthlyConfig struct {
	Days []int `json:"days"` // числа месяца (1-30)
}

// SpecificDatesConfig - конфигурация для конкретных дат
type SpecificDatesConfig struct {
	Dates []string `json:"dates"` // даты в формате "2006-01-02"
}

// EvenOddConfig - конфигурация для четных/нечетных дней
type EvenOddConfig struct {
	Parity string `json:"parity"` // "even" или "odd"
}

// RecurrenceConfig - общая конфигурация периодичности (для хранения в JSONB)
type RecurrenceConfig struct {
	// Daily
	Interval int `json:"interval,omitempty"`

	// Monthly
	Days []int `json:"days,omitempty"`

	// SpecificDates
	Dates []string `json:"dates,omitempty"`

	// EvenOdd
	Parity string `json:"parity,omitempty"`
}

// Validate проверяет корректность конфигурации в зависимости от типа
func (rc *RecurrenceConfig) Validate(recurrenceType RecurrenceType) error {
	switch recurrenceType {
	case RecurrenceDaily:
		if rc.Interval < 1 {
			return errors.New("daily recurrence requires interval >= 1")
		}

	case RecurrenceMonthly:
		if len(rc.Days) == 0 {
			return errors.New("monthly recurrence requires at least one day")
		}
		for _, day := range rc.Days {
			if day < 1 || day > 30 {
				return fmt.Errorf("invalid day %d: must be between 1 and 30", day)
			}
		}

	case RecurrenceSpecificDates:
		if len(rc.Dates) == 0 {
			return errors.New("specific dates recurrence requires at least one date")
		}
		for _, dateStr := range rc.Dates {
			if _, err := time.Parse("2006-01-02", dateStr); err != nil {
				return fmt.Errorf("invalid date format %s: expected YYYY-MM-DD", dateStr)
			}
		}

	case RecurrenceEvenOdd:
		if rc.Parity != "even" && rc.Parity != "odd" {
			return errors.New("even_odd recurrence requires parity 'even' or 'odd'")
		}

	case RecurrenceNone:
		// Нет конфигурации для разовых задач
		if rc.Interval != 0 || len(rc.Days) > 0 || len(rc.Dates) > 0 || rc.Parity != "" {
			return errors.New("none recurrence type should have no config")
		}
	}

	return nil
}

// ParseConfig парсит JSON из БД в структуру RecurrenceConfig
func ParseConfig(configJSON json.RawMessage) (*RecurrenceConfig, error) {
	if len(configJSON) == 0 {
		return &RecurrenceConfig{}, nil
	}

	var config RecurrenceConfig
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, fmt.Errorf("failed to parse recurrence config: %w", err)
	}

	return &config, nil
}

// MarshalConfig маршалит конфигурацию в JSON для БД
func MarshalConfig(config *RecurrenceConfig) (json.RawMessage, error) {
	if config == nil {
		return json.RawMessage("{}"), nil
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recurrence config: %w", err)
	}

	return json.RawMessage(data), nil
}
