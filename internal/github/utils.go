package github

import (
	"strings"

	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/utils"
)

func defaultLimit(limit int) int {
	if limit <= 0 {
		return models.DefaultGithubIssuesLimit 
	}
	return limit
}

func issueListLabels(options models.GithubIssueOptions) []string {
	labels := options.Labels
	if len(labels) == 0 && strings.TrimSpace(options.Label) != "" {
		labels = []string{options.Label}
	}
	result := []string{}
	seen := map[string]struct{}{}
	for _, label := range labels {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		key := strings.ToLower(label)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, label)
	}
	return result
}

func extractAuthor(value any) string {
	row, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	if login := utils.AsString(row["login"]); login != "" {
		return login
	}
	return utils.AsString(row["name"])
}

func extractActorLogins(value any) []string {
	users := extractActorUsers(value)
	out := make([]string, 0, len(users))
	for _, user := range users {
		if user.Login != "" {
			out = append(out, user.Login)
		}
	}
	return out
}

func extractActorUsers(value any) []models.GitHubUser {
	items, ok := value.([]any)
	if !ok {
		return []models.GitHubUser{}
	}
	out := make([]models.GitHubUser, 0, len(items))
	for _, item := range items {
		if login := extractAuthor(item); login != "" {
			row, _ := item.(map[string]any)
			out = append(out, models.GitHubUser{Login: login, ID: utils.AsInt64(firstNonNil(row["databaseId"], row["id"]))})
		}
	}
	return out
}

func extractLabelNames(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return []string{}
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if name := utils.AsString(row["name"]); name != "" {
			out = append(out, name)
		}
	}
	return out
}

func firstNonNil(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

	