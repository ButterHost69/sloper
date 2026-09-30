package worktree

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// FixSuffix is appended to a branch name to name the worktree an agent uses
// to address review findings.
const FixSuffix = "_fix"

// The worktree names BranchNameForIssue and ReviewWorktreeName produce, read
// back. The slug in a branch name is decorative — only the issue and PR
// numbers are load-bearing, because the database only learns a branch name
// after the agent has finished, while the directory exists for the whole run.
var (
	workNameRe   = regexp.MustCompile(`^sloper/issue-(\d+)(?:-|$)`)
	reviewNameRe = regexp.MustCompile(`^review-(\d+)-pr-(\d+)$`)
)

// Kind is what a worktree checkout is for.
type Kind string

const (
	KindWork    Kind = "work"
	KindReview  Kind = "review"
	KindFix     Kind = "fix"
	KindUnknown Kind = "unknown"
)

// Entry is one worktree directory found on disk.
type Entry struct {
	// RelPath is the path relative to the base dir, always slash-separated
	// ("sloper/issue-42-null-pointer"). Callers that report worktrees to a
	// remote console send this rather than the absolute path, which would
	// leak the host's home directory layout.
	RelPath string
	Kind    Kind
	// IssueNumber and PRNumber are 0 when RelPath does not follow one of the
	// known naming rules — a leftover directory, not sloper work.
	IssueNumber int64
	PRNumber    int64
	ModTime     time.Time
}

// ParseWorktreeName recovers the issue a worktree belongs to from its path
// relative to the base dir. It returns KindUnknown with zero numbers when the
// name is not one sloper created.
func ParseWorktreeName(rel string) (kind Kind, issueNumber, prNumber int64) {
	name := rel
	kind = KindWork
	if trimmed, ok := strings.CutSuffix(name, FixSuffix); ok {
		kind = KindFix
		name = trimmed
	}
	if m := reviewNameRe.FindStringSubmatch(name); m != nil {
		return KindReview, parseNumber(m[1]), parseNumber(m[2])
	}
	if m := workNameRe.FindStringSubmatch(name); m != nil {
		return kind, parseNumber(m[1]), 0
	}
	return KindUnknown, 0, 0
}

func parseNumber(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// Scan lists the worktree directories under baseDir, newest first.
//
// Branch names contain a slash ("sloper/issue-42-null-pointer"), so worktree
// directories sit one level below the base dir while review worktrees sit
// directly in it. Both layouts are covered. A missing base dir is not an
// error: it just means sloper has no worktrees.
func Scan(baseDir string) ([]Entry, error) {
	roots, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("worktree: scan %s: %w", baseDir, err)
	}

	var out []Entry
	add := func(rel string, mod time.Time) {
		kind, issue, pr := ParseWorktreeName(rel)
		out = append(out, Entry{
			RelPath:     rel,
			Kind:        kind,
			IssueNumber: issue,
			PRNumber:    pr,
			ModTime:     mod,
		})
	}

	for _, root := range roots {
		if !root.IsDir() {
			continue
		}
		rootPath := filepath.Join(baseDir, root.Name())

		// A review worktree ("review-42-pr-118") has no slash in its name.
		if isWorktreeDir(rootPath) {
			add(root.Name(), modTime(root))
			continue
		}

		// Otherwise root is a branch-name prefix ("sloper") holding the
		// worktree directories it created.
		children, err := os.ReadDir(rootPath)
		if err != nil {
			continue
		}
		for _, child := range children {
			if !child.IsDir() {
				continue
			}
			childPath := filepath.Join(rootPath, child.Name())
			if !isWorktreeDir(childPath) {
				continue
			}
			add(root.Name()+"/"+child.Name(), modTime(child))
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].ModTime.Equal(out[j].ModTime) {
			return out[i].RelPath < out[j].RelPath
		}
		return out[i].ModTime.After(out[j].ModTime)
	})
	return out, nil
}

// isWorktreeDir reports whether dir is a git checkout. Worktrees carry a .git
// file pointing at the parent repo; this only checks that it is there.
func isWorktreeDir(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

func modTime(e fs.DirEntry) time.Time {
	info, err := e.Info()
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}
