package models

import (
	"strings"
	"time"
)

// DockerOptions configures the docker CLI client used by the controller to run
// one container per attached repo.
type DockerOptions struct {
	Binary         string        // docker binary, defaults to "docker"
	Timeout        time.Duration // per-command deadline; 0 = no deadline
	MaxOutputBytes int           // cap on captured stdout/stderr per command
}

// PortMapping publishes one container port on the host.
type PortMapping struct {
	HostPort      int
	ContainerPort int
	BindAddr      string // host address to bind, defaults to 127.0.0.1
}

// VolumeMount attaches a named docker volume (or a host path when Source starts
// with "/") at ContainerPath.
type VolumeMount struct {
	Source        string
	ContainerPath string
	ReadOnly      bool
}

// ContainerSpec is the subset of `docker run` options the controller needs.
// Every container it starts is labeled, so a run of the controller can find and
// reconcile the containers it owns without a side channel.
type ContainerSpec struct {
	Name          string
	Image         string
	Command       []string
	Env           map[string]string
	Labels        map[string]string
	Ports         []PortMapping
	Volumes       []VolumeMount
	RestartPolicy string // no | on-failure | always | unless-stopped
	Network       string
}

// ContainerState is what docker inspect tells us about a container. Only the
// fields the controller reconciles on are modeled; unknown keys are ignored.
type ContainerState struct {
	Name      string
	ID        string
	Image     string
	State     string // created | running | paused | restarting | exited | dead
	Status    string // human readable, e.g. "Up 3 minutes"
	Health    string // starting | healthy | unhealthy, empty when no healthcheck
	ExitCode  int
	StartedAt string
	Labels    map[string]string
	HostPorts map[int]int // container port -> published host port
}

// Running reports whether the container is up right now.
func (s ContainerState) Running() bool { return s.State == "running" }

// Failed reports whether the container has stopped with a non-zero exit code.
func (s ContainerState) Failed() bool { return s.State == "exited" && s.ExitCode != 0 }

// HostPortFor returns the published host port for a container port, or 0 when
// that port is not published.
func (s ContainerState) HostPortFor(containerPort int) int {
	if s.HostPorts == nil {
		return 0
	}
	return s.HostPorts[containerPort]
}

// Repo status values stored in the repos table and reported to the console.
const (
	RepoStatusPending  = "pending"  // attached, container not created yet
	RepoStatusStarting = "starting" // container exists but is not running yet
	RepoStatusRunning  = "running"
	RepoStatusStopped  = "stopped"
	RepoStatusError    = "error" // the last reconcile attempt failed
)

// Desired state values stored in the repos table.
const (
	RepoDesiredRunning = "running"
	RepoDesiredStopped = "stopped"
)

// AttachedRepo is one repo the controller runs a container for. It is the JSON
// shape the console consumes, so its fields are tagged and snake_cased like the
// other API payloads.
type AttachedRepo struct {
	ID             int64  `json:"id"`
	Link           string `json:"link"` // owner/repo
	Name           string `json:"name"`
	CloneURL       string `json:"clone_url"`
	ContainerName  string `json:"container_name"`
	ContainerID    string `json:"container_id"`
	Image          string `json:"image"`
	HostPort       int    `json:"host_port"`
	APIURL         string `json:"api_url"` // sloper-web base URL of the repo instance
	Token          string `json:"token"`   // SLOPER_WEB_TOKEN of the repo instance
	DesiredState   string `json:"desired_state"`
	Status         string `json:"status"`
	StatusMessage  string `json:"status_message"`
	ContainerState string `json:"container_state"` // raw docker state when known
	RestartCount   int    `json:"restart_count"`
	AttachedAt     string `json:"attached_at"`
	UpdatedAt      string `json:"updated_at"`
}

// ContainerLabel keys every controller-managed container carries. The values are
// how the controller rediscovers its own containers after a restart.
const (
	LabelManaged = "sloper.managed"
	LabelRepo    = "sloper.repo"
	LabelPort    = "sloper.web.port"
)

// DefaultRepoContainerPort is the port sloper-web listens on inside every repo
// container (setup/compose.yml sets the same value).
const DefaultRepoContainerPort = 8080

// Slugify turns an arbitrary label into a docker-safe name fragment: lowercase
// alphanumerics and dashes, never empty, never starting with a dash.
func Slugify(value string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == '-' || r == '_' || r == '.' || r == '/' || r == ' ':
			if !lastDash && b.Len() > 0 {
				b.WriteRune('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "repo"
	}
	return out
}
