package scheduler

import (
	"context"
	"fmt"

	myGit "github.com/ButterHost69/sloper/internal/git"
	myGithub "github.com/ButterHost69/sloper/internal/github"
	"github.com/ButterHost69/sloper/internal/models"
)

type Scheduler struct {
	RepoName string
	RepoPath string
	RepoUrl  string
}

func New(repoPath string) *Scheduler {
	return &Scheduler{
		RepoPath: repoPath,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	// Get all issues from the repo
	ghClient := myGithub.NewGithubGateway(models.GithubOptions{
		CWD: s.RepoName,
	})

	gitClient := myGit.New(models.GitGatewayOptions{})
	repoName, err := gitClient.DetectGitHubRepo(context.Background(), s.RepoPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	s.RepoName = repoName

	output, err := ghClient.GetAllOpenIssuesRaw(ctx, models.GithubIssueOptions{})
	if err != nil {
		fmt.Println(err)
	}

	for _, out := range output {
		fmt.Println(out.Title)
		issue, err := ghClient.ViewIssue(ctx, models.ViewIssueInput{
			Repo:        s.RepoName,
			IssueNumber: out.Number,
			CWD:         s.RepoPath,
		})

		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(issue.String())
	}
}
