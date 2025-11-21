package commands

import (
	"fmt"
	"strings"

	"github.com/johnmirolha/facienda/internal/todo"
	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags",
	Long:  "Create, list, rename, and delete tags for organizing your tasks.",
}

var tagCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new tag",
	Long: `Create a new tag that can be associated with tasks.

Tag names must contain only lowercase letters, numbers, underscores, and hyphens.

Examples:
  facienda tag create work
  facienda tag create personal
  facienda tag create high-priority`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		tag, err := todo.NewTag(name)
		if err != nil {
			return err
		}

		if err := store.CreateTag(tag); err != nil {
			return err
		}

		fmt.Printf("✓ Tag created: %s\n", tag.Name)
		return nil
	},
}

var tagListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags",
	Long:  "Display all available tags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		tags, err := store.ListTags()
		if err != nil {
			return err
		}

		if len(tags) == 0 {
			fmt.Println("No tags found.")
			return nil
		}

		fmt.Printf("Tags (%d):\n", len(tags))
		for _, tag := range tags {
			count, err := store.CountTasksWithTag(tag.ID)
			if err != nil {
				return err
			}
			fmt.Printf("  • %s (%d task%s)\n", tag.Name, count, pluralize(count))
		}

		return nil
	},
}

var tagDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a tag",
	Long: `Delete a tag.

Note: You cannot delete a tag that is currently associated with tasks.
Remove the tag from all tasks first, then delete it.

Examples:
  facienda tag delete work
  facienda tag delete old-tag`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := todo.NormalizeTagName(args[0])

		tag, err := store.GetTagByName(name)
		if err != nil {
			return err
		}

		if err := store.DeleteTag(tag.ID); err != nil {
			if err == todo.ErrTagInUse {
				count, _ := store.CountTasksWithTag(tag.ID)
				return fmt.Errorf("cannot delete tag '%s': it is associated with %d task%s", name, count, pluralize(count))
			}
			return err
		}

		fmt.Printf("✓ Tag deleted: %s\n", name)
		return nil
	},
}

var tagRenameCmd = &cobra.Command{
	Use:   "rename [old-name] [new-name]",
	Short: "Rename a tag",
	Long: `Rename an existing tag.

The new name must follow the same rules: lowercase letters, numbers, underscores, and hyphens.

Examples:
  facienda tag rename work office
  facienda tag rename old_name new_name`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldName := todo.NormalizeTagName(args[0])
		newName := args[1]

		tag, err := store.GetTagByName(oldName)
		if err != nil {
			return err
		}

		// Validate the new name
		if err := todo.ValidateTagName(todo.NormalizeTagName(newName)); err != nil {
			return fmt.Errorf("invalid new tag name: %w", err)
		}

		tag.Name = todo.NormalizeTagName(newName)

		if err := store.UpdateTag(tag); err != nil {
			return err
		}

		fmt.Printf("✓ Tag renamed: %s → %s\n", oldName, tag.Name)
		return nil
	},
}

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// formatTags formats a slice of tags for display
func formatTags(tags []*todo.Tag) string {
	if len(tags) == 0 {
		return ""
	}

	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	return strings.Join(tagNames, ", ")
}

func init() {
	tagCmd.AddCommand(tagCreateCmd)
	tagCmd.AddCommand(tagListCmd)
	tagCmd.AddCommand(tagDeleteCmd)
	tagCmd.AddCommand(tagRenameCmd)
	rootCmd.AddCommand(tagCmd)
}
