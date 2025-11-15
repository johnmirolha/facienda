package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var completeCmd = &cobra.Command{
	Use:   "complete [task-id]",
	Short: "Mark a task as completed",
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

		task.Complete()
		if err := store.Update(task); err != nil {
			return err
		}

		fmt.Printf("✓ Task %d marked as completed\n", id)
		return nil
	},
}

var incompleteCmd = &cobra.Command{
	Use:   "incomplete [task-id]",
	Short: "Mark a task as incomplete",
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

		task.Incomplete()
		if err := store.Update(task); err != nil {
			return err
		}

		fmt.Printf("✓ Task %d marked as incomplete\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completeCmd)
	rootCmd.AddCommand(incompleteCmd)
}
