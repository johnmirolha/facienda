package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	editTitle   string
	editDetails string
)

var editCmd = &cobra.Command{
	Use:   "edit [task-id]",
	Short: "Edit task details",
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

		title := task.Title
		if editTitle != "" {
			title = editTitle
		}

		details := task.Details
		if cmd.Flags().Changed("details") {
			details = editDetails
		}

		if err := task.Update(title, details); err != nil {
			return err
		}

		if err := store.Update(task); err != nil {
			return err
		}

		fmt.Printf("✓ Task %d updated\n", id)
		return nil
	},
}

func init() {
	editCmd.Flags().StringVarP(&editTitle, "title", "t", "", "new task title")
	editCmd.Flags().StringVarP(&editDetails, "details", "m", "", "new task details")
	rootCmd.AddCommand(editCmd)
}
