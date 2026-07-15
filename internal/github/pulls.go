package github

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/utils"
)

type PRInfo struct {
	Number      int64
	Title       string
	Body        string
	State       string
	URL         string
	HeadRef     string
	HeadSHA     string
	BaseRef     string
	BaseSHA     string
	IsDraft     bool
	Author      string
	UpdatedAt   string
	Mergeable   string
	ReviewState string
}

type ReviewInfo struct {
	ID          int64
	Author      string
	State       string
	Body        string
	SubmittedAt string
}

func (g *GithubGateway) CreatePR(ctx context.Context, repo, title, body, head, base string) (int64, error) {
	_, repoPath := splitRepoHostname(repo)
	args := []string{"pr", "create", "--repo", repoPath,
		"--title", title, "--body", body,
		"--head", head, "--base", base}
	result, err := g.runGh(ctx, g.cwd, "", args...)
	if err != nil {
		return 0, fmt.Errorf("github: create pr: %w", err)
	}

	prURL := strings.TrimSpace(result.Stdout)
	prNumber := parsePRNumberFromURL(prURL)
	if prNumber == 0 {
		return 0, fmt.Errorf("github: create pr: could not parse PR number from URL: %s", prURL)
	}
	return prNumber, nil
}

func (g *GithubGateway) GetPR(ctx context.Context, repo string, prNumber int64) (*PRInfo, error) {
	hostname, repoPath := splitRepoHostname(repo)
	args := []string{"api", fmt.Sprintf("repos/%s/pulls/%d", repoPath, prNumber)}
	if hostname != "" {
		args = append(args, "--hostname", hostname)
	}
	result, err := g.runGh(ctx, g.cwd, "", args...)
	if err != nil {
		return nil, fmt.Errorf("github: get pr %d: %w", prNumber, err)
	}
	row, err := utils.DecodeJSONObject(result.Stdout)
	if err != nil {
		return nil, fmt.Errorf("github: decode pr %d: %w", prNumber, err)
	}

	head, _ := row["head"].(map[string]any)
	base, _ := row["base"].(map[string]any)

	return &PRInfo{
		Number:    utils.AsInt64(row["number"]),
		Title:     utils.AsString(row["title"]),
		Body:      utils.AsString(row["body"]),
		State:     utils.AsString(row["state"]),
		URL:       utils.FirstNonEmpty(utils.AsString(row["html_url"]), utils.AsString(row["url"])),
		HeadRef:   utils.AsString(firstNonNilMap(head, "ref")),
		HeadSHA:   utils.AsString(firstNonNilMap(head, "sha")),
		BaseRef:   utils.AsString(firstNonNilMap(base, "ref")),
		BaseSHA:   utils.AsString(firstNonNilMap(base, "sha")),
		IsDraft:   row["draft"] != nil && row["draft"] == true,
		Author:    extractAuthor(firstNonNil(row["user"], row["author"])),
		UpdatedAt: utils.FirstNonEmpty(utils.AsString(row["updated_at"]), utils.AsString(row["updatedAt"])),
	}, nil
}

func (g *GithubGateway) GetPRDiff(ctx context.Context, repo string, prNumber int64) (string, error) {
	_, repoPath := splitRepoHostname(repo)
	args := []string{"pr", "diff", fmt.Sprintf("%d", prNumber), "--repo", repoPath}
	result, err := g.runGhWithTimeout(ctx, g.cwd, "", 120*1000*1000*1000, args...)
	if err != nil {
		return "", fmt.Errorf("github: get pr diff %d: %w", prNumber, err)
	}
	return result.Stdout, nil
}

func (g *GithubGateway) GetPRComments(ctx context.Context, repo string, prNumber int64) ([]models.CommentInfo, error) {
	hostname, repoPath := splitRepoHostname(repo)
	args := []string{"api", "--paginate", "--slurp",
		fmt.Sprintf("repos/%s/issues/%d/comments", repoPath, prNumber)}
	if hostname != "" {
		args = append(args, "--hostname", hostname)
	}
	result, err := g.runGh(ctx, g.cwd, "", args...)
	if err != nil {
		return nil, fmt.Errorf("github: get pr comments %d: %w", prNumber, err)
	}
	rows, err := utils.DecodeJSONArrayOrPages(result.Stdout)
	if err != nil {
		return nil, fmt.Errorf("github: decode pr comments %d: %w", prNumber, err)
	}
	return extractCommentInfos(rows), nil
}

func (g *GithubGateway) GetPRReviews(ctx context.Context, repo string, prNumber int64) ([]ReviewInfo, error) {
	hostname, repoPath := splitRepoHostname(repo)
	args := []string{"api", "--paginate", "--slurp",
		fmt.Sprintf("repos/%s/pulls/%d/reviews", repoPath, prNumber)}
	if hostname != "" {
		args = append(args, "--hostname", hostname)
	}
	result, err := g.runGh(ctx, g.cwd, "", args...)
	if err != nil {
		return nil, fmt.Errorf("github: get pr reviews %d: %w", prNumber, err)
	}
	rows, err := utils.DecodeJSONArrayOrPages(result.Stdout)
	if err != nil {
		return nil, fmt.Errorf("github: decode pr reviews %d: %w", prNumber, err)
	}

	out := make([]ReviewInfo, 0, len(rows))
	for _, row := range rows {
		out = append(out, ReviewInfo{
			ID:          utils.AsInt64(firstNonNil(row["id"], row["databaseId"])),
			Author:      extractAuthor(firstNonNil(row["user"], row["author"])),
			State:       utils.AsString(row["state"]),
			Body:        utils.AsString(row["body"]),
			SubmittedAt: utils.FirstNonEmpty(utils.AsString(row["submitted_at"]), utils.AsString(row["submittedAt"])),
		})
	}
	return out, nil
}

func (g *GithubGateway) PostPRComment(ctx context.Context, repo string, prNumber int64, body string) error {
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("github: post pr comment: body is empty")
	}
	_, repoPath := splitRepoHostname(repo)
	args := []string{"pr", "comment", fmt.Sprintf("%d", prNumber),
		"--repo", repoPath, "--body", body}
	_, err := g.runGh(ctx, g.cwd, "", args...)
	return err
}

func (g *GithubGateway) MergePR(ctx context.Context, repo string, prNumber int64, method string) error {
	if method == "" {
		method = "squash"
	}
	_, repoPath := splitRepoHostname(repo)
	args := []string{"pr", "merge", fmt.Sprintf("%d", prNumber),
		"--repo", repoPath, "--" + method, "--delete-branch"}
	_, err := g.runGh(ctx, g.cwd, "", args...)
	return err
}

func (g *GithubGateway) GetPRForBranch(ctx context.Context, repo, branch string) (*PRInfo, error) {
	_, repoPath := splitRepoHostname(repo)
	args := []string{"pr", "list", "--repo", repoPath,
		"--head", branch, "--state", "open", "--json", "number,title,state,url,headRefName,baseRefName"}
	result, err := g.runGh(ctx, g.cwd, "", args...)
	if err != nil {
		return nil, fmt.Errorf("github: get pr for branch %s: %w", branch, err)
	}
	rows, err := utils.DecodeJSONArray(result.Stdout)
	if err != nil {
		return nil, fmt.Errorf("github: decode pr list: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	row := rows[0]
	return &PRInfo{
		Number:  utils.AsInt64(row["number"]),
		Title:   utils.AsString(row["title"]),
		State:   utils.AsString(row["state"]),
		URL:     utils.AsString(row["url"]),
		HeadRef: utils.AsString(row["headRefName"]),
		BaseRef: utils.AsString(row["baseRefName"]),
	}, nil
}

func parsePRNumberFromURL(url string) int64 {
	parts := strings.Split(strings.TrimRight(url, "/"), "/")
	for i, p := range parts {
		if p == "pull" && i+1 < len(parts) {
			n, err := strconv.ParseInt(parts[i+1], 10, 64)
			if err == nil {
				return n
			}
		}
	}
	return 0
}

func firstNonNilMap(m map[string]any, key string) any {
	if m == nil {
		return nil
	}
	return m[key]
}
