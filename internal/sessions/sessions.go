// Package sessions reads the pi coding-agent session files that sloper workers
// write, so the console can show what each stage is doing.
//
// pi stores a session as a JSONL file named "<timestamp>_<session-id>.jsonl"
// under the directory given to --session-dir (sloper passes ~/.sloper/sessions).
// Sloper hands every stage a deterministic id — sloper-issue-42-work — so the
// file name alone carries the issue and the stage.
package sessions

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// maxScanDepth bounds how far below the session root we look. pi writes flat
// files when --session-dir is set, but its default layout nests one directory
// per working directory.
const maxScanDepth = 3

// File is one session file on disk.
type File struct {
	Path         string    `json:"-"`
	Name         string    `json:"file"`
	SessionID    string    `json:"id"`
	IssueNumber  int64     `json:"issue_number"`
	Stage        string    `json:"stage"`
	FixIteration int       `json:"fix_iteration"`
	Size         int64     `json:"size_bytes"`
	ModTime      time.Time `json:"-"`
}

// ModifiedAt is the file mtime as an RFC 3339 string.
func (f File) ModifiedAt() string {
	if f.ModTime.IsZero() {
		return ""
	}
	return f.ModTime.UTC().Format(time.RFC3339)
}

// sloperSessionID matches the ids the scheduler hands to pi.
var sloperSessionID = regexp.MustCompile(`^sloper-issue-(\d+)-(spec|work|review|fix)(?:-(\d+))?$`)

// ParseFileName extracts the session id from a pi session file name. The
// timestamp prefix contains no underscore, so the first one separates it from
// the id; a file without a prefix is treated as the id itself.
func ParseFileName(name string) string {
	name = strings.TrimSuffix(filepath.Base(name), ".jsonl")
	if i := strings.Index(name, "_"); i >= 0 {
		return name[i+1:]
	}
	return name
}

// ParseSessionID splits a sloper session id into its parts. ok is false for
// ids sloper did not mint, which the console groups separately.
func ParseSessionID(id string) (issueNumber int64, stage string, fixIteration int, ok bool) {
	m := sloperSessionID.FindStringSubmatch(id)
	if m == nil {
		return 0, "", 0, false
	}
	issueNumber, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return 0, "", 0, false
	}
	if m[3] != "" {
		fixIteration, _ = strconv.Atoi(m[3])
	}
	return issueNumber, m[2], fixIteration, true
}

// Scan lists every session file below root, newest first. Unreadable
// subdirectories are skipped rather than failing the whole listing.
func Scan(root string) ([]File, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("sessions: empty session directory")
	}
	root = filepath.Clean(root)

	out := make([]File, 0)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == root {
				return err
			}
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			if path == root {
				return nil
			}
			if strings.HasPrefix(d.Name(), ".") || depth(root, path) >= maxScanDepth {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".jsonl") {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return nil
		}

		f := File{
			Path:      path,
			Name:      d.Name(),
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			SessionID: ParseFileName(d.Name()),
		}
		f.IssueNumber, f.Stage, f.FixIteration, _ = ParseSessionID(f.SessionID)
		out = append(out, f)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("sessions: scan %s: %w", root, err)
	}
	return out, nil
}

// depth counts path separators between root and path.
func depth(root, path string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return 0
	}
	return strings.Count(rel, string(filepath.Separator))
}
