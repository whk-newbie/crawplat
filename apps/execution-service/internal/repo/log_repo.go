package repo

import (
	"context"
	"time"

	"crawler-platform/apps/execution-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoLogRepository struct {
	collection *mongo.Collection
}

func NewMongoLogRepository(db *mongo.Database) *MongoLogRepository {
	return &MongoLogRepository{collection: db.Collection("execution_logs")}
}

func (r *MongoLogRepository) Init(_ context.Context, _ string) error {
	return nil
}

func (r *MongoLogRepository) Append(ctx context.Context, entry model.ExecutionLog) error {
	_, err := r.collection.InsertOne(ctx, bson.M{
		"_id":          entry.ID,
		"execution_id": entry.ExecutionID,
		"message":      entry.Message,
		"created_at":   entry.CreatedAt,
	})
	return err
}

func (r *MongoLogRepository) List(ctx context.Context, executionID string) ([]model.ExecutionLog, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"execution_id": executionID}, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type logDocument struct {
		ID          string    `bson:"_id"`
		ExecutionID string    `bson:"execution_id"`
		Message     string    `bson:"message"`
		CreatedAt   time.Time `bson:"created_at"`
	}

	var logs []model.ExecutionLog
	for cursor.Next(ctx) {
		var doc logDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		logs = append(logs, model.ExecutionLog{
			ID:          doc.ID,
			ExecutionID: doc.ExecutionID,
			Message:     doc.Message,
			CreatedAt:   doc.CreatedAt,
		})
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}
