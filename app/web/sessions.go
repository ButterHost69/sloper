package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/ButterHost69/sloper/internal/sessions"
	"github.com/ButterHost69/sloper/internal/storage"
)

// sessionLiveWindow is how recently a session file must have changed for the
// console to treat it as active. A long tool call can hold the file still for
// minutes, so a running pipeline run keeps a session live past this window.
const sessionLiveWindow = 120 * time.Second

// sessionLocator bounds how often the session directory is rescanned while the
// console polls a transcript.
type sessionLocator struct {
	mu     sync.Mutex
	at     time.Time
	byID   map[string]sessions.File
	loaded bool
}

const sessionLocateTTL = 2 * time.Second

func (l *sessionLocator) find(dir, id string) (sessions.File, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.loaded || time.Since(l.at) > sessionLocateTTL {
		files, err := sessions.Scan(dir)
		if err != nil {
			return sessions.File{}, false
		}
		byID := make(map[string]sessions.File, len(files))
		for _, f := range files {
			byID[f.SessionID] = f
		}
		l.byID = byID
		l.at = time.Now()
		l.loaded = true
	}
	f, ok := l.byID[id]
	return f, ok
}

// ─── DTOs ────────────────────────────────────────────────────────────

type sessionRunView struct {
	ID         int64  `json:"id"`
	Stage      string `json:"stage"`
	Status     string `json:"status"`
	StartedAt  string `json:"started_at"`
	EndedAt    string `json:"ended_at"`
	Error      string `json:"error_message,omitempty"`
	DurationMs int64  `json:"duration_ms,omitempty"`
}

type sessionPRView struct {
	Number      int64  `json:"number"`
	State       string `json:"state"`
	ReviewState string `json:"review_state,omitempty"`
	URL         string `json:"url,omitempty"`
}

type sessionView struct {
	ID           string           `json:"id"`
	File         string           `json:"file"`
	IssueNumber  int64            `json:"issue_number"`
	Stage        string           `json:"stage,omitempty"`
	FixIteration int              `json:"fix_iteration,omitempty"`
	SizeBytes    int64            `json:"size_bytes"`
	ModifiedAt   string           `json:"modified_at"`
	StartedAt    string           `json:"started_at,omitempty"`
	LastEntryAt  string           `json:"last_entry_at,omitempty"`
	AgeSeconds   int64            `json:"age_seconds"`
	Live         bool             `json:"live"`
	Summary      sessions.Summary `json:"summary"`
	Run          *sessionRunView  `json:"run,omitempty"`
	PR           *sessionPRView   `json:"pr,omitempty"`
	IssueTitle   string           `json:"issue_title,omitempty"`
	IssueState   string           `json:"issue_state,omitempty"`
	IssueStage   string           `json:"issue_stage,omitempty"`
	BranchName   string           `json:"branch_name,omitempty"`
}

type sessionGroupView struct {
	IssueNumber int64          `json:"issue_number"`
	Title       string         `json:"title,omitempty"`
	State       string         `json:"state,omitempty"`
	Stage       string         `json:"stage,omitempty"`
	BranchName  string         `json:"branch_name,omitempty"`
	PR          *sessionPRView `json:"pr,omitempty"`
	Sessions    []sessionView  `json:"sessions"`
	LiveCount   int            `json:"live_count"`
	UpdatedAt   string         `json:"updated_at,omitempty"`
	Unscoped    bool           `json:"unscoped,omitempty"`
}

// ─── Handlers ────────────────────────────────────────────────────────

// handleSessions lists every pi session on disk, grouped by issue (and by the
// pull request that issue produced).
func (s *webServer) handleSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	files, err := sessions.Scan(s.sessionDir)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusOK, map[string]any{
				"dir": s.sessionDir, "groups": []sessionGroupView{}, "sessions": []sessionView{},
				"count": 0, "live_count": 0, "generated_at": time.Now().UTC().Format(time.RFC3339),
			})
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	views, err := s.buildSessionViews(ctx, files)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	groups := groupSessions(views)
	liveCount := 0
	for _, v := range views {
		if v.Live {
			liveCount++
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"dir":          s.sessionDir,
		"groups":       groups,
		"sessions":     views,
		"count":        len(views),
		"live_count":   liveCount,
		"generated_at": time.Now().UTC().Format(time.RFC3339),
	})
}

// handleSessionDetail returns the newest window of one session's transcript.
func (s *webServer) handleSessionDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	file, ok := s.sessions.find(s.sessionDir, id)
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Errorf("session %q not found", id))
		return
	}

	limit := int(queryInt64(r, "limit"))
	chunk, err := sessions.ReadLast(file.Path, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	views, err := s.buildSessionViews(r.Context(), []sessions.File{file})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	view := views[0]
	// The transcript we just parsed is the freshest summary available.
	view.Summary = sessions.Summarize(chunk.Entries)
	s.sessionCache.Store(file, view.Summary)

	writeJSON(w, http.StatusOK, map[string]any{
		"session":         view,
		"entries":         chunk.Entries,
		"start_offset":    chunk.StartOffset,
		"next_offset":     chunk.NextOffset,
		"has_more_before": chunk.HasMoreBefore,
		"size_bytes":      chunk.Size,
		"modified_at":     chunk.ModifiedAt,
		"live":            view.Live,
		"skipped":         chunk.Skipped,
	})
}

// handleSessionEvents serves the incremental transcript feed: forward with
// ?offset=, backwards (scrollback) with ?before=.
func (s *webServer) handleSessionEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	file, ok := s.sessions.find(s.sessionDir, id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": fmt.Sprintf("session %q is no longer on disk", id),
			"gone":  true,
		})
		return
	}

	limit := int(queryInt64(r, "limit"))
	var (
		chunk *sessions.Chunk
		err   error
	)
	if before := queryInt64(r, "before"); before > 0 {
		chunk, err = sessions.ReadBefore(file.Path, before, limit)
	} else {
		chunk, err = sessions.ReadAfter(file.Path, queryInt64(r, "offset"), limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	live := time.Since(file.ModTime) <= sessionLiveWindow
	if run := s.matchRunForFile(r.Context(), file); run != nil {
		view := s.runView(*run)
		if run.Status == "running" && view.DurationMs < int64((30*time.Minute).Milliseconds()) {
			live = true
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"entries":      chunk.Entries,
		"start_offset": chunk.StartOffset,
		"next_offset":  chunk.NextOffset,
		"reset":        chunk.Reset,
		"has_more":     chunk.HasMore,
		"size_bytes":   chunk.Size,
		"modified_at":  chunk.ModifiedAt,
		"live":         live,
		"skipped":      chunk.Skipped,
	})
}

// ─── View building ───────────────────────────────────────────────────

func (s *webServer) buildSessionViews(ctx context.Context, files []sessions.File) ([]sessionView, error) {
	paths := make([]string, 0, len(files))
	numbers := make([]int64, 0, len(files))
	seen := make(map[int64]struct{}, len(files))
	for _, f := range files {
		paths = append(paths, f.Path)
		if f.IssueNumber > 0 {
			if _, ok := seen[f.IssueNumber]; !ok {
				seen[f.IssueNumber] = struct{}{}
				numbers = append(numbers, f.IssueNumber)
			}
		}
	}
	s.sessionCache.Retain(paths)

	issues, err := s.db.GetSessionIssues(ctx, numbers)
	if err != nil {
		return nil, err
	}
	runsByIssue, err := s.db.ListRunsByIssues(ctx, numbers)
	if err != nil {
		return nil, err
	}

	views := make([]sessionView, 0, len(files))
	for _, f := range files {
		summary, err := s.sessionCache.SummaryFor(f)
		if err != nil {
			// A file that vanished between scan and read is not an error worth
			// failing the whole listing for.
			continue
		}

		view := sessionView{
			ID:           f.SessionID,
			File:         f.Name,
			IssueNumber:  f.IssueNumber,
			Stage:        f.Stage,
			FixIteration: f.FixIteration,
			SizeBytes:    f.Size,
			ModifiedAt:   f.ModifiedAt(),
			StartedAt:    summary.StartedAt,
			LastEntryAt:  summary.LastEntryAt,
			AgeSeconds:   int64(time.Since(f.ModTime).Seconds()),
			Live:         time.Since(f.ModTime) <= sessionLiveWindow,
			Summary:      summary,
		}

		if issue, ok := issues[f.IssueNumber]; ok {
			view.IssueTitle = issue.Title
			view.IssueState = issue.State
			view.IssueStage = issue.Stage
			view.BranchName = issue.BranchName
			if issue.PRNumber > 0 {
				view.PR = &sessionPRView{
					Number:      issue.PRNumber,
					State:       displayPRState(issue.PRState, issue.PRMergedAt),
					ReviewState: issue.PRReviewState,
					URL:         issue.PRURL,
				}
			}
		}

		if run := matchRun(runsByIssue[f.IssueNumber], f.Stage, summary.StartedAt, f.ModTime); run != nil {
			view.Run = s.runView(*run)
			if run.Status == "running" {
				view.Live = true
			}
		}
		views = append(views, view)
	}

	sort.SliceStable(views, func(i, j int) bool {
		return views[i].ModifiedAt > views[j].ModifiedAt
	})
	return views, nil
}

func (s *webServer) runView(run storage.RunRecord) *sessionRunView {
	view := &sessionRunView{
		ID:        run.ID,
		Stage:     run.Stage,
		Status:    run.Status,
		StartedAt: run.StartedAt,
		EndedAt:   run.EndedAt,
		Error:     run.ErrorMessage,
	}
	if start, err := time.Parse(time.RFC3339, run.StartedAt); err == nil {
		end := time.Now()
		if run.EndedAt != "" {
			if parsed, err := time.Parse(time.RFC3339, run.EndedAt); err == nil {
				end = parsed
			}
		}
		view.DurationMs = end.Sub(start).Milliseconds()
	}
	return view
}

// matchRunForFile finds the pipeline run a single session file belongs to.
func (s *webServer) matchRunForFile(ctx context.Context, f sessions.File) *storage.RunRecord {
	if f.IssueNumber == 0 {
		return nil
	}
	runs, err := s.db.ListRunsByIssues(ctx, []int64{f.IssueNumber})
	if err != nil {
		return nil
	}
	summary, err := s.sessionCache.SummaryFor(f)
	if err != nil {
		return nil
	}
	return matchRun(runs[f.IssueNumber], f.Stage, summary.StartedAt, f.ModTime)
}

// matchRun pairs a session with the run of the same stage that overlaps it.
// Sessions and runs are both keyed by stage and start together, so the closest
// start time wins; a still-running run is preferred when two are close.
func matchRun(runs []storage.RunRecord, stage string, sessionStartedAt string, modTime time.Time) *storage.RunRecord {
	if len(runs) == 0 || stage == "" {
		return nil
	}
	var sessionStart time.Time
	if sessionStartedAt != "" {
		sessionStart, _ = time.Parse(time.RFC3339, sessionStartedAt)
	}
	if sessionStart.IsZero() {
		sessionStart = modTime
	}

	var best *storage.RunRecord
	bestScore := time.Duration(1<<62 - 1)
	for i := range runs {
		run := runs[i]
		if run.Stage != stage {
			continue
		}
		runStart, err := time.Parse(time.RFC3339, run.StartedAt)
		if err != nil {
			continue
		}
		score := runStart.Sub(sessionStart)
		if score < 0 {
			score = -score
		}
		// A run started after the session's last write cannot be its producer.
		if runStart.After(modTime.Add(2 * time.Minute)) {
			score += 24 * time.Hour
		}
		if run.Status == "running" {
			score -= time.Minute
		}
		if score < bestScore {
			bestScore = score
			best = &runs[i]
		}
	}
	return best
}

// groupSessions buckets sessions by issue, newest activity first. Sessions
// without a sloper issue number (a hand-run pi session in the same directory)
// land in one unscoped group.
func groupSessions(views []sessionView) []sessionGroupView {
	byIssue := make(map[int64]*sessionGroupView)
	order := make([]int64, 0)

	for _, view := range views {
		group, ok := byIssue[view.IssueNumber]
		if !ok {
			group = &sessionGroupView{
				IssueNumber: view.IssueNumber,
				Title:       view.IssueTitle,
				State:       view.IssueState,
				Stage:       view.IssueStage,
				BranchName:  view.BranchName,
				PR:          view.PR,
				Unscoped:    view.IssueNumber == 0,
			}
			if group.Unscoped {
				group.Title = "Sessions outside the pipeline"
			}
			byIssue[view.IssueNumber] = group
			order = append(order, view.IssueNumber)
		}
		group.Sessions = append(group.Sessions, view)
		if view.Live {
			group.LiveCount++
		}
		if view.ModifiedAt > group.UpdatedAt {
			group.UpdatedAt = view.ModifiedAt
		}
	}

	groups := make([]sessionGroupView, 0, len(order))
	for _, number := range order {
		group := byIssue[number]
		sort.SliceStable(group.Sessions, func(i, j int) bool {
			return sessionSortKey(group.Sessions[i]) < sessionSortKey(group.Sessions[j])
		})
		groups = append(groups, *group)
	}
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].LiveCount != groups[j].LiveCount {
			return groups[i].LiveCount > groups[j].LiveCount
		}
		return groups[i].UpdatedAt > groups[j].UpdatedAt
	})
	return groups
}

func displayPRState(state, mergedAt string) string {
	if state == "closed" && mergedAt != "" {
		return "merged"
	}
	return state
}

// sessionSortKey orders a stage's sessions chronologically, falling back to the
// last write when the header carried no timestamp.
func sessionSortKey(view sessionView) string {
	if view.StartedAt != "" {
		return view.StartedAt
	}
	return view.ModifiedAt
}

// sessionDir resolves where pi writes its session files. The scheduler passes
// --session-dir ~/.sloper/sessions; SLOPER_SESSION_DIR only matters when this
// server's home differs from the worker's, as when the two run in separate
// containers.
func sessionDir() string {
	if dir := os.Getenv("SLOPER_SESSION_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".sloper", "sessions")
	}
	return filepath.Join(home, ".sloper", "sessions")
}
