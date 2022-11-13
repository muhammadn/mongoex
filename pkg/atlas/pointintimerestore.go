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
    "strings"
    "net/url"
    "strconv"
)

func PointInTimeRestore(sourceProjectName string, targetClusterName string, pointInTimeSeconds int64, sourceClusterName string, targetProjectName string) error {
    // Validate target cluster, prevent accidental override source cluser
    if targetClusterName == sourceClusterName {
	fmt.Println("Target cluster name cannot be identical to Source cluster name")
	fmt.Println("Please double check in MongoDB Atlas")
        return nil
    }

    pubkey, privkey := config.ParseConfig()
    t := digest.NewTransport(pubkey, privkey)
    tc, err := t.Client()
    if err != nil {
        fmt.Println(err)
	return err
    }

    client := mongodbatlas.NewClient(tc)
    sourceProject, _, err := client.Projects.GetOneProjectByName(context.Background(), sourceProjectName)
    targetProject, _, err := client.Projects.GetOneProjectByName(context.Background(), targetProjectName)
    if err != nil {
	    fmt.Println(err)
            return err
    }
    fmt.Println(fmt.Sprintf("Source Project Name: %s\nSource Project ID: %s\nSource ClusterName %s\nTarget Project Name: %s\nTarget Project ID: %s\nTarget ClusterName: %s", sourceProjectName, sourceProject.ID, sourceClusterName, targetProjectName, targetProject.ID, targetClusterName))

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
        panic(err)
    }
    measurements, _, err := client.ProcessMeasurements.List(context.Background(), sourceProject.ID, hostname, port, mongoMeasurements)
    measurementVal := *measurements.Measurements[0].DataPoints[0].Value
    
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
                        return err
		}
		// progressBar
		bar.Add(1)

		if dc.StateName == "IDLE" {
			fmt.Println("Cluster has been created!")
                        fmt.Println(fmt.Sprintf("Cluster Srv Connection: %s", dc.ConnectionStrings.StandardSrv))
                        fmt.Println(fmt.Sprintf("Cluster Standard Connection: %s", dc.ConnectionStrings.Standard))
                        if dc.ConnectionStrings.PrivateSrv != "" {
                                fmt.Println(fmt.Sprintf("Cluster Srv Private Connection: %s", dc.ConnectionStrings.PrivateSrv))
			}
			if dc.ConnectionStrings.Private != "" {
                                fmt.Println(fmt.Sprintf("Cluster Standard Private Connection: %s", dc.ConnectionStrings.Private))
			}
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
            DeliveryType: "pointInTime",
    }

    //restoreJob, _, err := client.CloudProviderSnapshotRestoreJobs.Create(context.Background(), o, cloudProviderSnapshot)
    _, _, err = client.CloudProviderSnapshotRestoreJobs.Create(context.Background(), o, cloudProviderSnapshot)
    if err != nil {
            panic(err)
    }

    fmt.Println(fmt.Sprintf("Now doing Point-in-time-Recovery from EPOCH time: %d", pointInTimeSeconds))
    //fmt.Println(fmt.Sprintf("Restore PIT Job ID: %d", restoreJob.ID))
    fmt.Println("Please check MongoDB Atlas for progress")

    /* Disabled until we figure out whether should be check it synchronously
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

            if *gs.Failed == false {
                    fmt.Println(fmt.Sprintf("Finished restoring to %s cluster on %s", clusterName, targetProject.ID))
                    break
            }
            time.Sleep(15)
    }
    */
    // end of cluster restore code

    return nil
}
