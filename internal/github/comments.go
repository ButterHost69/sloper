package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/utils"
)

func (g *GithubGateway) PostIssueComment(ctx context.Context, repo string, issueNumber int64, body string) error {
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("github: post comment: body is empty")
	}
	_, repoPath := splitRepoHostname(repo)
	args := []string{"issue", "comment", fmt.Sprintf("%d", issueNumber),
		"--repo", repoPath, "--body", body}
	_, err := g.runGh(ctx, g.cwd, "", args...)
	return err
}

func (g *GithubGateway) AddIssueLabel(ctx context.Context, repo string, issueNumber int64, label string) error {
	_, repoPath := splitRepoHostname(repo)
	args := []string{"issue", "edit", fmt.Sprintf("%d", issueNumber),
		"--repo", repoPath, "--add-label", label}
	_, err := g.runGh(ctx, g.cwd, "", args...)
	return err
}

func (g *GithubGateway) RemoveIssueLabel(ctx context.Context, repo string, issueNumber int64, label string) error {
	_, repoPath := splitRepoHostname(repo)
	args := []string{"issue", "edit", fmt.Sprintf("%d", issueNumber),
		"--repo", repoPath, "--remove-label", label}
	_, err := g.runGh(ctx, g.cwd, "", args...)
	return err
}

func (g *GithubGateway) EditIssueBody(ctx context.Context, repo string, issueNumber int64, body string) error {
	_, repoPath := splitRepoHostname(repo)
	args := []string{"issue", "edit", fmt.Sprintf("%d", issueNumber),
		"--repo", repoPath, "--body", body}
	_, err := g.runGh(ctx, g.cwd, "", args...)
	return err
}

func (g *GithubGateway) GetIssueComments(ctx context.Context, repo string, issueNumber int64) ([]models.CommentInfo, error) {
	hostname, repoPath := splitRepoHostname(repo)
	args := []string{"api", "--paginate", "--slurp",
		fmt.Sprintf("repos/%s/issues/%d/comments", repoPath, issueNumber)}
	if hostname != "" {
		args = append(args, "--hostname", hostname)
	}
	result, err := g.runGh(ctx, g.cwd, "", args...)
	if err != nil {
		return nil, err
	}
	rows, err := utils.DecodeJSONArrayOrPages(result.Stdout)
	if err != nil {
		return nil, err
	}
	return extractCommentInfos(rows), nil
}

func (g *GithubGateway) CloseIssue(ctx context.Context, repo string, issueNumber int64) error {
	_, repoPath := splitRepoHostname(repo)
	args := []string{"issue", "close", fmt.Sprintf("%d", issueNumber), "--repo", repoPath}
	_, err := g.runGh(ctx, g.cwd, "", args...)
	return err
}
