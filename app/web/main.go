package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/storage"
	"github.com/ButterHost69/sloper/internal/version"
)

// webServer is the dashboard API server for a single sloper instance.
// It exposes read-only JSON endpoints over the local sloper sqlite database.
// The Next.js dashboard connects to one or more of these servers.
type webServer struct {
	db     *storage.Repositories
	start  time.Time
	repo   string
	gitURL string
	env    map[string]string
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Println("sloper-web", version.Version)
			os.Exit(0)
		}
	}

	fmt.Printf("sloper-web %s starting\n", version.Version)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	port := envOr("SLOPER_WEB_PORT", "8080")
	addr := envOr("SLOPER_WEB_ADDR", "0.0.0.0") + ":" + port

	dbPath := os.Getenv("SLOPER_DB_PATH")
	db, err := storage.OpenDB(dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "web: open database failed:", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := storage.Migrate(ctx, db); err != nil {
		fmt.Fprintln(os.Stderr, "web: database migration failed:", err)
		os.Exit(1)
	}

	s := &webServer{
		db:    storage.NewRepositories(db),
		start: time.Now(),
		repo:  envOr("SLOPER_REPO", detectRepoFromGit()),
		env:   safeEnv(),
	}
	s.gitURL = detectRepoURL()

	mux := http.NewServeMux()
	s.routes(mux)

	fmt.Printf("sloper dashboard API listening on http://%s (repo: %s)\n", addr, s.repo)

	srv := &http.Server{
		Addr:              addr,
		Handler:           cors(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintln(os.Stderr, "web: server error:", err)
		os.Exit(1)
	}
}

func (s *webServer) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/summary", s.handleSummary)
	mux.HandleFunc("GET /api/repo", s.handleRepo)
	mux.HandleFunc("GET /api/issues", s.handleIssues)
	mux.HandleFunc("GET /api/issues/{number}", s.handleIssueDetail)
	mux.HandleFunc("GET /api/issues/{number}/comments", s.handleIssueComments)
	mux.HandleFunc("GET /api/issues/{number}/runs", s.handleIssueRuns)
	mux.HandleFunc("GET /api/issues/{number}/events", s.handleIssueEvents)
	mux.HandleFunc("GET /api/pulls", s.handlePulls)
	mux.HandleFunc("GET /api/runs", s.handleRuns)
	mux.HandleFunc("GET /api/events", s.handleEvents)
	mux.HandleFunc("GET /api/metrics/activity", s.handleActivity)
}

// ─── Handlers ────────────────────────────────────────────────────────

func (s *webServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"repo":     s.repo,
		"version":  "dev",
		"go":       runtime.Version(),
		"uptime_s": int64(time.Since(s.start).Seconds()),
		"time":     time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *webServer) handleRepo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":    s.repo,
		"url":     s.gitURL,
		"db_size": dbSize(r.Context(), s.db),
		"agent":   s.env,
	})
}

func (s *webServer) handleSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stageCounts, err := s.db.ListIssuesByStage(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	stateCounts, err := s.db.ListIssuesByState(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	runStatuses, err := s.db.CountRunsByStatus(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	eventTypes, err := s.db.CountEventsByType(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	pulls, err := s.db.ListPRs(ctx, 5000)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	totalIssues := 0
	for _, c := range stageCounts {
		totalIssues += c
	}
	totalRuns := 0
	for _, c := range runStatuses {
		totalRuns += c
	}
	openPulls := 0
	for _, p := range pulls {
		if p.State == "open" {
			openPulls++
		}
	}

	eventTotal, err := s.db.CountEvents(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"repo": s.repo,
		"issues": map[string]any{
			"total":    totalIssues,
			"open":     stateCounts["open"],
			"closed":   stateCounts["closed"],
			"by_stage": stageCounts,
			"failed":   stageCounts[models.StageFailed],
			"merged":   stageCounts[models.StageMerged],
			"in_flight": stageCounts[models.StageApproved] + stageCounts[models.StageWorkDone] +
				stageCounts[models.StageReviewDone] + stageCounts[models.StageSpecOngoing],
		},
		"runs": map[string]any{
			"total":       totalRuns,
			"by_status":   runStatuses,
			"running":     runStatuses["running"],
			"failed":      runStatuses["failed"],
			"completed":   runStatuses["completed"],
			"interrupted": runStatuses["interrupted"],
		},
		"pulls": map[string]any{
			"total": len(pulls),
			"open":  openPulls,
		},
		"events": map[string]any{
			"total":   eventTotal,
			"by_type": eventTypes,
		},
	})
}

func (s *webServer) handleIssues(w http.ResponseWriter, r *http.Request) {
	limit, offset := pagination(r)
	issues, err := s.db.ListIssues(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]map[string]any, 0, len(issues))
	for _, it := range issues {
		out = append(out, issueJSON(it))
	}
	writeJSON(w, http.StatusOK, map[string]any{"issues": out, "count": len(out)})
}

func (s *webServer) handleIssueDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	num := pathInt(w, r, "number")
	if num == 0 {
		return
	}

	issue, err := s.db.GetIssue(ctx, num)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if issue == nil {
		writeError(w, http.StatusNotFound, fmt.Errorf("issue %d not found", num))
		return
	}

	comments, _ := s.db.ListCommentsByIssue(ctx, num)
	runs, _ := s.db.ListRunsByIssue(ctx, num)
	events, _ := s.db.ListEventsByIssue(ctx, num)

	var pr *map[string]any
	if issue.PRNumber != 0 {
		if p, err := s.db.GetPR(ctx, issue.PRNumber); err == nil && p != nil {
			m := prJSON(*p)
			pr = &m
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"issue":    issueJSON(*issue),
		"spec":     parsedSpec(issue.SpecJSON),
		"comments": comments,
		"runs":     runs,
		"events":   events,
		"pr":       pr,
	})
}

func (s *webServer) handleIssueComments(w http.ResponseWriter, r *http.Request) {
	num := pathInt(w, r, "number")
	if num == 0 {
		return
	}
	comments, err := s.db.ListCommentsByIssue(r.Context(), num)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"comments": comments})
}

func (s *webServer) handleIssueRuns(w http.ResponseWriter, r *http.Request) {
	num := pathInt(w, r, "number")
	if num == 0 {
		return
	}
	runs, err := s.db.ListRunsByIssue(r.Context(), num)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": runs})
}

func (s *webServer) handleIssueEvents(w http.ResponseWriter, r *http.Request) {
	num := pathInt(w, r, "number")
	if num == 0 {
		return
	}
	events, err := s.db.ListEventsByIssue(r.Context(), num)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events})
}

func (s *webServer) handlePulls(w http.ResponseWriter, r *http.Request) {
	limit, _ := pagination(r)
	pulls, err := s.db.ListPRs(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]map[string]any, 0, len(pulls))
	for _, p := range pulls {
		out = append(out, prJSON(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{"pulls": out, "count": len(out)})
}

func (s *webServer) handleRuns(w http.ResponseWriter, r *http.Request) {
	limit, offset := pagination(r)
	runs, err := s.db.ListRuns(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": runs, "count": len(runs)})
}

func (s *webServer) handleEvents(w http.ResponseWriter, r *http.Request) {
	limit, offset := pagination(r)
	events, err := s.db.ListEvents(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events, "count": len(events)})
}

func (s *webServer) handleActivity(w http.ResponseWriter, r *http.Request) {
	hours := 24
	if v := r.URL.Query().Get("hours"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 24*7 {
			hours = n
		}
	}
	buckets, err := s.db.ActivityByHour(r.Context(), hours)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"hours": hours, "buckets": buckets})
}

// ─── DTO helpers ─────────────────────────────────────────────────────

func issueJSON(it storage.IssueRecord) map[string]any {
	return map[string]any{
		"number":            it.Number,
		"title":             it.Title,
		"state":             it.State,
		"url":               it.URL,
		"author":            it.Author,
		"updated_at":        it.UpdatedAt,
		"labels":            it.Labels,
		"is_pull_request":   it.IsPullRequest,
		"stage":             it.Stage,
		"branch_name":       it.BranchName,
		"pr_number":         it.PRNumber,
		"last_comment_id":   it.LastCommentID,
		"review_iterations": it.ReviewIterations,
		"created_at":        it.CreatedAt,
		"first_seen_at":     it.FirstSeenAt,
		"updated_at_local":  it.UpdatedAtLocal,
	}
}

func prJSON(p storage.PRRecord) map[string]any {
	return map[string]any{
		"number":         p.Number,
		"issue_number":   p.IssueNumber,
		"title":          p.Title,
		"head_sha":       p.HeadSHA,
		"base_sha":       p.BaseSHA,
		"state":          p.State,
		"url":            p.URL,
		"updated_at":     p.UpdatedAt,
		"review_state":   p.ReviewState,
		"last_review_at": p.LastReviewAt,
	}
}

func parsedSpec(specJSON string) *models.SpecResult {
	return storage.ParseSpecJSON(specJSON)
}

// ─── HTTP helpers ────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

func pagination(r *http.Request) (limit, offset int) {
	limit = 100
	offset = 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 5000 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return
}

func pathInt(w http.ResponseWriter, r *http.Request, key string) int64 {
	v, err := strconv.ParseInt(r.PathValue(key), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid %s: %q", key, r.PathValue(key)))
		return 0
	}
	return v
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Instance-Id")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// safeEnv returns sanitized agent/repo config, never exposing API keys.
func safeEnv() map[string]string {
	return map[string]string{
		"model":    os.Getenv("AGENT_MODEL"),
		"provider": os.Getenv("AGENT_PROVIDER"),
		"bot_user": os.Getenv("GH_USERNAME"),
		"repo_env": os.Getenv("GH_REPO_LINK"),
	}
}

func dbSize(ctx context.Context, db *storage.Repositories) int64 {
	size, err := db.DBSize(ctx)
	if err != nil {
		return 0
	}
	return size
}

func detectRepoFromGit() string {
	if out, err := runCmd("git", "config", "--get", "remote.origin.url"); err == nil {
		out = strings.TrimSpace(out)
		out = strings.TrimSuffix(out, ".git")
		if i := strings.LastIndex(out, "/"); i >= 0 {
			return out[i+1:]
		}
		if out != "" {
			return out
		}
	}
	return "unknown-repo"
}

func detectRepoURL() string {
	out, err := runCmd("git", "config", "--get", "remote.origin.url")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}
