package mongoex

import (
        "mongoex/pkg/atlas"
        "github.com/spf13/cobra"
	_ "fmt"
	"strconv"
)

var automatedRestoreCmd = &cobra.Command{
    Use:   "automated",
    Aliases: []string{"auto"},
    Short:  "Create a temporary cluster and do a Automated Recovery to this cluster",
    //Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        projectName, _               := cmd.Flags().GetString("proj")
	diskSize, _                  := cmd.Flags().GetString("diskSize")
	tier, _                      := cmd.Flags().GetString("tier")
	clusterName, _               := cmd.Flags().GetString("clusterName")
	pubkey, _                    := cmd.Flags().GetString("pubkey")
	privkey, _                   := cmd.Flags().GetString("privkey")
        //atlas.AccessTest()
	// convert diskSize from string to float which is required
	diskSizef, err := strconv.ParseFloat(diskSize, 1)
	if err != nil {
                panic(err)
	}

	atlas.AutomatedRestore(projectName, diskSizef, tier, clusterName, pubkey, privkey)
        return nil
    },
}

func init() {
        tempClusterCmd.AddCommand(automatedRestoreCmd)
        automatedRestoreCmd.PersistentFlags().StringP("proj", "p", "", "MongoDB Project Name")
        automatedRestoreCmd.PersistentFlags().StringP("diskSize", "d", "", "Cluster disk size for target temporary cluster")
        automatedRestoreCmd.PersistentFlags().StringP("tier", "t", "", "Tier for temporary cluster")
        automatedRestoreCmd.PersistentFlags().StringP("clusterName", "c", "", "Name for temporary cluster")
        automatedRestoreCmd.PersistentFlags().StringP("pubkey", "", "", "Public MongoDB API Key")
        automatedRestoreCmd.PersistentFlags().StringP("privkey", "", "", "Private MongoDB API Key")
}
