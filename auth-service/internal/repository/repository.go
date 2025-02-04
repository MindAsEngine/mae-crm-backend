package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	models "auth-service/internal/domain"
)

type MongoRepository struct {
    db *mongo.Database
    client *mongo.Client
    logger *zap.Logger
}

func NewMongoRepository(logger *zap.Logger) *MongoRepository {
    return &MongoRepository{
		logger: logger,
	}
}

func (r *MongoRepository) ConnectMongo(ctx context.Context, uri string) error {
    var err error
    r.client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        return err
    }
    r.db = r.client.Database("authdb")
    return nil
}

func (r *MongoRepository) UsersCollection() *mongo.Collection {
    return r.client.Database("authdb").Collection("users")
}

func (r *MongoRepository) StoreRefreshToken(data *models.RefreshTokenData) error {
    if r.db == nil {
        return fmt.Errorf("db is nil")
    }
    
    r.logger.Debug("Storing refresh token", 
        zap.String("token", data.Token),
        zap.String("userId", data.UserID))
    
    _, err := r.db.Collection("refresh_tokens").InsertOne(context.Background(), data)
    return err
}

func (r *MongoRepository) GetRefreshToken(token string) (*models.RefreshTokenData, error) {
    var data models.RefreshTokenData
    err := r.db.Collection("refresh_tokens").FindOne(context.Background(), bson.M{"token": token}).Decode(&data)
    return &data, err
}

func (r *MongoRepository) DeleteRefreshTokensByUser(id string, new_token string) error {
    _, err := r.db.Collection("refresh_tokens").DeleteMany(context.Background(), bson.M{"user_id": id, "token": bson.M{"$ne": new_token}})
    return err
}