// 爬虫版本解析器。
// 负责根据 spiderID 和可选版本号解析执行所需的镜像、命令和认证引用。
// 提供两种实现：StaticSpiderVersionResolver（测试用静态配置）和 HTTPSpiderVersionResolver（调用爬虫管理服务 API）。
// 不负责执行创建或状态管理——仅返回版本信息，由调用方决定如何使用。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// StaticSpiderVersionResolver 使用硬编码的版本信息，主要用于测试场景。
// 不依赖外部服务，始终返回固定配置。
type StaticSpiderVersionResolver struct {
	Version         int
	RegistryAuthRef string
	Image           string
	Command         []string
}

// Resolve 返回静态配置的版本信息。如果 requestedVersion > 0 则使用请求的版本号，否则使用默认版本。
func (r *StaticSpiderVersionResolver) Resolve(_ context.Context, _ string, requestedVersion int) (int, string, string, []string, error) {
	version := r.Version
	if requestedVersion > 0 {
		version = requestedVersion
	}
	return version, strings.TrimSpace(r.RegistryAuthRef), strings.TrimSpace(r.Image), append([]string(nil), r.Command...), nil
}

// SpiderVersionResolver 定义爬虫版本解析接口。
// 返回值：(version, registryAuthRef, image, command, error)
// registryAuthRef 是私有镜像仓库的认证引用，用于 Kubernetes 拉取镜像时的 imagePullSecrets。
type SpiderVersionResolver interface {
	Resolve(ctx context.Context, spiderID string, requestedVersion int) (version int, registryAuthRef string, image string, command []string, err error)
}

// HTTPSpiderVersionResolver 通过 HTTP 调用爬虫管理服务获取版本信息。
// 调用 GET {baseURL}/api/v1/spiders/{spiderID}/versions，从返回的版本列表中按 requestedVersion 匹配。
type HTTPSpiderVersionResolver struct {
	baseURL string
	client  *http.Client
}

type spiderVersionPayload struct {
	Version         int      `json:"version"`
	RegistryAuthRef string   `json:"registryAuthRef"`
	Image           string   `json:"image"`
	Command         []string `json:"command"`
}

// NewHTTPSpiderVersionResolver 创建 HTTP 版本解析器。
// 如果 client 为 nil，使用 5 秒超时的默认 HTTP 客户端。
func NewHTTPSpiderVersionResolver(baseURL string, client *http.Client) *HTTPSpiderVersionResolver {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &HTTPSpiderVersionResolver{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
	}
}

// Resolve 通过 HTTP 查询爬虫管理服务，获取指定爬虫的版本信息。
//
// 请求路径：GET {baseURL}/api/v1/spiders/{spiderID}/versions
// 版本选择逻辑：
//   - 如果 requestedVersion = 0：使用版本列表中的第一个（最新版本）
//   - 如果 requestedVersion > 0：在列表中精确匹配
//   - 如果未找到匹配或列表为空：返回 ErrSpiderVersionNotFound
//
// 返回值验证：image 字段为空时也返回 ErrSpiderVersionNotFound，因为缺少镜像信息无法创建执行。
// registryAuthRef 可以为空字符串（公开镜像不需要认证）。
func (r *HTTPSpiderVersionResolver) Resolve(ctx context.Context, spiderID string, requestedVersion int) (int, string, string, []string, error) {
	spiderID = strings.TrimSpace(spiderID)
	if spiderID == "" {
		return 0, "", "", nil, ErrSpiderVersionNotFound
	}
	if r.baseURL == "" {
		return 0, "", "", nil, ErrSpiderVersionNotFound
	}
	endpoint := fmt.Sprintf("%s/api/v1/spiders/%s/versions", r.baseURL, url.PathEscape(spiderID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, "", "", nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return 0, "", "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return 0, "", "", nil, ErrSpiderVersionNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, "", "", nil, fmt.Errorf("resolve spider version failed: status %d", resp.StatusCode)
	}

	var versions []spiderVersionPayload
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return 0, "", "", nil, err
	}
	if len(versions) == 0 {
		return 0, "", "", nil, ErrSpiderVersionNotFound
	}

	selected := versions[0]
	if requestedVersion > 0 {
		found := false
		for _, version := range versions {
			if version.Version == requestedVersion {
				selected = version
				found = true
				break
			}
		}
		if !found {
			return 0, "", "", nil, ErrSpiderVersionNotFound
		}
	}
	image := strings.TrimSpace(selected.Image)
	if image == "" {
		return 0, "", "", nil, ErrSpiderVersionNotFound
	}
	return selected.Version, strings.TrimSpace(selected.RegistryAuthRef), image, append([]string(nil), selected.Command...), nil
}

var _ SpiderVersionResolver = (*StaticSpiderVersionResolver)(nil)
var _ SpiderVersionResolver = (*HTTPSpiderVersionResolver)(nil)
