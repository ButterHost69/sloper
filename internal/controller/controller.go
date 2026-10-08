// Package controller turns "attach this repo" into a running sloper instance.
//
// It owns the repos table, one docker container per attached repo, the host
// port that container publishes its dashboard API on, and a reconcile loop that
// keeps reality matching the table. The HTTP layer in app/controller is a thin
// shell over this package.
package controller

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ButterHost69/sloper/internal/docker"
	"github.com/ButterHost69/sloper/internal/logger"
	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/storage"
	"go.uber.org/zap"
)

const (
	// DefaultImage is the sloper-compatible image every repo container runs.
	// It is the setup/Dockerfile image, built with `make build-docker`.
	DefaultImage = "sloper-agent:latest"

	DefaultContainerPrefix = "sloper-repo"
	DefaultVolumePrefix    = "sloper"

	DefaultPortMin = 8080
	DefaultPortMax = 8180

	DefaultReconcileInterval = 10 * time.Second
	DefaultImagePullTimeout  = 10 * time.Minute
	DefaultStopTimeoutSec    = 10

	// DefaultImagePullRetryCooldown keeps a missing image from being re-pulled
	// on every reconcile tick. A typo in SLOPER_IMAGE would otherwise hammer the
	// registry forever.
	DefaultImagePullRetryCooldown = 2 * time.Minute
)

var (
	// ErrNotConfigured is returned when the controller cannot start a container
	// because the process itself is missing required configuration.
	ErrNotConfigured = errors.New("controller: missing required configuration")

	// ErrAlreadyAttached is returned when the same owner/repo is attached twice.
	ErrAlreadyAttached = errors.New("controller: repo already attached")

	// ErrBadDesiredState is returned for a desired state other than running/stopped.
	ErrBadDesiredState = errors.New("controller: desired state must be running or stopped")
)

// Docker is the container runtime the controller drives. *docker.Client
// implements it; tests substitute a fake.
type Docker interface {
	Available(ctx context.Context) error
	ImageExists(ctx context.Context, image string) (bool, error)
	Pull(ctx context.Context, image string, timeout time.Duration) error
	Run(ctx context.Context, spec models.ContainerSpec) (string, error)
	Start(ctx context.Context, name string) error
	Stop(ctx context.Context, name string, timeoutSec int) error
	Remove(ctx context.Context, name string, force bool) error
	RemoveVolume(ctx context.Context, name string) error
	Inspect(ctx context.Context, name string) (models.ContainerState, error)
	List(ctx context.Context, labels map[string]string) ([]models.ContainerState, error)
	Logs(ctx context.Context, name string, tail int) (string, error)
}

// Options configures a Service. Everything has a working default except DB and
// Docker.
type Options struct {
	DB     *storage.Repositories
	Docker Docker
	Log    *zap.Logger

	Image           string
	ContainerPrefix string
	VolumePrefix    string

	PortMin       int
	PortMax       int
	ContainerPort int

	// PublishAddr is the host address the instance API port is bound to
	// ("127.0.0.1" keeps repo instances off the network). PublicHost is the
	// hostname the console should use to reach that port.
	PublishAddr string
	PublicHost  string

	Network       string
	RestartPolicy string

	// InheritEnv is passed to every repo container: the GitHub token, the agent
	// model and key, and the commit identity. Keys are explicit so controller
	// only settings never leak into agent containers.
	InheritEnv map[string]string

	// RequiredEnv are the InheritEnv keys without which an attach is refused,
	// because the container could not clone the repo or run the agent.
	RequiredEnv []string

	ReconcileInterval time.Duration
	ImagePullTimeout  time.Duration

	// ImagePullRetryCooldown is how long a failed pull is remembered before the
	// controller tries that image again.
	ImagePullRetryCooldown time.Duration
}

// Service attaches repos and keeps their containers alive.
type Service struct {
	opts Options
	log  *zap.Logger

	kickCh      chan struct{}
	reconcileMu sync.Mutex

	pullMu       sync.Mutex
	pullInFlight map[string]bool
	pullErr      map[string]error
	pullErrAt    map[string]time.Time
	pullReady    map[string]bool

	baseCtx context.Context
}

// New applies defaults and returns a ready service.
func New(opts Options) *Service {
	if opts.Image == "" {
		opts.Image = DefaultImage
	}
	if opts.ContainerPrefix == "" {
		opts.ContainerPrefix = DefaultContainerPrefix
	}
	if opts.VolumePrefix == "" {
		opts.VolumePrefix = DefaultVolumePrefix
	}
	if opts.PortMin <= 0 {
		opts.PortMin = DefaultPortMin
	}
	if opts.PortMax <= 0 {
		opts.PortMax = DefaultPortMax
	}
	if opts.PortMax < opts.PortMin {
		opts.PortMax = opts.PortMin
	}
	if opts.ContainerPort <= 0 {
		opts.ContainerPort = models.DefaultRepoContainerPort
	}
	if opts.PublishAddr == "" {
		opts.PublishAddr = "127.0.0.1"
	}
	if opts.PublicHost == "" {
		opts.PublicHost = "localhost"
	}
	if opts.RestartPolicy == "" {
		opts.RestartPolicy = "unless-stopped"
	}
	if opts.ReconcileInterval <= 0 {
		opts.ReconcileInterval = DefaultReconcileInterval
	}
	if opts.ImagePullTimeout <= 0 {
		opts.ImagePullTimeout = DefaultImagePullTimeout
	}
	if opts.ImagePullRetryCooldown <= 0 {
		opts.ImagePullRetryCooldown = DefaultImagePullRetryCooldown
	}
	if opts.RequiredEnv == nil {
		opts.RequiredEnv = []string{"GH_TOKEN", "AGENT_MODEL"}
	}
	log := opts.Log
	if log == nil {
		log = logger.Default()
	}
	if log == nil {
		log = zap.NewNop()
	}

	return &Service{
		opts:         opts,
		log:          log,
		kickCh:       make(chan struct{}, 1),
		pullInFlight: map[string]bool{},
		pullErr:      map[string]error{},
		pullErrAt:    map[string]time.Time{},
		pullReady:    map[string]bool{},
	}
}

// Run reconciles once, then keeps reconciling on a ticker and whenever Attach,
// Restart or SetDesiredState asks for it. It blocks until ctx is cancelled.
func (s *Service) Run(ctx context.Context) {
	// ensureImage hands background pulls this context; it reads it under
	// pullMu, so publishing it takes the same lock.
	s.pullMu.Lock()
	s.baseCtx = ctx
	s.pullMu.Unlock()

	if err := s.Reconcile(ctx); err != nil {
		s.log.Warn("controller: initial reconcile failed", zap.Error(err))
	}
	s.warnOrphans(ctx)

	ticker := time.NewTicker(s.opts.ReconcileInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.log.Info("controller: stopping")
			return
		case <-ticker.C:
			if err := s.Reconcile(ctx); err != nil {
				s.log.Warn("controller: reconcile failed", zap.Error(err))
			}
		case <-s.kickCh:
			if err := s.Reconcile(ctx); err != nil {
				s.log.Warn("controller: reconcile failed", zap.Error(err))
			}
		}
	}
}

func (s *Service) kick() {
	select {
	case s.kickCh <- struct{}{}:
	default: // a reconcile is already queued
	}
}

// ─── Public operations ───────────────────────────────────────────────

// Attach validates a repo link, reserves its host port and records it. The
// container itself is created by the reconciler, so an attach never blocks on
// an image pull.
func (s *Service) Attach(ctx context.Context, rawLink string) (models.AttachedRepo, error) {
	var out models.AttachedRepo

	if missing := s.missingEnv(); len(missing) > 0 {
		return out, fmt.Errorf("%w: %s must be set for the controller so repo containers can clone and run the agent",
			ErrNotConfigured, strings.Join(missing, ", "))
	}
	if err := s.docker().Available(ctx); err != nil {
		return out, err
	}

	link, cloneURL, err := ParseRepoLink(rawLink)
	if err != nil {
		return out, err
	}

	// Everything from here to the insert mutates the repo set the reconciler
	// reads, so it runs under the same lock: otherwise a pass that snapshotted
	// the rows before this attach could write its stale view back, or two
	// attaches could be handed the same host port.
	s.reconcileMu.Lock()
	defer s.reconcileMu.Unlock()

	if _, err := s.opts.DB.GetRepoByLink(ctx, link); err == nil {
		return out, fmt.Errorf("%w: %s", ErrAlreadyAttached, link)
	} else if !errors.Is(err, storage.ErrRepoNotFound) {
		return out, err
	}

	port, err := s.allocatePort(ctx)
	if err != nil {
		return out, err
	}

	rec := storage.RepoRecord{
		Link:          link,
		Name:          RepoName(link),
		CloneURL:      cloneURL,
		Image:         s.opts.Image,
		HostPort:      port,
		APIToken:      s.apiToken(),
		DesiredState:  models.RepoDesiredRunning,
		Status:        models.RepoStatusPending,
		StatusMessage: "waiting for container",
	}
	rec.ContainerName = s.containerName(rec)

	id, err := s.opts.DB.InsertRepo(ctx, rec)
	if err != nil {
		if errors.Is(err, storage.ErrRepoLinkExists) {
			return out, fmt.Errorf("%w: %s", ErrAlreadyAttached, link)
		}
		return out, err
	}
	rec.ID = id

	s.log.Info("controller: repo attached",
		zap.String("repo", link),
		zap.String("container", rec.ContainerName),
		zap.Int("host_port", port))

	s.kick()

	stored, err := s.opts.DB.GetRepo(ctx, id)
	if err != nil {
		return out, err
	}
	return s.view(*stored), nil
}

// Detach stops and removes a repo's container and forgets the repo. Its data
// volumes survive unless purge is set, so re-attaching the same repo resumes
// where it left off.
func (s *Service) Detach(ctx context.Context, id int64, purge bool) error {
	s.reconcileMu.Lock()
	defer s.reconcileMu.Unlock()

	rec, err := s.opts.DB.GetRepo(ctx, id)
	if err != nil {
		return err
	}

	if rec.ContainerName != "" {
		if err := s.docker().Stop(ctx, rec.ContainerName, DefaultStopTimeoutSec); err != nil && !errors.Is(err, docker.ErrNotFound) {
			if errors.Is(err, docker.ErrUnavailable) {
				return err
			}
			s.log.Warn("controller: stop during detach failed",
				zap.String("container", rec.ContainerName), zap.Error(err))
		}
		if err := s.docker().Remove(ctx, rec.ContainerName, true); err != nil && !errors.Is(err, docker.ErrNotFound) {
			return err
		}
	}

	// Forget the repo first: once the container is gone, a row that outlives a
	// failed volume purge would be recreated by the next reconcile pass against
	// half-deleted data. A leaked volume is the better failure.
	if err := s.opts.DB.DeleteRepo(ctx, id); err != nil {
		return err
	}

	if purge {
		for _, volume := range s.volumeNames(*rec) {
			if err := s.docker().RemoveVolume(ctx, volume); err != nil && !errors.Is(err, docker.ErrNotFound) {
				return fmt.Errorf("controller: remove volume %s (repo already detached): %w", volume, err)
			}
		}
	}

	s.log.Info("controller: repo detached",
		zap.String("repo", rec.Link), zap.Bool("purged", purge))
	return nil
}

// Restart replaces a repo's container with a fresh one. The volumes are reused,
// so the clone, worktrees and pipeline database are kept.
func (s *Service) Restart(ctx context.Context, id int64) (models.AttachedRepo, error) {
	// Held across the container swap so a reconcile pass cannot race the
	// removal and start the old container back up.
	s.reconcileMu.Lock()
	defer s.reconcileMu.Unlock()

	rec, err := s.opts.DB.GetRepo(ctx, id)
	if err != nil {
		return models.AttachedRepo{}, err
	}

	if rec.ContainerName != "" {
		if err := s.docker().Stop(ctx, rec.ContainerName, DefaultStopTimeoutSec); err != nil && !errors.Is(err, docker.ErrNotFound) {
			s.log.Warn("controller: stop during restart failed",
				zap.String("container", rec.ContainerName), zap.Error(err))
		}
		// Force-removing also covers the "container exists but is wedged" case.
		if err := s.docker().Remove(ctx, rec.ContainerName, true); err != nil && !errors.Is(err, docker.ErrNotFound) {
			return models.AttachedRepo{}, err
		}
	}

	rec.DesiredState = models.RepoDesiredRunning
	rec.Status = models.RepoStatusPending
	rec.StatusMessage = "restarting"
	rec.ContainerID = ""
	rec.ContainerState = ""
	rec.RestartCount = 0
	if err := s.opts.DB.SetRepoDesiredState(ctx, id, models.RepoDesiredRunning); err != nil {
		return models.AttachedRepo{}, err
	}
	if err := s.opts.DB.UpdateRepoRuntime(ctx, *rec); err != nil {
		return models.AttachedRepo{}, err
	}

	// A restart is also the operator's "I just built the image" signal.
	s.retryImagePull(rec.Image)
	s.kick()

	stored, err := s.opts.DB.GetRepo(ctx, id)
	if err != nil {
		return models.AttachedRepo{}, err
	}
	return s.view(*stored), nil
}

// SetDesiredState records that the operator wants a repo instance running or
// stopped; the reconciler performs the transition.
func (s *Service) SetDesiredState(ctx context.Context, id int64, desired string) (models.AttachedRepo, error) {
	desired = strings.TrimSpace(strings.ToLower(desired))
	if desired != models.RepoDesiredRunning && desired != models.RepoDesiredStopped {
		return models.AttachedRepo{}, fmt.Errorf("%w: %q", ErrBadDesiredState, desired)
	}

	// The write is quick, but it must not land in the middle of a pass that
	// already snapshotted the old desired state.
	s.reconcileMu.Lock()
	defer s.reconcileMu.Unlock()

	if err := s.opts.DB.SetRepoDesiredState(ctx, id, desired); err != nil {
		return models.AttachedRepo{}, err
	}
	s.kick()

	stored, err := s.opts.DB.GetRepo(ctx, id)
	if err != nil {
		return models.AttachedRepo{}, err
	}
	return s.view(*stored), nil
}

// List returns every attached repo with its last observed container state.
func (s *Service) List(ctx context.Context) ([]models.AttachedRepo, error) {
	records, err := s.opts.DB.ListRepos(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]models.AttachedRepo, 0, len(records))
	for _, rec := range records {
		out = append(out, s.view(rec))
	}
	return out, nil
}

// Get returns one attached repo.
func (s *Service) Get(ctx context.Context, id int64) (models.AttachedRepo, error) {
	rec, err := s.opts.DB.GetRepo(ctx, id)
	if err != nil {
		return models.AttachedRepo{}, err
	}
	return s.view(*rec), nil
}

// Logs returns the tail of a repo container's log.
func (s *Service) Logs(ctx context.Context, id int64, tail int) (string, error) {
	rec, err := s.opts.DB.GetRepo(ctx, id)
	if err != nil {
		return "", err
	}
	if rec.ContainerName == "" {
		return "", fmt.Errorf("%w: repo %d has no container yet", docker.ErrNotFound, id)
	}
	return s.docker().Logs(ctx, rec.ContainerName, tail)
}

// Reconcile brings every repo container in line with its row: create what is
// missing, start what is stopped, stop what the operator stopped, and write the
// observed state back for the console.
func (s *Service) Reconcile(ctx context.Context) error {
	s.reconcileMu.Lock()
	defer s.reconcileMu.Unlock()

	records, err := s.opts.DB.ListRepos(ctx)
	if err != nil {
		return err
	}

	for _, rec := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		observed := s.reconcileOne(ctx, rec)
		if sameRuntimeState(rec, observed) {
			continue
		}
		if err := s.opts.DB.UpdateRepoRuntime(ctx, observed); err != nil {
			s.log.Warn("controller: persist reconcile state failed",
				zap.String("repo", observed.Link), zap.Error(err))
		}
	}
	return nil
}

func (s *Service) reconcileOne(ctx context.Context, rec storage.RepoRecord) storage.RepoRecord {
	if rec.DesiredState == models.RepoDesiredStopped {
		return s.reconcileStopped(ctx, rec)
	}
	return s.reconcileRunning(ctx, rec)
}

func (s *Service) reconcileRunning(ctx context.Context, rec storage.RepoRecord) storage.RepoRecord {
	state, err := s.docker().Inspect(ctx, rec.ContainerName)

	switch {
	case errors.Is(err, docker.ErrNotFound):
		return s.createContainer(ctx, rec)
	case err != nil:
		return fail(rec, "cannot inspect container: "+err.Error())
	case state.Running():
		rec.Status = models.RepoStatusRunning
		rec.StatusMessage = ""
		rec.ContainerState = state.State
		rec.ContainerID = shortID(state.ID)
		rec.RestartCount = 0
		if port := state.HostPortFor(s.opts.ContainerPort); port > 0 {
			rec.HostPort = port
		}
		return rec
	}

	// The container exists but is down. Try a plain start first; if that fails
	// the container is wedged, so replace it and let the volumes carry state.
	if err := s.docker().Start(ctx, rec.ContainerName); err != nil {
		s.log.Warn("controller: start failed, recreating container",
			zap.String("repo", rec.Link), zap.String("container", rec.ContainerName), zap.Error(err))
		_ = s.docker().Remove(ctx, rec.ContainerName, true)
		return s.createContainer(ctx, rec)
	}

	rec.Status = models.RepoStatusStarting
	rec.StatusMessage = ""
	// A container that keeps exiting (a bad token, a missing tool) would
	// otherwise look like a permanent "starting": count the restarts and say
	// why it went down, so the console shows the crash loop instead of hiding it.
	if state.Failed() {
		rec.RestartCount++
		rec.StatusMessage = fmt.Sprintf("container exited with code %d; restarting", state.ExitCode)
	}
	rec.ContainerState = state.State
	rec.ContainerID = shortID(state.ID)
	return rec
}

func (s *Service) reconcileStopped(ctx context.Context, rec storage.RepoRecord) storage.RepoRecord {
	state, err := s.docker().Inspect(ctx, rec.ContainerName)

	switch {
	case errors.Is(err, docker.ErrNotFound):
		rec.Status = models.RepoStatusStopped
		rec.StatusMessage = ""
		rec.ContainerState = ""
		rec.ContainerID = ""
		return rec
	case err != nil:
		return fail(rec, "cannot inspect container: "+err.Error())
	}

	// Stop anything docker does not already report as down. "restarting" (a
	// crash loop under unless-stopped) and "paused" are on their way up, and a
	// paused container cannot be stopped at all; removing it instead would
	// discard the container the operator only asked to pause.
	stopping := state.State != "exited" && state.State != "dead" && state.State != "created"
	if stopping {
		if err := s.docker().Stop(ctx, rec.ContainerName, DefaultStopTimeoutSec); err != nil {
			return fail(rec, "stop failed: "+err.Error())
		}
	}

	rec.Status = models.RepoStatusStopped
	rec.StatusMessage = ""
	if stopping {
		rec.ContainerState = "exited"
	} else {
		rec.ContainerState = state.State
	}
	rec.ContainerID = shortID(state.ID)
	return rec
}

func (s *Service) createContainer(ctx context.Context, rec storage.RepoRecord) storage.RepoRecord {
	ready, err := s.ensureImage(ctx, rec.Image)
	switch {
	case err != nil:
		return fail(rec, "image "+rec.Image+": "+err.Error())
	case !ready:
		rec.Status = models.RepoStatusPending
		rec.StatusMessage = "pulling image " + rec.Image
		return rec
	}

	spec := s.containerSpec(rec)
	id, err := s.docker().Run(ctx, spec)
	if err != nil {
		rec.RestartCount++
		return fail(rec, err.Error())
	}

	rec.Status = models.RepoStatusStarting
	rec.StatusMessage = ""
	rec.ContainerState = "created"
	rec.ContainerID = shortID(id)
	s.log.Info("controller: started repo container",
		zap.String("repo", rec.Link),
		zap.String("container", rec.ContainerName),
		zap.Int("host_port", rec.HostPort))
	return rec
}

func fail(rec storage.RepoRecord, message string) storage.RepoRecord {
	rec.Status = models.RepoStatusError
	rec.StatusMessage = truncate(message, 500)
	return rec
}

// ─── Container description ───────────────────────────────────────────

// containerName is the docker container for a repo.
func (s *Service) containerName(rec storage.RepoRecord) string {
	return fmt.Sprintf("%s-%s", s.opts.ContainerPrefix, s.repoSlug(rec.Link))
}

// repoSlug is the deterministic identity of a repo: its link plus a hash of
// that link. It deliberately does not use the row id, so detaching and
// re-attaching a repo lands on the same container name and the same volumes —
// the clone, worktrees and pipeline database survive a detach.
func (s *Service) repoSlug(link string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(link))))
	return fmt.Sprintf("%s-%s", models.Slugify(link), hex.EncodeToString(sum[:])[:8])
}

func (s *Service) volumeNames(rec storage.RepoRecord) []string {
	slug := s.repoSlug(rec.Link)
	return []string{
		fmt.Sprintf("%s-data-%s", s.opts.VolumePrefix, slug),
		fmt.Sprintf("%s-repo-%s", s.opts.VolumePrefix, slug),
	}
}

// containerSpec is the one place that decides what a repo instance is: the
// sloper image, the inherited credentials, a published dashboard API port, and
// two volumes that keep the clone and the pipeline database per repo.
func (s *Service) containerSpec(rec storage.RepoRecord) models.ContainerSpec {
	data, repo := s.volumeNames(rec)[0], s.volumeNames(rec)[1]

	env := make(map[string]string, len(s.opts.InheritEnv)+6)
	for k, v := range s.opts.InheritEnv {
		if strings.TrimSpace(v) != "" {
			env[k] = v
		}
	}
	env["GH_REPO_LINK"] = rec.Link
	env["SLOPER_REPO"] = rec.Name
	env["SLOPER_WEB_ADDR"] = "0.0.0.0"
	env["SLOPER_WEB_PORT"] = strconv.Itoa(s.opts.ContainerPort)
	env["SLOPER_DB_PATH"] = "/root/.sloper/sloper.sqlite"
	env["SLOPER_SESSION_DIR"] = "/root/.sloper/sessions"
	env["SLOPER_WORKTREE_DIR"] = "/root/.sloper/worktrees"
	if rec.APIToken != "" {
		env["SLOPER_WEB_TOKEN"] = rec.APIToken
	}

	return models.ContainerSpec{
		Name:  rec.ContainerName,
		Image: rec.Image,
		Env:   env,
		Labels: map[string]string{
			models.LabelManaged: "true",
			models.LabelRepo:    rec.Link,
			models.LabelPort:    strconv.Itoa(rec.HostPort),
		},
		Ports: []models.PortMapping{{
			HostPort:      rec.HostPort,
			ContainerPort: s.opts.ContainerPort,
			BindAddr:      s.opts.PublishAddr,
		}},
		Volumes: []models.VolumeMount{
			{Source: data, ContainerPath: "/root/.sloper"},
			{Source: repo, ContainerPath: "/root/repo"},
		},
		RestartPolicy: s.opts.RestartPolicy,
		Network:       s.opts.Network,
	}
}

// ─── Image handling ──────────────────────────────────────────────────

// ensureImage reports whether an image is ready locally, starting a background
// pull when it is not. A cold image must not stall the reconcile loop, so the
// caller gets (false, nil) and retries on the next tick; a failed pull is
// remembered for ImagePullRetryCooldown so the registry is not hammered on
// every tick.
//
// The image is passed in rather than read from the options: rows keep the image
// they were attached with, and changing SLOPER_IMAGE must not gate old rows on
// an image they do not run.
func (s *Service) ensureImage(ctx context.Context, image string) (bool, error) {
	if strings.TrimSpace(image) == "" {
		return false, errors.New("no image configured")
	}

	s.pullMu.Lock()
	if s.pullReady[image] {
		s.pullMu.Unlock()
		return true, nil
	}
	lastErr, lastErrAt := s.pullErr[image], s.pullErrAt[image]
	s.pullMu.Unlock()

	if lastErr != nil && time.Since(lastErrAt) < s.opts.ImagePullRetryCooldown {
		return false, lastErr
	}

	exists, err := s.docker().ImageExists(ctx, image)
	if err == nil && exists {
		s.markImageReady(image)
		return true, nil
	}
	if err != nil && !errors.Is(err, docker.ErrNotFound) {
		return false, err
	}

	s.pullMu.Lock()
	defer s.pullMu.Unlock()

	if s.pullInFlight[image] {
		return false, s.pullErr[image]
	}
	s.pullInFlight[image] = true
	s.log.Info("controller: pulling image", zap.String("image", image))

	base := s.baseCtx
	if base == nil {
		base = context.Background()
	}
	pullTimeout := s.opts.ImagePullTimeout
	go func() {
		pullCtx, cancel := context.WithTimeout(base, pullTimeout)
		defer cancel()

		err := s.docker().Pull(pullCtx, image, pullTimeout)

		s.pullMu.Lock()
		defer s.pullMu.Unlock()
		s.pullInFlight[image] = false
		s.pullErr[image] = err
		s.pullErrAt[image] = time.Now()
		if err == nil {
			s.pullReady[image] = true
			s.log.Info("controller: image ready", zap.String("image", image))
		} else {
			s.log.Error("controller: image pull failed",
				zap.String("image", image),
				zap.Duration("retry_in", s.opts.ImagePullRetryCooldown),
				zap.Error(err))
		}
	}()

	return false, s.pullErr[image]
}

// retryImagePull forgets what we know about an image so the next reconcile
// checks it from scratch. Restart uses it: after building, re-tagging or fixing
// the image, an operator retries without waiting out the cooldown.
func (s *Service) retryImagePull(image string) {
	s.pullMu.Lock()
	defer s.pullMu.Unlock()
	delete(s.pullErr, image)
	delete(s.pullErrAt, image)
	delete(s.pullReady, image)
}

func (s *Service) markImageReady(image string) {
	s.pullMu.Lock()
	defer s.pullMu.Unlock()
	s.pullReady[image] = true
	s.pullErr[image] = nil
	delete(s.pullErrAt, image)
}

// warnOrphans logs containers that carry our managed label but have no row.
// They are left alone: the operator can `docker rm` them, and deleting another
// deployment's containers would be worse than a stale container.
func (s *Service) warnOrphans(ctx context.Context) {
	containers, err := s.docker().List(ctx, map[string]string{models.LabelManaged: "true"})
	if err != nil {
		s.log.Debug("controller: orphan scan skipped", zap.Error(err))
		return
	}
	if len(containers) == 0 {
		return
	}

	known := map[string]bool{}
	if records, err := s.opts.DB.ListRepos(ctx); err == nil {
		for _, rec := range records {
			known[rec.ContainerName] = true
		}
	}
	for _, c := range containers {
		if !known[c.Name] {
			s.log.Warn("controller: found managed container with no repo row",
				zap.String("container", c.Name),
				zap.String("repo", c.Labels[models.LabelRepo]))
		}
	}
}

// ─── Ports and tokens ────────────────────────────────────────────────

// allocatePort reserves the first host port in the configured range that no
// repo row holds and nothing is listening on.
func (s *Service) allocatePort(ctx context.Context) (int, error) {
	// Every attached repo keeps the port it was given for as long as its row
	// exists, so a stopped or erroring repo can come back on the same URL.
	used := map[int]bool{}
	records, err := s.opts.DB.ListRepos(ctx)
	if err != nil {
		return 0, err
	}
	for _, rec := range records {
		if rec.HostPort > 0 {
			used[rec.HostPort] = true
		}
	}

	for port := s.opts.PortMin; port <= s.opts.PortMax; port++ {
		if used[port] || !portIsFree(s.opts.PublishAddr, port) {
			continue
		}
		return port, nil
	}
	return 0, fmt.Errorf("controller: no free port in range %d-%d", s.opts.PortMin, s.opts.PortMax)
}

// portIsFree probes the publish address so a port held by anything else on the
// host is never handed to a repo container.
func portIsFree(addr string, port int) bool {
	if addr == "0.0.0.0" || addr == "" {
		addr = ""
	}
	ln, err := net.Listen("tcp", net.JoinHostPort(addr, strconv.Itoa(port)))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// apiToken returns the bearer token the repo instance's dashboard API requires.
// An operator-provided SLOPER_WEB_TOKEN is reused verbatim, so the console can
// be pointed at a repo instance with the token it already knows.
func (s *Service) apiToken() string {
	if token := strings.TrimSpace(s.opts.InheritEnv["SLOPER_WEB_TOKEN"]); token != "" {
		return token
	}
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}

// ─── Helpers ─────────────────────────────────────────────────────────

func (s *Service) docker() Docker {
	if s.opts.Docker == nil {
		panic("controller: docker client is not configured")
	}
	return s.opts.Docker
}

func (s *Service) missingEnv() []string {
	var missing []string
	for _, key := range s.opts.RequiredEnv {
		if strings.TrimSpace(s.opts.InheritEnv[key]) == "" {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return missing
}

func (s *Service) view(rec storage.RepoRecord) models.AttachedRepo {
	apiURL := ""
	if rec.HostPort > 0 {
		apiURL = fmt.Sprintf("http://%s:%d", s.opts.PublicHost, rec.HostPort)
	}
	return models.AttachedRepo{
		ID:             rec.ID,
		Link:           rec.Link,
		Name:           rec.Name,
		CloneURL:       rec.CloneURL,
		ContainerName:  rec.ContainerName,
		ContainerID:    rec.ContainerID,
		Image:          rec.Image,
		HostPort:       rec.HostPort,
		APIURL:         apiURL,
		Token:          rec.APIToken,
		DesiredState:   rec.DesiredState,
		Status:         rec.Status,
		StatusMessage:  rec.StatusMessage,
		ContainerState: rec.ContainerState,
		RestartCount:   rec.RestartCount,
		AttachedAt:     rec.AttachedAt,
		UpdatedAt:      rec.UpdatedAt,
	}
}

// sameRuntimeState reports whether a reconcile pass changed anything worth
// writing, so a healthy controller does not rewrite every row every tick.
func sameRuntimeState(before, after storage.RepoRecord) bool {
	return before.Status == after.Status &&
		before.StatusMessage == after.StatusMessage &&
		before.ContainerID == after.ContainerID &&
		before.ContainerState == after.ContainerState &&
		before.HostPort == after.HostPort &&
		before.ContainerName == after.ContainerName &&
		before.RestartCount == after.RestartCount
}

func shortID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "…"
}
