package migrator

import (
    "context"
    "time"
    "log"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson"
    _ "go.mongodb.org/mongo-driver/mongo/readpref"
    "fmt"
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
        from_host, err := mongo.Connect(ctx, options.Client().ApplyURI(source))
        to_host, err := mongo.Connect(ctx, options.Client().ApplyURI(destination))

        for i := 0; i < len(collections); i++ {
                var results []bson.D

                source_collection := from_host.Database(databaseSource).Collection(collections[i])
                cursor, err := source_collection.Find(context.TODO(), bson.D{})
                if err != nil {
	                log.Panic(err)
                }

                defer cursor.Close(context.Background())

                destination_collection := to_host.Database(databaseDestination).Collection(collections[i])
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
        }

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
