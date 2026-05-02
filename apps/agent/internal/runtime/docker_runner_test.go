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
	cmd := buildDockerRunCommand(poller.ClaimedExecution{
		Image:          "crawler/go-echo:latest",
		Command:        []string{"./go-echo"},
		CPUCores:       1.5,
		MemoryMB:       768,
		TimeoutSeconds: 60,
	})
	want := []string{"docker", "run", "--rm", "--cpus", "1.5", "--memory", "768m", "crawler/go-echo:latest", "./go-echo"}
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
	}, nil)

	var logs []string
	err := runner.Run(context.Background(), poller.ClaimedExecution{
		ID:             "exec-1",
		Image:          "crawler/go-echo:latest",
		Command:        []string{"./go-echo"},
		CPUCores:       2,
		MemoryMB:       1024,
		TimeoutSeconds: 120,
	}, func(line string) {
		logs = append(logs, line)
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if gotName != "docker" {
		t.Fatalf("unexpected command name: %q", gotName)
	}
	if want := []string{"run", "--rm", "--cpus", "2", "--memory", "1024m", "crawler/go-echo:latest", "./go-echo"}; !reflect.DeepEqual(gotArgs, want) {
		t.Fatalf("unexpected command args: %#v", gotArgs)
	}
	if !reflect.DeepEqual(logs, []string{"line one", "line two"}) {
		t.Fatalf("unexpected logs: %#v", logs)
	}
}

func TestDockerRunnerReturnsCommandFailure(t *testing.T) {
	runner := NewDockerRunner(func(_ context.Context, _ string, _ ...string) command {
		return &fakeCommand{output: []byte("boom\n"), err: errors.New("exit status 1")}
	}, nil)

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

func TestDockerRunnerPerformsLoginAndPullWhenCredentialExists(t *testing.T) {
	var invocations [][]string
	runner := NewDockerRunner(func(_ context.Context, name string, args ...string) command {
		invocations = append(invocations, append([]string{name}, args...))
		switch args[0] {
		case "login":
			return &fakeCommand{output: []byte("login ok\n")}
		case "pull":
			return &fakeCommand{output: []byte("pull ok\n")}
		default:
			return &fakeCommand{output: []byte("run ok\n")}
		}
	}, map[string]RegistryCredential{
		"ghcr.io": {
			Username: "user",
			Password: "pass",
		},
	})

	var logs []string
	err := runner.Run(context.Background(), poller.ClaimedExecution{
		ID:              "exec-1",
		RegistryAuthRef: "ghcr.io",
		Image:           "ghcr.io/acme/crawler:latest",
		Command:         []string{"./crawler"},
	}, func(line string) {
		logs = append(logs, line)
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if len(invocations) != 3 {
		t.Fatalf("expected 3 docker invocations, got %#v", invocations)
	}
	if !reflect.DeepEqual(invocations[0], []string{"docker", "login", "ghcr.io", "-u", "user", "-p", "pass"}) {
		t.Fatalf("unexpected login invocation: %#v", invocations[0])
	}
	if !reflect.DeepEqual(invocations[1], []string{"docker", "pull", "ghcr.io/acme/crawler:latest"}) {
		t.Fatalf("unexpected pull invocation: %#v", invocations[1])
	}
	if !reflect.DeepEqual(invocations[2], []string{"docker", "run", "--rm", "ghcr.io/acme/crawler:latest", "./crawler"}) {
		t.Fatalf("unexpected run invocation: %#v", invocations[2])
	}
	if !reflect.DeepEqual(logs, []string{"login ok", "pull ok", "run ok"}) {
		t.Fatalf("unexpected logs: %#v", logs)
	}
}

func TestDockerRunnerUsesRegistryAuthRefOverImageHost(t *testing.T) {
	var invocations [][]string
	runner := NewDockerRunner(func(_ context.Context, name string, args ...string) command {
		invocations = append(invocations, append([]string{name}, args...))
		return &fakeCommand{}
	}, map[string]RegistryCredential{
		"my-ghcr": {
			Server:   "ghcr.io",
			Username: "user",
			Password: "pass",
		},
	})

	err := runner.Run(context.Background(), poller.ClaimedExecution{
		ID:              "exec-1",
		RegistryAuthRef: "my-ghcr",
		Image:           "ghcr.io/acme/crawler:latest",
		Command:         []string{"./crawler"},
	}, func(string) {})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(invocations) == 0 {
		t.Fatalf("expected invocations, got %#v", invocations)
	}
	if !reflect.DeepEqual(invocations[0], []string{"docker", "login", "ghcr.io", "-u", "user", "-p", "pass"}) {
		t.Fatalf("expected login by auth ref, got %#v", invocations[0])
	}
}
