package commands

import (
	"fmt"
	"time"

	"github.com/johnmirolha/facienda/internal/storage"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List current tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := store.List(storage.FilterCurrent)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks for today.")
			return nil
		}

		fmt.Printf("Tasks for %s:\n\n", time.Now().Format("2006-01-02"))
		for _, task := range tasks {
			status := "[ ]"
			if task.Completed {
				status = "[✓]"
			}

			title := task.Title
			if task.IsRecurring() {
				title = fmt.Sprintf("%s ↻", task.Title)
			}

			fmt.Printf("%s %d. %s", status, task.ID, title)
			if len(task.Tags) > 0 {
				fmt.Printf(" %s", formatTagList(task.Tags))
			}
			fmt.Println()

			if task.Details != "" {
				fmt.Printf("   %s\n", task.Details)
			}
			if task.IsRecurring() {
				fmt.Printf("   Recurs: %s\n", task.RecurrencePattern.String())
			}
		}

		return nil
	},
}

var pastCmd = &cobra.Command{
	Use:   "past",
	Short: "View past tasks (timeline)",
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := store.List(storage.FilterPast)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("No past tasks.")
			return nil
		}

		fmt.Println("Past tasks:")
		currentDate := ""
		for _, task := range tasks {
			taskDate := task.Date.Format("2006-01-02")
			if taskDate != currentDate {
				currentDate = taskDate
				fmt.Printf("\n%s:\n", currentDate)
			}

			status := "[ ]"
			if task.Completed {
				status = "[✓]"
			}

			title := task.Title
			if task.IsRecurring() {
				title = fmt.Sprintf("%s ↻", task.Title)
			}

			fmt.Printf("%s %d. %s", status, task.ID, title)
			if len(task.Tags) > 0 {
				fmt.Printf(" %s", formatTagList(task.Tags))
			}
			fmt.Println()

			if task.Details != "" {
				fmt.Printf("   %s\n", task.Details)
			}
			if task.IsRecurring() {
				fmt.Printf("   Recurs: %s\n", task.RecurrencePattern.String())
			}
		}

		return nil
	},
}

var futureCmd = &cobra.Command{
	Use:   "future",
	Short: "View future tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := store.List(storage.FilterFuture)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("No future tasks.")
			return nil
		}

		fmt.Println("Future tasks:")
		currentDate := ""
		for _, task := range tasks {
			taskDate := task.Date.Format("2006-01-02")
			if taskDate != currentDate {
				currentDate = taskDate
				fmt.Printf("\n%s:\n", currentDate)
			}

			status := "[ ]"
			if task.Completed {
				status = "[✓]"
			}

			title := task.Title
			if task.IsRecurring() {
				title = fmt.Sprintf("%s ↻", task.Title)
			}

			fmt.Printf("%s %d. %s", status, task.ID, title)
			if len(task.Tags) > 0 {
				fmt.Printf(" %s", formatTagList(task.Tags))
			}
			fmt.Println()

			if task.Details != "" {
				fmt.Printf("   %s\n", task.Details)
			}
			if task.IsRecurring() {
				fmt.Printf("   Recurs: %s\n", task.RecurrencePattern.String())
			}
		}

		return nil
	},
}

var (
	searchTag string
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search tasks by tag",
	Long: `Search for tasks with a specific tag.

Examples:
  facienda search --tag work
  facienda search -t personal`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if searchTag == "" {
			return fmt.Errorf("--tag flag is required")
		}

		tasks, err := store.ListByTag(searchTag, storage.FilterAll)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Printf("No tasks found with tag '%s'.\n", searchTag)
			return nil
		}

		fmt.Printf("Tasks with tag '%s':\n", searchTag)
		currentDate := ""
		for _, task := range tasks {
			taskDate := task.Date.Format("2006-01-02")
			if taskDate != currentDate {
				currentDate = taskDate
				fmt.Printf("\n%s:\n", currentDate)
			}

			status := "[ ]"
			if task.Completed {
				status = "[✓]"
			}

			title := task.Title
			if task.IsRecurring() {
				title = fmt.Sprintf("%s ↻", task.Title)
			}

			fmt.Printf("%s %d. %s", status, task.ID, title)
			if len(task.Tags) > 0 {
				fmt.Printf(" %s", formatTagList(task.Tags))
			}
			fmt.Println()

			if task.Details != "" {
				fmt.Printf("   %s\n", task.Details)
			}
			if task.IsRecurring() {
				fmt.Printf("   Recurs: %s\n", task.RecurrencePattern.String())
			}
		}

		return nil
	},
}

func init() {
	searchCmd.Flags().StringVarP(&searchTag, "tag", "t", "", "tag to search for")
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(pastCmd)
	rootCmd.AddCommand(futureCmd)
	rootCmd.AddCommand(searchCmd)
}
