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

