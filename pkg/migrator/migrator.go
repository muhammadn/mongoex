package mongoex

import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    _ "go.mongodb.org/mongo-driver/mongo/readpref"
)

func Migrator(source string, destination string) {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        from_host, err := mongo.Connect(ctx, options.Client().ApplyURI(source))
        to_host, err := mongo.Connect(ctx, options.Client().ApplyURI(destination))

        defer func() {
            if err = from_host.Disconnect(ctx); err != nil {
                panic(err)
            }

            if err = to_host.Disconnect(ctx); err != nil {
                panic(err)
            }
        }() 
}
