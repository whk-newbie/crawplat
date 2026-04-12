package heartbeat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type heartbeatRequest struct {
	Capabilities []string `json:"capabilities"`
}

func Run(ctx context.Context, baseURL, nodeName string) error {
	if err := postHeartbeat(ctx, baseURL, nodeName); err != nil {
		return fmt.Errorf("initial heartbeat: %w", err)
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := postHeartbeat(ctx, baseURL, nodeName); err != nil {
				log.Printf("heartbeat failed: %v", err)
			}
		}
	}
}

func postHeartbeat(ctx context.Context, baseURL, nodeName string) error {
	if err := validateNodeName(nodeName); err != nil {
		return err
	}

	payload, err := json.Marshal(heartbeatRequest{
		Capabilities: []string{"docker", "python", "go"},
	})
	if err != nil {
		return err
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/api/v1/nodes/" + nodeName + "/heartbeat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
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

func validateNodeName(nodeName string) error {
	if nodeName == "" {
		return errors.New("node name is required")
	}
	for _, r := range nodeName {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' {
			continue
		}
		return fmt.Errorf("invalid node name %q: use only letters, numbers, dash, underscore, or dot", nodeName)
	}
	return nil
}
