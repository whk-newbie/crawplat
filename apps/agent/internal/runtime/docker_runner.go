package runtime

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"crawler-platform/apps/agent/internal/poller"
)

type command interface {
	Start() error
	Wait() error
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
}

type commandFactory func(ctx context.Context, name string, args ...string) command

type DockerRunner struct {
	credentials map[string]RegistryCredential
	newCommand  commandFactory
}

type execCommand struct {
	*exec.Cmd
}

func (c *execCommand) StdoutPipe() (io.ReadCloser, error) {
	return c.Cmd.StdoutPipe()
}

func (c *execCommand) StderrPipe() (io.ReadCloser, error) {
	return c.Cmd.StderrPipe()
}

func NewDockerRunner(credentials map[string]RegistryCredential, factory commandFactory) *DockerRunner {
	if factory == nil {
		factory = func(ctx context.Context, name string, args ...string) command {
			return &execCommand{exec.CommandContext(ctx, name, args...)}
		}
	}
	return &DockerRunner{credentials: credentials, newCommand: factory}
}

func (r *DockerRunner) Run(ctx context.Context, exec poller.ClaimedExecution, onLog func(string)) error {
	cred, needsAuth := r.credentials[registryHostFromImage(exec.Image)]

	if needsAuth {
		if err := r.dockerLogin(ctx, cred); err != nil {
			return fmt.Errorf("docker login %s: %w", cred.Server, err)
		}
		if err := r.dockerPull(ctx, exec.Image); err != nil {
			return fmt.Errorf("docker pull %s: %w", exec.Image, err)
		}
	}

	return r.dockerRun(ctx, exec, onLog)
}

func (r *DockerRunner) dockerLogin(ctx context.Context, cred RegistryCredential) error {
	args := []string{"login", cred.Server, "-u", cred.Username, "-p", cred.Password}
	output, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func (r *DockerRunner) dockerPull(ctx context.Context, image string) error {
	output, err := exec.CommandContext(ctx, "docker", "pull", image).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func (r *DockerRunner) dockerRun(ctx context.Context, exec poller.ClaimedExecution, onLog func(string)) error {
	args := []string{"run", "--rm"}

	if exec.CpuCores > 0 {
		args = append(args, "--cpus", fmt.Sprintf("%.2f", exec.CpuCores))
	}
	if exec.MemoryMB > 0 {
		args = append(args, "--memory", fmt.Sprintf("%dm", exec.MemoryMB))
	}

	if exec.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(exec.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	args = append(args, exec.Image)
	args = append(args, exec.Command...)

	cmd := r.newCommand(ctx, "docker", args[0:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		scanLines(stdout, onLog)
	}()
	go func() {
		defer wg.Done()
		scanLines(stderr, onLog)
	}()
	wg.Wait()

	return cmd.Wait()
}

func scanLines(r io.Reader, onLog func(string)) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			onLog(line)
		}
	}
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
