package mongoex

import (
    "github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "mongoex",
		Short: "MongoDB migrate data in real-time",
		Long: `mongoex is a tool to migrate mongodb data in real time.
This tool helps to quickly do migrations to move data, especially from production to pre-prod for testing`,
	}
)

func Execute() error {
	return rootCmd.Execute()
}
