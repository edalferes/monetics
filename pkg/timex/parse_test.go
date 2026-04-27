package timex

import (
	"testing"
	"time"
)

func TestParseFlexibleDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		want    time.Time
	}{
		{name: "ISO date", input: "2026-01-15", want: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)},
		{name: "Brazilian format", input: "15/01/2026", want: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)},
		{name: "RFC3339", input: "2026-01-15T12:30:00Z", want: time.Date(2026, 1, 15, 12, 30, 0, 0, time.UTC)},
		{name: "no timezone", input: "2026-01-15T12:30:00", want: time.Date(2026, 1, 15, 12, 30, 0, 0, time.UTC)},
		{name: "invalid", input: "not-a-date", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFlexibleDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error mismatch: got %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
