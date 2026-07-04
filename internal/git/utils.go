package git

import "regexp"

func parseGitHubRepoFromRemoteURL(remoteURL string) string {
	if remoteURL == "" {
		return ""
	}

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^git@github\.com:(?P<repo>.+?)(?:\.git)?$`),
		regexp.MustCompile(`^ssh://git@github\.com/(?P<repo>.+?)(?:\.git)?$`),
		regexp.MustCompile(`^https://github\.com/(?P<repo>.+?)(?:\.git)?$`),
	}

	for _, pattern := range patterns {
		match := pattern.FindStringSubmatch(remoteURL)
		if match == nil {
			continue
		}
		index := pattern.SubexpIndex("repo")
		if index > 0 {
			return match[index]
		}
	}

	return ""
}