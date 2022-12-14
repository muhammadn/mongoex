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
        sourceClusterName, _         := cmd.Flags().GetString("sourceClusterName")
        targetProjectID, _           := cmd.Flags().GetString("targetProject")

	// convert diskSize from string to float which is required
	diskSizef, err := strconv.ParseFloat(diskSize, 1)
	if err != nil {
                panic(err)
	}

	atlas.AutomatedRestore(projectName, diskSizef, tier, clusterName, sourceClusterName, targetProjectID)
        return nil
    },
}

func init() {
        tempClusterCmd.AddCommand(automatedRestoreCmd)
        automatedRestoreCmd.Flags().StringP("proj", "p", "", "MongoDB Project Name")
        automatedRestoreCmd.Flags().StringP("diskSize", "d", "", "Cluster disk size for target temporary cluster")
        automatedRestoreCmd.Flags().StringP("tier", "t", "", "Tier for temporary cluster")
        automatedRestoreCmd.Flags().StringP("clusterName", "c", "", "Name for temporary cluster")
        automatedRestoreCmd.Flags().StringP("sourceClusterName", "", "", "Source MongoDB Cluster Name")
	automatedRestoreCmd.Flags().StringP("targetProject", "", "", "Target Project ID")

        automatedRestoreCmd.MarkFlagRequired("proj")
        automatedRestoreCmd.MarkFlagRequired("diskSize")
        automatedRestoreCmd.MarkFlagRequired("tier")
        automatedRestoreCmd.MarkFlagRequired("clusterName")
        automatedRestoreCmd.MarkFlagRequired("sourceClusterName")
        automatedRestoreCmd.MarkFlagRequired("targetProject")
}
