package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ─── List / read helpers used by the dashboard API ─────────────────────

const issueColumns = `number, title, state, url, author, updated_at, labels,
	is_pull_request, stage, COALESCE(spec_json, ''), COALESCE(branch_name, ''),
	COALESCE(pr_number, 0), last_comment_id, review_iterations,
	created_at, first_seen_at, updated_at_local`

const runColumns = `id, issue_number, stage, status, COALESCE(checkpoint_json, ''),
	COALESCE(agent_output, ''), COALESCE(agent_thinking, ''),
	COALESCE(shell_log, ''), started_at, COALESCE(ended_at, ''),
	COALESCE(error_message, '')`

func scanIssue(sc interface{ Scan(...any) error }) (*IssueRecord, error) {
	var rec IssueRecord
	var labelsJSON string
	var isPR int
	var prNum int64

	err := sc.Scan(
		&rec.Number, &rec.Title, &rec.State, &rec.URL, &rec.Author, &rec.UpdatedAt,
		&labelsJSON, &isPR, &rec.Stage, &rec.SpecJSON, &rec.BranchName,
		&prNum, &rec.LastCommentID, &rec.ReviewIterations,
		&rec.CreatedAt, &rec.FirstSeenAt, &rec.UpdatedAtLocal,
	)
	if err != nil {
		return nil, err
	}
	rec.IsPullRequest = isPR != 0
	rec.PRNumber = prNum
	_ = json.Unmarshal([]byte(labelsJSON), &rec.Labels)
	return &rec, nil
}

func scanRun(sc interface{ Scan(...any) error }) (*RunRecord, error) {
	var rec RunRecord
	if err := sc.Scan(&rec.ID, &rec.IssueNumber, &rec.Stage, &rec.Status,
		&rec.CheckpointJSON, &rec.AgentOutput, &rec.AgentThinking,
		&rec.ShellLog, &rec.StartedAt, &rec.EndedAt, &rec.ErrorMessage); err != nil {
		return nil, err
	}
	return &rec, nil
}

// ListIssues returns every cached issue, newest activity first.
func (r *Repositories) ListIssues(ctx context.Context, limit, offset int) ([]IssueRecord, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+issueColumns+`
		FROM issues
		ORDER BY updated_at_local DESC, number DESC
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("storage: list issues: %w", err)
	}
	defer rows.Close()

	out := make([]IssueRecord, 0)
	for rows.Next() {
		rec, err := scanIssue(rows)
		if err != nil {
			return nil, fmt.Errorf("storage: scan issue: %w", err)
		}
		out = append(out, *rec)
	}
	return out, rows.Err()
}

// ListIssuesFiltered returns issues optionally narrowed by stage and a free-text
// query. An empty stage ("all") and an empty query behave like ListIssues.
func (r *Repositories) ListIssuesFiltered(ctx context.Context, limit, offset int, stage, q string) ([]IssueRecord, error) {
	if limit <= 0 {
		limit = 500
	}
	where := ""
	args := []any{}
	if stage != "" && stage != "all" {
		where = " WHERE stage = ?"
		args = append(args, stage)
	}
	if q != "" {
		like := "%" + q + "%"
		if where == "" {
			where = " WHERE "
		} else {
			where += " AND "
		}
		where += "(title LIKE ? OR author LIKE ? OR CAST(number AS TEXT) LIKE ? OR labels LIKE ?)"
		args = append(args, like, like, like, like)
	}
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, `
		SELECT `+issueColumns+`
		FROM issues`+where+`
		ORDER BY updated_at_local DESC, number DESC
		LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, fmt.Errorf("storage: list filtered issues: %w", err)
	}
	defer rows.Close()

	out := make([]IssueRecord, 0)
	for rows.Next() {
		rec, err := scanIssue(rows)
		if err != nil {
			return nil, fmt.Errorf("storage: scan issue: %w", err)
		}
		out = append(out, *rec)
	}
	return out, rows.Err()
}

func (r *Repositories) ListIssuesByStage(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT stage, COUNT(*) FROM issues GROUP BY stage`)
	if err != nil {
		return nil, fmt.Errorf("storage: count issues by stage: %w", err)
	}
	defer rows.Close()

	out := make(map[string]int)
	for rows.Next() {
		var stage string
		var count int
		if err := rows.Scan(&stage, &count); err != nil {
			return nil, fmt.Errorf("storage: scan stage count: %w", err)
		}
		out[stage] = count
	}
	return out, rows.Err()
}

func (r *Repositories) ListIssuesByState(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT state, COUNT(*) FROM issues GROUP BY state`)
	if err != nil {
		return nil, fmt.Errorf("storage: count issues by state: %w", err)
	}
	defer rows.Close()

	out := make(map[string]int)
	for rows.Next() {
		var state string
		var count int
		if err := rows.Scan(&state, &count); err != nil {
			return nil, fmt.Errorf("storage: scan state count: %w", err)
		}
		out[state] = count
	}
	return out, rows.Err()
}

// ListCommentsByIssue returns all cached comments for an issue, oldest first.
func (r *Repositories) ListCommentsByIssue(ctx context.Context, issueNumber int64) ([]CommentRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, issue_number, author, body, created_at, processed, in_reply_to_id, replied_by_bot
		FROM issue_comments WHERE issue_number = ?
		ORDER BY created_at ASC`, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("storage: list comments for issue %d: %w", issueNumber, err)
	}
	defer rows.Close()

	out := make([]CommentRecord, 0)
	for rows.Next() {
		var rec CommentRecord
		var processed int
		var repliedByBot int
		if err := rows.Scan(&rec.ID, &rec.IssueNumber, &rec.Author, &rec.Body,
			&rec.CreatedAt, &processed, &rec.InReplyToID, &repliedByBot); err != nil {
			return nil, fmt.Errorf("storage: scan comment: %w", err)
		}
		rec.Processed = processed != 0
		rec.RepliedByBot = repliedByBot != 0
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *Repositories) ListPRs(ctx context.Context, limit int) ([]PRRecord, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT number, issue_number, title, head_sha, base_sha, state, url,
		       updated_at, COALESCE(merged_at, ''), review_state, COALESCE(last_review_at, '')
		FROM pull_requests
		ORDER BY number DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("storage: list prs: %w", err)
	}
	defer rows.Close()

	out := make([]PRRecord, 0)
	for rows.Next() {
		var rec PRRecord
		if err := rows.Scan(&rec.Number, &rec.IssueNumber, &rec.Title, &rec.HeadSHA,
			&rec.BaseSHA, &rec.State, &rec.URL, &rec.UpdatedAt, &rec.MergedAt,
			&rec.ReviewState, &rec.LastReviewAt); err != nil {
			return nil, fmt.Errorf("storage: scan pr: %w", err)
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// CountPRsByState returns state -> count for all cached pull requests.
// Merged PRs (closed on GitHub with a merged_at) are counted under "merged".
func (r *Repositories) CountPRsByState(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT CASE WHEN state = 'closed' AND merged_at != '' THEN 'merged' ELSE state END,
		       COUNT(*)
		FROM pull_requests GROUP BY 1`)
	if err != nil {
		return nil, fmt.Errorf("storage: count prs by state: %w", err)
	}
	defer rows.Close()

	out := make(map[string]int)
	for rows.Next() {
		var state string
		var count int
		if err := rows.Scan(&state, &count); err != nil {
			return nil, fmt.Errorf("storage: scan pr state count: %w", err)
		}
		out[state] = count
	}
	return out, rows.Err()
}

func (r *Repositories) ListRuns(ctx context.Context, limit, offset int) ([]RunRecord, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+runColumns+`
		FROM runs
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("storage: list runs: %w", err)
	}
	defer rows.Close()

	out := make([]RunRecord, 0)
	for rows.Next() {
		rec, err := scanRun(rows)
		if err != nil {
			return nil, fmt.Errorf("storage: scan run: %w", err)
		}
		out = append(out, *rec)
	}
	return out, rows.Err()
}

func (r *Repositories) ListRunsByIssue(ctx context.Context, issueNumber int64) ([]RunRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+runColumns+`
		FROM runs WHERE issue_number = ?
		ORDER BY id ASC`, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("storage: list runs for issue %d: %w", issueNumber, err)
	}
	defer rows.Close()

	out := make([]RunRecord, 0)
	for rows.Next() {
		rec, err := scanRun(rows)
		if err != nil {
			return nil, fmt.Errorf("storage: scan run: %w", err)
		}
		out = append(out, *rec)
	}
	return out, rows.Err()
}

func (r *Repositories) CountRunsByStatus(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT status, COUNT(*) FROM runs GROUP BY status`)
	if err != nil {
		return nil, fmt.Errorf("storage: count runs by status: %w", err)
	}
	defer rows.Close()

	out := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("storage: scan run status count: %w", err)
		}
		out[status] = count
	}
	return out, rows.Err()
}

type EventRecordFull struct {
	ID          int64          `json:"id"`
	IssueNumber int64          `json:"issue_number"`
	PRNumber    int64          `json:"pr_number"`
	EventType   string         `json:"event_type"`
	Stage       string         `json:"stage"`
	Message     string         `json:"message"`
	Context     map[string]any `json:"context"`
	CreatedAt   string         `json:"created_at"`
}

// CountEvents returns the total number of events in the log.
func (r *Repositories) CountEvents(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM event_logs`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("storage: count events: %w", err)
	}
	return n, nil
}

func (r *Repositories) ListEvents(ctx context.Context, limit, offset int) ([]EventRecordFull, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(issue_number, 0), COALESCE(pr_number, 0), event_type,
		       COALESCE(stage, ''), message, context_json, created_at
		FROM event_logs
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("storage: list events: %w", err)
	}
	defer rows.Close()

	out := make([]EventRecordFull, 0)
	for rows.Next() {
		var rec EventRecordFull
		var ctxJSON string
		if err := rows.Scan(&rec.ID, &rec.IssueNumber, &rec.PRNumber, &rec.EventType,
			&rec.Stage, &rec.Message, &ctxJSON, &rec.CreatedAt); err != nil {
			return nil, fmt.Errorf("storage: scan event: %w", err)
		}
		rec.Context = map[string]any{}
		_ = json.Unmarshal([]byte(ctxJSON), &rec.Context)
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *Repositories) ListEventsByIssue(ctx context.Context, issueNumber int64) ([]EventRecordFull, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(issue_number, 0), COALESCE(pr_number, 0), event_type,
		       COALESCE(stage, ''), message, context_json, created_at
		FROM event_logs WHERE issue_number = ?
		ORDER BY id ASC`, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("storage: list events for issue %d: %w", issueNumber, err)
	}
	defer rows.Close()

	out := make([]EventRecordFull, 0)
	for rows.Next() {
		var rec EventRecordFull
		var ctxJSON string
		if err := rows.Scan(&rec.ID, &rec.IssueNumber, &rec.PRNumber, &rec.EventType,
			&rec.Stage, &rec.Message, &ctxJSON, &rec.CreatedAt); err != nil {
			return nil, fmt.Errorf("storage: scan event: %w", err)
		}
		rec.Context = map[string]any{}
		_ = json.Unmarshal([]byte(ctxJSON), &rec.Context)
		out = append(out, rec)
	}
	return out, rows.Err()
}

// CountEventsByType returns event_type -> count for the last N events.
func (r *Repositories) CountEventsByType(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT event_type, COUNT(*) FROM event_logs GROUP BY event_type ORDER BY 2 DESC`)
	if err != nil {
		return nil, fmt.Errorf("storage: count events by type: %w", err)
	}
	defer rows.Close()

	out := make(map[string]int)
	for rows.Next() {
		var typ string
		var count int
		if err := rows.Scan(&typ, &count); err != nil {
			return nil, fmt.Errorf("storage: scan event type count: %w", err)
		}
		out[typ] = count
	}
	return out, rows.Err()
}

// ActivityByHour buckets events into the last `hours` hourly buckets, newest last.
// Missing buckets are filled with zeroes so charts render continuously.
func (r *Repositories) ActivityByHour(ctx context.Context, hours int) ([]ActivityBucket, error) {
	if hours <= 0 {
		hours = 24
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT strftime('%Y-%m-%dT%H:00:00Z', created_at) AS bucket, COUNT(*)
		FROM event_logs
		WHERE created_at >= datetime('now', ?)
		GROUP BY bucket
		ORDER BY bucket ASC`, fmt.Sprintf("-%d hours", hours))
	if err != nil {
		return nil, fmt.Errorf("storage: activity by hour: %w", err)
	}
	defer rows.Close()

	got := map[string]int{}
	for rows.Next() {
		var bucket string
		var count int
		if err := rows.Scan(&bucket, &count); err != nil {
			return nil, fmt.Errorf("storage: scan activity bucket: %w", err)
		}
		got[bucket] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]ActivityBucket, 0, hours)
	now := time.Now().UTC().Truncate(time.Hour)
	for i := hours - 1; i >= 0; i-- {
		b := now.Add(-time.Duration(i) * time.Hour)
		key := b.Format("2006-01-02T15:00:00Z")
		out = append(out, ActivityBucket{Time: key, Count: got[key]})
	}
	return out, nil
}

// DBSize returns the size of the underlying sqlite file in bytes.
func (r *Repositories) DBSize(ctx context.Context) (int64, error) {
	var size int64
	err := r.db.QueryRowContext(ctx, `PRAGMA page_count`).Scan(&size)
	if err != nil {
		return 0, err
	}
	var pageSize int64
	if err := r.db.QueryRowContext(ctx, `PRAGMA page_size`).Scan(&pageSize); err != nil {
		return 0, err
	}
	return size * pageSize, nil
}

// ActivityBucket is a single point on the activity timeline.
type ActivityBucket struct {
	Time  string `json:"time"`
	Count int    `json:"count"`
}
