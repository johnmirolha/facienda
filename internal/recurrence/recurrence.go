package recurrence

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidPattern = errors.New("invalid recurrence pattern")
	ErrInvalidDay     = errors.New("invalid day for monthly recurrence")
)

// Pattern represents a recurrence pattern
type Pattern string

const (
	PatternNone Pattern = ""
)

// ParsePattern parses a user-friendly recurrence string into a Pattern
// Supported formats:
// - "every monday", "every tuesday", etc.
// - "3rd of each month", "on 15th of every month", "15th of month", etc.
func ParsePattern(input string) (Pattern, error) {
	if input == "" {
		return PatternNone, nil
	}

	input = strings.ToLower(strings.TrimSpace(input))

	// Weekly pattern: "every monday", "every tuesday", etc.
	weeklyRegex := regexp.MustCompile(`^every\s+(monday|tuesday|wednesday|thursday|friday|saturday|sunday)$`)
	if matches := weeklyRegex.FindStringSubmatch(input); matches != nil {
		dayName := matches[1]
		return Pattern(fmt.Sprintf("weekly:%s", dayName)), nil
	}

	// Monthly pattern: "3rd of each month", "on 15th", "15th of month", etc.
	monthlyRegex := regexp.MustCompile(`^(?:on\s+)?(\d{1,2})(?:st|nd|rd|th)?(?:\s+of\s+(?:each|every)\s+month)?$`)
	if matches := monthlyRegex.FindStringSubmatch(input); matches != nil {
		dayNum, err := strconv.Atoi(matches[1])
		if err != nil || dayNum < 1 || dayNum > 31 {
			return "", ErrInvalidDay
		}
		return Pattern(fmt.Sprintf("monthly:%d", dayNum)), nil
	}

	return "", ErrInvalidPattern
}

// NextOccurrence calculates the next occurrence date after the given date
// based on the recurrence pattern
func (p Pattern) NextOccurrence(after time.Time) (time.Time, error) {
	if p == PatternNone {
		return time.Time{}, ErrInvalidPattern
	}

	parts := strings.Split(string(p), ":")
	if len(parts) != 2 {
		return time.Time{}, ErrInvalidPattern
	}

	patternType := parts[0]
	patternValue := parts[1]

	switch patternType {
	case "weekly":
		return nextWeeklyOccurrence(after, patternValue)
	case "monthly":
		dayNum, err := strconv.Atoi(patternValue)
		if err != nil {
			return time.Time{}, ErrInvalidPattern
		}
		return nextMonthlyOccurrence(after, dayNum)
	default:
		return time.Time{}, ErrInvalidPattern
	}
}

// nextWeeklyOccurrence finds the next occurrence of a specific weekday
func nextWeeklyOccurrence(after time.Time, dayName string) (time.Time, error) {
	targetWeekday := parseWeekday(dayName)
	if targetWeekday == -1 {
		return time.Time{}, ErrInvalidPattern
	}

	// Start from the day after 'after'
	current := after.AddDate(0, 0, 1)

	// Find the next occurrence of the target weekday
	for current.Weekday() != targetWeekday {
		current = current.AddDate(0, 0, 1)
	}

	// Set to start of day
	year, month, day := current.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, current.Location()), nil
}

// nextMonthlyOccurrence finds the next occurrence of a specific day of month
func nextMonthlyOccurrence(after time.Time, dayNum int) (time.Time, error) {
	if dayNum < 1 || dayNum > 31 {
		return time.Time{}, ErrInvalidDay
	}

	// Start with the next month
	year, month, _ := after.Date()
	current := time.Date(year, month, dayNum, 0, 0, 0, 0, after.Location())

	// If the date in current month has already passed, move to next month
	if !current.After(after) {
		current = current.AddDate(0, 1, 0)
	}

	// Handle months with fewer days (e.g., February, months with 30 days)
	// If the day doesn't exist in the target month, use the last day of that month
	for current.Day() != dayNum {
		// This means we've wrapped to the next month (e.g., Feb 31 -> Mar 3)
		// Go back to the last day of the previous month
		current = time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, current.Location()).AddDate(0, 0, -1)
	}

	return current, nil
}

// parseWeekday converts day name to time.Weekday
func parseWeekday(dayName string) time.Weekday {
	switch strings.ToLower(dayName) {
	case "sunday":
		return time.Sunday
	case "monday":
		return time.Monday
	case "tuesday":
		return time.Tuesday
	case "wednesday":
		return time.Wednesday
	case "thursday":
		return time.Thursday
	case "friday":
		return time.Friday
	case "saturday":
		return time.Saturday
	default:
		return -1
	}
}

// String returns a human-readable representation of the pattern
func (p Pattern) String() string {
	if p == PatternNone {
		return "none"
	}

	parts := strings.Split(string(p), ":")
	if len(parts) != 2 {
		return string(p)
	}

	switch parts[0] {
	case "weekly":
		return fmt.Sprintf("Every %s", strings.Title(parts[1]))
	case "monthly":
		return fmt.Sprintf("Day %s of each month", parts[1])
	default:
		return string(p)
	}
}

// IsRecurring returns true if the pattern represents a recurring task
func (p Pattern) IsRecurring() bool {
	return p != PatternNone
}
