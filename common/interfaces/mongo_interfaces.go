package interfaces

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClientProvider defines an interface for getting a MongoDB client
type MongoClientProvider interface {
	Client() *mongo.Client
	Database(name string) *mongo.Database
}

type MongoCollectionInterface interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
}

// MongoDatabaseInterface is an interface that describes the mongodb driver
type MongoDatabaseInterface interface {
	Collection(name string, opts ...*options.CollectionOptions) MongoCollectionInterface
}
