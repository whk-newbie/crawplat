package heartbeat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type heartbeatRequest struct {
	Capabilities []string `json:"capabilities"`
}

func Run(ctx context.Context, baseURL, nodeName string) error {
	if err := postHeartbeat(baseURL, nodeName); err != nil {
		return fmt.Errorf("initial heartbeat: %w", err)
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := postHeartbeat(baseURL, nodeName); err != nil {
				log.Printf("heartbeat failed: %v", err)
			}
		}
	}
}

func postHeartbeat(baseURL, nodeName string) error {
	payload, err := json.Marshal(heartbeatRequest{
		Capabilities: []string{"docker", "python", "go"},
	})
	if err != nil {
		return err
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/api/v1/nodes/" + url.PathEscape(nodeName) + "/heartbeat"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unexpected heartbeat status: %s", resp.Status)
	}

	return nil
}
