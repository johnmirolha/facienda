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
)

var addCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Add a new task",
	Long: `Add a new task to your todo list.

You can optionally specify a date or make the task recurring.

Examples:
  facienda add "Buy groceries"
  facienda add "Team meeting" --date 2025-11-20
  facienda add "Weekly report" --recur "every monday"
  facienda add "Pay rent" --recur "1st of each month"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]

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

			if err := store.Create(task); err != nil {
				return err
			}

			fmt.Printf("✓ Recurring task added (ID: %d)\n", task.ID)
			fmt.Printf("  Pattern: %s\n", pattern.String())
			fmt.Printf("  Next occurrence: %s\n", task.Date.Format("Mon, Jan 2, 2006"))
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

		if err := store.Create(task); err != nil {
			return err
		}

		fmt.Printf("✓ Task added (ID: %d)\n", task.ID)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addDate, "date", "d", "", "task date (YYYY-MM-DD, default: today)")
	addCmd.Flags().StringVarP(&addDetails, "details", "m", "", "task details")
	addCmd.Flags().StringVarP(&addRecur, "recur", "r", "", "recurrence pattern (e.g., 'every monday', '3rd of each month')")
	rootCmd.AddCommand(addCmd)
}
