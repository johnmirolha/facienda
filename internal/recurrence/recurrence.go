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
// - "1st weekday of the month", "first weekday of month", etc.
// - "2nd weekday of the month", "second weekday of month", etc.
// - "last weekend of the month", "last weekend of month", etc.
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

	// Nth weekday pattern: "1st weekday of the month", "2nd weekday of month", etc.
	nthWeekdayRegex := regexp.MustCompile(`^(?:(\d+)(?:st|nd|rd|th)|first|second|third|fourth|fifth)\s+weekday\s+of\s+(?:the\s+)?month$`)
	if matches := nthWeekdayRegex.FindStringSubmatch(input); matches != nil {
		var n int
		if matches[1] != "" {
			n, _ = strconv.Atoi(matches[1])
		} else {
			// Handle word forms
			switch {
			case strings.Contains(input, "first"):
				n = 1
			case strings.Contains(input, "second"):
				n = 2
			case strings.Contains(input, "third"):
				n = 3
			case strings.Contains(input, "fourth"):
				n = 4
			case strings.Contains(input, "fifth"):
				n = 5
			}
		}
		if n < 1 || n > 5 {
			return "", ErrInvalidPattern
		}
		return Pattern(fmt.Sprintf("monthly-nth-weekday:%d", n)), nil
	}

	// Last weekend pattern: "last weekend of the month", "last weekend of month"
	lastWeekendRegex := regexp.MustCompile(`^last\s+weekend\s+of\s+(?:the\s+)?month$`)
	if lastWeekendRegex.MatchString(input) {
		return Pattern("monthly-last-weekend"), nil
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

	// Handle special patterns without colons
	if p == "monthly-last-weekend" {
		return nextLastWeekendOccurrence(after)
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
	case "monthly-nth-weekday":
		n, err := strconv.Atoi(patternValue)
		if err != nil {
			return time.Time{}, ErrInvalidPattern
		}
		return nextNthWeekdayOccurrence(after, n)
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

// isWeekday returns true if the weekday is Monday-Friday
func isWeekday(weekday time.Weekday) bool {
	return weekday >= time.Monday && weekday <= time.Friday
}

// isWeekend returns true if the weekday is Saturday or Sunday
func isWeekend(weekday time.Weekday) bool {
	return weekday == time.Saturday || weekday == time.Sunday
}

// nextNthWeekdayOccurrence finds the next occurrence of the Nth weekday of a month
// For example, n=1 means the first weekday (Mon-Fri), n=2 means the second weekday, etc.
func nextNthWeekdayOccurrence(after time.Time, n int) (time.Time, error) {
	if n < 1 || n > 5 {
		return time.Time{}, ErrInvalidPattern
	}

	// Start with the first day of the current month
	year, month, _ := after.Date()
	current := time.Date(year, month, 1, 0, 0, 0, 0, after.Location())

	// Find the Nth weekday of the current month
	weekdayCount := 0
	for current.Month() == month {
		if isWeekday(current.Weekday()) {
			weekdayCount++
			if weekdayCount == n {
				// Found the Nth weekday of this month
				if current.After(after) {
					return current, nil
				}
				// The Nth weekday of this month has passed, move to next month
				break
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	// Move to the first day of next month
	current = time.Date(year, month, 1, 0, 0, 0, 0, after.Location()).AddDate(0, 1, 0)

	// Find the Nth weekday of the next month
	weekdayCount = 0
	for {
		if isWeekday(current.Weekday()) {
			weekdayCount++
			if weekdayCount == n {
				return current, nil
			}
		}
		current = current.AddDate(0, 0, 1)
	}
}

// nextLastWeekendOccurrence finds the next occurrence of the last weekend day of a month
// This will be either the last Saturday or last Sunday, whichever comes last
func nextLastWeekendOccurrence(after time.Time) (time.Time, error) {
	// Start with the current month
	year, month, _ := after.Date()

	// Get the last day of the current month
	firstOfNextMonth := time.Date(year, month, 1, 0, 0, 0, 0, after.Location()).AddDate(0, 1, 0)
	lastOfMonth := firstOfNextMonth.AddDate(0, 0, -1)

	// Find the last weekend day (Saturday or Sunday)
	current := lastOfMonth
	for !isWeekend(current.Weekday()) {
		current = current.AddDate(0, 0, -1)
	}

	// If this date is after 'after', return it
	if current.After(after) {
		return current, nil
	}

	// Otherwise, move to next month
	month = month + 1
	if month > 12 {
		month = 1
		year++
	}

	// Get the last day of the next month
	firstOfNextMonth = time.Date(year, month, 1, 0, 0, 0, 0, after.Location()).AddDate(0, 1, 0)
	lastOfMonth = firstOfNextMonth.AddDate(0, 0, -1)

	// Find the last weekend day
	current = lastOfMonth
	for !isWeekend(current.Weekday()) {
		current = current.AddDate(0, 0, -1)
	}

	return current, nil
}

// String returns a human-readable representation of the pattern
func (p Pattern) String() string {
	if p == PatternNone {
		return "none"
	}

	// Handle special patterns
	if p == "monthly-last-weekend" {
		return "Last weekend of each month"
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
	case "monthly-nth-weekday":
		ordinal := getOrdinal(parts[1])
		return fmt.Sprintf("%s weekday of each month", ordinal)
	default:
		return string(p)
	}
}

// getOrdinal converts a number string to its ordinal form
func getOrdinal(num string) string {
	switch num {
	case "1":
		return "1st"
	case "2":
		return "2nd"
	case "3":
		return "3rd"
	case "4":
		return "4th"
	case "5":
		return "5th"
	default:
		return num + "th"
	}
}

// IsRecurring returns true if the pattern represents a recurring task
func (p Pattern) IsRecurring() bool {
	return p != PatternNone
}
