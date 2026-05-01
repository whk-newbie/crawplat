package poller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const internalTokenHeader = "X-Internal-Token"

type ClaimedExecution struct {
	ID             string   `json:"id"`
	NodeID         string   `json:"nodeId,omitempty"`
	Status         string   `json:"status,omitempty"`
	Image          string   `json:"image"`
	Command        []string `json:"command"`
	CPUCores       float64  `json:"cpuCores"`
	MemoryMB       int      `json:"memoryMB"`
	TimeoutSeconds int      `json:"timeoutSeconds"`
}

type ExecutionClient interface {
	Claim(ctx context.Context, nodeID string) (ClaimedExecution, bool, error)
	Start(ctx context.Context, executionID, nodeID string) (ClaimedExecution, error)
	AppendLog(ctx context.Context, executionID, message string) error
	Complete(ctx context.Context, executionID string) error
	Fail(ctx context.Context, executionID, message string) error
}

type Runner interface {
	Run(ctx context.Context, exec ClaimedExecution, onLog func(string)) error
}

type Poller struct {
	client   ExecutionClient
	runner   Runner
	nodeID   string
	interval time.Duration
}

func New(client ExecutionClient, runner Runner, nodeID string, interval time.Duration) *Poller {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &Poller{
		client:   client,
		runner:   runner,
		nodeID:   nodeID,
		interval: interval,
	}
}

func (p *Poller) Run(ctx context.Context) error {
	if err := p.Tick(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := p.Tick(ctx); err != nil {
				return err
			}
		}
	}
}

func (p *Poller) Tick(ctx context.Context) error {
	exec, ok, err := p.client.Claim(ctx, p.nodeID)
	if err != nil || !ok {
		return err
	}

	exec, err = p.client.Start(ctx, exec.ID, p.nodeID)
	if err != nil {
		return fmt.Errorf("start execution %s: %w", exec.ID, err)
	}

	var logErr error
	runErr := p.runner.Run(ctx, exec, func(line string) {
		if logErr != nil || strings.TrimSpace(line) == "" {
			return
		}
		if err := p.client.AppendLog(ctx, exec.ID, line); err != nil {
			logErr = fmt.Errorf("append log for %s: %w", exec.ID, err)
		}
	})
	if logErr != nil && runErr == nil {
		runErr = logErr
	}

	if runErr != nil {
		if err := p.client.Fail(ctx, exec.ID, runErr.Error()); err != nil {
			return fmt.Errorf("fail execution %s: %w", exec.ID, err)
		}
		return nil
	}

	if err := p.client.Complete(ctx, exec.ID); err != nil {
		return fmt.Errorf("complete execution %s: %w", exec.ID, err)
	}
	return nil
}

type HTTPExecutionClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewExecutionClient(baseURL, token string) *HTTPExecutionClient {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "http://localhost:8085"
	}

	return &HTTPExecutionClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   strings.TrimSpace(token),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *HTTPExecutionClient) Claim(ctx context.Context, nodeID string) (ClaimedExecution, bool, error) {
	payload, err := json.Marshal(map[string]string{"nodeId": nodeID})
	if err != nil {
		return ClaimedExecution{}, false, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/v1/executions/claim", bytes.NewReader(payload))
	if err != nil {
		return ClaimedExecution{}, false, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setToken(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return ClaimedExecution{}, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return ClaimedExecution{}, false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return ClaimedExecution{}, false, fmt.Errorf("claim execution: unexpected status %s", resp.Status)
	}

	var exec ClaimedExecution
	if err := json.NewDecoder(resp.Body).Decode(&exec); err != nil {
		return ClaimedExecution{}, false, err
	}
	return exec, true, nil
}

func (c *HTTPExecutionClient) Start(ctx context.Context, executionID, nodeID string) (ClaimedExecution, error) {
	payload, err := json.Marshal(map[string]string{"nodeId": nodeID})
	if err != nil {
		return ClaimedExecution{}, err
	}
	var exec ClaimedExecution
	if err := c.postJSON(ctx, "/internal/v1/executions/"+executionID+"/start", payload, &exec); err != nil {
		return ClaimedExecution{}, err
	}
	return exec, nil
}

func (c *HTTPExecutionClient) AppendLog(ctx context.Context, executionID, message string) error {
	payload, err := json.Marshal(map[string]string{"message": message})
	if err != nil {
		return err
	}
	return c.postJSON(ctx, "/internal/v1/executions/"+executionID+"/logs", payload, nil)
}

func (c *HTTPExecutionClient) Complete(ctx context.Context, executionID string) error {
	return c.postJSON(ctx, "/internal/v1/executions/"+executionID+"/complete", nil, nil)
}

func (c *HTTPExecutionClient) Fail(ctx context.Context, executionID, message string) error {
	payload, err := json.Marshal(map[string]string{"error": message})
	if err != nil {
		return err
	}
	return c.postJSON(ctx, "/internal/v1/executions/"+executionID+"/fail", payload, nil)
}

func (c *HTTPExecutionClient) postJSON(ctx context.Context, path string, payload []byte, out any) error {
	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		body = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setToken(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%s: unexpected status %s", path, resp.Status)
	}
	if out == nil || resp.ContentLength == 0 {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *HTTPExecutionClient) setToken(req *http.Request) {
	if c.token != "" {
		req.Header.Set(internalTokenHeader, c.token)
	}
}
