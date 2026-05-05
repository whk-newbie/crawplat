// 日志数据仓库（MongoDB）。
// 管理执行日志的持久化，每条日志作为独立文档存储在 execution_logs 集合中。
// 不管理执行状态（由 PostgreSQL execution_repo 负责）或执行队列（由 Redis queue 负责）。
package repo

import (
	"context"
	"time"

	"crawler-platform/apps/execution-service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoLogRepository 封装了对 MongoDB execution_logs 集合的读写操作。
type MongoLogRepository struct {
	collection *mongo.Collection
}

// NewMongoLogRepository 创建日志仓库实例。db 应为已连接的 MongoDB Database 对象。
func NewMongoLogRepository(db *mongo.Database) *MongoLogRepository {
	return &MongoLogRepository{collection: db.Collection("execution_logs")}
}

// Init 初始化日志存储——MongoDB 不需要显式创建集合，当前为空操作。
// 保留该方法是为了满足 LogRepository 接口契约，便于未来可能的前置初始化（如创建索引）。
func (r *MongoLogRepository) Init(_ context.Context, _ string) error {
	return nil
}

// Append 向 execution_logs 集合插入一条日志文档。
// 日志按 execution_id 索引，按 created_at 排序，每条日志有独立的 UUID 作为 _id。
func (r *MongoLogRepository) Append(ctx context.Context, entry model.ExecutionLog) error {
	_, err := r.collection.InsertOne(ctx, bson.M{
		"_id":          entry.ID,
		"execution_id": entry.ExecutionID,
		"message":      entry.Message,
		"created_at":   entry.CreatedAt,
	})
	return err
}

// List 按 execution_id 查询并返回该执行的所有日志，按 created_at 升序排列。
// 返回值可能是空切片（无日志）而非 nil。
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
