package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var skipCmd = &cobra.Command{
	Use:   "skip [task-id]",
	Short: "Skip a task",
	Long: `Skip a task without marking it as completed.

If the task is recurring, this will automatically create the next occurrence.
Skipped tasks won't appear in the task list.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid task ID: %w", err)
		}

		task, err := store.GetByID(id)
		if err != nil {
			return err
		}

		// Check if this is a recurring task
		isRecurring := task.IsRecurring()

		task.Skip()
		if err := store.Update(task); err != nil {
			return err
		}

		fmt.Printf("⊘ Task %d skipped\n", id)

		// If recurring, generate the next instance
		if isRecurring {
			nextTask, err := task.GenerateNextInstance()
			if err != nil {
				return fmt.Errorf("failed to generate next instance: %w", err)
			}

			if nextTask != nil {
				if err := store.Create(nextTask); err != nil {
					return fmt.Errorf("failed to create next instance: %w", err)
				}

				fmt.Printf("✓ Next occurrence created (ID: %d) for %s\n",
					nextTask.ID,
					nextTask.Date.Format("Mon, Jan 2, 2006"))
			}
		}

		return nil
	},
}

var unskipCmd = &cobra.Command{
	Use:   "unskip [task-id]",
	Short: "Unskip a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid task ID: %w", err)
		}

		task, err := store.GetByID(id)
		if err != nil {
			return err
		}

		task.Unskip()
		if err := store.Update(task); err != nil {
			return err
		}

		fmt.Printf("✓ Task %d unskipped\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(skipCmd)
	rootCmd.AddCommand(unskipCmd)
}
