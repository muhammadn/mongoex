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
)

func AutomatedRestore(projectName string, diskSize float64, tier string, clusterName string, sourceCluster string, targetProjectId string) error {
        pubkey, privkey, _ := config.ParseConfig()
        t := digest.NewTransport(pubkey, privkey)
        tc, err := t.Client()
        if err != nil {
                fmt.Println(err)
	        return err
        }
    
        client := mongodbatlas.NewClient(tc)
        project, _, err := client.Projects.GetOneProjectByName(context.Background(), projectName)
        if err != nil {
	        fmt.Println(err)
		return err
        }
        fmt.Println(fmt.Sprintf("Project Name: %s\nProject ID: %s\nTemporary ClusterName: %s", projectName, project.ID, clusterName))
    
        /* TODO: add tempCluster creation in this code and restore in the snapshot loop
        // find snapshot
        snapshotParams := &mongodbatlas.SnapshotReqPathParameters{
                GroupID: project.ID,
                ClusterName: clusterName,
        }
        snapshots, _, err := client.CloudProviderSnapshots.GetAllCloudProviderSnapshots(context.Background(), snapshotParams, &mongodbatlas.ListOptions{})
        if err != nil {
                panic(err)
        }

        for i := 0; i < len(snapshots.Results); i++ {
                // get the first result (first snapshot) and exit loop
                if i == 0 {
                        // this would be the logic to create the cluster to restore
                        fmt.Println("Restoring from Snapshot ID: ", snapshots.Results[i].ID)
                        o := &mongodbatlas.SnapshotReqPathParameters{
                                GroupID:     project.ID,
                                ClusterName: clusterName,
                        }
                        cloudProviderSnapshot := &mongodbatlas.CloudProviderSnapshotRestoreJob{
                                SourceClusterName: "dummyCluster", // source cluster name
                                TargetGroupID: "targetProjectId", // change this later to be specifiable to restore to another project
                                TargetClusterName: clusterName, // target cluster is the one we're going to make
                                SnapshotID: snapshots.Results[i].ID,
                                DeliveryType: "automated",
                        }

                        result, _, err := client.CloudProviderSnapshotRestoreJobs.Create(context.Background(), o, cloudProviderSnapshot)
                        if err != nil {
                                panic(err)
                        }
                        fmt.Println("Result: ", result)
                        break
                }
        }
	*/

        // create new cluster
        // Code below should be move within the snapshot code above
        providerSettings := &mongodbatlas.ProviderSettings{
                ProviderName: "TENANT",
                InstanceSizeName: tier,
                BackingProviderName: "AWS",
                RegionName: "AP_SOUTHEAST_1",
        }
    
        // not needed
        /* replicationSpecs := []mongodbatlas.ReplicationSpec{
                {
                        NumShards: pointy.Int64(1),
                },
        } */
    
        regionsConfig := make(map[string]mongodbatlas.RegionsConfig)
        regionsConfig["AP_SOUTHEAST_1"] = mongodbatlas.RegionsConfig{ 
                ElectableNodes: pointy.Int64(3),
                Priority: pointy.Int64(7),
                ReadOnlyNodes: pointy.Int64(0),
        }
    
        cluster := &mongodbatlas.Cluster{
                Name: clusterName,
                //DiskSizeGB: pointy.Float64(float64(diskSize)),
                DiskSizeGB: pointy.Float64(diskSize),
                ClusterType: "REPLICASET",
                ProviderBackupEnabled: pointy.Bool(false),
                ProviderSettings: providerSettings,
                MongoDBMajorVersion: "4.2",
                NumShards: pointy.Int64(1),
                //ReplicationSpecs: replicationSpecs,
                ReplicationSpec: regionsConfig,
        }
    
        //fmt.Println(fmt.Sprintf("Cluster info: %s", cluster))
        _, _, err = client.Clusters.Create(context.Background(), project.ID, cluster)
        if err != nil {
                //fmt.Println(fmt.Sprintf("Error: %s", status))
		fmt.Println(err)
		return err
        }
    
        bar := progressbar.Default(
                    -1,
                    "Creating Temporary Cluster",
        )
    
        for {
                    c, _, err := client.Clusters.Get(context.Background(), project.ID, clusterName)
                    if err != nil {
                            fmt.Println(err)
			    return err
                    }
                    // progressBar
                    bar.Add(1)
    
                    if c.StateName == "IDLE" {
                            fmt.Println("Cluster has been created!")
                            fmt.Println(fmt.Sprintf("Cluster Srv Connection: %s", c.ConnectionStrings.StandardSrv))
                            fmt.Println(fmt.Sprintf("Cluster Standard Connection: %s", c.ConnectionStrings.Standard))
                            if c.ConnectionStrings.PrivateSrv != "" {
                                    fmt.Println(fmt.Sprintf("Cluster Srv Private Connection: %s", c.ConnectionStrings.PrivateSrv))
                            }
                            if c.ConnectionStrings.Private != "" {
                                    fmt.Println(fmt.Sprintf("Cluster Standard Private Connection: %s", c.ConnectionStrings.Private))
                            }
                            break
                    }
                    time.Sleep(15)
        }
    
        return nil
    }
