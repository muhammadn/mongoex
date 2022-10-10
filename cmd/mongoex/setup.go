package mongoex

import (
        "github.com/spf13/cobra"
	"mongoex/cmd/config"
)

var setupCmd = &cobra.Command{
    Use:   "setup",
    Short:  "creates mongoex config file",
    //Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
	config.SetupCreate()
        return nil
    },
}

func init() {
        rootCmd.AddCommand(setupCmd)
}
