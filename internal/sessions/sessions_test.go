package sessions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseFileNameStripsTimestampPrefix(t *testing.T) {
	cases := map[string]string{
		"2026-10-07T20-11-02-123Z_sloper-issue-42-work.jsonl": "sloper-issue-42-work",
		"sloper-issue-7-fix-2.jsonl":                          "sloper-issue-7-fix-2",
		"2026-10-07T20-11-02-123Z_019f2bb7-f3b8-7924.jsonl":   "019f2bb7-f3b8-7924",
	}
	for name, want := range cases {
		if got := ParseFileName(name); got != want {
			t.Errorf("ParseFileName(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestParseSessionID(t *testing.T) {
	issue, stage, fix, ok := ParseSessionID("sloper-issue-42-work")
	if !ok || issue != 42 || stage != "work" || fix != 0 {
		t.Fatalf("work session parsed as (%d, %q, %d, %v)", issue, stage, fix, ok)
	}
	issue, stage, fix, ok = ParseSessionID("sloper-issue-12-fix-3")
	if !ok || issue != 12 || stage != "fix" || fix != 3 {
		t.Fatalf("fix session parsed as (%d, %q, %d, %v)", issue, stage, fix, ok)
	}
	if _, _, _, ok := ParseSessionID("019f2bb7-f3b8-7924-9cc7-003529aafef8"); ok {
		t.Fatal("foreign session id should not parse as a sloper session")
	}
}

func TestScanFindsNestedFilesAndSkipsHiddenDirs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "2026-10-07T20-11-02-123Z_sloper-issue-1-work.jsonl"), sessionHeader("s1")+"\n")
	nested := filepath.Join(root, "--home-repo--")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(nested, "2026-10-07T20-11-02-123Z_sloper-issue-2-spec.jsonl"), sessionHeader("s2")+"\n")
	hidden := filepath.Join(root, ".cache")
	if err := os.MkdirAll(hidden, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(hidden, "2026-10-07T20-11-02-123Z_sloper-issue-3-spec.jsonl"), sessionHeader("s3")+"\n")
	writeFile(t, filepath.Join(root, "notes.txt"), "ignore me")

	files, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("Scan found %d files, want 2: %+v", len(files), files)
	}
	seen := map[int64]string{}
	for _, f := range files {
		seen[f.IssueNumber] = f.Stage
	}
	if seen[1] != "work" || seen[2] != "spec" {
		t.Fatalf("unexpected scan result: %+v", seen)
	}
}

func TestReadAfterLeavesPartialTrailingLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	first := sessionHeader("s1")
	second := entryLine(t, "message", map[string]any{
		"role":      "user",
		"content":   []map[string]any{{"type": "text", "text": "hello"}},
		"timestamp": 1783962377491,
	})
	writeFile(t, path, first+"\n"+second+"\n"+`{"type":"message","id":"partial"`)

	chunk, err := ReadAfter(path, 0, 10)
	if err != nil {
		t.Fatalf("ReadAfter: %v", err)
	}
	if len(chunk.Entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(chunk.Entries))
	}
	if chunk.NextOffset >= chunk.Size {
		t.Fatalf("NextOffset %d should stop before the partial line (size %d)", chunk.NextOffset, chunk.Size)
	}

	// pi finishes the line: the next poll must pick it up from the same offset.
	writeFile(t, path, first+"\n"+second+"\n"+`{"type":"message","id":"partial"}`+"\n")
	rest, err := ReadAfter(path, chunk.NextOffset, 10)
	if err != nil {
		t.Fatalf("ReadAfter (tail): %v", err)
	}
	if len(rest.Entries) != 1 || rest.Entries[0].ID != "partial" {
		t.Fatalf("tail entries = %+v, want the completed partial line", rest.Entries)
	}
	if rest.NextOffset != rest.Size {
		t.Fatalf("NextOffset %d != size %d", rest.NextOffset, rest.Size)
	}
	if rest.Reset {
		t.Fatal("tail read should not reset")
	}
}

func TestReadAfterResetsWhenOffsetIsPastEOF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	writeFile(t, path, sessionHeader("s1")+"\n")

	chunk, err := ReadAfter(path, 1<<20, 10)
	if err != nil {
		t.Fatalf("ReadAfter: %v", err)
	}
	if !chunk.Reset || chunk.NextOffset != 0 {
		t.Fatalf("Reset = %v, NextOffset = %d; want true and 0", chunk.Reset, chunk.NextOffset)
	}
}

func TestReadLastReturnsNewestWindowWithOffsets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	lines := []string{sessionHeader("s1")}
	for i := 0; i < 10; i++ {
		lines = append(lines, entryLine(t, "message", map[string]any{
			"role":      "user",
			"content":   []map[string]any{{"type": "text", "text": string(rune('a' + i))}},
			"timestamp": 1783962377491 + i,
		}))
	}
	writeFile(t, path, strings.Join(lines, "\n")+"\n")

	chunk, err := ReadLast(path, 3)
	if err != nil {
		t.Fatalf("ReadLast: %v", err)
	}
	if len(chunk.Entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(chunk.Entries))
	}
	if !chunk.HasMoreBefore {
		t.Fatal("HasMoreBefore should be true when the window is a tail")
	}
	if chunk.NextOffset != chunk.Size {
		t.Fatalf("NextOffset %d != size %d for a file ending in a newline", chunk.NextOffset, chunk.Size)
	}

	older, err := ReadBefore(path, chunk.StartOffset, 3)
	if err != nil {
		t.Fatalf("ReadBefore: %v", err)
	}
	if len(older.Entries) != 3 {
		t.Fatalf("older window has %d entries, want 3", len(older.Entries))
	}
	if older.NextOffset != chunk.StartOffset {
		t.Fatalf("older NextOffset %d != %d", older.NextOffset, chunk.StartOffset)
	}
}

func TestSummarizeTracksPendingToolAndUsage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	writeFile(t, path, strings.Join([]string{
		sessionHeader("sloper-issue-42-work"),
		entryLine(t, "message", map[string]any{
			"role":      "user",
			"content":   []map[string]any{{"type": "text", "text": "Fix the null pointer\nwith more detail"}},
			"timestamp": 1783962377491,
		}),
		entryLine(t, "message", map[string]any{
			"role":    "assistant",
			"content": []map[string]any{{"type": "text", "text": "Looking at the parser."}, {"type": "toolCall", "id": "call_1", "name": "bash", "arguments": map[string]any{"command": "go test ./..."}}},
			"model":   "deepseek-v4-pro", "provider": "opencode-go",
			"stopReason": "toolUse",
			"usage":      map[string]any{"totalTokens": 1200, "cost": map[string]any{"total": 0.42}},
			"timestamp":  1783962378000,
		}),
		entryLine(t, "message", map[string]any{
			"role": "toolResult", "toolCallId": "call_1", "toolName": "bash", "isError": false,
			"content": []map[string]any{{"type": "text", "text": "ok"}}, "timestamp": 1783962380000,
		}),
		entryLine(t, "message", map[string]any{
			"role":    "assistant",
			"content": []map[string]any{{"type": "thinking", "thinking": "next I should edit"}, {"type": "toolCall", "id": "call_2", "name": "edit", "arguments": map[string]any{"path": "internal/pipeline/pipeline.go"}}},
			"model":   "deepseek-v4-pro", "provider": "opencode-go",
			"stopReason": "toolUse",
			"usage":      map[string]any{"totalTokens": 300, "cost": map[string]any{"total": 0.1}},
			"timestamp":  1783962381000,
		}),
	}, "\n")+"\n")

	chunk, err := ReadAll(path)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	s := Summarize(chunk.Entries)

	if s.Title != "Fix the null pointer" {
		t.Errorf("Title = %q", s.Title)
	}
	if s.Messages != 4 || s.UserMessages != 1 || s.AssistantMessages != 2 || s.ToolResults != 1 {
		t.Errorf("counts = %+v", s)
	}
	if s.ToolCalls != 2 || s.ToolErrors != 0 {
		t.Errorf("tool calls = %d, errors = %d", s.ToolCalls, s.ToolErrors)
	}
	if s.TotalTokens != 1500 || s.CostUSD < 0.51 || s.CostUSD > 0.53 {
		t.Errorf("usage = %d tokens, $%.2f", s.TotalTokens, s.CostUSD)
	}
	if s.Model != "deepseek-v4-pro" || s.Provider != "opencode-go" {
		t.Errorf("model = %q/%q", s.Provider, s.Model)
	}
	if s.Current == nil || s.Current.Kind != "tool" {
		t.Fatalf("Current = %+v, want a pending tool", s.Current)
	}
	if len(s.Current.Tools) != 1 || s.Current.Tools[0].Name != "edit" {
		t.Fatalf("pending tools = %+v", s.Current.Tools)
	}
	if s.Current.Tools[0].Preview != "internal/pipeline/pipeline.go" {
		t.Errorf("preview = %q", s.Current.Tools[0].Preview)
	}
	if len(s.Tools) != 1 || s.Tools[0].Name != "edit" {
		t.Errorf("last message tools = %+v", s.Tools)
	}
	if s.LastEntryAt == "" || s.StartedAt == "" {
		t.Errorf("timestamps missing: %+v", s)
	}
}

func TestSummarizeReportsErrorStop(t *testing.T) {
	entries := []Entry{{
		Type:      "message",
		Timestamp: "2026-10-07T20:00:00.000Z",
		Message: &Message{
			Role:         "assistant",
			StopReason:   "error",
			ErrorMessage: "401: CreditsError",
		},
	}}
	s := Summarize(entries)
	if s.Current == nil || s.Current.Kind != "error" {
		t.Fatalf("Current = %+v, want an error", s.Current)
	}
}

func TestCacheRecomputesWhenFileChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	writeFile(t, path, sessionHeader("s1")+"\n")
	past := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(path, past, past); err != nil {
		t.Fatal(err)
	}

	cache := NewCache()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	file := File{Path: path, Size: info.Size(), ModTime: info.ModTime()}
	first, err := cache.SummaryFor(file)
	if err != nil {
		t.Fatalf("SummaryFor: %v", err)
	}
	if first.Messages != 0 {
		t.Fatalf("first summary = %+v, want no messages", first)
	}

	writeFile(t, path, sessionHeader("s1")+"\n"+entryLine(t, "message", map[string]any{
		"role": "user", "content": []map[string]any{{"type": "text", "text": "hi"}}, "timestamp": 1,
	})+"\n")
	logical := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(path, logical, logical); err != nil {
		t.Fatal(err)
	}
	info, err = os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	file = File{Path: path, Size: info.Size(), ModTime: info.ModTime()}
	second, err := cache.SummaryFor(file)
	if err != nil {
		t.Fatalf("SummaryFor (changed): %v", err)
	}
	if second.Messages != 1 || second.UserMessages != 1 {
		t.Fatalf("second summary = %+v, want one user message", second)
	}
}

func TestToolPreview(t *testing.T) {
	cases := []struct {
		name string
		args string
		want string
	}{
		{"bash", `{"command":"go test ./...\nmore"}`, "go test ./..."},
		{"read", `{"path":"main.go"}`, "main.go"},
		{"unknown", `{"zeta":"long value here","alpha":"short"}`, "short"},
		{"bash", ``, ""},
	}
	for _, tc := range cases {
		if got := ToolPreview(tc.name, json.RawMessage(tc.args)); got != tc.want {
			t.Errorf("ToolPreview(%q, %s) = %q, want %q", tc.name, tc.args, got, tc.want)
		}
	}
}

func TestMessageContentAcceptsPlainString(t *testing.T) {
	var m Message
	if err := json.Unmarshal([]byte(`{"role":"user","content":"hello there"}`), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(m.Content) != 1 || m.Content[0].Text != "hello there" {
		t.Fatalf("content = %+v", m.Content)
	}
}

func sessionHeader(id string) string {
	return `{"type":"session","version":3,"id":"` + id + `","timestamp":"2026-10-07T20:00:00.000Z","cwd":"/repo"}`
}

func entryLine(t *testing.T, typ string, extra map[string]any) string {
	t.Helper()
	entry := map[string]any{"type": typ, "id": "e" + time.Now().Format("150405.000000"), "parentId": nil, "timestamp": "2026-10-07T20:00:01.000Z"}
	if typ == "message" {
		entry["message"] = extra
	} else {
		for k, v := range extra {
			entry[k] = v
		}
	}
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
