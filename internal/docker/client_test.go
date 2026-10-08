package docker

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ButterHost69/sloper/internal/models"
)

// recordRunner captures the command line and returns canned results, so tests
// can assert on the exact docker invocation without a daemon.
type recordRunner struct {
	calls   [][]string
	results []models.ShellResult
	errs    []error
	idx     int
}

func (r *recordRunner) run(_ context.Context, opts models.ShellOptions) (models.ShellResult, error) {
	r.calls = append(r.calls, append([]string{opts.Command}, opts.Args...))
	i := r.idx
	r.idx++
	if i < len(r.results) {
		res := r.results[i]
		var err error
		if i < len(r.errs) {
			err = r.errs[i]
		}
		return res, err
	}
	return models.ShellResult{}, nil
}

func shellFailure(stderr string) error {
	return &models.ShellCommandExecutionError{
		Message: "Command exited with code 1",
		Result:  models.ShellResult{ExitCode: 1, Stderr: stderr},
	}
}

func newTestClient(r *recordRunner) *Client {
	return New(models.DockerOptions{Binary: "docker", Timeout: time.Second}).WithRunner(r.run)
}

func TestRunBuildsFullCommand(t *testing.T) {
	r := &recordRunner{results: []models.ShellResult{{Stdout: "abc123\n"}}}
	client := newTestClient(r)

	id, err := client.Run(context.Background(), models.ContainerSpec{
		Name:  "sloper-repo-owner-repo-1",
		Image: "sloper-agent:latest",
		Env: map[string]string{
			"GH_REPO_LINK":     "owner/repo",
			"SLOPER_WEB_PORT":  "8080",
			"SLOPER_WEB_TOKEN": "tok",
		},
		Labels: map[string]string{
			models.LabelManaged: "true",
			models.LabelRepo:    "owner/repo",
		},
		Ports: []models.PortMapping{{HostPort: 8081, ContainerPort: 8080, BindAddr: "127.0.0.1"}},
		Volumes: []models.VolumeMount{
			{Source: "sloper-data-owner-repo-1", ContainerPath: "/root/.sloper"},
			{Source: "sloper-repo-owner-repo-1", ContainerPath: "/root/repo"},
		},
		RestartPolicy: "unless-stopped",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if id != "abc123" {
		t.Errorf("container id = %q, want abc123", id)
	}

	got := strings.Join(r.calls[0], " ")
	wants := []string{
		"docker run -d --name sloper-repo-owner-repo-1",
		"--restart unless-stopped",
		"--label sloper.managed=true --label sloper.repo=owner/repo",
		"--env GH_REPO_LINK=owner/repo",
		"--env SLOPER_WEB_PORT=8080",
		"--env SLOPER_WEB_TOKEN=tok",
		"--publish 127.0.0.1:8081:8080",
		"--volume sloper-data-owner-repo-1:/root/.sloper",
		"--volume sloper-repo-owner-repo-1:/root/repo",
		"sloper-agent:latest",
	}
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Errorf("docker run command missing %q:\n%s", want, got)
		}
	}
}

func TestRunRequiresNameAndImage(t *testing.T) {
	client := newTestClient(&recordRunner{})
	if _, err := client.Run(context.Background(), models.ContainerSpec{Image: "x"}); err == nil {
		t.Error("expected an error when the container name is missing")
	}
	if _, err := client.Run(context.Background(), models.ContainerSpec{Name: "x"}); err == nil {
		t.Error("expected an error when the image is missing")
	}
}

func TestInspectParsesState(t *testing.T) {
	jsonl := `{"Id":"0123456789abcdef","Name":"/sloper-repo-owner-repo-1",` +
		`"State":{"Status":"running","Running":true,"ExitCode":0,"StartedAt":"2026-10-09T00:00:00Z",` +
		`"Health":{"Status":"healthy"}},` +
		`"Config":{"Image":"sloper-agent:latest","Labels":{"sloper.managed":"true","sloper.repo":"owner/repo"}},` +
		`"NetworkSettings":{"Ports":{"8080/tcp":[{"HostIp":"127.0.0.1","HostPort":"8081"}]}}}`

	r := &recordRunner{results: []models.ShellResult{{Stdout: jsonl}}}
	client := newTestClient(r)

	state, err := client.Inspect(context.Background(), "sloper-repo-owner-repo-1")
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if state.Name != "sloper-repo-owner-repo-1" {
		t.Errorf("name = %q", state.Name)
	}
	if !state.Running() {
		t.Errorf("state = %q, want running", state.State)
	}
	if state.Health != "healthy" {
		t.Errorf("health = %q, want healthy", state.Health)
	}
	if state.Labels[models.LabelRepo] != "owner/repo" {
		t.Errorf("labels = %v", state.Labels)
	}
	if got := state.HostPortFor(8080); got != 8081 {
		t.Errorf("published port = %d, want 8081", got)
	}
}

func TestInspectMissingContainerIsNotFound(t *testing.T) {
	containerErr := "Error response from daemon: No such container: sloper-x\n"
	imageErr := "Error response from daemon: No such image: missing:latest\n"
	r := &recordRunner{
		results: []models.ShellResult{
			{ExitCode: 1, Stderr: containerErr},
			{ExitCode: 1, Stderr: imageErr},
		},
		errs: []error{shellFailure(containerErr), shellFailure(imageErr)},
	}
	client := newTestClient(r)

	_, err := client.Inspect(context.Background(), "sloper-x")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Inspect error = %v, want ErrNotFound", err)
	}

	exists, err := client.ImageExists(context.Background(), "missing:latest")
	if err != nil {
		t.Fatalf("ImageExists: %v", err)
	}
	if exists {
		t.Error("ImageExists = true for a missing image")
	}
}

func TestAvailableReportsUnreachableDaemon(t *testing.T) {
	stderr := "Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?\n"
	r := &recordRunner{
		results: []models.ShellResult{{ExitCode: 1, Stderr: stderr}},
		errs:    []error{shellFailure(stderr)},
	}
	client := newTestClient(r)

	if err := client.Available(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Available error = %v, want ErrUnavailable", err)
	}
}

func TestMissingBinaryIsUnavailable(t *testing.T) {
	runner := func(context.Context, models.ShellOptions) (models.ShellResult, error) {
		// shell.Run fails at cmd.Start, so the result is empty and the reason
		// only exists in the Go error.
		return models.ShellResult{}, errors.New(`error occured in executing cmd: exec: "docker": executable file not found in $PATH`)
	}
	client := New(models.DockerOptions{}).WithRunner(runner)

	if err := client.Available(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Available error = %v, want ErrUnavailable", err)
	}
}

func TestListFiltersByLabelAndInspectsEachID(t *testing.T) {
	jsonl := `{"Id":"aaa","Name":"/one","State":{"Status":"running","Running":true},"Config":{"Labels":{}}}`
	r := &recordRunner{results: []models.ShellResult{
		{Stdout: "aaa\nbbb\n"},
		{Stdout: jsonl},
		{Stdout: jsonl},
	}}
	client := newTestClient(r)

	states, err := client.List(context.Background(), map[string]string{models.LabelManaged: "true"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(states) != 2 {
		t.Fatalf("got %d containers, want 2", len(states))
	}

	ps := strings.Join(r.calls[0], " ")
	if !strings.Contains(ps, "docker ps --all --no-trunc") {
		t.Errorf("unexpected ps command: %s", ps)
	}
	if !strings.Contains(ps, "--filter label=sloper.managed=true") {
		t.Errorf("ps command is not label filtered: %s", ps)
	}
}

func TestParseContainerPort(t *testing.T) {
	cases := map[string]int{"8080/tcp": 8080, "8080": 8080, "9090/udp": 9090, "nope": 0}
	for input, want := range cases {
		if got := parseContainerPort(input); got != want {
			t.Errorf("parseContainerPort(%q) = %d, want %d", input, got, want)
		}
	}
}
