package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ButterHost69/sloper/internal/models"
)

type Repositories struct {
	db *sql.DB
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{db: db}
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// ─── Issue Repository ─────────────────────────────────────────────────

type IssueRecord struct {
	Number           int64
	Title            string
	State            string
	URL              string
	Author           string
	UpdatedAt        string
	Labels           []string
	IsPullRequest    bool
	Stage            string
	SpecJSON         string
	BranchName       string
	PRNumber         int64
	LastCommentID    int64
	ReviewIterations int
	CreatedAt        string
	FirstSeenAt      string
	UpdatedAtLocal   string
}

func (r *Repositories) GetIssue(ctx context.Context, number int64) (*IssueRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT number, title, state, url, author, updated_at, labels,
		       is_pull_request, stage, COALESCE(spec_json, ''), COALESCE(branch_name, ''),
		       COALESCE(pr_number, 0), last_comment_id, review_iterations,
		       created_at, first_seen_at, updated_at_local
		FROM issues WHERE number = ?
	`, number)

	var rec IssueRecord
	var labelsJSON string
	var isPR int
	var prNum int64

	err := row.Scan(
		&rec.Number, &rec.Title, &rec.State, &rec.URL, &rec.Author, &rec.UpdatedAt,
		&labelsJSON, &isPR, &rec.Stage, &rec.SpecJSON, &rec.BranchName,
		&prNum, &rec.LastCommentID, &rec.ReviewIterations,
		&rec.CreatedAt, &rec.FirstSeenAt, &rec.UpdatedAtLocal,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("storage: get issue %d: %w", number, err)
	}

	rec.IsPullRequest = isPR != 0
	rec.PRNumber = prNum
	_ = json.Unmarshal([]byte(labelsJSON), &rec.Labels)

	return &rec, nil
}

func (r *Repositories) UpsertIssue(ctx context.Context, rec IssueRecord) error {
	labelsJSON, _ := json.Marshal(rec.Labels)
	isPR := 0
	if rec.IsPullRequest {
		isPR = 1
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO issues (number, title, state, url, author, updated_at, labels,
		                    is_pull_request, stage, spec_json, branch_name, pr_number,
		                    last_comment_id, review_iterations, created_at, first_seen_at,
		                    updated_at_local)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(number) DO UPDATE SET
			title = excluded.title,
			state = excluded.state,
			url = excluded.url,
			author = excluded.author,
			updated_at = excluded.updated_at,
			labels = excluded.labels,
			is_pull_request = excluded.is_pull_request,
			updated_at_local = excluded.updated_at_local
	`,
		rec.Number, rec.Title, rec.State, rec.URL, rec.Author, rec.UpdatedAt, string(labelsJSON),
		isPR, rec.Stage, nullableString(rec.SpecJSON), nullableString(rec.BranchName),
		rec.PRNumber, rec.LastCommentID, rec.ReviewIterations, rec.CreatedAt, rec.FirstSeenAt,
		now(),
	)
	if err != nil {
		return fmt.Errorf("storage: upsert issue %d: %w", rec.Number, err)
	}
	return nil
}

func (r *Repositories) UpdateIssueStage(ctx context.Context, number int64, stage string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE issues SET stage = ?, updated_at_local = ? WHERE number = ?",
		stage, now(), number,
	)
	if err != nil {
		return fmt.Errorf("storage: update issue %d stage: %w", number, err)
	}
	return nil
}

func (r *Repositories) UpdateIssueSpec(ctx context.Context, number int64, specJSON string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE issues SET spec_json = ?, updated_at_local = ? WHERE number = ?",
		specJSON, now(), number,
	)
	if err != nil {
		return fmt.Errorf("storage: update issue %d spec: %w", number, err)
	}
	return nil
}

func (r *Repositories) UpdateIssuePR(ctx context.Context, number int64, prNumber int64, branchName string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE issues SET pr_number = ?, branch_name = ?, updated_at_local = ? WHERE number = ?",
		prNumber, branchName, now(), number,
	)
	if err != nil {
		return fmt.Errorf("storage: update issue %d PR: %w", number, err)
	}
	return nil
}

func (r *Repositories) UpdateIssueLastCommentID(ctx context.Context, number int64, commentID int64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE issues SET last_comment_id = ?, updated_at_local = ? WHERE number = ?",
		commentID, now(), number,
	)
	if err != nil {
		return fmt.Errorf("storage: update issue %d last comment: %w", number, err)
	}
	return nil
}

func (r *Repositories) IncrementReviewIterations(ctx context.Context, number int64) (int, error) {
	_, err := r.db.ExecContext(ctx,
		"UPDATE issues SET review_iterations = review_iterations + 1 WHERE number = ?",
		number,
	)
	if err != nil {
		return 0, fmt.Errorf("storage: increment review iterations %d: %w", number, err)
	}

	rec, err := r.GetIssue(ctx, number)
	if err != nil || rec == nil {
		return 0, err
	}
	return rec.ReviewIterations, nil
}

func (r *Repositories) GetIssuesByStage(ctx context.Context, stage string) ([]IssueRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT number, title, state, url, author, updated_at, labels,
		       is_pull_request, stage, COALESCE(spec_json, ''), COALESCE(branch_name, ''),
		       COALESCE(pr_number, 0), last_comment_id, review_iterations,
		       created_at, first_seen_at, updated_at_local
		FROM issues WHERE stage = ?
	`, stage)
	if err != nil {
		return nil, fmt.Errorf("storage: get issues by stage %s: %w", stage, err)
	}
	defer rows.Close()

	var out []IssueRecord
	for rows.Next() {
		var rec IssueRecord
		var labelsJSON string
		var isPR int
		var prNum int64

		if err := rows.Scan(
			&rec.Number, &rec.Title, &rec.State, &rec.URL, &rec.Author, &rec.UpdatedAt,
			&labelsJSON, &isPR, &rec.Stage, &rec.SpecJSON, &rec.BranchName,
			&prNum, &rec.LastCommentID, &rec.ReviewIterations,
			&rec.CreatedAt, &rec.FirstSeenAt, &rec.UpdatedAtLocal,
		); err != nil {
			return nil, fmt.Errorf("storage: scan issue: %w", err)
		}
		rec.IsPullRequest = isPR != 0
		rec.PRNumber = prNum
		_ = json.Unmarshal([]byte(labelsJSON), &rec.Labels)
		out = append(out, rec)
	}
	return out, rows.Err()
}

// ─── Issue Comment Repository ─────────────────────────────────────────

type CommentRecord struct {
	ID          int64
	IssueNumber int64
	Author      string
	Body        string
	CreatedAt   string
	Processed   bool
}

func (r *Repositories) InsertComment(ctx context.Context, c CommentRecord) error {
	processed := 0
	if c.Processed {
		processed = 1
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO issue_comments (id, issue_number, author, body, created_at, processed)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET body = excluded.body, created_at = excluded.created_at
	`, c.ID, c.IssueNumber, c.Author, c.Body, c.CreatedAt, processed)
	if err != nil {
		return fmt.Errorf("storage: insert comment %d: %w", c.ID, err)
	}
	return nil
}

func (r *Repositories) GetUnprocessedComments(ctx context.Context, issueNumber int64) ([]CommentRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, issue_number, author, body, created_at, processed
		FROM issue_comments WHERE issue_number = ? AND processed = 0
		ORDER BY created_at ASC
	`, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("storage: get unprocessed comments for issue %d: %w", issueNumber, err)
	}
	defer rows.Close()

	var out []CommentRecord
	for rows.Next() {
		var rec CommentRecord
		var processed int
		if err := rows.Scan(&rec.ID, &rec.IssueNumber, &rec.Author, &rec.Body,
			&rec.CreatedAt, &processed); err != nil {
			return nil, fmt.Errorf("storage: scan comment: %w", err)
		}
		rec.Processed = processed != 0
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *Repositories) MarkCommentProcessed(ctx context.Context, commentID int64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE issue_comments SET processed = 1 WHERE id = ?", commentID)
	if err != nil {
		return fmt.Errorf("storage: mark comment %d processed: %w", commentID, err)
	}
	return nil
}

func (r *Repositories) HasComment(ctx context.Context, commentID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM issue_comments WHERE id = ?)", commentID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("storage: check comment %d: %w", commentID, err)
	}
	return exists, nil
}

// ─── Pull Request Repository ──────────────────────────────────────────

type PRRecord struct {
	Number       int64
	IssueNumber  int64
	Title        string
	HeadSHA      string
	BaseSHA      string
	State        string
	URL          string
	UpdatedAt    string
	ReviewState  string
	LastReviewAt string
}

func (r *Repositories) UpsertPR(ctx context.Context, rec PRRecord) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO pull_requests (number, issue_number, title, head_sha, base_sha,
		                           state, url, updated_at, review_state, last_review_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(number) DO UPDATE SET
			issue_number = excluded.issue_number,
			title = excluded.title,
			head_sha = excluded.head_sha,
			base_sha = excluded.base_sha,
			state = excluded.state,
			url = excluded.url,
			updated_at = excluded.updated_at,
			review_state = excluded.review_state
	`, rec.Number, rec.IssueNumber, rec.Title, rec.HeadSHA, rec.BaseSHA,
		rec.State, rec.URL, rec.UpdatedAt, rec.ReviewState,
		nullableString(rec.LastReviewAt))
	if err != nil {
		return fmt.Errorf("storage: upsert pr %d: %w", rec.Number, err)
	}
	return nil
}

func (r *Repositories) GetPR(ctx context.Context, prNumber int64) (*PRRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT number, issue_number, title, head_sha, base_sha, state, url,
		       updated_at, review_state, COALESCE(last_review_at, '')
		FROM pull_requests WHERE number = ?
	`, prNumber)

	var rec PRRecord
	err := row.Scan(&rec.Number, &rec.IssueNumber, &rec.Title, &rec.HeadSHA,
		&rec.BaseSHA, &rec.State, &rec.URL, &rec.UpdatedAt, &rec.ReviewState,
		&rec.LastReviewAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("storage: get pr %d: %w", prNumber, err)
	}
	return &rec, nil
}

// ─── Run Repository ───────────────────────────────────────────────────

type RunRecord struct {
	ID             int64
	IssueNumber    int64
	Stage          string
	Status         string
	CheckpointJSON string
	AgentOutput    string
	AgentThinking  string
	ShellLog       string
	StartedAt      string
	EndedAt        string
	ErrorMessage   string
}

func (r *Repositories) StartRun(ctx context.Context, issueNumber int64, stage string) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO runs (issue_number, stage, status, started_at)
		VALUES (?, ?, 'running', ?)
	`, issueNumber, stage, now())
	if err != nil {
		return 0, fmt.Errorf("storage: start run for issue %d: %w", issueNumber, err)
	}
	id, _ := res.LastInsertId()
	return id, nil
}

func (r *Repositories) CompleteRun(ctx context.Context, runID int64, agentOutput, agentThinking, shellLog string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE runs SET status = 'completed', agent_output = ?, agent_thinking = ?,
		                shell_log = ?, ended_at = ?
		WHERE id = ?
	`, agentOutput, agentThinking, shellLog, now(), runID)
	if err != nil {
		return fmt.Errorf("storage: complete run %d: %w", runID, err)
	}
	return nil
}

func (r *Repositories) FailRun(ctx context.Context, runID int64, errMsg string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE runs SET status = 'failed', error_message = ?, ended_at = ?
		WHERE id = ?
	`, errMsg, now(), runID)
	if err != nil {
		return fmt.Errorf("storage: fail run %d: %w", runID, err)
	}
	return nil
}

func (r *Repositories) GetLatestRun(ctx context.Context, issueNumber int64) (*RunRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, issue_number, stage, status, COALESCE(checkpoint_json, ''),
		       COALESCE(agent_output, ''), COALESCE(agent_thinking, ''),
		       COALESCE(shell_log, ''), started_at, COALESCE(ended_at, ''),
		       COALESCE(error_message, '')
		FROM runs WHERE issue_number = ?
		ORDER BY id DESC LIMIT 1
	`, issueNumber)

	var rec RunRecord
	err := row.Scan(&rec.ID, &rec.IssueNumber, &rec.Stage, &rec.Status,
		&rec.CheckpointJSON, &rec.AgentOutput, &rec.AgentThinking,
		&rec.ShellLog, &rec.StartedAt, &rec.EndedAt, &rec.ErrorMessage)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("storage: get latest run for issue %d: %w", issueNumber, err)
	}
	return &rec, nil
}

func (r *Repositories) GetInterruptedRuns(ctx context.Context) ([]RunRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, issue_number, stage, status, COALESCE(checkpoint_json, ''),
		       COALESCE(agent_output, ''), COALESCE(agent_thinking, ''),
		       COALESCE(shell_log, ''), started_at, COALESCE(ended_at, ''),
		       COALESCE(error_message, '')
		FROM runs WHERE status = 'running'
	`)
	if err != nil {
		return nil, fmt.Errorf("storage: get interrupted runs: %w", err)
	}
	defer rows.Close()

	var out []RunRecord
	for rows.Next() {
		var rec RunRecord
		if err := rows.Scan(&rec.ID, &rec.IssueNumber, &rec.Stage, &rec.Status,
			&rec.CheckpointJSON, &rec.AgentOutput, &rec.AgentThinking,
			&rec.ShellLog, &rec.StartedAt, &rec.EndedAt, &rec.ErrorMessage); err != nil {
			return nil, fmt.Errorf("storage: scan run: %w", err)
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *Repositories) MarkRunsInterrupted(ctx context.Context) (int, error) {
	res, err := r.db.ExecContext(ctx,
		"UPDATE runs SET status = 'interrupted' WHERE status = 'running'")
	if err != nil {
		return 0, fmt.Errorf("storage: mark runs interrupted: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// ─── Event Log Repository ─────────────────────────────────────────────

type EventRecord struct {
	IssueNumber int64
	PRNumber    int64
	EventType   string
	Stage       string
	Message     string
	Context     map[string]any
}

func (r *Repositories) AppendEvent(ctx context.Context, evt EventRecord) error {
	ctxJSON, _ := json.Marshal(evt.Context)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO event_logs (issue_number, pr_number, event_type, stage, message, context_json)
		VALUES (?, ?, ?, ?, ?, ?)
	`, nullableInt64(evt.IssueNumber), nullableInt64(evt.PRNumber),
		evt.EventType, nullableString(evt.Stage), evt.Message, string(ctxJSON))
	if err != nil {
		return fmt.Errorf("storage: append event: %w", err)
	}
	return nil
}

// ─── Helpers ──────────────────────────────────────────────────────────

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullableInt64(n int64) any {
	if n == 0 {
		return nil
	}
	return n
}

// IssueRecordFromModel converts a models.IssueDetail + stage info into a DB record.
func IssueRecordFromModel(issue models.IssueDetail, stage string) IssueRecord {
	return IssueRecord{
		Number:        issue.Number,
		Title:         issue.Title,
		State:         issue.State,
		URL:           issue.URL,
		Author:        issue.Author,
		UpdatedAt:     issue.UpdatedAt,
		Labels:        issue.Labels,
		IsPullRequest: issue.IsPullRequest,
		Stage:         stage,
		CreatedAt:     issue.CreatedAt,
	}
}

// IssueRecordFromSummary converts a models.GithubIssueSummary into a DB record.
func IssueRecordFromSummary(summary models.GithubIssueSummary) IssueRecord {
	return IssueRecord{
		Number:        summary.Number,
		Title:         summary.Title,
		State:         summary.State,
		URL:           summary.URL,
		Author:        summary.Author,
		UpdatedAt:     summary.UpdatedAt,
		Labels:        summary.Labels,
		IsPullRequest: summary.IsPullRequest,
	}
}

// SpecJSON helper: marshal a SpecResult to JSON string.
func SpecJSON(spec *models.SpecResult) string {
	if spec == nil {
		return ""
	}
	b, err := json.Marshal(spec)
	if err != nil {
		return ""
	}
	return string(b)
}

// ParseSpecJSON unmarshals a spec JSON string back into a SpecResult.
func ParseSpecJSON(specJSON string) *models.SpecResult {
	if specJSON == "" {
		return nil
	}
	var spec models.SpecResult
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		return nil
	}
	return &spec
}
