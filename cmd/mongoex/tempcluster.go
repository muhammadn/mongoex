package mongoex

import (
        "github.com/spf13/cobra"
)

var tempClusterCmd = &cobra.Command{
    Use:   "tempcluster",
    Aliases: []string{"tmpcls"},
    Short:  "Create a temporary cluster",
    //Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        return nil
    },
}

func init() {
        atlasCmd.AddCommand(tempClusterCmd)
}
