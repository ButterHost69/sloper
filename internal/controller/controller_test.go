package controller

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ButterHost69/sloper/internal/docker"
	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/storage"
	"go.uber.org/zap"
)

// fakeDocker is an in-memory docker daemon: enough behavior for the reconcile
// paths the controller actually uses.
type fakeDocker struct {
	mu sync.Mutex

	available bool
	images    map[string]bool
	pulled    []string
	pullErr   error

	containers map[string]models.ContainerState
	runErr     error

	removeVolumeErr error

	runs           []models.ContainerSpec
	starts         []string
	stops          []string
	removed        []string
	removedVolumes []string
}

func newFakeDocker() *fakeDocker {
	return &fakeDocker{
		available:  true,
		images:     map[string]bool{},
		containers: map[string]models.ContainerState{},
	}
}

func (f *fakeDocker) Available(context.Context) error {
	if !f.available {
		return fmt.Errorf("%w: fake daemon down", docker.ErrUnavailable)
	}
	return nil
}

func (f *fakeDocker) ImageExists(_ context.Context, image string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.images[image], nil
}

func (f *fakeDocker) Pull(_ context.Context, image string, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pulled = append(f.pulled, image)
	if f.pullErr != nil {
		return f.pullErr
	}
	f.images[image] = true
	return nil
}

func (f *fakeDocker) Run(_ context.Context, spec models.ContainerSpec) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.runErr != nil {
		return "", f.runErr
	}
	f.runs = append(f.runs, spec)
	id := fmt.Sprintf("id%06d", len(f.runs))
	f.containers[spec.Name] = models.ContainerState{
		Name: spec.Name, ID: id, State: "created", Image: spec.Image, Labels: spec.Labels,
	}
	return id, nil
}

func (f *fakeDocker) Start(_ context.Context, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.starts = append(f.starts, name)
	state, ok := f.containers[name]
	if !ok {
		return fmt.Errorf("%w: %s", docker.ErrNotFound, name)
	}
	state.State = "running"
	f.containers[name] = state
	return nil
}

func (f *fakeDocker) Stop(_ context.Context, name string, _ int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stops = append(f.stops, name)
	if _, ok := f.containers[name]; !ok {
		return fmt.Errorf("%w: %s", docker.ErrNotFound, name)
	}
	return nil
}

func (f *fakeDocker) Remove(_ context.Context, name string, _ bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.removed = append(f.removed, name)
	if _, ok := f.containers[name]; !ok {
		return fmt.Errorf("%w: %s", docker.ErrNotFound, name)
	}
	delete(f.containers, name)
	return nil
}

func (f *fakeDocker) RemoveVolume(_ context.Context, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.removedVolumes = append(f.removedVolumes, name)
	return f.removeVolumeErr
}

func (f *fakeDocker) Inspect(_ context.Context, name string) (models.ContainerState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	state, ok := f.containers[name]
	if !ok {
		return models.ContainerState{}, fmt.Errorf("%w: %s", docker.ErrNotFound, name)
	}
	return state, nil
}

func (f *fakeDocker) List(_ context.Context, _ map[string]string) ([]models.ContainerState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]models.ContainerState, 0, len(f.containers))
	for _, state := range f.containers {
		out = append(out, state)
	}
	return out, nil
}

func (f *fakeDocker) Logs(context.Context, string, int) (string, error) {
	return "fake logs", nil
}

// setRunning marks a container as up, the way a real `docker inspect` would.
func (f *fakeDocker) setRunning(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	state := f.containers[name]
	state.State = "running"
	state.ID = "abcdef123456"
	state.HostPorts = map[int]int{models.DefaultRepoContainerPort: 18080}
	f.containers[name] = state
}

func (f *fakeDocker) setExited(name string, exitCode int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	state := f.containers[name]
	state.State = "exited"
	state.ExitCode = exitCode
	f.containers[name] = state
}

func (f *fakeDocker) hasContainer(name string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.containers[name]
	return ok
}

func (f *fakeDocker) lastRun() models.ContainerSpec {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.runs) == 0 {
		return models.ContainerSpec{}
	}
	return f.runs[len(f.runs)-1]
}

// newTestService wires a service against a temp sqlite file and a fake daemon.
func newTestService(t *testing.T, fake *fakeDocker) *Service {
	t.Helper()

	db, err := storage.OpenDB(filepath.Join(t.TempDir(), "controller.sqlite"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := storage.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	return New(Options{
		DB:                storage.NewRepositories(db),
		Docker:            fake,
		Log:               zap.NewNop(),
		Image:             "sloper-agent:test",
		PortMin:           18080,
		PortMax:           18089,
		PublishAddr:       "127.0.0.1",
		PublicHost:        "localhost",
		InheritEnv:        map[string]string{"GH_TOKEN": "gh-token", "AGENT_MODEL": "anthropic/claude"},
		ReconcileInterval: time.Hour, // tests drive Reconcile directly
	})
}

func TestAttachValidatesConfiguration(t *testing.T) {
	svc := newTestService(t, newFakeDocker())
	svc.opts.InheritEnv = map[string]string{"AGENT_MODEL": "anthropic/claude"}

	_, err := svc.Attach(context.Background(), "owner/repo")
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("Attach error = %v, want ErrNotConfigured", err)
	}
	if !strings.Contains(err.Error(), "GH_TOKEN") {
		t.Errorf("error should name the missing variable: %v", err)
	}
}

func TestAttachRejectsBadAndDuplicateLinks(t *testing.T) {
	svc := newTestService(t, newFakeDocker())

	if _, err := svc.Attach(context.Background(), "not-a-repo"); !errors.Is(err, ErrInvalidLink) {
		t.Fatalf("Attach error = %v, want ErrInvalidLink", err)
	}

	if _, err := svc.Attach(context.Background(), "owner/repo"); err != nil {
		t.Fatalf("Attach: %v", err)
	}
	if _, err := svc.Attach(context.Background(), "https://github.com/owner/repo.git"); !errors.Is(err, ErrAlreadyAttached) {
		t.Fatalf("second Attach error = %v, want ErrAlreadyAttached", err)
	}
}

func TestAttachRecordsRepoWithPortAndToken(t *testing.T) {
	fake := newFakeDocker()
	svc := newTestService(t, fake)

	repo, err := svc.Attach(context.Background(), "ButterHost69/sloper")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}

	if repo.Link != "ButterHost69/sloper" || repo.Name != "sloper" {
		t.Errorf("unexpected repo identity: %+v", repo)
	}
	if repo.Status != models.RepoStatusPending {
		t.Errorf("status = %q, want pending", repo.Status)
	}
	if repo.HostPort < 18080 || repo.HostPort > 18089 {
		t.Errorf("host port %d outside the configured range", repo.HostPort)
	}
	if repo.APIURL != fmt.Sprintf("http://localhost:%d", repo.HostPort) {
		t.Errorf("api url = %q", repo.APIURL)
	}
	if repo.Token == "" {
		t.Error("expected a generated instance token")
	}
	if !strings.Contains(repo.ContainerName, "sloper-repo-butterhost69-sloper-") {
		t.Errorf("container name = %q", repo.ContainerName)
	}
}

func TestReconcileCreatesContainerFromRow(t *testing.T) {
	fake := newFakeDocker()
	fake.images["sloper-agent:test"] = true
	svc := newTestService(t, fake)

	repo, err := svc.Attach(context.Background(), "ButterHost69/sloper")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	spec := fake.lastRun()
	if spec.Name != repo.ContainerName {
		t.Errorf("container name = %q, want %q", spec.Name, repo.ContainerName)
	}
	if spec.Image != "sloper-agent:test" {
		t.Errorf("image = %q", spec.Image)
	}
	if spec.Env["SLOPER_REPO_LINK"] != "ButterHost69/sloper" {
		t.Errorf("SLOPER_REPO_LINK = %q", spec.Env["SLOPER_REPO_LINK"])
	}
	// The repo link is assigned per container, never configured through the
	// controller's own environment.
	if _, ok := spec.Env["GH_REPO_LINK"]; ok {
		t.Error("GH_REPO_LINK leaked into the container environment")
	}
	if spec.Env["GH_TOKEN"] != "gh-token" || spec.Env["AGENT_MODEL"] != "anthropic/claude" {
		t.Errorf("credentials were not inherited: %v", spec.Env)
	}
	if spec.Env["SLOPER_WEB_TOKEN"] != repo.Token {
		t.Errorf("SLOPER_WEB_TOKEN = %q, want the generated token", spec.Env["SLOPER_WEB_TOKEN"])
	}
	if spec.Env["SLOPER_WEB_ADDR"] != "0.0.0.0" {
		t.Errorf("SLOPER_WEB_ADDR = %q, want 0.0.0.0", spec.Env["SLOPER_WEB_ADDR"])
	}
	if spec.Labels[models.LabelManaged] != "true" || spec.Labels[models.LabelRepo] != "ButterHost69/sloper" {
		t.Errorf("labels = %v", spec.Labels)
	}
	if len(spec.Ports) != 1 || spec.Ports[0].HostPort != repo.HostPort || spec.Ports[0].ContainerPort != models.DefaultRepoContainerPort {
		t.Errorf("ports = %+v", spec.Ports)
	}
	if len(spec.Volumes) != 2 ||
		spec.Volumes[0].ContainerPath != "/root/.sloper" ||
		spec.Volumes[1].ContainerPath != "/root/repo" {
		t.Errorf("volumes = %+v", spec.Volumes)
	}

	stored, err := svc.Get(context.Background(), repo.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if stored.Status != models.RepoStatusStarting {
		t.Errorf("status = %q, want starting", stored.Status)
	}

	// The next pass sees the container up and reports it as running.
	fake.setRunning(repo.ContainerName)
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	stored, _ = svc.Get(context.Background(), repo.ID)
	if stored.Status != models.RepoStatusRunning {
		t.Errorf("status = %q, want running", stored.Status)
	}
	if stored.ContainerID != "abcdef123456" {
		t.Errorf("container id = %q", stored.ContainerID)
	}
}

func TestReconcilePullsMissingImageInBackground(t *testing.T) {
	fake := newFakeDocker() // image is not present
	svc := newTestService(t, fake)

	repo, err := svc.Attach(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	stored, _ := svc.Get(context.Background(), repo.ID)
	if stored.Status != models.RepoStatusPending || !strings.Contains(stored.StatusMessage, "pulling image") {
		t.Errorf("status = %q / %q, want a pending pull", stored.Status, stored.StatusMessage)
	}
	if len(fake.runs) != 0 {
		t.Error("a container was created before the image was ready")
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && !fake.images["sloper-agent:test"] {
		time.Sleep(5 * time.Millisecond)
	}
	if !fake.images["sloper-agent:test"] {
		t.Fatal("the image was never pulled")
	}

	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile after pull: %v", err)
	}
	if len(fake.runs) != 1 {
		t.Fatalf("containers created = %d, want 1", len(fake.runs))
	}
}

func TestReconcileStartsStoppedContainerThenRecreatesWhenStartFails(t *testing.T) {
	fake := newFakeDocker()
	fake.images["sloper-agent:test"] = true
	svc := newTestService(t, fake)

	repo, err := svc.Attach(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	fake.setExited(repo.ContainerName, 137)

	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if len(fake.starts) != 1 {
		t.Fatalf("start calls = %d, want 1", len(fake.starts))
	}
	stored, _ := svc.Get(context.Background(), repo.ID)
	if stored.Status != models.RepoStatusStarting {
		t.Errorf("status = %q, want starting", stored.Status)
	}
	// A crash loop must be visible, not a permanent silent "starting".
	if !strings.Contains(stored.StatusMessage, "exited with code 137") {
		t.Errorf("status message = %q, want the exit code surfaced", stored.StatusMessage)
	}
	if stored.RestartCount != 1 {
		t.Errorf("restart count = %d, want 1", stored.RestartCount)
	}

	// A container that vanishes mid-flight is recreated from the row.
	if err := fake.Remove(context.Background(), repo.ContainerName, true); err != nil {
		t.Fatalf("fake remove: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if len(fake.runs) != 2 {
		t.Errorf("containers created = %d, want 2 after the container disappeared", len(fake.runs))
	}
}

func TestFailedImagePullBacksOff(t *testing.T) {
	fake := newFakeDocker()
	fake.pullErr = errors.New("pull access denied for sloper-agent")
	svc := newTestService(t, fake)

	repo, err := svc.Attach(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	waitFor(t, func() bool {
		fake.mu.Lock()
		defer fake.mu.Unlock()
		return len(fake.pulled) == 1
	}, "the first pull attempt")

	// Every further pass must reuse the remembered failure instead of pulling.
	for i := 0; i < 3; i++ {
		if err := svc.Reconcile(context.Background()); err != nil {
			t.Fatalf("Reconcile: %v", err)
		}
	}
	fake.mu.Lock()
	pulls := len(fake.pulled)
	fake.mu.Unlock()
	if pulls != 1 {
		t.Fatalf("pull attempts = %d, want 1 while the cooldown holds", pulls)
	}

	stored, _ := svc.Get(context.Background(), repo.ID)
	if stored.Status != models.RepoStatusError || !strings.Contains(stored.StatusMessage, "pull access denied") {
		t.Errorf("status = %q / %q, want the pull error surfaced", stored.Status, stored.StatusMessage)
	}

	// Restart is the operator's retry signal: it clears the cooldown.
	if _, err := svc.Restart(context.Background(), repo.ID); err != nil {
		t.Fatalf("Restart: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	waitFor(t, func() bool {
		fake.mu.Lock()
		defer fake.mu.Unlock()
		return len(fake.pulled) == 2
	}, "a pull retry after restart")
}

func waitFor(t *testing.T, condition func() bool, what string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestReAttachReusesContainerNameAndVolumes(t *testing.T) {
	fake := newFakeDocker()
	fake.images["sloper-agent:test"] = true
	svc := newTestService(t, fake)

	first, err := svc.Attach(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	firstSpec := fake.lastRun()

	if err := svc.Detach(context.Background(), first.ID, false); err != nil {
		t.Fatalf("Detach: %v", err)
	}

	second, err := svc.Attach(context.Background(), "https://github.com/owner/repo.git")
	if err != nil {
		t.Fatalf("re-Attach: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	secondSpec := fake.lastRun()

	// Identity comes from the link, not the row id, so a re-attached repo keeps
	// its clone, worktrees and pipeline database instead of starting over.
	if second.ContainerName != first.ContainerName {
		t.Errorf("container name changed across re-attach: %q -> %q", first.ContainerName, second.ContainerName)
	}
	for i := range firstSpec.Volumes {
		if secondSpec.Volumes[i] != firstSpec.Volumes[i] {
			t.Errorf("volume %d changed across re-attach: %+v -> %+v", i, firstSpec.Volumes[i], secondSpec.Volumes[i])
		}
	}
}

func TestConcurrentAttachesGetDistinctPorts(t *testing.T) {
	fake := newFakeDocker()
	fake.images["sloper-agent:test"] = true
	svc := newTestService(t, fake)

	const n = 4
	ports := make([]int, n)
	errs := make([]error, n)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			repo, err := svc.Attach(context.Background(), fmt.Sprintf("owner/repo-%d", i))
			if err != nil {
				errs[i] = err
				return
			}
			ports[i] = repo.HostPort
		}(i)
	}
	wg.Wait()

	seen := map[int]bool{}
	for i, err := range errs {
		if err != nil {
			t.Fatalf("attach %d: %v", i, err)
		}
		if seen[ports[i]] {
			t.Fatalf("port %d was handed out twice: %v", ports[i], ports)
		}
		seen[ports[i]] = true
	}
}

func TestPurgeFailureStillForgetsTheRepo(t *testing.T) {
	fake := newFakeDocker()
	fake.images["sloper-agent:test"] = true
	svc := newTestService(t, fake)

	repo, err := svc.Attach(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	fake.removeVolumeErr = errors.New("volume is in use")
	err = svc.Detach(context.Background(), repo.ID, true)
	if err == nil {
		t.Fatal("Detach with a failing volume purge should report the failure")
	}

	// A row that outlived a half-finished purge would be recreated by the next
	// reconcile pass, so the repo must be gone even when the purge fails.
	if _, err := svc.Get(context.Background(), repo.ID); !errors.Is(err, storage.ErrRepoNotFound) {
		t.Errorf("Get after failed purge = %v, want ErrRepoNotFound", err)
	}
	if fake.hasContainer(repo.ContainerName) {
		t.Error("container survived detach")
	}
}

func TestReconcileUsesTheImageTheRowWasAttachedWith(t *testing.T) {
	fake := newFakeDocker()
	fake.images["sloper-agent:test"] = true
	svc := newTestService(t, fake)

	repo, err := svc.Attach(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}

	// A row keeps the image it was attached with, whatever SLOPER_IMAGE says
	// later: that is the image the controller must check and pull.
	svc.opts.Image = "sloper-agent:v2"
	fake.images["sloper-agent:v2"] = true

	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if len(fake.runs) != 1 {
		t.Fatalf("containers created = %d, want 1", len(fake.runs))
	}
	if got := fake.lastRun().Image; got != "sloper-agent:test" {
		t.Errorf("image = %q, want the row's image", got)
	}

	stored, _ := svc.Get(context.Background(), repo.ID)
	if stored.Status != models.RepoStatusStarting {
		t.Errorf("status = %q, want starting", stored.Status)
	}
}

func TestReconcileRecordsDockerFailure(t *testing.T) {
	fake := newFakeDocker()
	fake.images["sloper-agent:test"] = true
	fake.runErr = errors.New("port is already allocated")
	svc := newTestService(t, fake)

	repo, err := svc.Attach(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	stored, _ := svc.Get(context.Background(), repo.ID)
	if stored.Status != models.RepoStatusError {
		t.Fatalf("status = %q, want error", stored.Status)
	}
	if !strings.Contains(stored.StatusMessage, "port is already allocated") {
		t.Errorf("status message = %q", stored.StatusMessage)
	}
	if stored.RestartCount != 1 {
		t.Errorf("restart count = %d, want 1", stored.RestartCount)
	}
}

func TestDesiredStateStoppedStopsTheContainer(t *testing.T) {
	fake := newFakeDocker()
	fake.images["sloper-agent:test"] = true
	svc := newTestService(t, fake)

	repo, err := svc.Attach(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	fake.setRunning(repo.ContainerName)

	if _, err := svc.SetDesiredState(context.Background(), repo.ID, models.RepoDesiredStopped); err != nil {
		t.Fatalf("SetDesiredState: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(fake.stops) != 1 {
		t.Errorf("stop calls = %d, want 1", len(fake.stops))
	}
	stored, _ := svc.Get(context.Background(), repo.ID)
	if stored.Status != models.RepoStatusStopped {
		t.Errorf("status = %q, want stopped", stored.Status)
	}
	if !fake.hasContainer(repo.ContainerName) {
		t.Error("a stopped repo keeps its container so it can be started again")
	}

	if _, err := svc.SetDesiredState(context.Background(), repo.ID, "paused"); !errors.Is(err, ErrBadDesiredState) {
		t.Errorf("SetDesiredState error = %v, want ErrBadDesiredState", err)
	}
}

func TestDetachRemovesContainerAndPurgesVolumes(t *testing.T) {
	fake := newFakeDocker()
	fake.images["sloper-agent:test"] = true
	svc := newTestService(t, fake)

	repo, err := svc.Attach(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if err := svc.Detach(context.Background(), repo.ID, false); err != nil {
		t.Fatalf("Detach: %v", err)
	}
	if fake.hasContainer(repo.ContainerName) {
		t.Error("container still present after detach")
	}
	if len(fake.removedVolumes) != 0 {
		t.Errorf("volumes removed without purge: %v", fake.removedVolumes)
	}
	if _, err := svc.Get(context.Background(), repo.ID); !errors.Is(err, storage.ErrRepoNotFound) {
		t.Errorf("Get after detach = %v, want ErrRepoNotFound", err)
	}

	// Attaching again reuses the freed port; purging removes the volumes from
	// the earlier attach.
	again, err := svc.Attach(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("re-Attach: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if err := svc.Detach(context.Background(), again.ID, true); err != nil {
		t.Fatalf("Detach with purge: %v", err)
	}
	if len(fake.removedVolumes) != 2 {
		t.Errorf("volumes removed = %v, want the data and repo volumes", fake.removedVolumes)
	}
}

func TestRestartReplacesTheContainerAndKeepsVolumes(t *testing.T) {
	fake := newFakeDocker()
	fake.images["sloper-agent:test"] = true
	svc := newTestService(t, fake)

	repo, err := svc.Attach(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	firstRun := fake.lastRun()

	if _, err := svc.Restart(context.Background(), repo.ID); err != nil {
		t.Fatalf("Restart: %v", err)
	}
	if err := svc.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if len(fake.runs) != 2 {
		t.Fatalf("containers created = %d, want 2", len(fake.runs))
	}
	secondRun := fake.lastRun()
	if secondRun.Volumes[0] != firstRun.Volumes[0] || secondRun.Volumes[1] != firstRun.Volumes[1] {
		t.Errorf("restart changed the volumes: %+v -> %+v", firstRun.Volumes, secondRun.Volumes)
	}
}

func TestAttachFailsWhenDockerIsUnavailable(t *testing.T) {
	fake := newFakeDocker()
	fake.available = false
	svc := newTestService(t, fake)

	if _, err := svc.Attach(context.Background(), "owner/repo"); !errors.Is(err, docker.ErrUnavailable) {
		t.Fatalf("Attach error = %v, want ErrUnavailable", err)
	}
}
