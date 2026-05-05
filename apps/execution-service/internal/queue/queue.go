// Redis 执行队列。
// 管理两个 Redis List：executions:pending（待认领）和 executions:inflight（已被认领但未完成）。
// 提供原子认领（LMOVE）、确认删除（LREM）和事务性释放（TxPipeline：inflight→pending）操作。
// 不关心执行状态转换——只负责队列原语，状态机逻辑由 service 层在 ClaimNext/Complete/Fail 中调用。
package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const (
	// executionQueue 待执行队列——新创建的 pending 执行通过 Enqueue 加入此队列尾部。
	executionQueue         = "executions:pending"
	// executionInflightQueue 运行中队列——Claim 将任务从 pending 原子移动到 inflight，Complete/Fail 时 Ack 从此队列删除。
	executionInflightQueue = "executions:inflight"
)

// RedisQueue 封装了对 Redis 执行队列的操作。
type RedisQueue struct {
	client *redis.Client
}

// NewRedisQueue 创建队列实例。client 必须已完成连接，调用方负责生命周期管理。
func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{client: client}
}

// Enqueue 将执行 ID 加入 pending 队列尾部（FIFO 入队）。
// 在 Create 流程中，执行写入 PostgreSQL 和初始化 MongoDB 日志之后调用。
func (q *RedisQueue) Enqueue(ctx context.Context, executionID string) error {
	return q.client.RPush(ctx, executionQueue, executionID).Err()
}

// Claim 使用 LMOVE 原子地从 pending 队列头部取出一个 ID 并移入 inflight 队列。
// 返回值：如果队列为空，返回 ("", nil)；如果 Redis 出错，返回 ("", err)。
// 原子性保证：Claim 不会丢失任务——ID 要么留在 pending，要么安全地移入 inflight。
func (q *RedisQueue) Claim(ctx context.Context) (string, error) {
	id, err := q.client.LMove(ctx, executionQueue, executionInflightQueue, "LEFT", "RIGHT").Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

// Ack 从 inflight 队列中删除指定执行 ID，表示该任务已终态完成（succeeded 或 failed）。
// 在 Complete 或 Fail 成功将状态写入 PostgreSQL 后调用，确保 inflight 队列不残留已完成任务。
func (q *RedisQueue) Ack(ctx context.Context, executionID string) error {
	return q.client.LRem(ctx, executionInflightQueue, 1, executionID).Err()
}

// Release 将执行从 inflight 队列移回 pending 队列头部，使用 Redis TxPipeline 保证原子性。
// 调用场景：Claim 后 MarkRunning 失败时，需要将任务归还队列以便其他节点重试。
// 为什么用 LPush 而非 RPush：放回头部优先重试，避免因同一任务反复失败导致饥饿。
func (q *RedisQueue) Release(ctx context.Context, executionID string) error {
	pipe := q.client.TxPipeline()
	pipe.LRem(ctx, executionInflightQueue, 1, executionID)
	pipe.LPush(ctx, executionQueue, executionID)
	_, err := pipe.Exec(ctx)
	return err
}
