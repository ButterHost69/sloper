package models

import (
	"context"
	"encoding/json"
	"time"
)

const (
	DefaultGithubIssuesLimit = 30
	DefaultGhCommandTimeout  = 60 * time.Second
)

type GithubOptions struct {
	GHPath            string
	CWD               string
	Now               func() time.Time
	DiscoveryCacheTTL time.Duration
	GHRun             func(context.Context, ShellOptions) (ShellResult, error)
	// ReviewSubmitDiagnostic func(event string, fields map[string]any)
}

type GitHubUser struct {
	Login string
	ID    int64
}

type GithubIssueOptions struct {
	Repo     string
	CWD      string
	Limit    int
	Assignee string
	Label    string
	Labels   []string
}

type GithubIssueSummary struct {
	Number            int64
	Title             string
	Body              string
	URL               string
	State             string
	UpdatedAt         string
	Author            string
	AuthorAssociation string
	Assignees         []string
	AssigneeUsers     []GitHubUser
	Labels            []string
	IsPullRequest     bool
}

type ViewIssueInput struct {
	Repo        string
	IssueNumber int64
	CWD         string
}

type IssueDetail struct {
	Number            int64
	Title             string
	Body              string
	URL               string
	State             string
	StateReason       string
	CreatedAt         string
	UpdatedAt         string
	ClosedAt          string
	Author            string
	AuthorAssociation string
	Assignees         []string
	AssigneeUsers     []GitHubUser
	Labels            []string
	IsPullRequest     bool
	CommentCount      int
	Comments          []CommentInfo
}

func (d IssueDetail) String() string {
	m := make(map[string]any)

	if d.Number != 0 {
		m["number"] = d.Number
	}
	if d.Title != "" {
		m["title"] = d.Title
	}
	if d.Body != "" {
		m["body"] = d.Body
	}
	if d.URL != "" {
		m["url"] = d.URL
	}
	if d.State != "" {
		m["state"] = d.State
	}
	if d.StateReason != "" {
		m["state_reason"] = d.StateReason
	}
	if d.CreatedAt != "" {
		m["created_at"] = d.CreatedAt
	}
	if d.UpdatedAt != "" {
		m["updated_at"] = d.UpdatedAt
	}
	if d.ClosedAt != "" {
		m["closed_at"] = d.ClosedAt
	}
	if d.Author != "" {
		m["author"] = d.Author
	}
	if d.AuthorAssociation != "" {
		m["author_association"] = d.AuthorAssociation
	}
	if len(d.Assignees) > 0 {
		m["assignees"] = d.Assignees
	}
	if len(d.AssigneeUsers) > 0 {
		m["assignee_users"] = d.AssigneeUsers
	}
	if len(d.Labels) > 0 {
		m["labels"] = d.Labels
	}
	if d.IsPullRequest {
		m["is_pull_request"] = true
	}
	if d.CommentCount > 0 || len(d.Comments) > 0 {
		m["comment_count"] = d.CommentCount
	}
	if len(d.Comments) > 0 {
		m["comments"] = d.Comments
	}

	b, _ := json.Marshal(m)
	return string(b)
}

type CommentInfo struct {
	ID                int64
	Author            string
	AuthorAssociation string
	Body              string
	CreatedAt         string
	UpdatedAt         string
	URL               string
}
