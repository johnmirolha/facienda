package recurrence

import (
	"testing"
	"time"
)

func TestParsePattern(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        Pattern
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "empty pattern",
			input:   "",
			want:    PatternNone,
			wantErr: false,
		},
		{
			name:    "every monday",
			input:   "every monday",
			want:    "weekly:monday",
			wantErr: false,
		},
		{
			name:    "every Tuesday with capital",
			input:   "Every Tuesday",
			want:    "weekly:tuesday",
			wantErr: false,
		},
		{
			name:    "every friday",
			input:   "every friday",
			want:    "weekly:friday",
			wantErr: false,
		},
		{
			name:    "3rd of each month",
			input:   "3rd of each month",
			want:    "monthly:3",
			wantErr: false,
		},
		{
			name:    "on 15th",
			input:   "on 15th",
			want:    "monthly:15",
			wantErr: false,
		},
		{
			name:    "1st of every month",
			input:   "1st of every month",
			want:    "monthly:1",
			wantErr: false,
		},
		{
			name:    "31st",
			input:   "31st",
			want:    "monthly:31",
			wantErr: false,
		},
		{
			name:        "invalid day name",
			input:       "every funday",
			want:        "",
			wantErr:     true,
			expectedErr: ErrInvalidPattern,
		},
		{
			name:        "invalid day number",
			input:       "32nd of each month",
			want:        "",
			wantErr:     true,
			expectedErr: ErrInvalidDay,
		},
		{
			name:        "invalid day number zero",
			input:       "0th of each month",
			want:        "",
			wantErr:     true,
			expectedErr: ErrInvalidDay,
		},
		{
			name:        "completely invalid",
			input:       "whenever I feel like it",
			want:        "",
			wantErr:     true,
			expectedErr: ErrInvalidPattern,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePattern(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.expectedErr != nil && err != tt.expectedErr {
				t.Errorf("ParsePattern() error = %v, expectedErr %v", err, tt.expectedErr)
			}
			if got != tt.want {
				t.Errorf("ParsePattern() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPattern_NextOccurrence_Weekly(t *testing.T) {
	tests := []struct {
		name    string
		pattern Pattern
		after   time.Time
		wantDay time.Weekday
		want    time.Time
	}{
		{
			name:    "next monday from sunday",
			pattern: "weekly:monday",
			after:   time.Date(2025, 11, 9, 12, 0, 0, 0, time.UTC), // Sunday
			wantDay: time.Monday,
			want:    time.Date(2025, 11, 10, 0, 0, 0, 0, time.UTC), // Monday
		},
		{
			name:    "next monday from monday",
			pattern: "weekly:monday",
			after:   time.Date(2025, 11, 10, 12, 0, 0, 0, time.UTC), // Monday
			wantDay: time.Monday,
			want:    time.Date(2025, 11, 17, 0, 0, 0, 0, time.UTC), // Next Monday
		},
		{
			name:    "next friday from monday",
			pattern: "weekly:friday",
			after:   time.Date(2025, 11, 10, 12, 0, 0, 0, time.UTC), // Monday
			wantDay: time.Friday,
			want:    time.Date(2025, 11, 14, 0, 0, 0, 0, time.UTC), // Friday
		},
		{
			name:    "next sunday from saturday",
			pattern: "weekly:sunday",
			after:   time.Date(2025, 11, 15, 12, 0, 0, 0, time.UTC), // Saturday
			wantDay: time.Sunday,
			want:    time.Date(2025, 11, 16, 0, 0, 0, 0, time.UTC), // Sunday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.pattern.NextOccurrence(tt.after)
			if err != nil {
				t.Errorf("NextOccurrence() error = %v", err)
				return
			}
			if got.Weekday() != tt.wantDay {
				t.Errorf("NextOccurrence() weekday = %v, want %v", got.Weekday(), tt.wantDay)
			}

			// Check exact date
			if !got.Equal(tt.want) {
				t.Errorf("NextOccurrence() = %v, want %v", got, tt.want)
			}

			// Check that time is set to start of day
			if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 {
				t.Errorf("NextOccurrence() time = %02d:%02d:%02d, want 00:00:00",
					got.Hour(), got.Minute(), got.Second())
			}
		})
	}
}

func TestPattern_NextOccurrence_Monthly(t *testing.T) {
	tests := []struct {
		name      string
		pattern   Pattern
		after     time.Time
		wantYear  int
		wantMonth time.Month
		wantDay   int
	}{
		{
			name:      "15th from 10th same month",
			pattern:   "monthly:15",
			after:     time.Date(2025, 11, 10, 12, 0, 0, 0, time.UTC),
			wantYear:  2025,
			wantMonth: time.November,
			wantDay:   15,
		},
		{
			name:      "15th from 15th same month",
			pattern:   "monthly:15",
			after:     time.Date(2025, 11, 15, 12, 0, 0, 0, time.UTC),
			wantYear:  2025,
			wantMonth: time.December,
			wantDay:   15,
		},
		{
			name:      "15th from 20th wraps to next month",
			pattern:   "monthly:15",
			after:     time.Date(2025, 11, 20, 12, 0, 0, 0, time.UTC),
			wantYear:  2025,
			wantMonth: time.December,
			wantDay:   15,
		},
		{
			name:      "1st of next month",
			pattern:   "monthly:1",
			after:     time.Date(2025, 11, 30, 12, 0, 0, 0, time.UTC),
			wantYear:  2025,
			wantMonth: time.December,
			wantDay:   1,
		},
		{
			name:      "31st wraps year",
			pattern:   "monthly:31",
			after:     time.Date(2025, 12, 15, 12, 0, 0, 0, time.UTC),
			wantYear:  2025,
			wantMonth: time.December,
			wantDay:   31,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.pattern.NextOccurrence(tt.after)
			if err != nil {
				t.Errorf("NextOccurrence() error = %v", err)
				return
			}

			if got.Year() != tt.wantYear {
				t.Errorf("NextOccurrence() year = %d, want %d", got.Year(), tt.wantYear)
			}
			if got.Month() != tt.wantMonth {
				t.Errorf("NextOccurrence() month = %v, want %v", got.Month(), tt.wantMonth)
			}
			if got.Day() != tt.wantDay {
				t.Errorf("NextOccurrence() day = %d, want %d", got.Day(), tt.wantDay)
			}

			// Check that time is set to start of day
			if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 {
				t.Errorf("NextOccurrence() time = %02d:%02d:%02d, want 00:00:00",
					got.Hour(), got.Minute(), got.Second())
			}
		})
	}
}

func TestPattern_String(t *testing.T) {
	tests := []struct {
		name    string
		pattern Pattern
		want    string
	}{
		{
			name:    "none pattern",
			pattern: PatternNone,
			want:    "none",
		},
		{
			name:    "weekly monday",
			pattern: "weekly:monday",
			want:    "Every Monday",
		},
		{
			name:    "weekly friday",
			pattern: "weekly:friday",
			want:    "Every Friday",
		},
		{
			name:    "monthly 15",
			pattern: "monthly:15",
			want:    "Day 15 of each month",
		},
		{
			name:    "monthly 1",
			pattern: "monthly:1",
			want:    "Day 1 of each month",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pattern.String(); got != tt.want {
				t.Errorf("Pattern.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPattern_IsRecurring(t *testing.T) {
	tests := []struct {
		name    string
		pattern Pattern
		want    bool
	}{
		{
			name:    "none pattern",
			pattern: PatternNone,
			want:    false,
		},
		{
			name:    "weekly pattern",
			pattern: "weekly:monday",
			want:    true,
		},
		{
			name:    "monthly pattern",
			pattern: "monthly:15",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pattern.IsRecurring(); got != tt.want {
				t.Errorf("Pattern.IsRecurring() = %v, want %v", got, tt.want)
			}
		})
	}
}
