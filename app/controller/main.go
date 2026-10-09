// Command controller is the sloper backend that owns attached repos.
//
// It is the only process that talks to the docker daemon: the console (or curl)
// posts a repo link to it, and it records the repo, starts one container for it,
// publishes that container's dashboard API on a host port, and keeps it alive.
// Each repo container runs the same sloper + sloper-web pair as any other sloper
// deployment, so everything downstream (issues, pipeline, sessions) is
// unchanged.
package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ButterHost69/sloper/internal/controller"
	"github.com/ButterHost69/sloper/internal/docker"
	"github.com/ButterHost69/sloper/internal/logger"
	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/storage"
	"github.com/ButterHost69/sloper/internal/version"
	"go.uber.org/zap"
)

// inheritedEnvKeys are the variables a controller process passes down to every
// repo container. The allowlist is explicit so controller-only settings (ports,
// image, db path) never leak into an agent container.
var inheritedEnvKeys = []string{
	"GH_TOKEN",
	"GH_USERNAME",
	"GH_EMAIL",
	"AGENT_MODEL",
	"AGENT_KEY",
	"AGENT_PROVIDER",
	"AGENT_API",
	"SLOPER_WEB_TOKEN",
}

type controllerServer struct {
	svc    *controller.Service
	docker *docker.Client
	start  time.Time
	token  string
	image  string
	log    *zap.Logger
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Println("sloper-controller", version.Version)
			os.Exit(0)
		}
	}

	fmt.Printf("sloper-controller %s starting\n", version.Version)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logDir := os.Getenv("SLOPER_CONTROLLER_LOG_DIR")
	if logDir == "" {
		if home, err := os.UserHomeDir(); err == nil {
			logDir = filepath.Join(home, ".sloper", "controller-logs")
		}
	}
	if err := logger.Init(logDir); err != nil {
		fmt.Fprintln(os.Stderr, "controller: file logger unavailable:", err)
	}
	log := logger.Default()

	dbPath := os.Getenv("SLOPER_CONTROLLER_DB_PATH")
	if dbPath == "" {
		if home, err := os.UserHomeDir(); err == nil {
			dbPath = filepath.Join(home, ".sloper", "controller.sqlite")
		}
	}

	db, err := storage.OpenDB(dbPath)
	if err != nil {
		fatal(log, "open database failed", err)
	}
	defer db.Close()

	if err := storage.Migrate(ctx, db); err != nil {
		fatal(log, "database migration failed", err)
	}

	portMin, portMax := parsePortRange(envOr("SLOPER_PORT_RANGE", "8080-8180"))
	image := envOr("SLOPER_IMAGE", controller.DefaultImage)
	bindAddr := envOr("SLOPER_CONTROLLER_ADDR", "127.0.0.1")
	token := os.Getenv("SLOPER_CONTROLLER_TOKEN")
	dockerClient := docker.New(models.DockerOptions{})

	// Attaching a repo starts a container from an image of the operator's
	// choosing, so an unauthenticated controller on a reachable address is a
	// remote container-spawn API. Refuse that combination instead of warning.
	if token == "" && !isLoopbackAddr(bindAddr) {
		fatal(log, "refusing to serve an unauthenticated controller on a non-loopback address",
			fmt.Errorf("SLOPER_CONTROLLER_ADDR=%s needs SLOPER_CONTROLLER_TOKEN (or bind 127.0.0.1)", bindAddr))
	}

	svc := controller.New(controller.Options{
		DB:                storage.NewRepositories(db),
		Docker:            dockerClient,
		Log:               log,
		Image:             image,
		PortMin:           portMin,
		PortMax:           portMax,
		PublishAddr:       envOr("SLOPER_PUBLISH_ADDR", "127.0.0.1"),
		PublicHost:        envOr("SLOPER_PUBLIC_HOST", "localhost"),
		Network:           os.Getenv("SLOPER_NETWORK"),
		RestartPolicy:     envOr("SLOPER_RESTART_POLICY", "unless-stopped"),
		InheritEnv:        inheritedEnv(),
		ReconcileInterval: secondsEnv("SLOPER_CONTROLLER_RECONCILE_SECONDS", controller.DefaultReconcileInterval),
	})

	log.Info("controller: configured",
		zap.String("db", dbPath),
		zap.String("image", image),
		zap.Int("port_min", portMin),
		zap.Int("port_max", portMax),
		zap.String("publish_addr", envOr("SLOPER_PUBLISH_ADDR", "127.0.0.1")),
		zap.String("public_host", envOr("SLOPER_PUBLIC_HOST", "localhost")))

	// The repo container's entrypoint runs `git config --global user.name
	// "$GH_USERNAME"`, so a missing identity only shows up later, as agent
	// commits that git refuses to author.
	for _, key := range []string{"GH_USERNAME", "GH_EMAIL"} {
		if os.Getenv(key) == "" {
			log.Warn("controller: "+key+" is not set; commits made inside repo containers may fail",
				zap.String("variable", key))
		}
	}

	go svc.Run(ctx)

	addr := bindAddr + ":" + envOr("SLOPER_CONTROLLER_PORT", "9090")

	s := &controllerServer{
		svc:    svc,
		docker: dockerClient,
		start:  time.Now(),
		token:  token,
		image:  image,
		log:    log,
	}

	mux := http.NewServeMux()
	s.routes(mux)

	fmt.Printf("sloper-controller listening on http://%s\n", addr)
	if s.token != "" {
		fmt.Println("controller: bearer token auth enabled (SLOPER_CONTROLLER_TOKEN)")
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           cors(s.auth(mux)),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fatal(log, "server error", err)
	}

	logger.Sync()
}

func (s *controllerServer) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/repos", s.handleListRepos)
	mux.HandleFunc("POST /api/repos", s.handleAttachRepo)
	mux.HandleFunc("GET /api/repos/{id}", s.handleGetRepo)
	mux.HandleFunc("PATCH /api/repos/{id}", s.handlePatchRepo)
	mux.HandleFunc("DELETE /api/repos/{id}", s.handleDetachRepo)
	mux.HandleFunc("POST /api/repos/{id}/restart", s.handleRestartRepo)
	mux.HandleFunc("GET /api/repos/{id}/logs", s.handleRepoLogs)
}

// ─── Handlers ────────────────────────────────────────────────────────

func (s *controllerServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	dockerStatus := "ok"
	dockerErr := ""
	if err := s.docker.Available(r.Context()); err != nil {
		dockerStatus = "unavailable"
		dockerErr = err.Error()
	}

	repos, err := s.svc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	byRepoStatus := map[string]int{}
	for _, repo := range repos {
		byRepoStatus[repo.Status]++
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":         "ok",
		"version":        version.Version,
		"uptime_s":       int(time.Since(s.start).Seconds()),
		"time":           time.Now().UTC().Format(time.RFC3339),
		"docker":         dockerStatus,
		"docker_error":   dockerErr,
		"image":          s.image,
		"repos":          len(repos),
		"repos_by_state": byRepoStatus,
	})
}

func (s *controllerServer) handleListRepos(w http.ResponseWriter, r *http.Request) {
	repos, err := s.svc.List(r.Context())
	if err != nil {
		writeError(w, statusForError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"repos": repos, "count": len(repos)})
}

func (s *controllerServer) handleGetRepo(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	repo, err := s.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, statusForError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, repo)
}

type attachRequest struct {
	Link string `json:"link"`
}

func (s *controllerServer) handleAttachRepo(w http.ResponseWriter, r *http.Request) {
	var req attachRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	repo, err := s.svc.Attach(r.Context(), req.Link)
	if err != nil {
		writeError(w, statusForError(err), err)
		return
	}
	writeJSON(w, http.StatusCreated, repo)
}

type patchRequest struct {
	DesiredState string `json:"desired_state"`
}

func (s *controllerServer) handlePatchRepo(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req patchRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	repo, err := s.svc.SetDesiredState(r.Context(), id, req.DesiredState)
	if err != nil {
		writeError(w, statusForError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, repo)
}

func (s *controllerServer) handleDetachRepo(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	purge := r.URL.Query().Get("purge") == "true"

	if err := s.svc.Detach(r.Context(), id, purge); err != nil {
		writeError(w, statusForError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "detached", "id": id, "purged": purge})
}

func (s *controllerServer) handleRestartRepo(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	repo, err := s.svc.Restart(r.Context(), id)
	if err != nil {
		writeError(w, statusForError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, repo)
}

func (s *controllerServer) handleRepoLogs(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	tail := 200
	if v := r.URL.Query().Get("tail"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 5000 {
			tail = n
		}
	}
	logs, err := s.svc.Logs(r.Context(), id, tail)
	if err != nil {
		writeError(w, statusForError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "tail": tail, "logs": logs})
}

// ─── HTTP helpers ────────────────────────────────────────────────────

// statusForError maps controller/storage/docker errors onto HTTP statuses so
// the console can tell "you typed the link wrong" from "docker is down".
func statusForError(err error) int {
	switch {
	case errors.Is(err, controller.ErrInvalidLink),
		errors.Is(err, controller.ErrNotConfigured),
		errors.Is(err, controller.ErrBadDesiredState):
		return http.StatusBadRequest
	case errors.Is(err, controller.ErrAlreadyAttached):
		return http.StatusConflict
	case errors.Is(err, storage.ErrRepoNotFound), errors.Is(err, docker.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, docker.ErrUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	body := http.MaxBytesReader(w, r.Body, 64*1024)
	defer body.Close()

	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("request body is empty")
		}
		return fmt.Errorf("invalid request body: %w", err)
	}
	return nil
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid repo id %q", r.PathValue("id")))
		return 0, false
	}
	return id, true
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// auth guards every route with a bearer token when SLOPER_CONTROLLER_TOKEN is
// set. Attaching a repo starts a container, so the controller should not be
// exposed without one.
func (s *controllerServer) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.token != "" {
			want := "Bearer " + s.token
			got := r.Header.Get("Authorization")
			if subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
				w.Header().Set("WWW-Authenticate", "Bearer")
				writeError(w, http.StatusUnauthorized, errors.New("missing or invalid bearer token"))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// ─── Config helpers ──────────────────────────────────────────────────

func inheritedEnv() map[string]string {
	out := make(map[string]string, len(inheritedEnvKeys))
	for _, key := range inheritedEnvKeys {
		if v := os.Getenv(key); v != "" {
			out[key] = v
		}
	}
	return out
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// isLoopbackAddr reports whether an address only accepts connections from this
// machine: "127.0.0.1", "::1", "localhost", or a bare port-less host of the
// same kind.
func isLoopbackAddr(addr string) bool {
	host := strings.TrimSpace(addr)
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	if host == "" || strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func parsePortRange(value string) (int, int) {
	minStr, maxStr, ok := strings.Cut(strings.TrimSpace(value), "-")
	if !ok {
		return controller.DefaultPortMin, controller.DefaultPortMax
	}
	minPort, errMin := strconv.Atoi(strings.TrimSpace(minStr))
	maxPort, errMax := strconv.Atoi(strings.TrimSpace(maxStr))
	if errMin != nil || errMax != nil || minPort <= 0 || maxPort < minPort || maxPort > 65535 {
		return controller.DefaultPortMin, controller.DefaultPortMax
	}
	return minPort, maxPort
}

func secondsEnv(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(v)
	if err != nil || seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

func fatal(log *zap.Logger, message string, err error) {
	log.Error("controller: "+message, zap.Error(err))
	fmt.Fprintln(os.Stderr, "controller:", message+":", err)
	os.Exit(1)
}
