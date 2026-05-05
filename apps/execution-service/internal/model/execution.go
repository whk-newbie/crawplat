// 执行服务的数据模型定义。
// Execution 对应 PostgreSQL executions 表，ExecutionLog 对应 MongoDB execution_logs 集合。
// 该文件只定义数据结构，不包含任何业务逻辑、数据库操作或 API 序列化细节。
package model

import "time"

// Execution 表示一次爬虫执行的完整状态记录。
// 状态机：pending → running → succeeded / failed（见 service 层的 transitionToRunning / Complete / Fail）。
// 重试通过 RetryOfExecutionID 字段形成执行链，MaterializeRetry 负责将 failed 执行物化为新的 pending 执行。
type Execution struct {
	ID              string `json:"id"`
	ProjectID       string `json:"projectId"`
	SpiderID        string `json:"spiderId"`
	SpiderVersion   string `json:"spiderVersion,omitempty"`
	RegistryAuthRef string `json:"registryAuthRef,omitempty"`
	// NodeID 是认领该执行的工作节点标识，在 ClaimNext 时写入，用于后续 Start/Complete/Fail 的节点校验。
	NodeID string `json:"nodeId,omitempty"`
	// Status 状态机字段：pending（初始）→ running（已认领）→ succeeded（完成）/ failed（失败）。
	Status string `json:"status"`
	// TriggerSource 触发来源：manual（手动）、scheduled（定时）、retry（重试物化）。
	TriggerSource string `json:"triggerSource"`
	Image         string   `json:"image"`
	Command       []string `json:"command"`
	// CpuCores 请求的 CPU 核数，MemoryMB 请求的内存（MB），TimeoutSeconds 执行超时时间。
	CpuCores       float64 `json:"cpuCores,omitempty"`
	MemoryMB       int     `json:"memoryMB,omitempty"`
	TimeoutSeconds int     `json:"timeoutSeconds,omitempty"`
	// RetryLimit 最大重试次数，RetryCount 当前已重试次数。
	// 只有当 retry_limit > retry_count 时，该执行才是有效的重试候选。
	RetryLimit    int `json:"retryLimit"`
	RetryCount    int `json:"retryCount"`
	// RetryDelaySeconds 失败后等待多少秒才能重试，MaterializeRetry 在筛选候选时会检查 finished_at + retry_delay_seconds <= now。
	RetryDelaySeconds int    `json:"retryDelaySeconds"`
	// RetryOfExecutionID 指向重试链中的原始失败执行 ID，用于追踪执行血缘关系。
	RetryOfExecutionID string `json:"retryOfExecutionId,omitempty"`
	ErrorMessage        string `json:"errorMessage,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	StartedAt           *time.Time `json:"startedAt,omitempty"`
	FinishedAt          *time.Time `json:"finishedAt,omitempty"`
	// RetriedAt 是重试物化的乐观锁时间戳——ClaimNextRetryCandidate 在原子 UPDATE 中写入，防止并发重试同一执行。
	RetriedAt *time.Time     `json:"retriedAt,omitempty"`
	Logs      []ExecutionLog `json:"logs,omitempty"`
}

// ExecutionLog 表示一条执行日志记录，以独立文档形式存储在 MongoDB 的 execution_logs 集合中。
type ExecutionLog struct {
	ID          string    `json:"id"`
	ExecutionID string    `json:"executionId"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"createdAt"`
}
