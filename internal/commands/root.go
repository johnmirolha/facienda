package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/johnmirolha/facienda/internal/storage"
	"github.com/spf13/cobra"
)

var (
	dbPath  string
	store   storage.Storage
	rootCmd = &cobra.Command{
		Use:   "facienda",
		Short: "A console-based TODO application",
		Long:  "Facienda is a simple and efficient console TODO app for managing your tasks.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			store, err = storage.NewSQLiteStorage(dbPath)
			if err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if store != nil {
				return store.Close()
			}
			return nil
		},
	}
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	defaultDB := filepath.Join(home, ".facienda.db")

	rootCmd.PersistentFlags().StringVar(&dbPath, "db", defaultDB, "path to SQLite database file")
}

func Execute() error {
	return rootCmd.Execute()
}
