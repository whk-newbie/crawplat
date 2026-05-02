package runtime

import (
	"context"
	"errors"
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

type RegistryCredential struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type DockerRunner struct {
	newCommand commandFactory
	creds      map[string]RegistryCredential
}

func NewDockerRunner(factory commandFactory, creds map[string]RegistryCredential) *DockerRunner {
	if factory == nil {
		factory = func(ctx context.Context, name string, args ...string) command {
			return exec.CommandContext(ctx, name, args...)
		}
	}
	return &DockerRunner{newCommand: factory, creds: creds}
}

func (r *DockerRunner) Run(ctx context.Context, exec poller.ClaimedExecution, onLog func(string)) error {
	runCtx := ctx
	cancel := func() {}
	if exec.TimeoutSeconds > 0 {
		runCtx, cancel = context.WithTimeout(ctx, time.Duration(exec.TimeoutSeconds)*time.Second)
	}
	defer cancel()

	if cred, ok := r.findCredential(exec); ok {
		login := buildDockerLoginCommand(cred)
		if err := r.runCommand(runCtx, login, onLog); err != nil {
			return err
		}
		pull := buildDockerPullCommand(exec.Image)
		if err := r.runCommand(runCtx, pull, onLog); err != nil {
			return err
		}
	}
	if err := r.runCommand(runCtx, buildDockerRunCommand(exec), onLog); err != nil {
		return err
	}
	return nil
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

func buildDockerPullCommand(image string) []string {
	return []string{"docker", "pull", image}
}

func buildDockerLoginCommand(cred RegistryCredential) []string {
	return []string{"docker", "login", cred.Server, "-u", cred.Username, "-p", cred.Password}
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

func (r *DockerRunner) runCommand(ctx context.Context, cmd []string, onLog func(string)) error {
	if len(cmd) == 0 {
		return errors.New("empty command")
	}
	output, err := r.newCommand(ctx, cmd[0], cmd[1:]...).CombinedOutput()
	for _, line := range splitOutputLines(string(output)) {
		onLog(line)
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}

func (r *DockerRunner) findCredential(exec poller.ClaimedExecution) (RegistryCredential, bool) {
	if len(r.creds) == 0 {
		return RegistryCredential{}, false
	}
	if ref := strings.TrimSpace(strings.ToLower(exec.RegistryAuthRef)); ref != "" {
		cred, ok := r.creds[ref]
		if ok {
			if strings.TrimSpace(cred.Server) == "" {
				cred.Server = ref
			}
			return cred, true
		}
	}
	host := registryHostFromImage(exec.Image)
	cred, ok := r.creds[host]
	if ok {
		if strings.TrimSpace(cred.Server) == "" {
			cred.Server = host
		}
		return cred, true
	}
	return RegistryCredential{}, false
}

func registryHostFromImage(image string) string {
	trimmed := strings.TrimSpace(image)
	if trimmed == "" {
		return "docker.io"
	}
	first := trimmed
	if slash := strings.IndexByte(trimmed, '/'); slash >= 0 {
		first = trimmed[:slash]
	} else {
		return "docker.io"
	}
	if strings.Contains(first, ".") || strings.Contains(first, ":") || first == "localhost" {
		return first
	}
	return "docker.io"
}
