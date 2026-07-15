package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/ButterHost69/sloper/internal/models"
)

func (g *GitGateway) CreateBranch(ctx context.Context, cwd, name, base string) error {
	args := []string{"checkout", "-b", name}
	if base != "" {
		args = append(args, base)
	}
	return g.runGit(ctx, models.GitRunOptions{Cwd: cwd, Args: args})
}

func (g *GitGateway) CheckoutBranch(ctx context.Context, cwd, name string) error {
	return g.runGit(ctx, models.GitRunOptions{Cwd: cwd, Args: []string{"checkout", name}})
}

func (g *GitGateway) CheckoutDetach(ctx context.Context, cwd, ref string) error {
	return g.runGit(ctx, models.GitRunOptions{Cwd: cwd, Args: []string{"checkout", "--detach", ref}})
}

func (g *GitGateway) CommitAll(ctx context.Context, cwd, message string) error {
	if err := g.runGit(ctx, models.GitRunOptions{Cwd: cwd, Args: []string{"add", "-A"}}); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	return g.runGit(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"commit", "-m", message},
	})
}

func (g *GitGateway) Push(ctx context.Context, cwd, remote, branch string) error {
	args := []string{"push", "-u", remote, branch}
	if remote == "" {
		remote = "origin"
	}
	args = []string{"push", "-u", remote, branch}
	return g.runGit(ctx, models.GitRunOptions{Cwd: cwd, Args: args})
}

func (g *GitGateway) PushRef(ctx context.Context, cwd, remote, localRef, remoteRef string) error {
	if remote == "" {
		remote = "origin"
	}
	return g.runGit(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"push", remote, fmt.Sprintf("%s:%s", localRef, remoteRef)},
	})
}

func (g *GitGateway) ForcePush(ctx context.Context, cwd, remote, branch string) error {
	if remote == "" {
		remote = "origin"
	}
	return g.runGit(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"push", "--force-with-lease", "-u", remote, branch},
	})
}

func (g *GitGateway) GetDiff(ctx context.Context, cwd, base, head string) (string, error) {
	var args []string
	if base != "" && head != "" {
		args = []string{"diff", base + "..." + head}
	} else if base != "" {
		args = []string{"diff", base}
	} else {
		args = []string{"diff", "HEAD"}
	}
	res, err := g.runGitResult(ctx, models.GitRunOptions{Cwd: cwd, Args: args, MaxOutputBytes: 1024 * 1024})
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	return res.Result.Stdout, nil
}

func (g *GitGateway) GetCurrentBranch(ctx context.Context, cwd string) (string, error) {
	res, err := g.runGitResult(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"rev-parse", "--abbrev-ref", "HEAD"},
	})
	if err != nil {
		return "", fmt.Errorf("git current branch: %w", err)
	}
	return strings.TrimSpace(res.Result.Stdout), nil
}

func (g *GitGateway) GetHeadSHA(ctx context.Context, cwd string) (string, error) {
	res, err := g.runGitResult(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"rev-parse", "HEAD"},
	})
	if err != nil {
		return "", fmt.Errorf("git head sha: %w", err)
	}
	return strings.TrimSpace(res.Result.Stdout), nil
}

func (g *GitGateway) HasUncommittedChanges(ctx context.Context, cwd string) (bool, error) {
	res, err := g.runGitResult(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"status", "--porcelain"},
	})
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}
	return strings.TrimSpace(res.Result.Stdout) != "", nil
}

func (g *GitGateway) FetchBranch(ctx context.Context, cwd, remote, branch string) error {
	if remote == "" {
		remote = "origin"
	}
	return g.runGit(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"fetch", remote, branch},
	})
}

func (g *GitGateway) FetchAll(ctx context.Context, cwd string) error {
	return g.runGit(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"fetch", "--all"},
	})
}

func (g *GitGateway) AddWorktree(ctx context.Context, cwd, worktreePath, branchOrRef string, newBranch string) error {
	args := []string{"worktree", "add"}
	if newBranch != "" {
		args = append(args, "-b", newBranch)
	}
	args = append(args, worktreePath, branchOrRef)
	return g.runGit(ctx, models.GitRunOptions{Cwd: cwd, Args: args})
}

func (g *GitGateway) AddWorktreeDetached(ctx context.Context, cwd, worktreePath, ref string) error {
	return g.runGit(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"worktree", "add", "--detach", worktreePath, ref},
	})
}

func (g *GitGateway) RemoveWorktree(ctx context.Context, cwd, worktreePath string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, worktreePath)
	return g.runGit(ctx, models.GitRunOptions{Cwd: cwd, Args: args})
}

func (g *GitGateway) PruneWorktrees(ctx context.Context, cwd string) error {
	return g.runGit(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"worktree", "prune"},
	})
}

func (g *GitGateway) ListWorktrees(ctx context.Context, cwd string) (string, error) {
	res, err := g.runGitResult(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"worktree", "list", "--porcelain"},
	})
	if err != nil {
		return "", fmt.Errorf("git worktree list: %w", err)
	}
	return res.Result.Stdout, nil
}

func (g *GitGateway) GetRemotes(ctx context.Context, cwd string) (string, error) {
	res, err := g.runGitResult(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"remote", "-v"},
	})
	if err != nil {
		return "", fmt.Errorf("git remote: %w", err)
	}
	return res.Result.Stdout, nil
}

func (g *GitGateway) GetDefaultBranch(ctx context.Context, cwd string) (string, error) {
	res, err := g.runGitResult(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"symbolic-ref", "--short", "refs/remotes/origin/HEAD"},
	})
	if err != nil {
		return "main", nil
	}
	branch := strings.TrimSpace(res.Result.Stdout)
	parts := strings.SplitN(branch, "/", 2)
	if len(parts) == 2 {
		return parts[1], nil
	}
	return branch, nil
}

func (g *GitGateway) ResetHard(ctx context.Context, cwd, ref string) error {
	return g.runGit(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"reset", "--hard", ref},
	})
}

func (g *GitGateway) CountCommits(ctx context.Context, cwd, base, head string) (int, error) {
	res, err := g.runGitResult(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"rev-list", "--count", base + ".." + head},
	})
	if err != nil {
		return 0, fmt.Errorf("git rev-list: %w", err)
	}
	var count int
	if _, err := fmt.Sscanf(strings.TrimSpace(res.Result.Stdout), "%d", &count); err != nil {
		return 0, fmt.Errorf("git rev-list: parse count: %w", err)
	}
	return count, nil
}

func (g *GitGateway) BranchExists(ctx context.Context, cwd, branchName string) (bool, error) {
	res, err := g.runGitResult(ctx, models.GitRunOptions{
		Cwd:  cwd,
		Args: []string{"branch", "--list", branchName},
	})
	if err != nil {
		return false, fmt.Errorf("git branch list: %w", err)
	}
	return strings.TrimSpace(res.Result.Stdout) != "", nil
}

func (g *GitGateway) DeleteBranch(ctx context.Context, cwd, branchName string, force bool) error {
	args := []string{"branch", "-D", branchName}
	if !force {
		args = []string{"branch", "-d", branchName}
	}
	return g.runGit(ctx, models.GitRunOptions{Cwd: cwd, Args: args})
}
