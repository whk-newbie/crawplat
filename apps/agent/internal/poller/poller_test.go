package poller

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeExecutionClient struct {
	claimed       ClaimedExecution
	claimOK       bool
	claimErr      error
	started       []string
	logs          map[string][]string
	completed     []string
	failed        map[string]string
	startErr      error
	appendLogErr  error
	completeErr   error
	failErr       error
}

func (c *fakeExecutionClient) Claim(_ context.Context, nodeID string) (ClaimedExecution, bool, error) {
	if c.claimErr != nil {
		return ClaimedExecution{}, false, c.claimErr
	}
	if !c.claimOK {
		return ClaimedExecution{}, false, nil
	}
	return c.claimed, true, nil
}

func (c *fakeExecutionClient) Start(_ context.Context, executionID, nodeID string) (ClaimedExecution, error) {
	if c.startErr != nil {
		return ClaimedExecution{}, c.startErr
	}
	c.started = append(c.started, executionID+":"+nodeID)
	return c.claimed, nil
}

func (c *fakeExecutionClient) AppendLog(_ context.Context, executionID, message string) error {
	if c.appendLogErr != nil {
		return c.appendLogErr
	}
	if c.logs == nil {
		c.logs = map[string][]string{}
	}
	c.logs[executionID] = append(c.logs[executionID], message)
	return nil
}

func (c *fakeExecutionClient) Complete(_ context.Context, executionID string) error {
	if c.completeErr != nil {
		return c.completeErr
	}
	c.completed = append(c.completed, executionID)
	return nil
}

func (c *fakeExecutionClient) Fail(_ context.Context, executionID, message string) error {
	if c.failErr != nil {
		return c.failErr
	}
	if c.failed == nil {
		c.failed = map[string]string{}
	}
	c.failed[executionID] = message
	return nil
}

type fakeRunner struct {
	lastExecutionID string
	lastImage       string
	lastCommand     []string
	logLines        []string
	err             error
}

func (r *fakeRunner) Run(_ context.Context, exec ClaimedExecution, onLog func(string)) error {
	r.lastExecutionID = exec.ID
	r.lastImage = exec.Image
	r.lastCommand = append([]string(nil), exec.Command...)
	for _, line := range r.logLines {
		onLog(line)
	}
	return r.err
}

func TestTickReturnsNilWhenNoExecutionExists(t *testing.T) {
	p := New(&fakeExecutionClient{}, &fakeRunner{}, "node-1", 5)

	if err := p.Tick(context.Background()); err != nil {
		t.Fatalf("Tick returned error: %v", err)
	}
}

func TestPollerClaimsAndRunsExecution(t *testing.T) {
	client := &fakeExecutionClient{
		claimOK: true,
		claimed: ClaimedExecution{
			ID:      "exec-1",
			Image:   "crawler/go-echo:latest",
			Command: []string{"./go-echo"},
		},
	}
	runner := &fakeRunner{logLines: []string{"hello", "world"}}
	p := New(client, runner, "node-1", 5)

	if err := p.Tick(context.Background()); err != nil {
		t.Fatalf("Tick returned error: %v", err)
	}
	if runner.lastExecutionID != "exec-1" {
		t.Fatalf("expected exec-1 to run, got %q", runner.lastExecutionID)
	}
	if !reflect.DeepEqual(client.started, []string{"exec-1:node-1"}) {
		t.Fatalf("unexpected start calls: %+v", client.started)
	}
	if got := client.logs["exec-1"]; !reflect.DeepEqual(got, []string{"hello", "world"}) {
		t.Fatalf("unexpected logs: %+v", got)
	}
	if !reflect.DeepEqual(client.completed, []string{"exec-1"}) {
		t.Fatalf("unexpected completed executions: %+v", client.completed)
	}
}

func TestTickMarksExecutionFailedWhenRunnerFails(t *testing.T) {
	client := &fakeExecutionClient{
		claimOK: true,
		claimed: ClaimedExecution{
			ID:      "exec-1",
			Image:   "crawler/go-echo:latest",
			Command: []string{"./go-echo"},
		},
	}
	runner := &fakeRunner{err: errors.New("exit status 1")}
	p := New(client, runner, "node-1", 5)

	if err := p.Tick(context.Background()); err != nil {
		t.Fatalf("Tick returned error: %v", err)
	}
	if got := client.failed["exec-1"]; got != "exit status 1" {
		t.Fatalf("unexpected failed message: %q", got)
	}
	if len(client.completed) != 0 {
		t.Fatalf("expected no completion, got %+v", client.completed)
	}
}

func TestTickReturnsErrorWhenCompletionCannotBeReported(t *testing.T) {
	client := &fakeExecutionClient{
		claimOK:     true,
		completeErr: errors.New("complete failed"),
		claimed: ClaimedExecution{
			ID:      "exec-1",
			Image:   "crawler/go-echo:latest",
			Command: []string{"./go-echo"},
		},
	}
	p := New(client, &fakeRunner{}, "node-1", 5)

	err := p.Tick(context.Background())
	if err == nil || err.Error() != "complete execution exec-1: complete failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}
