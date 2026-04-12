package mongox

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	return mongo.Connect(ctx, options.Client().ApplyURI(uri))
}
