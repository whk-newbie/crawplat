package mongox

import (
	"context"
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	if strings.TrimSpace(uri) == "" {
		return nil, errors.New("mongo uri is required")
	}
	return mongo.Connect(ctx, options.Client().ApplyURI(uri))
}
