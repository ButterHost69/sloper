package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ButterHost69/sloper/internal/sessions"
	"github.com/ButterHost69/sloper/internal/storage"
)

func TestSessionsAPI(t *testing.T) {
	dir := t.TempDir()
	sessionDir := filepath.Join(dir, "sessions")
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}

	longAgo := time.Now().Add(-2 * time.Hour)
	specFile := writeSession(t, sessionDir, "2026-10-07T18-00-00-000Z_sloper-issue-42-spec.jsonl", strings.Join([]string{
		header("sloper-issue-42-spec", "2026-10-07T18:00:00.000Z"),
		message("u1", "user", map[string]any{
			"content":   []map[string]any{{"type": "text", "text": "Spec out issue 42"}},
			"timestamp": 1783962000000,
		}),
		message("a1", "assistant", map[string]any{
			"content":    []map[string]any{{"type": "text", "text": "Here is the plan."}},
			"model":      "deepseek-v4-pro",
			"provider":   "opencode-go",
			"stopReason": "stop",
			"usage":      map[string]any{"totalTokens": 900, "cost": map[string]any{"total": 0.11}},
			"timestamp":  1783962060000,
		}),
	}, "\n")+"\n")
	touch(t, specFile, longAgo)

	workFile := writeSession(t, sessionDir, "2026-10-07T20-00-00-000Z_sloper-issue-42-work.jsonl", strings.Join([]string{
		header("sloper-issue-42-work", "2026-10-07T20:00:00.000Z"),
		message("u1", "user", map[string]any{
			"content":   []map[string]any{{"type": "text", "text": "Implement the plan"}},
			"timestamp": 1783963000000,
		}),
		message("a1", "assistant", map[string]any{
			"content": []map[string]any{
				{"type": "thinking", "thinking": "I should run the tests first."},
				{"type": "text", "text": "Running the pipeline tests."},
				{"type": "toolCall", "id": "call_live", "name": "bash", "arguments": map[string]any{"command": "go test ./internal/pipeline/..."}},
			},
			"model":      "deepseek-v4-pro",
			"provider":   "opencode-go",
			"stopReason": "toolUse",
			"usage":      map[string]any{"totalTokens": 1500, "cost": map[string]any{"total": 0.2}},
			"timestamp":  1783963060000,
		}),
	}, "\n")+"\n")

	foreignFile := writeSession(t, sessionDir, "2026-10-07T17-00-00-000Z_019f2bb7-f3b8-7924.jsonl", header("019f2bb7-f3b8-7924", "2026-10-07T17:00:00.000Z")+"\n")
	touch(t, foreignFile, longAgo)

	dbPath := filepath.Join(dir, "sloper.sqlite")
	db, err := storage.OpenDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	if err := storage.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	repo := storage.NewRepositories(db)
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.ExecContext(ctx, `INSERT INTO issues
		(number,title,state,url,author,updated_at,stage,branch_name,pr_number,created_at,updated_at_local)
		VALUES (42,'Fix the null pointer','open','https://example.test/issues/42','anorak',?,'work-done','sloper/issue-42-null-pointer',118,?,?)`,
		now, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO pull_requests
		(number,issue_number,title,head_sha,base_sha,state,url,updated_at,review_state)
		VALUES (118,42,'Fix the null pointer','abc','def','open','https://example.test/pull/118',?,'changes_requested')`, now); err != nil {
		t.Fatal(err)
	}
	specRun, err := repo.StartRun(ctx, 42, "spec")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.CompleteRun(ctx, specRun, "spec output", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.StartRun(ctx, 42, "work"); err != nil {
		t.Fatal(err)
	}

	server := &webServer{
		db:           repo,
		sessionDir:   sessionDir,
		sessionCache: sessions.NewCache(),
	}
	mux := http.NewServeMux()
	server.routes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	t.Run("list groups sessions by issue", func(t *testing.T) {
		var body struct {
			Groups []struct {
				IssueNumber int64 `json:"issue_number"`
				Title       string
				PR          *struct {
					Number int64 `json:"number"`
					State  string
				} `json:"pr"`
				LiveCount int `json:"live_count"`
				Sessions  []struct {
					ID      string `json:"id"`
					Stage   string `json:"stage"`
					Live    bool   `json:"live"`
					Summary struct {
						Model   string `json:"model"`
						Current *struct {
							Kind  string `json:"kind"`
							Tools []struct {
								Name    string `json:"name"`
								Preview string `json:"preview"`
							} `json:"tools"`
						} `json:"current"`
					} `json:"summary"`
				} `json:"sessions"`
			} `json:"groups"`
			Count     int `json:"count"`
			LiveCount int `json:"live_count"`
		}
		getJSON(t, ts.URL+"/api/sessions", &body)

		if body.Count != 3 {
			t.Fatalf("count = %d, want 3", body.Count)
		}
		if body.LiveCount != 1 {
			t.Fatalf("live_count = %d, want 1", body.LiveCount)
		}

		var issue42 *struct {
			IssueNumber int64 `json:"issue_number"`
			Title       string
			PR          *struct {
				Number int64 `json:"number"`
				State  string
			} `json:"pr"`
			LiveCount int `json:"live_count"`
			Sessions  []struct {
				ID      string `json:"id"`
				Stage   string `json:"stage"`
				Live    bool   `json:"live"`
				Summary struct {
					Model   string `json:"model"`
					Current *struct {
						Kind  string `json:"kind"`
						Tools []struct {
							Name    string `json:"name"`
							Preview string `json:"preview"`
						} `json:"tools"`
					} `json:"current"`
				} `json:"summary"`
			} `json:"sessions"`
		}
		unscoped := false
		for i := range body.Groups {
			g := &body.Groups[i]
			switch g.IssueNumber {
			case 42:
				issue42 = g
			case 0:
				unscoped = true
			}
		}
		if issue42 == nil {
			t.Fatal("issue 42 group missing")
		}
		if !unscoped {
			t.Fatal("unscoped group missing")
		}
		if issue42.PR == nil || issue42.PR.Number != 118 || issue42.PR.State != "open" {
			t.Fatalf("PR = %+v, want #118 open", issue42.PR)
		}
		if len(issue42.Sessions) != 2 {
			t.Fatalf("issue 42 has %d sessions, want 2", len(issue42.Sessions))
		}
		if issue42.Sessions[0].Stage != "spec" || issue42.Sessions[1].Stage != "work" {
			t.Fatalf("stage order = %q, %q; want spec then work",
				issue42.Sessions[0].Stage, issue42.Sessions[1].Stage)
		}
		if issue42.Sessions[0].Live {
			t.Error("the finished spec session should not be live")
		}
		work := issue42.Sessions[1]
		if !work.Live {
			t.Error("the running work session should be live")
		}
		if work.Summary.Model != "deepseek-v4-pro" {
			t.Errorf("model = %q", work.Summary.Model)
		}
		if work.Summary.Current == nil || work.Summary.Current.Kind != "tool" {
			t.Fatalf("current = %+v, want a running tool", work.Summary.Current)
		}
		if got := work.Summary.Current.Tools[0].Preview; !strings.Contains(got, "go test") {
			t.Errorf("pending tool preview = %q", got)
		}
	})

	t.Run("detail returns the transcript", func(t *testing.T) {
		var body struct {
			Session struct {
				ID    string `json:"id"`
				Stage string `json:"stage"`
				Live  bool   `json:"live"`
			} `json:"session"`
			Entries []struct {
				Type    string `json:"type"`
				Message *struct {
					Role string `json:"role"`
				} `json:"message"`
			} `json:"entries"`
			NextOffset    int64 `json:"next_offset"`
			HasMoreBefore bool  `json:"has_more_before"`
		}
		getJSON(t, ts.URL+"/api/sessions/sloper-issue-42-work", &body)
		if body.Session.ID != "sloper-issue-42-work" || body.Session.Stage != "work" {
			t.Fatalf("session = %+v", body.Session)
		}
		if len(body.Entries) != 3 {
			t.Fatalf("entries = %d, want 3", len(body.Entries))
		}
		if body.HasMoreBefore {
			t.Error("a whole-file read should not claim older entries")
		}
		if body.NextOffset == 0 {
			t.Error("next_offset should point at the end of the transcript")
		}
	})

	t.Run("events stream only what is new", func(t *testing.T) {
		var first struct {
			NextOffset int64 `json:"next_offset"`
		}
		getJSON(t, ts.URL+"/api/sessions/sloper-issue-42-work", &first)

		var empty struct {
			Entries []json.RawMessage `json:"entries"`
		}
		getJSON(t, fmt.Sprintf("%s/api/sessions/sloper-issue-42-work/events?offset=%d", ts.URL, first.NextOffset), &empty)
		if len(empty.Entries) != 0 {
			t.Fatalf("expected no new entries, got %d", len(empty.Entries))
		}

		result := message("r1", "toolResult", map[string]any{
			"toolCallId": "call_live",
			"toolName":   "bash",
			"isError":    false,
			"content":    []map[string]any{{"type": "text", "text": "ok"}},
			"timestamp":  1783963200000,
		})
		appendLine(t, workFile, result)

		var next struct {
			Entries []struct {
				Type    string `json:"type"`
				Message *struct {
					Role string `json:"role"`
				} `json:"message"`
			} `json:"entries"`
		}
		getJSON(t, fmt.Sprintf("%s/api/sessions/sloper-issue-42-work/events?offset=%d", ts.URL, first.NextOffset), &next)
		if len(next.Entries) != 1 || next.Entries[0].Message == nil || next.Entries[0].Message.Role != "toolResult" {
			t.Fatalf("stream entries = %+v, want one toolResult", next.Entries)
		}
	})

	t.Run("missing and traversal ids are 404", func(t *testing.T) {
		for _, id := range []string{"nope", "..%2F..%2Fetc%2Fpasswd", "sloper-issue-42-spec%00"} {
			res, err := http.Get(ts.URL + "/api/sessions/" + id)
			if err != nil {
				t.Fatal(err)
			}
			_ = res.Body.Close()
			if res.StatusCode != http.StatusNotFound {
				t.Errorf("GET %s = %d, want 404", id, res.StatusCode)
			}
		}
	})
}

func TestGroupSessionsKeepsLiveGroupsFirst(t *testing.T) {
	groups := groupSessions([]sessionView{
		{IssueNumber: 1, ModifiedAt: "2026-10-07T20:00:00Z", StartedAt: "2026-10-07T19:00:00Z"},
		{IssueNumber: 2, ModifiedAt: "2026-10-07T18:00:00Z", StartedAt: "2026-10-07T18:00:00Z", Live: true},
	})
	if len(groups) != 2 || groups[0].IssueNumber != 2 {
		t.Fatalf("groups = %+v, want the live group first", groups)
	}
}

func TestMatchRunPicksTheSameStageAndTime(t *testing.T) {
	runs := []storage.RunRecord{
		{ID: 1, Stage: "spec", Status: "completed", StartedAt: "2026-10-07T18:00:00Z"},
		{ID: 2, Stage: "work", Status: "running", StartedAt: "2026-10-07T20:00:00Z"},
	}
	got := matchRun(runs, "work", "2026-10-07T20:00:01Z", time.Date(2026, 10, 7, 20, 5, 0, 0, time.UTC))
	if got == nil || got.ID != 2 {
		t.Fatalf("matchRun = %+v, want run 2", got)
	}
	if got := matchRun(runs, "review", "2026-10-07T20:00:01Z", time.Now()); got != nil {
		t.Fatalf("matchRun for an unrun stage = %+v, want nil", got)
	}
}

// ─── helpers ─────────────────────────────────────────────────────────

func getJSON(t *testing.T, url string, dst any) {
	t.Helper()
	res, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET %s = %d", url, res.StatusCode)
	}
	if err := json.NewDecoder(res.Body).Decode(dst); err != nil {
		t.Fatalf("decode %s: %v", url, err)
	}
}

func writeSession(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func appendLine(t *testing.T, path, line string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(line + "\n"); err != nil {
		t.Fatal(err)
	}
}

func touch(t *testing.T, path string, when time.Time) {
	t.Helper()
	if err := os.Chtimes(path, when, when); err != nil {
		t.Fatal(err)
	}
}

func header(id, timestamp string) string {
	return fmt.Sprintf(`{"type":"session","version":3,"id":%q,"timestamp":%q,"cwd":"/repo"}`, id, timestamp)
}

func message(id, role string, fields map[string]any) string {
	msg := map[string]any{"role": role}
	for k, v := range fields {
		msg[k] = v
	}
	entry := map[string]any{
		"type":      "message",
		"id":        id,
		"parentId":  nil,
		"timestamp": "2026-10-07T20:00:01.000Z",
		"message":   msg,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		panic(err)
	}
	return string(data)
}
