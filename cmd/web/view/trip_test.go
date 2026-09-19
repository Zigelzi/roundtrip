package view

import (
	"testing"
	"time"
)

// Trips show as a Finnish-style date range, written as short as it stays
// unambiguous (01/S2, S7).
func TestFormatDateRange(t *testing.T) {
	tests := []struct {
		name      string
		departure string
		days      int64
		want      string
	}{
		{"same month", "2026-09-20", 3, "20.–22.9.2026"},
		{"across months", "2026-11-30", 4, "30.11.–3.12.2026"},
		{"across years", "2026-12-30", 5, "30.12.2026–3.1.2027"},
		{"one day", "2026-09-20", 1, "20.9.2026"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			departure, err := time.Parse("2006-01-02", tt.departure)
			if err != nil {
				t.Fatal(err)
			}
			if got := FormatDateRange(departure, tt.days); got != tt.want {
				t.Errorf("FormatDateRange(%s, %d) = %q, want %q", tt.departure, tt.days, got, tt.want)
			}
		})
	}
}
