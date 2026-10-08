package worktree

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseWorktreeName(t *testing.T) {
	cases := []struct {
		rel    string
		kind   Kind
		issue  int64
		pr     int64
		reason string
	}{
		{"sloper/issue-42-null-pointer", KindWork, 42, 0, "work worktree"},
		{"sloper/issue-7-a", KindWork, 7, 0, "single digit issue"},
		{"sloper/issue-42", KindWork, 42, 0, "no slug"},
		{"sloper/issue-42-2024-regression", KindWork, 42, 0, "digits in the slug"},
		{"sloper/issue-42-null-pointer_fix", KindFix, 42, 0, "fix worktree"},
		{"sloper/issue-42_fix", KindFix, 42, 0, "fix worktree without a slug"},
		{"review-42-pr-118", KindReview, 42, 118, "review worktree"},
		{"review-7-pr-8", KindReview, 7, 8, "review worktree, small numbers"},
		{"", KindUnknown, 0, 0, "empty name"},
		{"scratch", KindUnknown, 0, 0, "unrelated directory"},
		{"sloper/issue-abc-null", KindUnknown, 0, 0, "no issue number"},
		{"review-42-pr", KindUnknown, 0, 0, "truncated review name"},
		{"sloper/issue-42-x/extra", KindWork, 42, 0, "issue number still readable"},
	}

	for _, tc := range cases {
		kind, issue, pr := ParseWorktreeName(tc.rel)
		if kind != tc.kind || issue != tc.issue || pr != tc.pr {
			t.Errorf("ParseWorktreeName(%q) = (%q, %d, %d), want (%q, %d, %d) — %s",
				tc.rel, kind, issue, pr, tc.kind, tc.issue, tc.pr, tc.reason)
		}
	}
}

// A slug is built from lowercase letters, digits and dashes only, so it can
// never end in "_fix" and the fix suffix is unambiguous.
func TestParseWorktreeNameFixSuffixIsUnambiguous(t *testing.T) {
	if kind, issue, _ := ParseWorktreeName("sloper/issue-42-fix-the-fix_fix"); kind != KindFix || issue != 42 {
		t.Errorf("got (%q, %d), want (%q, 42)", kind, issue, KindFix)
	}
}

// Worktrees are checked out of the parent repo, so the directory is a git
// worktree only when it carries a .git entry. This uses a temp dir rather than
// a real repo, so Scan must not depend on git being runnable.
func writeFakeWorktree(t *testing.T, base, rel string, mod time.Time) {
	t.Helper()
	dir := filepath.Join(base, filepath.FromSlash(rel))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	gitFile := filepath.Join(dir, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: /repo/.git/worktrees/x\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", gitFile, err)
	}
	if !mod.IsZero() {
		if err := os.Chtimes(dir, mod, mod); err != nil {
			t.Fatalf("chtimes %s: %v", dir, err)
		}
	}
}

func TestScanFindsBothLayouts(t *testing.T) {
	base := t.TempDir()

	oldest := time.Now().Add(-2 * time.Hour)
	newest := time.Now()
	middle := time.Now().Add(-1 * time.Hour)

	// Branch names contain a slash, so these sit one level down.
	writeFakeWorktree(t, base, "sloper/issue-42-null-pointer", oldest)
	writeFakeWorktree(t, base, "sloper/issue-7-a_fix", newest)
	// Review worktree names have no slash, so this one sits in the base dir.
	writeFakeWorktree(t, base, "review-42-pr-118", middle)

	// Neither of these is a worktree and both must be skipped.
	if err := os.MkdirAll(filepath.Join(base, "sloper", "not-a-worktree"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, "sloper", "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := Scan(base)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	want := []struct {
		rel   string
		kind  Kind
		issue int64
		pr    int64
	}{
		{"sloper/issue-7-a_fix", KindFix, 7, 0}, // newest first
		{"review-42-pr-118", KindReview, 42, 118},
		{"sloper/issue-42-null-pointer", KindWork, 42, 0},
	}
	if len(got) != len(want) {
		t.Fatalf("Scan returned %d entries, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].RelPath != w.rel || got[i].Kind != w.kind ||
			got[i].IssueNumber != w.issue || got[i].PRNumber != w.pr {
			t.Errorf("entry %d = {%q %q %d %d}, want {%q %q %d %d}",
				i, got[i].RelPath, got[i].Kind, got[i].IssueNumber, got[i].PRNumber,
				w.rel, w.kind, w.issue, w.pr)
		}
	}
}

// Sloper creates no worktrees before its first run, and a missing base
// directory is the normal state rather than a failure.
func TestScanMissingBaseDir(t *testing.T) {
	got, err := Scan(filepath.Join(t.TempDir(), "worktrees"))
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Scan returned %d entries, want 0", len(got))
	}
}
