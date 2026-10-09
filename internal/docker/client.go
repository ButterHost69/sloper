// Package docker drives the docker CLI to run one container per attached repo.
//
// Sloper already talks to GitHub and git by shelling out to their CLIs, so the
// container runtime follows the same pattern instead of pulling in the docker
// SDK: the controller needs `run`, `start`, `stop`, `rm`, `inspect` and `ps`,
// and shelling out keeps the binary dependency-free and the exact command lines
// visible in the log.
package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/shell"
)

// ErrNotFound is returned by Inspect when docker has no such container.
var ErrNotFound = errors.New("docker: container not found")

// ErrUnavailable is returned when the docker daemon cannot be reached.
var ErrUnavailable = errors.New("docker: daemon unavailable")

// Runner runs one command and returns its captured output. It exists so tests
// can assert on the exact docker command line without a daemon.
type Runner func(ctx context.Context, opts models.ShellOptions) (models.ShellResult, error)

// Client is a thin, typed wrapper over the docker CLI.
type Client struct {
	bin     string
	timeout time.Duration
	maxOut  int
	run     Runner
}

// New builds a client. The binary defaults to "docker" on PATH.
func New(opts models.DockerOptions) *Client {
	bin := strings.TrimSpace(opts.Binary)
	if bin == "" {
		bin = "docker"
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	maxOut := opts.MaxOutputBytes
	if maxOut <= 0 {
		maxOut = 512 * 1024
	}
	return &Client{bin: bin, timeout: timeout, maxOut: maxOut, run: shell.Run}
}

// WithRunner swaps the command runner (used by tests).
func (c *Client) WithRunner(r Runner) *Client {
	if r != nil {
		c.run = r
	}
	return c
}

// exec runs the docker CLI with a deadline and captured output.
func (c *Client) exec(ctx context.Context, timeout time.Duration, args ...string) (models.ShellResult, error) {
	if timeout <= 0 {
		timeout = c.timeout
	}
	res, err := c.run(ctx, models.ShellOptions{
		Command:          c.bin,
		Args:             args,
		Timeout:          timeout,
		MaxCapturedBytes: c.maxOut,
	})
	if err != nil {
		return res, wrapDockerError(args, res, err)
	}
	return res, nil
}

// wrapDockerError classifies the CLI's stderr so callers can branch on
// "missing container" and "no daemon" without parsing messages themselves. The
// Go error text is part of the haystack because a missing docker binary never
// reaches the CLI: shell.Run fails at start and returns an empty result.
func wrapDockerError(args []string, res models.ShellResult, err error) error {
	haystack := res.Stderr + res.Stdout
	if err != nil {
		haystack += " " + err.Error()
	}
	stderr := strings.ToLower(haystack)
	switch {
	case strings.Contains(stderr, "no such container"),
		strings.Contains(stderr, "no such object"),
		strings.Contains(stderr, "no such image"),
		strings.Contains(stderr, "no such volume"):
		return fmt.Errorf("%w: docker %s", ErrNotFound, strings.Join(args, " "))
	case strings.Contains(stderr, "cannot connect to the docker daemon"),
		strings.Contains(stderr, "is the docker daemon running"),
		strings.Contains(stderr, "docker: not found"),
		strings.Contains(stderr, "docker\": executable file not found"),
		strings.Contains(stderr, "executable file not found"):
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return fmt.Errorf("docker %s: %w", strings.Join(args, " "), err)
}

// Available checks that the docker CLI exists and its daemon answers.
func (c *Client) Available(ctx context.Context) error {
	if _, err := c.exec(ctx, 15*time.Second, "version", "--format", "{{.Server.Version}}"); err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("%w: %v", ErrUnavailable, err)
		}
		return err
	}
	return nil
}

// ImageExists reports whether an image is present locally.
func (c *Client) ImageExists(ctx context.Context, image string) (bool, error) {
	_, err := c.exec(ctx, 20*time.Second, "image", "inspect", "--format", "{{.Id}}", image)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	return false, err
}

// Pull fetches an image. Pulling is slow, so callers decide the deadline.
func (c *Client) Pull(ctx context.Context, image string, timeout time.Duration) error {
	_, err := c.exec(ctx, timeout, "pull", image)
	return err
}

// Run creates and starts a detached container, returning its id. Volumes are
// created by docker on demand, so a repo container is a single command.
func (c *Client) Run(ctx context.Context, spec models.ContainerSpec) (string, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return "", errors.New("docker: container name is required")
	}
	if strings.TrimSpace(spec.Image) == "" {
		return "", errors.New("docker: image is required")
	}

	args := []string{"run", "-d", "--name", spec.Name}

	policy := spec.RestartPolicy
	if policy == "" {
		policy = "unless-stopped"
	}
	args = append(args, "--restart", policy)

	for _, key := range sortedKeys(spec.Labels) {
		args = append(args, "--label", key+"="+spec.Labels[key])
	}
	for _, key := range sortedKeys(spec.Env) {
		args = append(args, "--env", key+"="+spec.Env[key])
	}
	for _, p := range spec.Ports {
		bind := p.BindAddr
		if bind == "" {
			bind = "127.0.0.1"
		}
		args = append(args, "--publish", fmt.Sprintf("%s:%d:%d", bind, p.HostPort, p.ContainerPort))
	}
	for _, v := range spec.Volumes {
		mount := v.Source + ":" + v.ContainerPath
		if v.ReadOnly {
			mount += ":ro"
		}
		args = append(args, "--volume", mount)
	}
	if spec.Network != "" {
		args = append(args, "--network", spec.Network)
	}

	args = append(args, spec.Image)
	args = append(args, spec.Command...)

	res, err := c.exec(ctx, 0, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(res.Stdout), nil
}

// Start starts an existing container.
func (c *Client) Start(ctx context.Context, name string) error {
	_, err := c.exec(ctx, 60*time.Second, "start", name)
	return err
}

// Stop stops a running container, giving it timeoutSec to exit cleanly.
func (c *Client) Stop(ctx context.Context, name string, timeoutSec int) error {
	if timeoutSec <= 0 {
		timeoutSec = 10
	}
	_, err := c.exec(ctx, time.Duration(timeoutSec+20)*time.Second, "stop", "--time", strconv.Itoa(timeoutSec), name)
	return err
}

// Remove deletes a container. Named volumes are never touched here: the repo's
// data lives in them and RemoveVolume is the only thing that deletes it.
func (c *Client) Remove(ctx context.Context, name string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, name)
	_, err := c.exec(ctx, 60*time.Second, args...)
	return err
}

// RemoveVolume deletes a named volume. Used only when a repo is detached with
// purge, because the volume holds the repo clone, worktrees and pipeline DB.
func (c *Client) RemoveVolume(ctx context.Context, name string) error {
	_, err := c.exec(ctx, 60*time.Second, "volume", "rm", "--force", name)
	return err
}

// Logs returns the tail of a container's combined stdout/stderr.
func (c *Client) Logs(ctx context.Context, name string, tail int) (string, error) {
	if tail <= 0 {
		tail = 200
	}
	res, err := c.exec(ctx, 30*time.Second, "logs", "--tail", strconv.Itoa(tail), name)
	if err != nil {
		return "", err
	}
	return res.Stdout + res.Stderr, nil
}

// inspectResult mirrors the subset of `docker inspect` JSON we consume.
type inspectResult struct {
	ID    string `json:"Id"`
	Name  string `json:"Name"`
	State struct {
		Status     string `json:"Status"`
		Running    bool   `json:"Running"`
		ExitCode   int    `json:"ExitCode"`
		StartedAt  string `json:"StartedAt"`
		FinishedAt string `json:"FinishedAt"`
		Health     *struct {
			Status string `json:"Status"`
		} `json:"Health"`
	} `json:"State"`
	Config struct {
		Image  string            `json:"Image"`
		Labels map[string]string `json:"Labels"`
	} `json:"Config"`
	NetworkSettings struct {
		Ports map[string][]struct {
			HostIP   string `json:"HostIp"`
			HostPort string `json:"HostPort"`
		} `json:"Ports"`
	} `json:"NetworkSettings"`
}

// Inspect returns the current state of one container.
func (c *Client) Inspect(ctx context.Context, name string) (models.ContainerState, error) {
	res, err := c.exec(ctx, 30*time.Second, "inspect", "--type", "container", "--format", "{{json .}}", name)
	if err != nil {
		return models.ContainerState{}, err
	}
	return parseInspect(res.Stdout)
}

func parseInspect(stdout string) (models.ContainerState, error) {
	line := strings.TrimSpace(stdout)
	if line == "" {
		return models.ContainerState{}, errors.New("docker: empty inspect output")
	}
	var raw inspectResult
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return models.ContainerState{}, fmt.Errorf("docker: decode inspect: %w", err)
	}

	state := models.ContainerState{
		Name:      strings.TrimPrefix(raw.Name, "/"),
		ID:        raw.ID,
		Image:     raw.Config.Image,
		State:     raw.State.Status,
		ExitCode:  raw.State.ExitCode,
		StartedAt: raw.State.StartedAt,
		Labels:    raw.Config.Labels,
		HostPorts: map[int]int{},
	}
	if raw.State.Health != nil {
		state.Health = raw.State.Health.Status
	}
	if state.State == "" {
		if raw.State.Running {
			state.State = "running"
		} else {
			state.State = "exited"
		}
	}
	for portSpec, bindings := range raw.NetworkSettings.Ports {
		containerPort := parseContainerPort(portSpec)
		if containerPort == 0 {
			continue
		}
		for _, b := range bindings {
			if p, err := strconv.Atoi(b.HostPort); err == nil && p > 0 {
				state.HostPorts[containerPort] = p
				break
			}
		}
	}
	return state, nil
}

// parseContainerPort extracts 8080 from "8080/tcp".
func parseContainerPort(spec string) int {
	spec = strings.TrimSpace(spec)
	if i := strings.Index(spec, "/"); i >= 0 {
		spec = spec[:i]
	}
	p, err := strconv.Atoi(spec)
	if err != nil {
		return 0
	}
	return p
}

// List returns the containers carrying every label in the filter.
func (c *Client) List(ctx context.Context, labels map[string]string) ([]models.ContainerState, error) {
	args := []string{"ps", "--all", "--no-trunc", "--format", "{{.ID}}"}
	for _, key := range sortedKeys(labels) {
		args = append(args, "--filter", "label="+key+"="+labels[key])
	}
	res, err := c.exec(ctx, 30*time.Second, args...)
	if err != nil {
		return nil, err
	}

	var out []models.ContainerState
	for _, id := range strings.Fields(res.Stdout) {
		state, err := c.Inspect(ctx, id)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				continue // removed between ps and inspect
			}
			return nil, err
		}
		out = append(out, state)
	}
	return out, nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
