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
			fmt.Printf("%s %d. %s\n", status, task.ID, task.Title)
			if task.Details != "" {
				fmt.Printf("   %s\n", task.Details)
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
			fmt.Printf("%s %d. %s\n", status, task.ID, task.Title)
			if task.Details != "" {
				fmt.Printf("   %s\n", task.Details)
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
			fmt.Printf("%s %d. %s\n", status, task.ID, task.Title)
			if task.Details != "" {
				fmt.Printf("   %s\n", task.Details)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(pastCmd)
	rootCmd.AddCommand(futureCmd)
}
