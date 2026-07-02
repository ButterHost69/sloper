package models

import (
	"context"
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
