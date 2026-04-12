package heartbeat

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type heartbeatRequest struct {
	Capabilities []string `json:"capabilities"`
}

func Run(ctx context.Context, baseURL, nodeName string) error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			postHeartbeat(baseURL, nodeName)
		}
	}
}

func postHeartbeat(baseURL, nodeName string) {
	payload, _ := json.Marshal(heartbeatRequest{
		Capabilities: []string{"docker", "python", "go"},
	})

	endpoint := strings.TrimRight(baseURL, "/") + "/api/v1/nodes/" + nodeName + "/heartbeat"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return
	}
}
