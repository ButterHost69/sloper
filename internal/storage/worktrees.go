package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// WorktreeIssue is the database half of a worktree: the label the console shows
// next to a worktree directory. Every field is read from the issues and
// pull_requests tables on demand — worktrees get no table of their own because
// their directories are transient while these rows are permanent.
type WorktreeIssue struct {
	Number        int64
	Title         string
	State         string
	Stage         string
	BranchName    string
	PRNumber      int64
	PRState       string
	PRReviewState string
	UpdatedAt     string
	RunStage      string
	RunStatus     string
}

// worktreeIssueSelect joins an issue to its PR and to the most recent pipeline
// run, which is what tells the console whether an agent is working on it.
const worktreeIssueSelect = `
	SELECT i.number, i.title, i.state, i.stage, COALESCE(i.branch_name, ''),
	       COALESCE(i.pr_number, 0), COALESCE(p.state, ''), COALESCE(p.review_state, ''),
	       i.updated_at_local,
	       COALESCE((SELECT r.stage FROM runs r
	                  WHERE r.issue_number = i.number ORDER BY r.id DESC LIMIT 1), ''),
	       COALESCE((SELECT r.status FROM runs r
	                  WHERE r.issue_number = i.number ORDER BY r.id DESC LIMIT 1), '')
	FROM issues i
	LEFT JOIN pull_requests p ON p.number = i.pr_number`

// GetWorktreeIssues resolves the issues behind a set of worktree directories,
// keyed by issue number. Issues with no cached row are simply absent, which
// leaves the worktree visible but unlabelled rather than hiding it.
func (r *Repositories) GetWorktreeIssues(ctx context.Context, numbers []int64) (map[int64]WorktreeIssue, error) {
	out := make(map[int64]WorktreeIssue, len(numbers))
	for _, batch := range chunkNumbers(numbers, 200) {
		placeholders := make([]string, len(batch))
		args := make([]any, len(batch))
		for i, n := range batch {
			placeholders[i] = "?"
			args[i] = n
		}
		rows, err := r.db.QueryContext(ctx,
			worktreeIssueSelect+" WHERE i.number IN ("+strings.Join(placeholders, ",")+")", args...)
		if err != nil {
			return nil, fmt.Errorf("storage: get worktree issues: %w", err)
		}
		records, err := scanWorktreeIssues(rows)
		if err != nil {
			return nil, err
		}
		for _, rec := range records {
			out[rec.Number] = rec
		}
	}
	return out, nil
}

// ListOpenWorktreeIssues returns the issues that still have unmerged work: a
// branch that was pushed and a PR that is neither closed nor merged. The
// branch name is written only after an agent finishes its work, so this set
// never includes a run that is still in progress.
func (r *Repositories) ListOpenWorktreeIssues(ctx context.Context, limit int) ([]WorktreeIssue, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx,
		worktreeIssueSelect+`
		WHERE COALESCE(i.branch_name, '') != ''
		  AND i.state != 'closed'
		  AND i.stage != 'merged'
		ORDER BY i.updated_at_local DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("storage: list open worktree issues: %w", err)
	}
	return scanWorktreeIssues(rows)
}

func scanWorktreeIssues(rows *sql.Rows) ([]WorktreeIssue, error) {
	defer rows.Close()
	out := make([]WorktreeIssue, 0)
	for rows.Next() {
		var rec WorktreeIssue
		if err := rows.Scan(&rec.Number, &rec.Title, &rec.State, &rec.Stage, &rec.BranchName,
			&rec.PRNumber, &rec.PRState, &rec.PRReviewState, &rec.UpdatedAt,
			&rec.RunStage, &rec.RunStatus); err != nil {
			return nil, fmt.Errorf("storage: scan worktree issue: %w", err)
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// chunkNumbers splits issue numbers into deduped batches small enough to stay
// well under SQLite's bound-parameter limit.
func chunkNumbers(numbers []int64, size int) [][]int64 {
	seen := make(map[int64]bool, len(numbers))
	unique := make([]int64, 0, len(numbers))
	for _, n := range numbers {
		if n <= 0 || seen[n] {
			continue
		}
		seen[n] = true
		unique = append(unique, n)
	}
	if len(unique) == 0 {
		return nil
	}
	var out [][]int64
	for start := 0; start < len(unique); start += size {
		end := start + size
		if end > len(unique) {
			end = len(unique)
		}
		out = append(out, unique[start:end])
	}
	return out
}
