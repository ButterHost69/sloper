package scheduler

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ButterHost69/sloper/internal/agent"
	myGit "github.com/ButterHost69/sloper/internal/git"
	myGithub "github.com/ButterHost69/sloper/internal/github"
	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/pipeline"
)

// Scheduler orchestrates the issue → spec → work → review → merge pipeline.
type Scheduler struct {
	RepoName string
	RepoPath string
	RepoUrl  string
}

func New(repoPath string) *Scheduler {
	return &Scheduler{RepoPath: repoPath}
}

func (s *Scheduler) Start(ctx context.Context) {
	// ── GitHub client ────────────────────────────────────────────
	ghClient := myGithub.NewGithubGateway(models.GithubOptions{CWD: s.RepoPath})

	// ── Detect repo name from git remote ─────────────────────────
	gitClient := myGit.New(models.GitGatewayOptions{})
	repoName, err := gitClient.DetectGitHubRepo(context.Background(), s.RepoPath)
	if err != nil {
		log.Printf("scheduler: detect repo: %v", err)
		return
	}
	s.RepoName = repoName
	log.Printf("scheduler: repo = %s", s.RepoName)

	// ── Agent gateway + pipeline ─────────────────────────────────
	agentGw := agent.NewAgentGateway(models.AgentOptions{
		CWD:      s.RepoPath,
		Model:    os.Getenv("AGENT_MODEL"),
		Thinking: "high",
	})
	pl := pipeline.New(agentGw)

	// ── Fetch open issues ────────────────────────────────────────
	issues, err := ghClient.GetAllOpenIssuesRaw(ctx, models.GithubIssueOptions{
		Repo:  s.RepoName,
		CWD:   s.RepoPath,
		Limit: 10,
	})
	if err != nil {
		log.Printf("scheduler: list issues: %v", err)
		return
	}

	log.Printf("scheduler: found %d open issues", len(issues))

	for _, summary := range issues {
		if err := ctx.Err(); err != nil {
			return
		}
		if err := s.processOne(ctx, ghClient, pl, summary); err != nil {
			log.Printf("scheduler: issue #%d: %v", summary.Number, err)
		}
	}
}

// processOne runs the full pipeline for a single issue.
func (s *Scheduler) processOne(
	ctx context.Context,
	gh *myGithub.GithubGateway,
	pl *pipeline.Pipeline,
	summary models.GithubIssueSummary,
) error {
	issue, err := gh.ViewIssue(ctx, models.ViewIssueInput{
		Repo:        s.RepoName,
		IssueNumber: summary.Number,
		CWD:         s.RepoPath,
	})
	if err != nil {
		return fmt.Errorf("view issue: %w", err)
	}

	triaged := hasLabel(issue.Labels, models.TRIAGED_LABEL)
	log.Printf("scheduler: #%d %q — triaged=%v labels=%v",
		issue.Number, issue.Title, triaged, issue.Labels)

	// ─── STAGE 1: SPEC ───────────────────────────────────────────
	if !triaged {
		log.Printf("scheduler: #%d → SPEC stage", issue.Number)
		spec, err := pl.SpecIssue(ctx, issue)
		if err != nil {
			return fmt.Errorf("spec: %w", err)
		}
		log.Printf("scheduler: #%d spec summary: %s", issue.Number, spec.Summary)
		log.Printf("scheduler: #%d files: %v", issue.Number, spec.FilesToChange)

		// TODO: add "triaged" label, post spec as comment
		_ = spec
	}

	// ─── STAGE 2: WORK ───────────────────────────────────────────
	// TODO: run ImplementFix only if spec exists and is approved
	// work, err := pl.ImplementFix(ctx, spec)

	// ─── STAGE 3: REVIEW ─────────────────────────────────────────
	// TODO: get PR diff, run ReviewPR
	// review, err := pl.ReviewPR(ctx, diff)
	// if !review.Approved {
	//     work, err = pl.FixReviewIssues(ctx, review)
	//     // loop review
	// }

	// ─── STAGE 4: MERGE ──────────────────────────────────────────
	// TODO: merge PR if approved

	return nil
}

func hasLabel(labels []string, target string) bool {
	for _, l := range labels {
		if l == target {
			return true
		}
	}
	return false
}
