package mongoex

import (
        "mongoex/pkg/atlas"
        "github.com/spf13/cobra"
	_ "fmt"
	"strconv"
	"time"
)

var pointInTimeCmd = &cobra.Command{
    Use:   "pointintime",
    Aliases: []string{"pit"},
    Short:  "Create a temporary cluster and do a Point-In-Time-Recovery to this cluster",
    //Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        projectName, _               := cmd.Flags().GetString("proj")
	diskSize, _                  := cmd.Flags().GetString("diskSize")
	tier, _                      := cmd.Flags().GetString("tier")
	clusterName, _               := cmd.Flags().GetString("clusterName")
	pubkey, _                    := cmd.Flags().GetString("pubkey")
	privkey, _                   := cmd.Flags().GetString("privkey")
	pointInTimeSeconds, _        := cmd.Flags().GetString("time")
        //atlas.AccessTest()
	// convert diskSize from string to float which is required
	diskSizef, err := strconv.ParseFloat(diskSize, 1)
	if err != nil {
                panic(err)
	}

	var timeSeconds int64
	if pointInTimeSeconds == "" {
		timeSeconds = time.Now().Unix()
	} else {
		timeSeconds, err = strconv.ParseInt(pointInTimeSeconds, 10, 64)
                if err != nil {
                        panic(err)
                }
	}
	atlas.PointInTimeRestore(projectName, diskSizef, tier, clusterName, pubkey, privkey, timeSeconds)
        return nil
    },
}

func init() {
        tempClusterCmd.AddCommand(pointInTimeCmd)
        pointInTimeCmd.PersistentFlags().StringP("proj", "p", "", "MongoDB Project Name")
        pointInTimeCmd.PersistentFlags().StringP("diskSize", "d", "", "Cluster disk size for target temporary cluster")
        pointInTimeCmd.PersistentFlags().StringP("tier", "t", "", "Tier for temporary cluster")
        pointInTimeCmd.PersistentFlags().StringP("clusterName", "c", "", "Name for temporary cluster")
        pointInTimeCmd.PersistentFlags().StringP("pubkey", "", "", "Public MongoDB API Key")
        pointInTimeCmd.PersistentFlags().StringP("privkey", "", "", "Private MongoDB API Key")
	pointInTimeCmd.PersistentFlags().StringP("time", "", "", "Point-in-time since epoch")
}
