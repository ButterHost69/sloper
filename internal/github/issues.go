package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/utils"
)

func (g *GithubGateway) GetAllOpenIssuesRaw(ctx context.Context, options models.GithubIssueOptions) ([]models.GithubIssueSummary, error) {
	args := []string{"issue", "list", "--repo", options.Repo, "--state", "open", "--limit", fmt.Sprintf("%d", defaultLimit(options.Limit))}
	if strings.TrimSpace(options.Assignee) != "" {
		args = append(args, "--assignee", options.Assignee)
	}
	for _, label := range issueListLabels(options) {
		args = append(args, "--label", label)
	}

	args = append(args, "--json", strings.Join([]string{"number", "title", "body", "url", "state", "updatedAt", "author", "assignees", "labels"}, ","))

	result, err := g.runGh(ctx, options.CWD, "", args...)
	if err != nil {
		return nil, err
	}
	rows, err := utils.DecodeJSONArray(result.Stdout)
	if err != nil {
		return nil, err
	}
	out := make([]models.GithubIssueSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.GithubIssueSummary{
			Number:            utils.AsInt64(row["number"]),
			Title:             utils.AsString(row["title"]),
			Body:              utils.AsString(row["body"]),
			URL:               utils.AsString(row["url"]),
			State:             utils.AsString(row["state"]),
			UpdatedAt:         utils.AsString(row["updatedAt"]),
			Author:            extractAuthor(row["author"]),
			AuthorAssociation: utils.AsString(row["authorAssociation"]),
			Assignees:         extractActorLogins(row["assignees"]),
			AssigneeUsers:     extractActorUsers(row["assignees"]),
			Labels:            extractLabelNames(row["labels"]),
		})
	}
	return out, nil

}

func (g *GithubGateway) ViewIssue(ctx context.Context, input models.ViewIssueInput) (models.IssueDetail, error) {
	hostname, repo := splitRepoHostname(input.Repo)
	args := []string{"api", fmt.Sprintf("repos/%s/issues/%d", repo, input.IssueNumber)}
	if hostname != "" {
		args = append(args, "--hostname", hostname)
	}
	result, err := g.runGh(ctx, input.CWD, "", args...)
	if err != nil {
		return models.IssueDetail{}, err
	}
	row, err := utils.DecodeJSONObject(result.Stdout)
	if err != nil {
		return models.IssueDetail{}, err
	}
	commentArgs := []string{"api", "--paginate", "--slurp", fmt.Sprintf("repos/%s/issues/%d/comments", repo, input.IssueNumber)}
	if hostname != "" {
		commentArgs = append(commentArgs, "--hostname", hostname)
	}
	commentsResult, err := g.runGh(ctx, input.CWD, "", commentArgs...)
	if err != nil {
		return models.IssueDetail{}, err
	}
	commentRows, err := utils.DecodeJSONArrayOrPages(commentsResult.Stdout)
	if err != nil {
		return models.IssueDetail{}, err
	}
	return models.IssueDetail{
		Number:            utils.AsInt64(row["number"]),
		Title:             utils.AsString(row["title"]),
		Body:              utils.AsString(row["body"]),
		URL:               utils.FirstNonEmpty(utils.AsString(row["html_url"]), utils.AsString(row["url"])),
		State:             utils.AsString(row["state"]),
		StateReason:       utils.FirstNonEmpty(utils.AsString(row["state_reason"]), utils.AsString(row["stateReason"])),
		CreatedAt:         utils.FirstNonEmpty(utils.AsString(row["created_at"]), utils.AsString(row["createdAt"])),
		UpdatedAt:         utils.FirstNonEmpty(utils.AsString(row["updated_at"]), utils.AsString(row["updatedAt"])),
		ClosedAt:          utils.FirstNonEmpty(utils.AsString(row["closed_at"]), utils.AsString(row["closedAt"])),
		Author:            extractAuthor(firstNonNil(row["user"], row["author"])),
		AuthorAssociation: utils.AsString(row["author_association"]),
		Assignees:         extractActorLogins(row["assignees"]),
		AssigneeUsers:     extractActorUsers(row["assignees"]),
		Labels:            extractLabelNames(row["labels"]),
		IsPullRequest:     row["pull_request"] != nil,
		CommentCount:      len(commentRows),
		Comments:          extractCommentInfos(commentRows),
	}, nil
}