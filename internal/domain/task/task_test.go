package task

import (
	"testing"
)

func TestStatus_Valid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"valid new", StatusNew, true},
		{"valid in_progress", StatusInProgress, true},
		{"valid done", StatusDone, true},
		{"invalid status", Status("invalid"), false},
		{"empty status", Status(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Valid(); got != tt.want {
				t.Errorf("Status.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecurrenceConfig_Validate(t *testing.T) {
	tests := []struct {
		name           string
		recurrenceType RecurrenceType
		config         RecurrenceConfig
		wantErr        bool
	}{
		{
			name:           "valid daily",
			recurrenceType: RecurrenceDaily,
			config:         RecurrenceConfig{Interval: 1},
			wantErr:        false,
		},
		{
			name:           "invalid daily - interval 0",
			recurrenceType: RecurrenceDaily,
			config:         RecurrenceConfig{Interval: 0},
			wantErr:        true,
		},
		{
			name:           "valid monthly",
			recurrenceType: RecurrenceMonthly,
			config:         RecurrenceConfig{Days: []int{5, 20}},
			wantErr:        false,
		},
		{
			name:           "invalid monthly - empty days",
			recurrenceType: RecurrenceMonthly,
			config:         RecurrenceConfig{Days: []int{}},
			wantErr:        true,
		},
		{
			name:           "invalid monthly - day 31",
			recurrenceType: RecurrenceMonthly,
			config:         RecurrenceConfig{Days: []int{31}},
			wantErr:        true,
		},
		{
			name:           "valid specific_dates",
			recurrenceType: RecurrenceSpecificDates,
			config:         RecurrenceConfig{Dates: []string{"2026-05-01", "2026-05-15"}},
			wantErr:        false,
		},
		{
			name:           "invalid specific_dates - empty",
			recurrenceType: RecurrenceSpecificDates,
			config:         RecurrenceConfig{Dates: []string{}},
			wantErr:        true,
		},
		{
			name:           "valid even_odd even",
			recurrenceType: RecurrenceEvenOdd,
			config:         RecurrenceConfig{Parity: "even"},
			wantErr:        false,
		},
		{
			name:           "valid even_odd odd",
			recurrenceType: RecurrenceEvenOdd,
			config:         RecurrenceConfig{Parity: "odd"},
			wantErr:        false,
		},
		{
			name:           "invalid even_odd",
			recurrenceType: RecurrenceEvenOdd,
			config:         RecurrenceConfig{Parity: "invalid"},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate(tt.recurrenceType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
