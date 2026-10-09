package controller

import (
	"errors"
	"testing"
)

func TestParseRepoLink(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		wantLink  string
		wantClone string
		wantErr   error
	}{
		{name: "owner slash repo", input: "ButterHost69/sloper", wantLink: "ButterHost69/sloper", wantClone: "https://github.com/ButterHost69/sloper.git"},
		{name: "trimmed spaces", input: "  ButterHost69/sloper  ", wantLink: "ButterHost69/sloper", wantClone: "https://github.com/ButterHost69/sloper.git"},
		{name: "https url", input: "https://github.com/ButterHost69/sloper", wantLink: "ButterHost69/sloper", wantClone: "https://github.com/ButterHost69/sloper.git"},
		{name: "https url with git suffix", input: "https://github.com/ButterHost69/sloper.git", wantLink: "ButterHost69/sloper", wantClone: "https://github.com/ButterHost69/sloper.git"},
		{name: "https url with trailing slash", input: "https://github.com/ButterHost69/sloper/", wantLink: "ButterHost69/sloper", wantClone: "https://github.com/ButterHost69/sloper.git"},
		{name: "ssh url", input: "git@github.com:ButterHost69/sloper.git", wantLink: "ButterHost69/sloper", wantClone: "https://github.com/ButterHost69/sloper.git"},
		{name: "host prefix without scheme", input: "github.com/ButterHost69/sloper", wantLink: "ButterHost69/sloper", wantClone: "https://github.com/ButterHost69/sloper.git"},
		{name: "extra path is ignored", input: "https://github.com/ButterHost69/sloper/tree/main", wantLink: "ButterHost69/sloper", wantClone: "https://github.com/ButterHost69/sloper.git"},
		{name: "dotted repo name", input: "owner/repo.js", wantLink: "owner/repo.js", wantClone: "https://github.com/owner/repo.js.git"},
		{name: "empty", input: "", wantErr: ErrInvalidLink},
		{name: "only owner", input: "ButterHost69", wantErr: ErrInvalidLink},
		{name: "gitlab is rejected", input: "https://gitlab.com/owner/repo", wantErr: ErrInvalidLink},
		{name: "spaces are rejected", input: "owner/re po", wantErr: ErrInvalidLink},
		{name: "leading dash repo is rejected", input: "owner/-repo", wantErr: ErrInvalidLink},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			link, clone, err := ParseRepoLink(tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("ParseRepoLink(%q) error = %v, want %v", tc.input, err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRepoLink(%q) unexpected error: %v", tc.input, err)
			}
			if link != tc.wantLink {
				t.Errorf("link = %q, want %q", link, tc.wantLink)
			}
			if clone != tc.wantClone {
				t.Errorf("clone = %q, want %q", clone, tc.wantClone)
			}
		})
	}
}

func TestRepoName(t *testing.T) {
	if got := RepoName("ButterHost69/sloper"); got != "sloper" {
		t.Errorf("RepoName = %q, want sloper", got)
	}
	if got := RepoName("sloper"); got != "sloper" {
		t.Errorf("RepoName = %q, want sloper", got)
	}
}
