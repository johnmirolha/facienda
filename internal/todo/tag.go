package todo

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrEmptyTagName      = errors.New("tag name cannot be empty")
	ErrInvalidTagName    = errors.New("tag name can only contain lowercase letters, numbers, underscores, and hyphens")
	ErrTagNameTooLong    = errors.New("tag name cannot exceed 50 characters")
	ErrTooManyTags       = errors.New("task cannot have more than 5 tags")
	ErrTagNotFound       = errors.New("tag not found")
	ErrTagAlreadyExists  = errors.New("tag already exists")
	ErrTagInUse          = errors.New("tag is in use by one or more tasks")
)

// Tag represents a tag that can be associated with tasks
type Tag struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

// tagNameRegex matches valid tag names: lowercase letters, numbers, underscore, and hyphen
var tagNameRegex = regexp.MustCompile(`^[a-z0-9_-]+$`)

// ValidateTagName validates a tag name according to the rules
func ValidateTagName(name string) error {
	if name == "" {
		return ErrEmptyTagName
	}

	if len(name) > 50 {
		return ErrTagNameTooLong
	}

	if !tagNameRegex.MatchString(name) {
		return ErrInvalidTagName
	}

	return nil
}

// NormalizeTagName converts a tag name to lowercase and trims whitespace
func NormalizeTagName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// NewTag creates a new tag with validation
func NewTag(name string) (*Tag, error) {
	name = NormalizeTagName(name)

	if err := ValidateTagName(name); err != nil {
		return nil, err
	}

	return &Tag{
		Name:      name,
		CreatedAt: time.Now(),
	}, nil
}

// ValidateTaskTags checks if the number of tags is within the allowed limit
func ValidateTaskTags(tags []*Tag) error {
	if len(tags) > 5 {
		return ErrTooManyTags
	}
	return nil
}
