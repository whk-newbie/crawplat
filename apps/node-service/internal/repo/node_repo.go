// 文件职责：基于 Redis 的节点仓库实现（实现 service.Repository 接口）。
// 负责：
//   1. 心跳数据的持久化（JSON 序列化后写入 Redis key，带 TTL 过期）。
//   2. 在线节点索引维护（通过 Redis Set `nodes:online` 记录所有活跃节点的 id）。
//   3. 在线节点列表查询（从 Set 读取 id 列表，再逐个 Get key，懒清理已过期的成员）。
// 在线/离线判定机制：
//   - 每个心跳对应的 Redis key（`nodes:{id}`）设置了 30 秒 TTL。
//   - 如果节点持续发送心跳（在 TTL 内），key 持续存在，节点保持"在线"。
//   - 如果节点停止发送心跳超过 TTL，key 被 Redis 自动删除，该节点从 ListOnline 结果中消失，即视为"离线"。
//   - Redis Set 中的过期引用通过懒清理机制移除（在 ListOnline 中如果发现 key 不存在，则从 Set 中删除）。
// 与谁交互：依赖 go-redis/v9 客户端，被 main.go 注入到 NodeService。
// 不负责：API 路由处理、业务逻辑（由 api、service 包负责）、Postgres 存储逻辑。
package repo

import (
	"context"
	"encoding/json"
	"time"

	"crawler-platform/apps/node-service/internal/service"
	"github.com/redis/go-redis/v9"
)

const (
	nodeKeyPrefix = "nodes:"   // Redis key 前缀，生成格式为 nodes:{id}
	nodeIndexKey  = "nodes:online" // Redis Set key，存储所有在线节点的 id
)

// RedisRepository 使用 Redis 双数据结构存储节点状态：
//   - String key（nodes:{id}）：存储 Node JSON（带 TTL），用于获取节点完整信息
//   - Set（nodes:online）：存储在线节点 id，用于快速列出所有在线节点
type RedisRepository struct {
	client *redis.Client
	ttl    time.Duration // 心跳过期间隔（如 30s），同时也是 Redis key 的 TTL
}

// NewRedisRepository 创建 Redis 仓库实例。
// ttl 参数同时决定了心跳过期时间和 Redis key 的过期时间 —— 超过 ttl 未收到心跳的节点被视为离线。
func NewRedisRepository(client *redis.Client, ttl time.Duration) *RedisRepository {
	return &RedisRepository{client: client, ttl: ttl}
}

// UpsertHeartbeat 处理一次心跳写入/更新。
// 操作分两步（原子性由 Redis 单次命令保证，但两步之间非原子）：
//   1. SET nodes:{id} = NodeJSON, TTL = ttl —— 存储节点快照，设置过期时间
//   2. SADD nodes:online {id} —— 将节点 id 加入在线集合
// 返回构建的 Node 快照，状态硬编码为 "online"，LastSeenAt 为当前时间。
func (r *RedisRepository) UpsertHeartbeat(ctx context.Context, name string, capabilities []string) (service.Node, error) {
	node := service.Node{
		ID:           name,
		Name:         name,
		Status:       "online",
		Capabilities: append([]string(nil), capabilities...),
		LastSeenAt:   time.Now(),
	}

	payload, err := json.Marshal(node)
	if err != nil {
		return service.Node{}, err
	}

	key := nodeKeyPrefix + name
	if err := r.client.Set(ctx, key, payload, r.ttl).Err(); err != nil {
		return service.Node{}, err
	}
	if err := r.client.SAdd(ctx, nodeIndexKey, name).Err(); err != nil {
		return service.Node{}, err
	}
	return node, nil
}

// ListOnline 返回当前在线的节点列表。
// 步骤：
//   1. SMEMBERS nodes:online —— 获取在线集合中的所有节点 id
//   2. 对每个 id 执行 GET nodes:{id} —— 获取节点快照
//   3. 如果 key 不存在（已过期），从 nodes:online Set 中移除该 id（懒清理）
// 结果：只返回 key 仍在 TTL 有效期内的节点 —— 这些是最后一次心跳在 ttl 时间内的节点。
func (r *RedisRepository) ListOnline(ctx context.Context) ([]service.Node, error) {
	ids, err := r.client.SMembers(ctx, nodeIndexKey).Result()
	if err != nil {
		return nil, err
	}

	nodes := make([]service.Node, 0, len(ids))
	for _, id := range ids {
		payload, err := r.client.Get(ctx, nodeKeyPrefix+id).Result()
		if err == redis.Nil {
			_ = r.client.SRem(ctx, nodeIndexKey, id).Err()
			continue
		}
		if err != nil {
			return nil, err
		}

		var node service.Node
		if err := json.Unmarshal([]byte(payload), &node); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}
