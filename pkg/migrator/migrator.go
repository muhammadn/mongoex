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
)

// MigrateAll is just a dummy connect and disconnect, 
// does not do anything else but just returns true if successful
func MigrateAll(source string, destination string) bool {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        from_host, err := mongo.Connect(ctx, options.Client().ApplyURI(source))
        to_host, err := mongo.Connect(ctx, options.Client().ApplyURI(destination))

        defer func() {
            if err = from_host.Disconnect(ctx); err != nil {
                fmt.Println(err)
            }

            if err = to_host.Disconnect(ctx); err != nil {
                fmt.Println(err)
            }
        }() 

        return true
}

func MigrateCollections(source string, destination string, databaseSource string, databaseDestination string, collections []string) bool {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        source_client, err := mongo.Connect(ctx, options.Client().ApplyURI(source))
        destination_client, err := mongo.Connect(ctx, options.Client().ApplyURI(destination))

        for i := 0; i < len(collections); i++ {
                var results []bson.D

                source_collection := source_client.Database(databaseSource).Collection(collections[i])
                cursor, err := source_collection.Find(context.TODO(), bson.D{})
                if err != nil {
	                log.Panic(err)
                }

                defer cursor.Close(context.Background())

                destination_collection := destination_client.Database(databaseDestination).Collection(collections[i])
                if err = cursor.All(context.TODO(), &results); err != nil {
	                log.Panic(err)
                }

                for _, result := range results {
                        inserts, err := destination_collection.InsertMany(context.TODO(), []interface{}{result})
                        if err != nil {
                                log.Panic(err)
                        }
     	                fmt.Println(inserts)
                }

                GetSetIndexes(source_collection, destination_collection)
        }

        defer func() {
            if err = source_client.Disconnect(ctx); err != nil {
                fmt.Println(err)
            }

            if err = destination_client.Disconnect(ctx); err != nil {
                fmt.Println(err)
            }
        }()

        return true
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
                                                        fmt.Println("Indexes().CreateOne() ERROR:", err)
                                                        os.Exit(1) // exit in case of error
                                                } else {
                                                        // API call returns string of the index name
                                                        fmt.Println("CreateOne() index:", ind)
                                                        fmt.Println("CreateOne() type:", reflect.TypeOf(ind), "\n")
                                                }
                                        default: // else add regular index
                                                mod := mongo.IndexModel{
                                                        Keys: bson.M{
                                                                k2: v2, // index in ascending order
                                                        }, Options: nil,
                                                }
                                                ind, err := destination_collection.Indexes().CreateOne(context.TODO(), mod)
                                                if err != nil {
                                                        fmt.Println("Indexes().CreateOne() ERROR:", err)
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
