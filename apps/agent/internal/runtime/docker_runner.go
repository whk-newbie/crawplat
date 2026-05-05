package runtime

import (
	"context"
	"os/exec"
	"strings"

	"crawler-platform/apps/agent/internal/poller"
)

type command interface {
	CombinedOutput() ([]byte, error)
}

type commandFactory func(ctx context.Context, name string, args ...string) command

type DockerRunner struct {
	newCommand commandFactory
}

func NewDockerRunner(factory commandFactory) *DockerRunner {
	if factory == nil {
		factory = func(ctx context.Context, name string, args ...string) command {
			return exec.CommandContext(ctx, name, args...)
		}
	}
	return &DockerRunner{newCommand: factory}
}

func (r *DockerRunner) Run(ctx context.Context, exec poller.ClaimedExecution, onLog func(string)) error {
	cmd := buildDockerRunCommand(exec.Image, exec.Command)
	output, err := r.newCommand(ctx, cmd[0], cmd[1:]...).CombinedOutput()
	for _, line := range splitOutputLines(string(output)) {
		onLog(line)
	}
	return err
}

func buildDockerRunCommand(image string, command []string) []string {
	args := []string{"docker", "run", "--rm", image}
	return append(args, command...)
}

func splitOutputLines(output string) []string {
	var lines []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}
