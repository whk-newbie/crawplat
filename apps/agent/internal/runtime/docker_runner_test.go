package runtime

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"crawler-platform/apps/agent/internal/poller"
)

type fakeCommand struct {
	output []byte
	err    error
}

func (c *fakeCommand) CombinedOutput() ([]byte, error) {
	return c.output, c.err
}

func TestDockerRunnerBuildsExpectedCommand(t *testing.T) {
	cmd := buildDockerRunCommand("crawler/go-echo:latest", []string{"./go-echo"})
	want := []string{"docker", "run", "--rm", "crawler/go-echo:latest", "./go-echo"}
	if !reflect.DeepEqual(cmd, want) {
		t.Fatalf("unexpected docker command: %#v", cmd)
	}
}

func TestDockerRunnerEmitsOutputLines(t *testing.T) {
	var gotName string
	var gotArgs []string
	runner := NewDockerRunner(func(_ context.Context, name string, args ...string) command {
		gotName = name
		gotArgs = append([]string(nil), args...)
		return &fakeCommand{output: []byte("line one\nline two\n")}
	})

	var logs []string
	err := runner.Run(context.Background(), poller.ClaimedExecution{
		ID:      "exec-1",
		Image:   "crawler/go-echo:latest",
		Command: []string{"./go-echo"},
	}, func(line string) {
		logs = append(logs, line)
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if gotName != "docker" {
		t.Fatalf("unexpected command name: %q", gotName)
	}
	if want := []string{"run", "--rm", "crawler/go-echo:latest", "./go-echo"}; !reflect.DeepEqual(gotArgs, want) {
		t.Fatalf("unexpected command args: %#v", gotArgs)
	}
	if !reflect.DeepEqual(logs, []string{"line one", "line two"}) {
		t.Fatalf("unexpected logs: %#v", logs)
	}
}

func TestDockerRunnerReturnsCommandFailure(t *testing.T) {
	runner := NewDockerRunner(func(_ context.Context, _ string, _ ...string) command {
		return &fakeCommand{output: []byte("boom\n"), err: errors.New("exit status 1")}
	})

	var logs []string
	err := runner.Run(context.Background(), poller.ClaimedExecution{
		ID:      "exec-1",
		Image:   "crawler/go-echo:latest",
		Command: []string{"./go-echo"},
	}, func(line string) {
		logs = append(logs, line)
	})
	if err == nil || err.Error() != "exit status 1" {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(logs, []string{"boom"}) {
		t.Fatalf("unexpected logs: %#v", logs)
	}
}
