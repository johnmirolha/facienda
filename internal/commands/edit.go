package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	editTitle   string
	editDetails string
	editTags    []string
)

var editCmd = &cobra.Command{
	Use:   "edit [task-id]",
	Short: "Edit task details",
	Long: `Edit a task's title, details, or tags.

Examples:
  facienda edit 5 --title "New title"
  facienda edit 5 --details "Updated details"
  facienda edit 5 --tags work,urgent
  facienda edit 5 --tags ""  (removes all tags)`,
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

		// Update tags if specified
		if cmd.Flags().Changed("tags") {
			tags, err := resolveTags(editTags)
			if err != nil {
				return err
			}
			if err := task.SetTags(tags); err != nil {
				return err
			}
		}

		if err := store.Update(task); err != nil {
			return err
		}

		fmt.Printf("✓ Task %d updated\n", id)
		if cmd.Flags().Changed("tags") {
			if len(task.Tags) > 0 {
				fmt.Printf("  Tags: %s\n", formatTagList(task.Tags))
			} else {
				fmt.Printf("  Tags: (none)\n")
			}
		}
		return nil
	},
}

func init() {
	editCmd.Flags().StringVarP(&editTitle, "title", "t", "", "new task title")
	editCmd.Flags().StringVarP(&editDetails, "details", "m", "", "new task details")
	editCmd.Flags().StringSliceVar(&editTags, "tags", []string{}, "tags (comma-separated)")
	rootCmd.AddCommand(editCmd)
}
