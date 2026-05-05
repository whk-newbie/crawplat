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

type StaticSpiderVersionResolver struct {
	Version         int
	RegistryAuthRef string
	Image           string
	Command         []string
}

func (r *StaticSpiderVersionResolver) Resolve(_ context.Context, _ string, requestedVersion int) (int, string, string, []string, error) {
	version := r.Version
	if requestedVersion > 0 {
		version = requestedVersion
	}
	return version, strings.TrimSpace(r.RegistryAuthRef), strings.TrimSpace(r.Image), append([]string(nil), r.Command...), nil
}

type SpiderVersionResolver interface {
	Resolve(ctx context.Context, spiderID string, requestedVersion int) (version int, registryAuthRef string, image string, command []string, err error)
}

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

func NewHTTPSpiderVersionResolver(baseURL string, client *http.Client) *HTTPSpiderVersionResolver {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &HTTPSpiderVersionResolver{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
	}
}

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
