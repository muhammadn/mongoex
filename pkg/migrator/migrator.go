package migrator

import (
    "context"
    "time"
    "log"
    "os"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    _ "go.mongodb.org/mongo-driver/mongo/readpref"
    "fmt"
    "reflect"
    "github.com/schollz/progressbar/v3"
)

// migrate all is per database basis
func MigrateAll(source string, destination string, databaseSource string, databaseDestination string, dropCollection bool) (bool, error) {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        source_client, err := mongo.Connect(ctx, options.Client().ApplyURI(source))
        destination_client, err := mongo.Connect(ctx, options.Client().ApplyURI(destination))

        defer func() (bool, error) {
            if err = source_client.Disconnect(ctx); err != nil {
                fmt.Println(err)
                return false, err
            }

            if err = destination_client.Disconnect(ctx); err != nil {
                fmt.Println(err)
                return false, err
            }

            return true, nil
        }() 
 
        source_database := source_client.Database(databaseSource)
        source_collections, err := source_database.ListCollectionNames(
                context.TODO(),
                bson.D{},
        )
        if err != nil {
                fmt.Println(err)
		return false, err
        }

        var total_estimate int64
        for i := 0; i < len(source_collections); i++ {
                source_collection_estimate, err := source_client.Database(databaseSource).Collection(source_collections[i]).EstimatedDocumentCount(context.Background())
                if err != nil {
			fmt.Println(err)
			return false, err
                }

                total_estimate = source_collection_estimate + total_estimate
        }

        fmt.Println("Estimated Documents: ", total_estimate)
        bar := progressbar.Default(total_estimate, "Copying Documents")

        for i := 0; i < len(source_collections); i++ {
                var results []bson.D
 
                source_collection := source_database.Collection(source_collections[i])
                cursor, err := source_collection.Find(context.TODO(), bson.D{})
                if err != nil {
			fmt.Println(err)
			return false, err
                }

                destination_collection := destination_client.Database(databaseDestination).Collection(source_collections[i])
                // if --dropCollection flag is used
                if dropCollection {
                        destination_client.Database(databaseDestination).Collection(source_collections[i]).Drop(context.TODO())
                }
                if err = cursor.All(context.TODO(), &results); err != nil {
			fmt.Println(err)
			return false, err
                }

                for _, result := range results {
                        //inserts, err := destination_collection.InsertMany(context.TODO(), []interface{}{result})
                        _, err := destination_collection.InsertMany(context.TODO(), []interface{}{result})
                        if err != nil {
				fmt.Println(err)
				return false, err
                        }

                        bar.Add(1)
                        //fmt.Println(inserts)
                }

		// Disable copying of indexes for now
                //GetSetIndexes(source_collection, destination_collection)
                fmt.Println(fmt.Sprintf("Finish copying %s collection", source_collections[i]))
        }

        return true, nil
}

// migrate collections is per database and selective collections only
func MigrateCollections(source string, destination string, databaseSource string, databaseDestination string, collections []string, dropCollection bool) (bool, error) {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        source_client, err := mongo.Connect(ctx, options.Client().ApplyURI(source))
        destination_client, err := mongo.Connect(ctx, options.Client().ApplyURI(destination))

        var total_estimate int64
        for i := 0; i < len(collections); i++ {
                source_collection_estimate, err := source_client.Database(databaseSource).Collection(collections[i]).EstimatedDocumentCount(context.Background())
                if err != nil {
                        fmt.Println(err)
			return false, err
                }

                total_estimate = source_collection_estimate + total_estimate
                fmt.Println("Estimated Documents: ", total_estimate)
        }

	bar := progressbar.Default(total_estimate, "Copying Documents")

        for i := 0; i < len(collections); i++ {
                var results []bson.D

                source_collection := source_client.Database(databaseSource).Collection(collections[i])
                cursor, err := source_collection.Find(context.TODO(), bson.D{})
                if err != nil {
	                fmt.Println(err)
			return false, err
                }

                defer cursor.Close(context.Background())

                destination_collection := destination_client.Database(databaseDestination).Collection(collections[i])
                // if --dropCollection flag is used
                if dropCollection {
                        destination_client.Database(databaseDestination).Collection(collections[i]).Drop(context.TODO())
                }
                if err = cursor.All(context.TODO(), &results); err != nil {
	                fmt.Println(err)
			return false, err
                }

                for _, result := range results {
                        //inserts, err := destination_collection.InsertMany(context.TODO(), []interface{}{result})
                        _, err := destination_collection.InsertMany(context.TODO(), []interface{}{result})
                        if err != nil {
				fmt.Println(err)
				return false, err
                        }

                        bar.Add(1)
     	                //fmt.Println(inserts)
                }
 
		// Disable index copying for now
                //GetSetIndexes(source_collection, destination_collection)
		fmt.Println(fmt.Sprintf("Finish copying %s collection", collections[i]))
        }

        defer func() (bool, error) {
            if err = source_client.Disconnect(ctx); err != nil {
                fmt.Println(err)
                return false, err
            }

            if err = destination_client.Disconnect(ctx); err != nil {
                fmt.Println(err)
                return false, err
            }

            return true, nil
        }()

        return true, nil
}

func GetSetIndexes(source_collection *mongo.Collection, destination_collection *mongo.Collection) {
        fmt.Println("Copying Indexes...")
        indexView := source_collection.Indexes()
        opts := options.ListIndexes().SetMaxTime(2 * time.Second)
        cursor, err := indexView.List(context.TODO(), opts)

        if err != nil {
                log.Fatal(err)
        }

        var result []bson.M
        if err = cursor.All(context.TODO(), &result); err != nil {
                log.Fatal(err)
        }

        for _, v := range result {
                for k1, v1 := range v {
                        if reflect.ValueOf(v1).Kind() == reflect.Map {
                                v1a := v1.(primitive.M)
                                fmt.Printf("%v: {\n", k1)
                                for k2, v2 := range v1a {
                                        // initial work we add unique indexes only first
                                        // will add other types of indexes later
                                        switch k1 {
                                        case "unique":
                                                // unique indexes
                                                mod := mongo.IndexModel{
                                                        Keys: bson.M{
                                                                k2: v2, // index in ascending order
                                                        }, Options: options.Index().SetUnique(true),
                                                }
                                                ind, err := destination_collection.Indexes().CreateOne(context.TODO(), mod)
                                                if err != nil {
                                                        fmt.Println("Indexes().CreateOne() Unique ERROR:", err)
                                                        os.Exit(1) // exit in case of error
                                                } else {
                                                        // API call returns string of the index name
                                                        fmt.Println("CreateOne() index:", ind)
                                                        fmt.Println("CreateOne() type:", reflect.TypeOf(ind), "\n")
                                                }
                                        default: // else add regular index
					        fmt.Printf(fmt.Sprintf("k2: %s, v2 %s", k2, v2))
                                                mod := mongo.IndexModel{
                                                        Keys: bson.M{
                                                                k2: v2, // index in ascending order
                                                        }, Options: nil,
                                                }
                                                ind, err := destination_collection.Indexes().CreateOne(context.TODO(), mod)
                                                if err != nil {
                                                        fmt.Println("Indexes().CreateOne() Regular ERROR:", err)
                                                        os.Exit(1) // exit in case of error
                                                } else {
                                                        // API call returns string of the index name
                                                        fmt.Println("CreateOne() index:", ind)
                                                        fmt.Println("CreateOne() type:", reflect.TypeOf(ind), "\n")
                                                }
                                        }

                                        //fmt.Printf("  %v: %v\n", k2, v2)
                                }
                                fmt.Printf("}\n")
                        } else {
                                fmt.Printf("%v: %v\n", k1, v1)
                        }
                }
                fmt.Println()
        }
}
