package github

import (
	"context"
	"strings"
	"time"

	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/shell"
)

type GithubGateway struct {
	ghPath  string
	cwd     string
	ghShell func(context.Context, models.ShellOptions) (models.ShellResult, error)
}

// TODO: I dont like the way the cwd is handled, it can be overidden even though we know the cwd before hand, find a way to improve it (if needed).
//
//	Look into if there are any needs for it to be overridable
func NewGithubGateway(options models.GithubOptions) *GithubGateway {
	ghPath := strings.TrimSpace(options.GHPath)
	if ghPath == "" {
		ghPath = "gh"
	}

	ghShell := options.GHRun
	if ghShell == nil {
		ghShell = shell.Run
	}

	if strings.TrimSpace(options.CWD) == "" {
		// print some error / show in some logs and shit
	}

	return &GithubGateway{
		ghPath:  ghPath,
		cwd:     options.CWD,
		ghShell: ghShell,
	}
}

func (g *GithubGateway) runGh(ctx context.Context, cwd, stdin string, args ...string) (models.ShellResult, error) {
	return g.runGhWithTimeout(ctx, cwd, stdin, models.DefaultGhCommandTimeout, args...)
}

func (g *GithubGateway) runGhWithTimeout(ctx context.Context, cwd, stdin string, timeout time.Duration, args ...string) (models.ShellResult, error) {
	if strings.TrimSpace(cwd) == "" {
		cwd = g.cwd
	}

	result, err := g.ghShell(
		ctx,
		models.ShellOptions{Command: g.ghPath, Args: args, CWD: cwd, Stdin: stdin, Timeout: timeout},
	)

	if err != nil && isTransientGitHubMessage(strings.Join([]string{err.Error(), result.Stdout, result.Stderr}, "\n")) {
		return result, &TransientError{Err: err}
	}
	return result, err
}
