package todo

import (
	"testing"
)

func TestValidateTagName(t *testing.T) {
	tests := []struct {
		name    string
		tagName string
		wantErr error
	}{
		{
			name:    "valid lowercase",
			tagName: "work",
			wantErr: nil,
		},
		{
			name:    "valid with numbers",
			tagName: "project123",
			wantErr: nil,
		},
		{
			name:    "valid with underscore",
			tagName: "home_office",
			wantErr: nil,
		},
		{
			name:    "valid with hyphen",
			tagName: "high-priority",
			wantErr: nil,
		},
		{
			name:    "valid complex",
			tagName: "my_tag-123",
			wantErr: nil,
		},
		{
			name:    "empty name",
			tagName: "",
			wantErr: ErrEmptyTagName,
		},
		{
			name:    "uppercase letters",
			tagName: "Work",
			wantErr: ErrInvalidTagName,
		},
		{
			name:    "with spaces",
			tagName: "work home",
			wantErr: ErrInvalidTagName,
		},
		{
			name:    "with special characters",
			tagName: "work@home",
			wantErr: ErrInvalidTagName,
		},
		{
			name:    "too long",
			tagName: "this_is_a_very_long_tag_name_that_exceeds_fifty_characters_limit",
			wantErr: ErrTagNameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTagName(tt.tagName)
			if err != tt.wantErr {
				t.Errorf("ValidateTagName(%q) error = %v, want %v", tt.tagName, err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeTagName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase",
			input:    "work",
			expected: "work",
		},
		{
			name:     "uppercase to lowercase",
			input:    "WORK",
			expected: "work",
		},
		{
			name:     "mixed case",
			input:    "WoRk",
			expected: "work",
		},
		{
			name:     "with leading/trailing spaces",
			input:    "  work  ",
			expected: "work",
		},
		{
			name:     "complex",
			input:    "  My-Tag_123  ",
			expected: "my-tag_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeTagName(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeTagName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNewTag(t *testing.T) {
	tests := []struct {
		name    string
		tagName string
		wantErr error
	}{
		{
			name:    "valid tag",
			tagName: "work",
			wantErr: nil,
		},
		{
			name:    "normalized uppercase",
			tagName: "WORK",
			wantErr: nil,
		},
		{
			name:    "empty name",
			tagName: "",
			wantErr: ErrEmptyTagName,
		},
		{
			name:    "invalid characters",
			tagName: "work@home",
			wantErr: ErrInvalidTagName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, err := NewTag(tt.tagName)
			if err != tt.wantErr {
				t.Errorf("NewTag(%q) error = %v, want %v", tt.tagName, err, tt.wantErr)
			}
			if err == nil {
				expectedName := NormalizeTagName(tt.tagName)
				if tag.Name != expectedName {
					t.Errorf("NewTag(%q).Name = %q, want %q", tt.tagName, tag.Name, expectedName)
				}
				if tag.CreatedAt.IsZero() {
					t.Error("NewTag().CreatedAt is zero")
				}
			}
		})
	}
}

func TestValidateTaskTags(t *testing.T) {
	tests := []struct {
		name     string
		tagCount int
		wantErr  error
	}{
		{
			name:     "no tags",
			tagCount: 0,
			wantErr:  nil,
		},
		{
			name:     "one tag",
			tagCount: 1,
			wantErr:  nil,
		},
		{
			name:     "five tags (max)",
			tagCount: 5,
			wantErr:  nil,
		},
		{
			name:     "six tags (too many)",
			tagCount: 6,
			wantErr:  ErrTooManyTags,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := make([]*Tag, tt.tagCount)
			for i := 0; i < tt.tagCount; i++ {
				tags[i] = &Tag{ID: int64(i + 1), Name: "tag"}
			}

			err := ValidateTaskTags(tags)
			if err != tt.wantErr {
				t.Errorf("ValidateTaskTags(%d tags) error = %v, want %v", tt.tagCount, err, tt.wantErr)
			}
		})
	}
}

func TestTaskSetTags(t *testing.T) {
	task := &Task{
		Title: "Test task",
	}

	// Test valid tags
	tags := []*Tag{
		{ID: 1, Name: "work"},
		{ID: 2, Name: "urgent"},
	}

	err := task.SetTags(tags)
	if err != nil {
		t.Errorf("SetTags() error = %v, want nil", err)
	}

	if len(task.Tags) != 2 {
		t.Errorf("len(task.Tags) = %d, want 2", len(task.Tags))
	}

	// Test too many tags
	tooManyTags := make([]*Tag, 6)
	for i := 0; i < 6; i++ {
		tooManyTags[i] = &Tag{ID: int64(i + 1), Name: "tag"}
	}

	err = task.SetTags(tooManyTags)
	if err != ErrTooManyTags {
		t.Errorf("SetTags(6 tags) error = %v, want %v", err, ErrTooManyTags)
	}
}
