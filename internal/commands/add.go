package commands

import (
	"fmt"
	"time"

	"github.com/johnmirolha/facienda/internal/recurrence"
	"github.com/johnmirolha/facienda/internal/todo"
	"github.com/spf13/cobra"
)

var (
	addDate    string
	addDetails string
	addRecur   string
	addTags    []string
)

var addCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Add a new task",
	Long: `Add a new task to your todo list.

You can optionally specify a date, tags, or make the task recurring.

Examples:
  facienda add "Buy groceries"
  facienda add "Team meeting" --date 2025-11-20
  facienda add "Weekly report" --recur "every monday" --tags work,important
  facienda add "Pay rent" --recur "1st of each month" --tags bills`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]

		// Resolve tags
		tags, err := resolveTags(addTags)
		if err != nil {
			return err
		}

		// Handle recurring tasks
		if addRecur != "" {
			pattern, err := recurrence.ParsePattern(addRecur)
			if err != nil {
				return fmt.Errorf("invalid recurrence pattern: %w\nExamples: 'every monday', '3rd of each month'", err)
			}

			task, err := todo.NewRecurringTask(title, addDetails, pattern)
			if err != nil {
				return err
			}

			task.Tags = tags

			if err := store.Create(task); err != nil {
				return err
			}

			fmt.Printf("✓ Recurring task added (ID: %d)\n", task.ID)
			fmt.Printf("  Pattern: %s\n", pattern.String())
			fmt.Printf("  Next occurrence: %s\n", task.Date.Format("Mon, Jan 2, 2006"))
			if len(tags) > 0 {
				fmt.Printf("  Tags: %s\n", formatTagList(tags))
			}
			return nil
		}

		// Handle regular tasks
		date := time.Now()
		if addDate != "" {
			parsedDate, err := time.Parse("2006-01-02", addDate)
			if err != nil {
				return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
			}
			date = parsedDate
		}

		task, err := todo.NewTask(title, addDetails, date)
		if err != nil {
			return err
		}

		task.Tags = tags

		if err := store.Create(task); err != nil {
			return err
		}

		fmt.Printf("✓ Task added (ID: %d)\n", task.ID)
		if len(tags) > 0 {
			fmt.Printf("  Tags: %s\n", formatTagList(tags))
		}
		return nil
	},
}

// resolveTags resolves tag names to tag objects, creating them if needed
func resolveTags(tagNames []string) ([]*todo.Tag, error) {
	if len(tagNames) == 0 {
		return nil, nil
	}

	var tags []*todo.Tag
	for _, name := range tagNames {
		name = todo.NormalizeTagName(name)
		if name == "" {
			continue
		}

		// Try to get existing tag
		tag, err := store.GetTagByName(name)
		if err == todo.ErrTagNotFound {
			// Create new tag
			tag, err = todo.NewTag(name)
			if err != nil {
				return nil, fmt.Errorf("invalid tag '%s': %w", name, err)
			}
			if err := store.CreateTag(tag); err != nil {
				return nil, fmt.Errorf("failed to create tag '%s': %w", name, err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("failed to get tag '%s': %w", name, err)
		}

		tags = append(tags, tag)
	}

	if err := todo.ValidateTaskTags(tags); err != nil {
		return nil, err
	}

	return tags, nil
}

// formatTagList formats a slice of tags for display
func formatTagList(tags []*todo.Tag) string {
	if len(tags) == 0 {
		return ""
	}

	names := make([]string, len(tags))
	for i, tag := range tags {
		names[i] = tag.Name
	}
	return fmt.Sprintf("[%s]", joinStrings(names, ", "))
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func init() {
	addCmd.Flags().StringVarP(&addDate, "date", "d", "", "task date (YYYY-MM-DD, default: today)")
	addCmd.Flags().StringVarP(&addDetails, "details", "m", "", "task details")
	addCmd.Flags().StringVarP(&addRecur, "recur", "r", "", "recurrence pattern (e.g., 'every monday', '3rd of each month')")
	addCmd.Flags().StringSliceVarP(&addTags, "tags", "t", []string{}, "tags (comma-separated, e.g., 'work,important')")
	rootCmd.AddCommand(addCmd)
}
