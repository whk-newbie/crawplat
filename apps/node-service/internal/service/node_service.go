// 文件职责：节点服务的领域模型与业务逻辑层。
// 负责：
//   1. 定义核心数据类型：Node（节点状态）、NodeHeartbeat（心跳历史记录）、NodeExecution（执行记录）、
//      ExecutionQuery（执行记录过滤条件）。
//   2. 定义 Repository 接口（抽象存储层，支持 Redis/Postgres/内存三种实现）。
//   3. 定义 CatalogRepository 扩展接口（Postgres 实现额外支持：节点详情、心跳历史、执行记录过滤）。
//   4. 提供 NodeService 业务逻辑（心跳处理、节点列表）。
//   5. 内置 memoryRepository（内存存储），用于测试和独立运行（无需 Redis/Postgres）。
// 与谁交互：被 api 层调用，通过 Repository 接口调用 repo 层（依赖反转，service 不依赖具体存储实现）。
// 不负责：HTTP 路由与请求处理（由 api 包负责）、具体存储逻辑（由 repo 包负责）。
// 在线/离线判定逻辑：心跳上报时状态固定为 "online"，离线判定由存储层的 TTL 过期自动实现 ——
//   Redis 模式下 key 过期后节点从在线集合中消失，内存模式下不自动过期（仅用于测试）。
package service

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Node 表示一个爬虫节点的完整快照。
// ID 是节点的唯一标识（等同于 Name），Status 为 "online" 或 "offline"，
// LastSeenAt 记录最后一次心跳到达的时间，Capabilities 记录节点的能力标签（如 ["docker","python","go"]）。
type Node struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Capabilities []string  `json:"capabilities"`
	LastSeenAt   time.Time `json:"lastSeenAt"`
}

// NodeHeartbeat 表示一次心跳的历史记录条目。
// 用于节点详情接口返回心跳时间序列，可据此推断节点的活跃时段和在线会话。
type NodeHeartbeat struct {
	SeenAt       time.Time `json:"seenAt"`
	Capabilities []string  `json:"capabilities"`
}

// ExecutionQuery 封装节点执行记录的过滤条件。
// Status：按执行状态过滤（如 "succeeded", "failed", "running"）。
// From/To：按创建时间范围过滤（CreatedAt 在此区间内）。
// Limit/Offset：分页参数。
// 所有字段均为可选，未设置时对应 SQL 条件被省略（动态 WHERE 拼接）。
type ExecutionQuery struct {
	Status string
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}

// NodeExecution 表示一次爬虫任务的执行记录。
// ID 为执行记录唯一 ID，ProjectID/SpiderID 标识所属项目和爬虫，
// Status 为执行状态（succeeded/failed/running 等），TriggerSource 为触发来源（manual/scheduled/webhook 等），
// CreatedAt/StartedAt/FinishedAt 记录时间戳（StartedAt/FinishedAt 可能为空，表示尚未开始或未结束）。
type NodeExecution struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"projectId"`
	SpiderID      string     `json:"spiderId"`
	Status        string     `json:"status"`
	TriggerSource string     `json:"triggerSource"`
	CreatedAt     time.Time  `json:"createdAt"`
	StartedAt     *time.Time `json:"startedAt,omitempty"`
	FinishedAt    *time.Time `json:"finishedAt,omitempty"`
}

// ErrNodeNotFound 在 Postgres 仓库中根据 id 查找节点不存在时返回。
var ErrNodeNotFound = errors.New("node not found")

// NodeService 是节点服务的核心业务逻辑聚合。
// 通过 Repository 接口与存储层解耦，不直接依赖具体的 Redis 或 Postgres 实现。
type NodeService struct {
	repo Repository
}

// Repository 定义了节点存储的最小接口。
// 所有存储后端（Redis、内存、Postgres）都必须实现此接口。
// UpsertHeartbeat：写入/更新心跳，同时更新 LastSeenAt，标记节点在线。
// ListOnline：返回当前在线的节点列表。
type Repository interface {
	UpsertHeartbeat(ctx context.Context, name string, capabilities []string) (Node, error)
	ListOnline(ctx context.Context) ([]Node, error)
}

// CatalogRepository 是 Postgres 仓库的扩展接口，在 Repository 基础之上提供：
//   - 节点目录管理（UpsertCatalog/ListCatalog）：持久化节点元信息，不依赖 TTL。
//   - 按 ID 查询节点详情（GetByID）：返回单个节点的元信息，节点不存在时返回 ErrNodeNotFound。
//   - 心跳历史查询（ListHeartbeatHistory）：按时间倒序返回指定数量的心跳记录，用于绘制活跃度曲线。
//   - 执行记录过滤查询（ListRecentExecutions）：支持按状态、时间范围、分页过滤节点上的任务执行记录。
type CatalogRepository interface {
	Repository
	UpsertCatalog(ctx context.Context, name string, capabilities []string, seenAt time.Time) (Node, error)
	ListCatalog(ctx context.Context) ([]Node, error)
	GetByID(ctx context.Context, nodeID string) (Node, error)
	ListHeartbeatHistory(ctx context.Context, nodeID string, limit int) ([]NodeHeartbeat, error)
	ListRecentExecutions(ctx context.Context, nodeID string, query ExecutionQuery) ([]NodeExecution, error)
}

// memoryRepository 是基于内存的 Repository 实现，主要用于单元测试。
// 注意：内存模式不模拟心跳过期 —— 节点写入后永久保持 "online" 状态，
//       因此在线/离线判定测试需要使用 Redis 仓库（miniredis）。
type memoryRepository struct {
	mu    sync.Mutex
	nodes map[string]Node
}

// UpsertHeartbeat 内存模式下：创建 Node 快照（状态硬编码为 "online"），存入 map。
// Capabilities 会被拷贝一份以切断外部引用，防止数据污染。
func (r *memoryRepository) UpsertHeartbeat(_ context.Context, name string, capabilities []string) (Node, error) {
	node := Node{
		ID:           name,
		Name:         name,
		Status:       "online",
		Capabilities: append([]string(nil), capabilities...),
		LastSeenAt:   time.Now(),
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.nodes == nil {
		r.nodes = make(map[string]Node)
	}
	r.nodes[name] = node
	return node, nil
}

// ListOnline 内存模式下：返回 map 中所有节点的切片，不区分在线/离线。
// 注意：内存模式下所有已上报心beat的节点都会返回，无 TTL 过期机制。
func (r *memoryRepository) ListOnline(_ context.Context) ([]Node, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	nodes := make([]Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// NewNodeService 创建 NodeService 实例，支持可选的 Repository 注入。
// 如果传入了非 nil 的 Repository，则使用该实现（如 Redis 或 Postgres 仓库）。
// 如果未传或传入 nil，则默认使用内存仓库（memoryRepository），方便测试和快速启动。
func NewNodeService(repos ...Repository) *NodeService {
	if len(repos) > 0 && repos[0] != nil {
		return &NodeService{repo: repos[0]}
	}
	return &NodeService{repo: &memoryRepository{nodes: make(map[string]Node)}}
}

// Heartbeat 处理节点心跳上报。
// 委托给底层 Repository 的 UpsertHeartbeat，存储层负责：
//   - 将节点状态设为 "online"
//   - 更新时间戳为当前时间
//   - 在 Redis 模式下设置 TTL（30 秒），过期后自动被视为离线
// 返回更新后的 Node 快照。
func (s *NodeService) Heartbeat(name string, capabilities []string) (Node, error) {
	return s.repo.UpsertHeartbeat(context.Background(), name, capabilities)
}

// List 返回当前在线的节点列表。
// 委托给底层 Repository 的 ListOnline，各存储后端的语义：
//   - Redis 模式：只返回 key 未过期的节点（TTL 内收到过心跳的节点）
//   - 内存模式：返回所有历史记录过的节点（无过期机制）
func (s *NodeService) List() ([]Node, error) {
	return s.repo.ListOnline(context.Background())
}
