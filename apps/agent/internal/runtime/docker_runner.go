package runtime

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

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
	runCtx := ctx
	cancel := func() {}
	if exec.TimeoutSeconds > 0 {
		runCtx, cancel = context.WithTimeout(ctx, time.Duration(exec.TimeoutSeconds)*time.Second)
	}
	defer cancel()

	cmd := buildDockerRunCommand(exec)
	output, err := r.newCommand(runCtx, cmd[0], cmd[1:]...).CombinedOutput()
	for _, line := range splitOutputLines(string(output)) {
		onLog(line)
	}
	if runCtx.Err() != nil {
		return runCtx.Err()
	}
	return err
}

func buildDockerRunCommand(exec poller.ClaimedExecution) []string {
	args := []string{"docker", "run", "--rm"}
	if exec.CPUCores > 0 {
		args = append(args, "--cpus", strconv.FormatFloat(exec.CPUCores, 'f', -1, 64))
	}
	if exec.MemoryMB > 0 {
		args = append(args, "--memory", fmt.Sprintf("%dm", exec.MemoryMB))
	}
	args = append(args, exec.Image)
	return append(args, exec.Command...)
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
