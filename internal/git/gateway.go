package git

import (
	"context"
	"strings"
	"time"

	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/shell"
)

const (
	defaultGitMaxOutputBytes = 256 * 1024
)

type GitGateway struct {
	gitPath string
	now     func() time.Time
}

func New(options models.GitGatewayOptions) *GitGateway {
	gitPath := strings.TrimSpace(options.GitPath)
	if gitPath == "" {
		gitPath = "git"
	}

	now := options.Now
	if now == nil {
		now = time.Now
	}

	return &GitGateway{
		gitPath: gitPath,
		now:     now,
	}
}

// Overrides if no policy exisits
var DefaultGitRetryPolicy = &models.DefaultGitRetryPolicy{
	MaxAttemptsLimit: 3,
	BaseDelay:        50 * time.Millisecond,
	MaxDelay:         500 * time.Millisecond,
}

// Does not provide the output
func (g *GitGateway) runGit(ctx context.Context, opt models.GitRunOptions) error {
	_, err := g.runGitResult(ctx, opt)
	return err
}

// Provides Output
func (g *GitGateway) runGitResult(ctx context.Context, opt models.GitRunOptions) (models.GitResult, error) {
	return g.runGitResultWithOptions(ctx, models.GitRunOptions{
		Cwd:  opt.Cwd,
		Env:  opt.Env,
		Args: opt.Args,
	})
}

func (g *GitGateway) runGitResultWithOptions(ctx context.Context, opts models.GitRunOptions) (models.GitResult, error) {
	if opts.MaxOutputBytes == 0 {
		opts.MaxOutputBytes = defaultGitMaxOutputBytes
	}
	if opts.RetryPolicy == nil {
		// Fallback to the DefaultGitRetryPolicy
		opts.RetryPolicy = DefaultGitRetryPolicy
	}

	if opts.Trace.OnStart != nil {
		opts.Trace.OnStart(opts.Args)
	}
	start := time.Now()

	var (
		lastRes models.GitResult
		errs    []error
	)

	for attempt := 0; ; attempt++ {
		res, err := g.runGitResultOnce(ctx, opts)
		lastRes = res

		if err == nil {
			lastRes.Retries = attempt
			if opts.Trace.OnComplete != nil {
				opts.Trace.OnComplete(opts.Args, lastRes, nil, time.Since(start))
			}
			return lastRes, nil
		}

		errs = append(errs, err)

		if attempt >= opts.RetryPolicy.MaxAttempts() {
			break
		}

		// If not retry -> break
		if !opts.RetryPolicy.ShouldRetry(opts.Args, err, attempt) {
			break
		}

		if opts.Trace.OnRetry != nil {
			opts.Trace.OnRetry(opts.Args, attempt+1, err)
		}

		delay := opts.RetryPolicy.Delay(attempt)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			if opts.Trace.OnComplete != nil {
				opts.Trace.OnComplete(opts.Args, lastRes, ctx.Err(), time.Since(start))
			}
			return lastRes, ctx.Err()
		case <-timer.C:
		}
	}

	retryErr := &models.GitRetryError{
		Command:  opts.Args,
		Attempts: len(errs),
		Errors:   errs,
	}
	lastRes.Retries = len(errs) - 1

	if opts.Trace.OnComplete != nil {
		opts.Trace.OnComplete(opts.Args, lastRes, retryErr, time.Since(start))
	}
	return lastRes, retryErr
}

func (g *GitGateway) runGitResultOnce(ctx context.Context, opts models.GitRunOptions) (models.GitResult, error) {
	if opts.MaxOutputBytes == 0 {
		opts.MaxOutputBytes = defaultGitMaxOutputBytes
	}

	shellOpts := models.ShellOptions{
		Command:          g.gitPath,
		Args:             opts.Args,
		CWD:              opts.Cwd,
		Env:              opts.Env,
		Timeout:          opts.Timeout,
		MaxCapturedBytes: opts.MaxOutputBytes,
	}

	result, err := shell.Run(ctx, shellOpts)
	gitRes := models.GitResult{Result: result}
	if err == nil {
		return gitRes, nil
	}

	return gitRes, &models.GitError{
		Args:   opts.Args,
		Result: result,
		Cause:  err,
	}
}
