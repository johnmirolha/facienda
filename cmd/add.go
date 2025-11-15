package cmd

import (
	"fmt"
	"time"

	"github.com/johnmirolha/facienda/internal/todo"
	"github.com/spf13/cobra"
)

var (
	addDate    string
	addDetails string
)

var addCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Add a new task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]

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
	rootCmd.AddCommand(addCmd)
}
