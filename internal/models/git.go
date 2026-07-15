package models

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type GitGatewayOptions struct {
	GitPath string
	Now func() time.Time
}

type GitRunOptions struct {
	Cwd            string
	Env            map[string]string
	Args           []string
	Timeout        time.Duration
	MaxOutputBytes int
	RetryPolicy    RetryPolicy
	Trace          GitTrace
}

// GitTrace hooks let callers instrument every git invocation without changing the core functions.
// Used for metrics collection and other custom logic
// (Look into a better logging method that does it all, so we dont need something like this)
type GitTrace struct {
	OnStart    func(args []string)
	OnRetry    func(args []string, attempt int, err error)
	OnComplete func(args []string, result GitResult, err error, duration time.Duration)
}

// Implements interface RetryPolicy
type DefaultGitRetryPolicy struct {
	MaxAttemptsLimit int
	BaseDelay        time.Duration
	MaxDelay         time.Duration
}

func (p *DefaultGitRetryPolicy) ShouldRetry(args []string, err error, attempt int) bool {
	if isPermanentGitError(err) {
		return false
	}
	return true
}

func isPermanentGitError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	permanent := []string{
		"already exists",
		"not found",
		"not a git repository",
		"invalid",
		"permission denied",
		"Permission denied",
		"authentication failed",
		"could not read Username",
		"could not read Password",
		"no such file or directory",
		"not a valid object",
		"does not match any",
	}
	for _, p := range permanent {
		if contains(msg, p) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (p *DefaultGitRetryPolicy) Delay(attempt int) time.Duration {
	// Exponential backoff with jitter: 50ms, ~75ms, ~150ms, capped at maxDelay.
	d := p.BaseDelay * (1 << attempt)
	if d > p.MaxDelay || d <= 0 {
		d = p.MaxDelay
	}
	if d > 0 {
		maxJitter := d / 2
		if maxJitter > 0 {
			jitter, _ := rand.Int(rand.Reader, big.NewInt(int64(maxJitter)))
			d += time.Duration(jitter.Int64())
		}
	}
	return d
}

func (p *DefaultGitRetryPolicy) MaxAttempts() int { return p.MaxAttemptsLimit }

// type Repositories struct {
// 	Projects             *ProjectsRepository
// 	Loops                *LoopsRepository
// 	Runs                 *RunsRepository
// 	AgentExecutions      *AgentExecutionsRepository
// 	PullRequestSnapshots *PullRequestSnapshotsRepository
// 	Events               *EventsRepository
// 	Locks                *LocksRepository
// 	Queue                *QueueRepository
// 	Notifications        *NotificationsRepository
// 	Worktrees            *WorktreesRepository
// 	WebhookForwarders    *WebhookForwardersRepository
// 	WebhookTunnelHooks   *WebhookTunnelHooksRepository
// }

type GitResult struct {
	Result    ShellResult
	Retries   int  // number of retry attempts that were made
}

type GitError struct {
	Args   []string
	Result ShellResult
	Cause  error // Contains the root cause, ShellExecutionError, context.Cancelled
}

// Concats, stdout and stderr to the error message
func (e *GitError) Error() string {
	msg := fmt.Sprintf("error=%s", e.Cause.Error())
	if e.Result.Stderr != "" {
		msg = fmt.Sprintf("stderr=%s %s", strings.TrimSpace(e.Result.Stderr), msg)

	} else if e.Result.Stdout != "" {
		msg = fmt.Sprintf("stdout=%s %s", strings.TrimSpace(e.Result.Stdout), msg)
	}

	return fmt.Sprintf("git %s: %s", strings.Join(e.Args, " "), msg)
}

func (e *GitError) Unwrap() error { return e.Cause }

type GitRetryError struct {
	Command  []string
	Attempts int
	Errors   []error
}

func (e *GitRetryError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "git %s: failed after %d attempt(s):", strings.Join(e.Command, " "), e.Attempts)
	for i, err := range e.Errors {
		fmt.Fprintf(&b, "\n  attempt %d: %v", i+1, err)
	}
	return b.String()
}

// Unwrap returns the last attempt error 
// so errors.As still works for shell.CommandExecutionError, GitError.
func (e *GitRetryError) Unwrap() error {
	if len(e.Errors) == 0 {
		return nil
	}
	return e.Errors[len(e.Errors)-1]
}
