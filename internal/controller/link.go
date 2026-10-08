package controller

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrInvalidLink marks a repo link the controller cannot turn into owner/repo.
var ErrInvalidLink = errors.New("controller: invalid repo link")

var (
	ownerPattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9._-]{0,38})$`)
	repoPattern  = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9._-]{0,99})$`)
)

// ParseRepoLink accepts the forms an operator might paste and returns the
// canonical `owner/repo` plus its https clone URL:
//
//	owner/repo
//	github.com/owner/repo
//	https://github.com/owner/repo(.git)
//	git@github.com:owner/repo.git
//	https://github.com/owner/repo/tree/main   (extra path is ignored)
//
// Only GitHub is supported because every other part of sloper (gh CLI, issues,
// PRs) is GitHub-specific.
func ParseRepoLink(raw string) (link string, cloneURL string, err error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", "", fmt.Errorf("%w: empty link", ErrInvalidLink)
	}

	value = strings.TrimSuffix(value, "/")
	value = strings.TrimSuffix(value, ".git")

	switch {
	case strings.HasPrefix(value, "git@"):
		// git@github.com:owner/repo
		rest := strings.TrimPrefix(value, "git@")
		host, path, ok := strings.Cut(rest, ":")
		if !ok {
			return "", "", fmt.Errorf("%w: %q", ErrInvalidLink, raw)
		}
		if !isGitHubHost(host) {
			return "", "", fmt.Errorf("%w: only github.com is supported, got %q", ErrInvalidLink, host)
		}
		value = path
	case strings.Contains(value, "://"):
		scheme, rest, _ := strings.Cut(value, "://")
		if scheme != "http" && scheme != "https" && scheme != "ssh" {
			return "", "", fmt.Errorf("%w: unsupported scheme %q", ErrInvalidLink, scheme)
		}
		host, path, ok := strings.Cut(rest, "/")
		if !ok {
			return "", "", fmt.Errorf("%w: %q", ErrInvalidLink, raw)
		}
		host = stripUserInfo(host)
		if !isGitHubHost(host) {
			return "", "", fmt.Errorf("%w: only github.com is supported, got %q", ErrInvalidLink, host)
		}
		value = path
	default:
		// github.com/owner/repo without a scheme.
		if host, path, ok := strings.Cut(value, "/"); ok && isGitHubHost(host) {
			value = path
		}
	}

	parts := strings.Split(strings.Trim(value, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("%w: expected owner/repo, got %q", ErrInvalidLink, raw)
	}
	owner, repo := parts[0], parts[1]
	if !ownerPattern.MatchString(owner) {
		return "", "", fmt.Errorf("%w: invalid owner %q", ErrInvalidLink, owner)
	}
	if !repoPattern.MatchString(repo) {
		return "", "", fmt.Errorf("%w: invalid repo %q", ErrInvalidLink, repo)
	}

	link = owner + "/" + repo
	return link, "https://github.com/" + link + ".git", nil
}

func isGitHubHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "github.com" || host == "www.github.com"
}

func stripUserInfo(host string) string {
	if i := strings.LastIndex(host, "@"); i >= 0 {
		return host[i+1:]
	}
	return host
}

// RepoName returns the repo part of an owner/repo link.
func RepoName(link string) string {
	if _, repo, ok := strings.Cut(link, "/"); ok {
		return repo
	}
	return link
}
