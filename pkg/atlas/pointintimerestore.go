package atlas

import (
    "context"
    "time"

    "github.com/mongodb-forks/digest"
    "go.mongodb.org/atlas/mongodbatlas"
    "github.com/mwielbut/pointy"
    "fmt"
    "github.com/schollz/progressbar/v3"
    "mongoex/cmd/config"
    "mongoex/pkg/slack"
    "strings"
    "net/url"
    "strconv"
    "errors"
)

func PointInTimeRestore(sourceProjectName string, targetClusterName string, pointInTimeSeconds int64, sourceClusterName string, targetProjectName string) error {
    // key handler
    pubkey, privkey, slackWebhookUrl := config.ParseConfig()
    t := digest.NewTransport(pubkey, privkey)
    tc, err := t.Client()
    if err != nil {
        fmt.Println(err)
        return err
    }

    // Validate target cluster, prevent accidental override source cluser
    if targetClusterName == sourceClusterName && sourceProjectName == targetProjectName {
	err := errors.New("Target cluster name cannot be identical to Source cluster name\nPlease double check in MongoDB Atlas")
        slack.Notification("Target cluster name cannot be identical to Source cluster name\nPlease double check in MongoDB Atlas", slackWebhookUrl)
        return err
    }

    // create new client
    client := mongodbatlas.NewClient(tc)
    sourceProject, _, err := client.Projects.GetOneProjectByName(context.Background(), sourceProjectName)
    targetProject, _, err := client.Projects.GetOneProjectByName(context.Background(), targetProjectName)
    if err != nil {
	    fmt.Println(err)
            return err
    }

    // snapshot params for cloudbackup snapshots
    /*
    cbs := &mongodbatlas.SnapshotReqPathParameters{
            GroupID:     sourceProject.ID,
            ClusterName: sourceClusterName,
    }

    // begin list all snapshots
    //lo := &mongodbatlas.ListOptions{}

    cloudBackupSnapshots, _, err := client.CloudProviderSnapshots.GetAllCloudProviderSnapshots(context.Background(), cbs, lo)
    //fmt.Println("cloudbackup snapshots: ", cloudBackupSnapshots.Results)
    for i := 0; i < len(cloudBackupSnapshots.Results); i++ {
            fmt.Println(cloudBackupSnapshots.Results[i])
    } 
    // end of list all snapshiots
    */

    //, check the snapshot policy
    /* Still a WIP to find out the oldest snapshot we can retrieve from PIT
    snapshotPolicy, _, err := client.CloudProviderSnapshotBackupPolicies.Get(context.Background(), sourceProject.ID, sourceClusterName)
    if err != nil {
            fmt.Println(err)
	    return err
    }

    timeNowUnix     := time.Now().Unix()
    secondsUnixDays := *snapshotPolicy.RestoreWindowDays * int64(86400)
    backwardsTime   := timeNowUnix - secondsUnixDays // go back start of date earliest snapshot
    //fmt.Println(fmt.Sprintf("timeNowUnix: %d, backwardsTime: %d, pointInTimeSeconds: %d", timeNowUnix, backwardsTime, pointInTimeSeconds))

    if pointInTimeSeconds > backwardsTime && pointInTimeSeconds < timeNowUnix {
            fmt.Println("\nSource Snapshot exists! ✅")
    } else {
	    err := errors.New("\nError! Snapshot for that time is not available")
	    return err
    }
    */

    fmt.Println(fmt.Sprintf("Source Project Name: %s\nSource Project ID: %s\nSource ClusterName %s\n\nTarget Project Name: %s\nTarget Project ID: %s\nTarget ClusterName: %s\n", sourceProjectName, sourceProject.ID, sourceClusterName, targetProjectName, targetProject.ID, targetClusterName))

    // sc = source cluster
    sc, _, err := client.Clusters.Get(context.Background(), sourceProject.ID, sourceClusterName)
    if err != nil {
                fmt.Println(err)
                return err
    }
    fmt.Println(fmt.Sprintf("Source Cluster %s, Disk size: %.2fGB", sourceClusterName, *sc.DiskSizeGB))

    srcMongoURI      := strings.Split(sc.MongoURI, ",")
    firstMongoURI    := srcMongoURI[0] // get the first mongodb string
    finalMongoURI, _ := url.Parse(firstMongoURI)
    //fmt.Println("MongoURI: ", finalMongoURI.Host)
    //fmt.Println("ConnectionStrings: ", *sc.ConnectionStrings)
    
    mongoMeasurements := &mongodbatlas.ProcessMeasurementListOptions{
            Granularity: "PT24H", // 24hrs
            Period:      "PT128H", // 128hrs
            M:           []string{"DB_DATA_SIZE_TOTAL"},
    }

    hostnameAndPort := strings.Split(finalMongoURI.Host, ":") // finalMongoURI.Host is mongodb-shard-00-00.mongodb.net:27017, we split that string here
    hostname        := hostnameAndPort[0]
    // convert port to integer
    port, err := strconv.Atoi(hostnameAndPort[1])
    if err != nil {
        slack.Notification(fmt.Sprintf("Error converting port number string to integer with error: %s", err), slackWebhookUrl)
        panic(err)
    }	

    measurements, _, err := client.ProcessMeasurements.List(context.Background(), sourceProject.ID, hostname, port, mongoMeasurements)
    if measurements.Measurements == nil {
            err = errors.New("There is no measurements to determine the data size")
	    return err
    }

    // for disk size measurement data points, it returns an array/slice and we need to pick
    // the largest value (data size in cluster) in the slice
    var measurementVal float32
    measurementDataPoints := measurements.Measurements[0].DataPoints
    for j := 0; j < len(measurementDataPoints) ; j++ {
	k := j + 1 // forward lookup for k index

	if k > 0 && k < len(measurementDataPoints) {
	        currentDataPoint := *measurements.Measurements[0].DataPoints[j].Value // first ordered data in slice
	        nextDataPoint    := *measurements.Measurements[0].DataPoints[k].Value // next data in slice
		//fmt.Println("currentDataPoint: ", currentDataPoint)
                //fmt.Println("nextDataPoint: ", nextDataPoint)
	        if int64(nextDataPoint) >= int64(currentDataPoint) {
                        measurementVal = *measurements.Measurements[0].DataPoints[k].Value
			//fmt.Println("measurementVal: ", measurementVal)
	        }
        }
    }

    // convert float exponent to integer
    diskUsage := int64(measurementVal)

    // set which tiers to use
    var tier string
    switch true {
    // more than 128GB
    case diskUsage > 137438953472:
            tier = "M20"
            break
    // more than 256GB
    case diskUsage > 274877906944:
            tier = "M30"
            break
    // more than 512GB
    case diskUsage > 549755813888:
            tier = "M40"
            break
    // more than 1TB
    case diskUsage > 1099511627776:
            tier = "M50"
            break
    // if not more than 128GB or not in sizes above
    // max disk size in MongoDB Atlas is 4TB anyway, the max is M50
    // default is M10 which is the smallest dedicated instance
    default:
            tier = "M10"
    }

    fmt.Println("Tier to be used (according to data size): ", tier)

    // create new cluster
    providerSettings := &mongodbatlas.ProviderSettings{
            //ProviderName: "TENANT", // BackingProviderName and ProviderName is only used for M0,M2,M5 - We need M10 and above
            //BackingProviderName: "AWS",
            ProviderName: "AWS",
            InstanceSizeName: tier,
            RegionName: "AP_SOUTHEAST_1",
    }

    regionsConfig := make(map[string]mongodbatlas.RegionsConfig)
    regionsConfig["AP_SOUTHEAST_1"] = mongodbatlas.RegionsConfig{ 
            ElectableNodes: pointy.Int64(3),
	    Priority: pointy.Int64(7),
	    ReadOnlyNodes: pointy.Int64(0),
    }

    cluster := &mongodbatlas.Cluster{
	    Name: targetClusterName,
	    DiskSizeGB: sc.DiskSizeGB,
	    ClusterType: "REPLICASET",
	    ProviderBackupEnabled: pointy.Bool(false),
	    ProviderSettings: providerSettings,
	    MongoDBMajorVersion: "4.4",
	    NumShards: pointy.Int64(1),
	    //ReplicationSpecs: replicationSpecs,
	    ReplicationSpec: regionsConfig,
    }

    _, _, err = client.Clusters.Create(context.Background(), targetProject.ID, cluster)
    if err != nil {
            slack.Notification(fmt.Sprintf("\nProblem creating target cluster %s on project %s with error: %s", targetClusterName, targetProjectName, err), slackWebhookUrl)
            fmt.Println(err)
	    return err
    }

    bar := progressbar.Default(
		-1,
		"Creating Temporary Cluster",
    )

    for {
                // dc = destination cluster
                dc, _, err := client.Clusters.Get(context.Background(), targetProject.ID, targetClusterName)
		if err != nil {
			fmt.Println(err)
                        slack.Notification(fmt.Sprintf("\nProblem creating target cluster %s on project %s", targetClusterName, targetProjectName), slackWebhookUrl)
                        return err
		}

		// progressBar
		bar.Add(1)

		if dc.StateName == "IDLE" {
			fmt.Println("\nCluster has been created!")
                        fmt.Println(fmt.Sprintf("Cluster Srv Connection: %s", dc.ConnectionStrings.StandardSrv))
                        fmt.Println(fmt.Sprintf("Cluster Standard Connection: %s", dc.ConnectionStrings.Standard))

                        slack.Notification(fmt.Sprintf("\nCluster %s has been created", targetClusterName), slackWebhookUrl)
                        slack.Notification(fmt.Sprintf("\nCluster %s Srv Connection: %s", targetClusterName, dc.ConnectionStrings.StandardSrv), slackWebhookUrl)
                        slack.Notification(fmt.Sprintf("\nCluster %s Standard Connection: %s", targetClusterName, dc.ConnectionStrings.Standard), slackWebhookUrl)

                        if dc.ConnectionStrings.PrivateSrv != "" {
                                fmt.Println(fmt.Sprintf("Cluster Srv Private Connection: %s", dc.ConnectionStrings.PrivateSrv))
                                slack.Notification(fmt.Sprintf("\nCluster %s Srv Private Connection: %s", targetClusterName, dc.ConnectionStrings.PrivateSrv), slackWebhookUrl)
			}
			if dc.ConnectionStrings.Private != "" {
                                fmt.Println(fmt.Sprintf("Cluster Standard Private Connection: %s", dc.ConnectionStrings.Private))
                                slack.Notification(fmt.Sprintf("\nCluster %s Standard Private Connection: %s", targetClusterName, dc.ConnectionStrings.Private), slackWebhookUrl)
			}

                        slack.Notification(fmt.Sprintf("Please wait while we do a Point-In-Time Restore to %s", targetClusterName), slackWebhookUrl)
			break
		}
		time.Sleep(15)
    }

    // this would be the logic to create the cluster to restore
    // start of cluster restore code
    o := &mongodbatlas.SnapshotReqPathParameters{
            GroupID:     sourceProject.ID,
            ClusterName: sourceClusterName,
    }

    cloudProviderSnapshot := &mongodbatlas.CloudProviderSnapshotRestoreJob{
            TargetGroupID: targetProject.ID, // change this later to be specifiable to restore to another project
            TargetClusterName: targetClusterName, // target cluster is the one we're going to make
            PointInTimeUTCSeconds: pointInTimeSeconds, // UNIX epoch time in seconds
            //OplogTs: pointInTimeSeconds, 
            //OplogInc: 1,
            DeliveryType: "pointInTime",
    }

    restoreJob, _, err := client.CloudProviderSnapshotRestoreJobs.Create(context.Background(), o, cloudProviderSnapshot)
    //_, _, err = client.CloudProviderSnapshotRestoreJobs.Create(context.Background(), o, cloudProviderSnapshot)
    if err != nil {
            slack.Notification(fmt.Sprintf("\nProblem doing PIT restore to %s cluster on %s with error: %s", targetClusterName, targetProjectName, err), slackWebhookUrl)
            slack.Notification(fmt.Sprintf("Deleting target cluster %s on project %s", targetClusterName, targetProjectName), slackWebhookUrl)

            _, cerr := client.Clusters.Delete(context.Background(), targetProject.ID, targetClusterName)
            if cerr != nil {
                    fmt.Println(cerr)
                    slack.Notification(fmt.Sprintf("Error deleting target cluster %s with error: %s", targetClusterName, cerr), slackWebhookUrl)
                    return err
            }

            panic(err)
    }

    fmt.Println(fmt.Sprintf("\nNow doing Point-in-time-Recovery from EPOCH time: %d", pointInTimeSeconds))
    //fmt.Println(fmt.Sprintf("Restore PIT Job ID: %d", restoreJob.ID))
    //fmt.Println("Please check MongoDB Atlas for progress")
    // end of point in time restore code

    // Checks the restoration progress
    p := &mongodbatlas.SnapshotReqPathParameters{
            GroupID:     sourceProject.ID,
            ClusterName: sourceClusterName,
            JobID:       restoreJob.ID,
    }

    bar = progressbar.Default(
                -1,
                "Restoring from Point-In-Time Recovery",
    )

    for {
            // gs/get snapshot
            gs, _, err := client.CloudProviderSnapshotRestoreJobs.Get(context.Background(), p)
            if err != nil {
                    panic(err)
            }

            // progressBar
            bar.Add(1)

	    // once we finished, break the loop
	    if gs.FinishedAt != "" {
                    fmt.Println(fmt.Sprintf("\nFinished restoring to %s cluster on %s", targetClusterName, targetProjectName))
                    slack.Notification(fmt.Sprintf("Finished restoring from *Project %s* on *cluster %s* to *Project %s* on *cluster %s*", sourceProjectName, sourceClusterName, targetProjectName, targetClusterName), slackWebhookUrl)
                    break
            }
            time.Sleep(15)
    }
    // end of cluster restore code

    return nil
}
