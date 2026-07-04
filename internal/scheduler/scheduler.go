package scheduler

import (
	"context"
	"fmt"

	githubClient "github.com/ButterHost69/sloper/internal/github"
	"github.com/ButterHost69/sloper/internal/models"
)

type Scheduler struct {
	RepoName string
}

func New(repoName string) *Scheduler {
	return &Scheduler{
		RepoName: repoName,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	// Get all issues from the repo
	ghClient := githubClient.NewGithubGateway(models.GithubOptions{
		CWD: s.RepoName,
	})

	output, err := ghClient.GetAllOpenIssuesRaw(ctx, models.GithubIssueOptions{})
	if err != nil {
		fmt.Println(err)
	}
	
	for _, out := range output {
		fmt.Println(out.Title)
	}
}
