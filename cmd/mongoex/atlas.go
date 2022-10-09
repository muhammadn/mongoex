package mongoex

import (
        "github.com/spf13/cobra"
)

var atlasCmd = &cobra.Command{
    Use:   "atlas",
    Aliases: []string{"atl"},
    Short:  "Atlas commands",
    Long: "This command is used to control MongoDB Atlas from the API",
    //Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        return nil
    },
}

func init() {
        rootCmd.AddCommand(atlasCmd)
}
