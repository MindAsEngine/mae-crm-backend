package repository

import (
    "context"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func ConnectMongo(ctx context.Context, uri string) error {
    var err error
    Client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        return err
    }
    return nil
}

func UsersCollection() *mongo.Collection {
    return Client.Database("authdb").Collection("users")
}