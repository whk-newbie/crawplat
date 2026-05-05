package runtime

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"crawler-platform/apps/agent/internal/poller"
)

type fakeReadCloser struct {
	*strings.Reader
}

func (f *fakeReadCloser) Close() error { return nil }

type fakeCommand struct {
	stdoutLines []string
	stderrLines []string
	startErr    error
	waitErr     error
}

func (c *fakeCommand) Start() error {
	return c.startErr
}

func (c *fakeCommand) Wait() error {
	return c.waitErr
}

func (c *fakeCommand) StdoutPipe() (io.ReadCloser, error) {
	return &fakeReadCloser{strings.NewReader(strings.Join(c.stdoutLines, "\n"))}, nil
}

func (c *fakeCommand) StderrPipe() (io.ReadCloser, error) {
	return &fakeReadCloser{strings.NewReader(strings.Join(c.stderrLines, "\n"))}, nil
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
	runner := NewDockerRunner(nil, func(_ context.Context, name string, args ...string) command {
		gotName = name
		gotArgs = append([]string(nil), args...)
		return &fakeCommand{stdoutLines: []string{"line one", "line two"}}
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
	runner := NewDockerRunner(nil, func(_ context.Context, _ string, _ ...string) command {
		return &fakeCommand{stdoutLines: []string{"boom"}, waitErr: errors.New("exit status 1")}
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

func TestDockerRunnerAppliesResourceLimits(t *testing.T) {
	var gotArgs []string
	runner := NewDockerRunner(nil, func(_ context.Context, _ string, args ...string) command {
		gotArgs = append([]string(nil), args...)
		return &fakeCommand{}
	})

	err := runner.Run(context.Background(), poller.ClaimedExecution{
		ID:             "exec-1",
		Image:          "crawler/go-echo:latest",
		Command:        []string{"./go-echo"},
		CpuCores:       2.0,
		MemoryMB:       512,
		TimeoutSeconds: 300,
	}, func(line string) {})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	wantArgs := []string{"run", "--rm", "--cpus", "2.00", "--memory", "512m", "crawler/go-echo:latest", "./go-echo"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("unexpected args with resource limits: %#v", gotArgs)
	}
}

func TestDockerRunnerCollectsStderrOutput(t *testing.T) {
	runner := NewDockerRunner(nil, func(_ context.Context, _ string, _ ...string) command {
		return &fakeCommand{
			stdoutLines: []string{"stdout line"},
			stderrLines: []string{"stderr line"},
		}
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
	if !contains(logs, "stdout line") || !contains(logs, "stderr line") {
		t.Fatalf("expected both stdout and stderr lines, got: %#v", logs)
	}
}

func TestParseRegistryCredentialsNilForEmpty(t *testing.T) {
	creds, err := ParseRegistryCredentials("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds != nil {
		t.Fatalf("expected nil for empty input, got: %+v", creds)
	}
}

func TestParseRegistryCredentialsParsesValidJSON(t *testing.T) {
	raw := `{"docker.io": {"username": "user", "password": "pass"}}`
	creds, err := ParseRegistryCredentials(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c, ok := creds["docker.io"]; !ok || c.Username != "user" || c.Password != "pass" {
		t.Fatalf("unexpected credentials: %+v", creds)
	}
}

func TestSplitOutputLinesSkipsEmpty(t *testing.T) {
	lines := splitOutputLines("a\n\nb\n  \nc")
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("unexpected lines: %#v", lines)
	}
}

func contains(lines []string, needle string) bool {
	for _, l := range lines {
		if strings.Contains(l, needle) {
			return true
		}
	}
	return false
}
